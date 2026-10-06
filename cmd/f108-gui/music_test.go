//go:build linux

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/parsiya/f108-pro/pkg/aula"
)

func tone(freq float64, n int) []float64 {
	x := make([]float64, n)
	for i := range x {
		x[i] = 0.3 * math.Sin(2*math.Pi*freq*float64(i)/musicRate)
	}
	return x
}

func levelsFor(freq, gain float64) [musicColumns]byte {
	bins := bandBins(musicWindow, musicRate)
	win := hann(musicWindow)
	x := tone(freq, musicWindow)
	spec := make([]complex128, musicWindow)
	for i := range spec {
		spec[i] = complex(x[i]*win[i], 0)
	}
	fft(spec)
	var smooth [musicColumns]float64
	var lv [musicColumns]byte
	for frame := 0; frame < 30; frame++ { // deja asentarse el suavizado
		lv = computeLevels(spec, bins, gain, &smooth)
	}
	return lv
}

func argmax(lv [musicColumns]byte) int {
	best := 0
	for i, v := range lv {
		if v > lv[best] {
			best = i
		}
	}
	return best
}

func TestToneLandsInExpectedColumn(t *testing.T) {
	cases := []struct {
		freq    float64
		columns []int
		min     byte
	}{
		{100, []int{2, 3}, 4},
		{500, []int{8}, 4},
		{2000, []int{13}, 3},
		{8000, []int{18}, 2},
	}
	for _, c := range cases {
		lv := levelsFor(c.freq, 1)
		col := argmax(lv)
		ok := false
		for _, want := range c.columns {
			ok = ok || col == want
		}
		if !ok || lv[col] < c.min {
			t.Errorf("%v Hz: columna %d nivel %d, esperaba columnas %v con nivel >= %d (%v)", c.freq, col, lv[col], c.columns, c.min, lv)
		}
	}
}

func TestSilenceIsZero(t *testing.T) {
	bins := bandBins(musicWindow, musicRate)
	spec := make([]complex128, musicWindow)
	var smooth [musicColumns]float64
	if lv := computeLevels(spec, bins, 1, &smooth); lv != [musicColumns]byte{} {
		t.Errorf("el silencio debe dar todo cero, dio %v", lv)
	}
}

func TestGainScalesLevels(t *testing.T) {
	low, high := levelsFor(2000, 0.3), levelsFor(2000, 2)
	if low[13] >= high[13] {
		t.Errorf("mas amplitud debe dar mas nivel: %d vs %d", low[13], high[13])
	}
}

func TestFFTMatchesDFT(t *testing.T) {
	const n = 64
	a := make([]complex128, n)
	for i := range a {
		a[i] = complex(math.Sin(float64(i)*0.7)+0.1*float64(i%5), 0)
	}
	want := make([]complex128, n)
	for k := 0; k < n; k++ {
		for i := 0; i < n; i++ {
			ang := -2 * math.Pi * float64(k*i) / n
			want[k] += a[i] * complex(math.Cos(ang), math.Sin(ang))
		}
	}
	fft(a)
	for k := range a {
		if d := a[k] - want[k]; real(d)*real(d)+imag(d)*imag(d) > 1e-12 {
			t.Fatalf("bin %d: %v vs %v", k, a[k], want[k])
		}
	}
}

// ---- bucle completo con un teclado y un audio simulados ----

type frame struct {
	fore, back, bright byte
	levels             [musicColumns]byte
}

type fakeKeyboard struct {
	mu     sync.Mutex
	frames []frame
	closed bool
	err    error
}

func (f *fakeKeyboard) SendMusicFrame(fore, back, bright byte, levels [21]byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.frames = append(f.frames, frame{fore, back, bright, levels})
	return f.err
}

func (f *fakeKeyboard) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakeKeyboard) snapshot() []frame {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]frame(nil), f.frames...)
}

// pcmTone genera `frames` fotogramas de audio (48 kHz, mono, s16le) con un tono; 0 Hz = silencio.
func pcmTone(freq float64, frames int) []byte {
	n := frames * (musicRate / musicFPS)
	buf := make([]byte, 2*n)
	for i := 0; i < n; i++ {
		v := 0.0
		if freq > 0 {
			v = 0.3 * math.Sin(2*math.Pi*freq*float64(i)/musicRate)
		}
		binary.LittleEndian.PutUint16(buf[2*i:], uint16(int16(v*32767)))
	}
	return buf
}

func runLoop(t *testing.T, kb *fakeKeyboard, cfg musicCfg, pcm []byte, lastSeen time.Time) {
	t.Helper()
	music.mu.Lock()
	music.running, music.cfg, music.err, music.lastSeen = true, cfg, "", lastSeen
	music.mu.Unlock()
	done := make(chan struct{})
	go musicLoop(context.Background(), done, kb, nil, bytes.NewReader(pcm))
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("el bucle no termino")
	}
}

