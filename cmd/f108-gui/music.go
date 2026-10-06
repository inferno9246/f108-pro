//go:build linux

package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/parsiya/f108-pro/pkg/aula"
)

// Modo musica por el dongle 2.4G (paquete 0B 1C). Formato reconstruido de capturas del
// software oficial: 21 niveles (0-5), uno por columna, enviados ~30 veces por segundo.
// El teclado no los guarda: si dejan de llegar, vuelve a su iluminacion normal.

const (
	musicColumns   = 21
	musicRate      = 48000
	musicFPS       = 30
	musicWindow    = 2048
	musicMaxLevel  = 5.0
	musicTilt      = 10.0 // dB extra hacia los agudos
	musicKeepalive = 250 * time.Millisecond
	musicIdleLimit = 20 * time.Second
)

type musicCfg struct {
	Fore   int     `json:"fore"`   // efecto principal (1-7)
	Back   int     `json:"back"`   // fondo (0-8)
	Bright int     `json:"bright"` // brillo del fondo (0-100)
	Gain   float64 `json:"gain"`   // amplitud: cuanto reaccionan las barras
	Source string  `json:"source"` // "monitor" (lo que suena) o "mic"
}

func (c musicCfg) validate() error {
	switch {
	case c.Fore < 1 || c.Fore > 7:
		return errors.New("el efecto debe estar entre 1 y 7")
	case c.Back < 0 || c.Back > 12:
		return errors.New("el fondo debe estar entre 0 y 12")
	case c.Bright < 0 || c.Bright > 100:
		return errors.New("el brillo del fondo debe estar entre 0 y 100")
	case c.Gain < 0.1 || c.Gain > 4:
		return errors.New("la amplitud debe estar entre 0.1 y 4")
	case c.Source != "monitor" && c.Source != "mic":
		return errors.New("la fuente debe ser \"monitor\" o \"mic\"")
	}
	return nil
}

// frameSender es lo que el bucle necesita del teclado (permite probarlo sin hardware).
type frameSender interface {
	SendMusicFrame(fore, back, bright byte, levels [21]byte) error
	Close() error
}

var music struct {
	mu       sync.Mutex
	running  bool
	cancel   context.CancelFunc
	done     chan struct{}
	cfg      musicCfg
	lastSeen time.Time
	err      string
}

var musicSent, musicAcked atomic.Int64

// ---- analisis de audio (puro, probado en music_test.go) ----

func hann(n int) []float64 {
	w := make([]float64, n)
	for i := range w {
		w[i] = 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n-1))
	}
	return w
}

// fft: transformada de Fourier iterativa (radix-2). len(a) debe ser potencia de 2.
func fft(a []complex128) {
	n := len(a)
	j := 0
	for i := 1; i < n; i++ {
		bit := n >> 1
		for ; j&bit != 0; bit >>= 1 {
			j ^= bit
		}
		j ^= bit
		if i < j {
			a[i], a[j] = a[j], a[i]
		}
	}
	for length := 2; length <= n; length <<= 1 {
		ang := -2 * math.Pi / float64(length)
		wl := complex(math.Cos(ang), math.Sin(ang))
		for i := 0; i < n; i += length {
			w := complex(1, 0)
			for k := 0; k < length/2; k++ {
				u := a[i+k]
				v := a[i+k+length/2] * w
				a[i+k] = u + v
				a[i+k+length/2] = u - v
				w *= wl
			}
		}
	}
}

// bandBins reparte los bins de la FFT en 21 bandas logaritmicas entre 40 Hz y 16 kHz.
func bandBins(size, rate int) [musicColumns][]int {
	var out [musicColumns][]int
	edge := func(i int) float64 { return 40 * math.Pow(16000.0/40.0, float64(i)/musicColumns) }
	for i := 0; i < musicColumns; i++ {
		lo, hi := edge(i), edge(i+1)
		for k := 1; k <= size/2; k++ {
			if f := float64(k) * float64(rate) / float64(size); f >= lo && f < hi {
				out[i] = append(out[i], k)
			}
		}
		if len(out[i]) == 0 { // banda mas estrecha que un bin: el mas cercano
			mid := math.Sqrt(lo * hi)
			best, bestD := 1, math.MaxFloat64
			for k := 1; k <= size/2; k++ {
				if d := math.Abs(float64(k)*float64(rate)/float64(size) - mid); d < bestD {
					best, bestD = k, d
				}
			}
			out[i] = []int{best}
		}
	}
	return out
}

// computeLevels convierte el espectro en 21 niveles (0-5). smooth conserva el estado
// entre fotogramas: sube de golpe y baja suavemente.
func computeLevels(spec []complex128, bins [musicColumns][]int, gain float64, smooth *[musicColumns]float64) [musicColumns]byte {
	var out [musicColumns]byte
	for i := 0; i < musicColumns; i++ {
		var sum float64
		for _, k := range bins[i] {
			sum += cmplx.Abs(spec[k]) / float64(len(spec))
		}
		m := sum / float64(len(bins[i]))
		db := 20*math.Log10(m+1e-9) + musicTilt*float64(i)/float64(musicColumns-1)
		lvl := (db + 75) / 50 * musicMaxLevel * gain
		lvl = math.Max(0, math.Min(musicMaxLevel, lvl))
		if lvl > smooth[i] {
			smooth[i] = lvl
		} else {
			smooth[i] = smooth[i]*0.75 + lvl*0.25
		}
		out[i] = byte(math.Round(smooth[i]))
	}
	return out
}