func TestLoopToneThenFinalZeroFrame(t *testing.T) {
	kb := &fakeKeyboard{}
	cfg := musicCfg{Fore: 3, Back: 6, Bright: 40, Gain: 1, Source: "monitor"}
	pcm := append(pcmTone(100, 30), pcmTone(0, 30)...) // 1 s de tono grave y 1 s de silencio
	runLoop(t, kb, cfg, pcm, time.Now())

	fr := kb.snapshot()
	if len(fr) < 3 {
		t.Fatalf("pocos fotogramas: %d", len(fr))
	}
	if len(fr) > 25 {
		t.Errorf("deberia omitir fotogramas repetidos, pero envio %d", len(fr))
	}
	peak := 0
	for _, f := range fr {
		if f.fore != 3 || f.back != 6 || f.bright != 40 {
			t.Fatalf("ajustes mal enviados: %+v", f)
		}
		if a := argmax(f.levels); f.levels[a] >= 3 && (a < 1 || a > 4) { // los fotogramas de arranque son planos
			t.Errorf("el tono de 100 Hz cayo en la columna %d: %v", a, f.levels)
		}
		if int(f.levels[2]) > peak {
			peak = int(f.levels[2])
		}
	}
	if peak < 3 {
		t.Errorf("nivel maximo bajo en los graves: %d", peak)
	}
	if last := fr[len(fr)-1]; last.levels != [musicColumns]byte{} {
		t.Errorf("el ultimo fotograma debe ser todo cero: %v", last.levels)
	}
	if !kb.closed {
		t.Error("no se cerro el dispositivo")
	}
	music.mu.Lock()
	running, errMsg := music.running, music.err
	music.mu.Unlock()
	if running {
		t.Error("el estado debe quedar detenido")
	}
	if errMsg == "" {
		t.Error("al acabar el audio debe quedar un mensaje de error")
	}
}

func TestLoopStopsWhenPageGoesAway(t *testing.T) {
	kb := &fakeKeyboard{}
	cfg := musicCfg{Fore: 2, Bright: 100, Gain: 1, Source: "monitor"}
	runLoop(t, kb, cfg, pcmTone(500, 30), time.Now().Add(-time.Minute))
	music.mu.Lock()
	errMsg := music.err
	music.mu.Unlock()
	if errMsg == "" || len(kb.snapshot()) != 1 { // solo el fotograma final
		t.Errorf("deberia detenerse sin enviar audio: err=%q fotogramas=%d", errMsg, len(kb.snapshot()))
	}
}

func TestLoopSurvivesLostACKButStopsOnRealError(t *testing.T) {
	kb := &fakeKeyboard{err: aula.ErrNoACK}
	runLoop(t, kb, musicCfg{Fore: 2, Bright: 100, Gain: 1, Source: "monitor"}, pcmTone(500, 10), time.Now())
	if len(kb.snapshot()) < 2 {
		t.Errorf("un ACK perdido no debe detener el envio: %d fotogramas", len(kb.snapshot()))
	}

	kb2 := &fakeKeyboard{err: context.DeadlineExceeded}
	runLoop(t, kb2, musicCfg{Fore: 2, Bright: 100, Gain: 1, Source: "monitor"}, pcmTone(500, 10), time.Now())
	music.mu.Lock()
	errMsg := music.err
	music.mu.Unlock()
	if errMsg == "" || len(kb2.snapshot()) != 2 { // el que fallo + el final
		t.Errorf("un error real debe detener el bucle: err=%q fotogramas=%d", errMsg, len(kb2.snapshot()))
	}
}

func TestConfigValidation(t *testing.T) {
	ok := musicCfg{Fore: 2, Back: 0, Bright: 100, Gain: 1, Source: "monitor"}
	if err := ok.validate(); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []musicCfg{
		{Fore: 0, Back: 0, Bright: 100, Gain: 1, Source: "monitor"},
		{Fore: 8, Back: 0, Bright: 100, Gain: 1, Source: "monitor"},
		{Fore: 2, Back: 9, Bright: 100, Gain: 1, Source: "monitor"},
		{Fore: 2, Back: 0, Bright: 101, Gain: 1, Source: "monitor"},
		{Fore: 2, Back: 0, Bright: 100, Gain: 0, Source: "monitor"},
		{Fore: 2, Back: 0, Bright: 100, Gain: 1, Source: "otra"},
	} {
		if bad.validate() == nil {
			t.Errorf("deberia rechazar %+v", bad)
		}
	}
}