// ---- control ----

func startCapture(ctx context.Context, source string) (*exec.Cmd, io.Reader, error) {
	dev := "@DEFAULT_MONITOR@"
	if source == "mic" {
		dev = "@DEFAULT_SOURCE@"
	}
	cmd := exec.CommandContext(ctx, "parec", "--format=s16le", fmt.Sprintf("--rate=%d", musicRate),
		"--channels=1", "--latency-msec=20", "-d", dev)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("no se pudo iniciar parec (¿esta instalado libpulse?): %w", err)
	}
	return cmd, out, nil
}

func musicStartPayload(raw json.RawMessage) (any, error) {
	cfg := musicCfg{Fore: 2, Bright: 100, Gain: 1, Source: "monitor"}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	music.mu.Lock()
	defer music.mu.Unlock()
	music.lastSeen = time.Now()

	if music.running {
		if music.cfg.Source != cfg.Source {
			return nil, errors.New("para cambiar la fuente de audio, detén el modo música y vuelve a activarlo")
		}
		music.cfg = cfg // ajuste en vivo: efecto, fondo, brillo y amplitud
		return map[string]bool{"running": true}, nil
	}

	wdev, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd, pcm, err := startCapture(ctx, cfg.Source)
	if err != nil {
		cancel()
		wdev.Close()
		return nil, err
	}

	music.running, music.cancel, music.cfg, music.err = true, cancel, cfg, ""
	music.done = make(chan struct{})
	musicSent.Store(0)
	musicAcked.Store(0)
	go musicLoop(ctx, music.done, wdev, cmd, pcm)
	return map[string]bool{"running": true}, nil
}

func musicStop() (any, error) {
	music.mu.Lock()
	cancel, done := music.cancel, music.done
	music.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}
	return map[string]bool{"running": false}, nil
}

// musicStatus tambien actua como "latido" de la pagina: si deja de llamarse, el motor se detiene.
func musicStatus() (any, error) {
	music.mu.Lock()
	defer music.mu.Unlock()
	music.lastSeen = time.Now()
	return map[string]any{
		"running": music.running,
		"error":   music.err,
		"sent":    musicSent.Load(),
		"acked":   musicAcked.Load(),
	}, nil
}

func musicFail(msg string) {
	music.mu.Lock()
	music.err = msg
	music.mu.Unlock()
}

func musicLoop(ctx context.Context, done chan struct{}, dev frameSender, cmd *exec.Cmd, pcm io.Reader) {
	defer func() {
		if cmd != nil {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			_ = cmd.Wait()
		}
		music.mu.Lock()
		cfg := music.cfg
		music.mu.Unlock()
		deviceMu.Lock() // fotograma final a cero y fuera
		_ = dev.SendMusicFrame(byte(cfg.Fore), byte(cfg.Back), byte(cfg.Bright), [musicColumns]byte{})
		deviceMu.Unlock()
		dev.Close()
		music.mu.Lock()
		music.running = false
		music.mu.Unlock()
		close(done)
	}()

	hop := musicRate / musicFPS
	raw := make([]byte, hop*2)
	samples := make([]float64, musicWindow)
	win := hann(musicWindow)
	bins := bandBins(musicWindow, musicRate)
	spec := make([]complex128, musicWindow)
	var smooth [musicColumns]float64
	var last [musicColumns]byte
	var lastCfg musicCfg
	var lastSend time.Time

	for ctx.Err() == nil {
		if _, err := io.ReadFull(pcm, raw); err != nil {
			if ctx.Err() == nil {
				musicFail("no se pudo leer el audio (¿esta instalado parec y hay una salida de sonido?)")
			}
			return
		}
		music.mu.Lock()
		cfg, idle := music.cfg, time.Since(music.lastSeen)
		music.mu.Unlock()
		if idle > musicIdleLimit {
			musicFail("detenido: la página dejó de responder")
			return
		}

		copy(samples, samples[hop:])
		for i := 0; i < hop; i++ {
			samples[musicWindow-hop+i] = float64(int16(binary.LittleEndian.Uint16(raw[2*i:]))) / 32768
		}
		for i := range spec {
			spec[i] = complex(samples[i]*win[i], 0)
		}
		fft(spec)
		levels := computeLevels(spec, bins, cfg.Gain, &smooth)

		if levels == last && cfg == lastCfg && time.Since(lastSend) < musicKeepalive {
			continue // sin cambios: solo un latido cada 250 ms para mantener el modo
		}
		deviceMu.Lock()
		err := dev.SendMusicFrame(byte(cfg.Fore), byte(cfg.Back), byte(cfg.Bright), levels)
		deviceMu.Unlock()
		musicSent.Add(1)
		switch {
		case err == nil:
			musicAcked.Add(1)
		case errors.Is(err, aula.ErrNoACK):
			// un fotograma perdido no importa
		default:
			musicFail("error al enviar al teclado: " + err.Error())
			return
		}
		last, lastCfg, lastSend = levels, cfg, time.Now()
	}
}
