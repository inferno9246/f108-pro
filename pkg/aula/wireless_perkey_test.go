//go:build linux

package aula

import (
	"os"
	"strings"
	"syscall"
	"testing"
)

// Salida de wl_tools.py (validado byte a byte contra una captura del software oficial) para:
// todas las teclas (0,0,40), Esc (255,0,0) y A (0,255,0), brillo 5. Cada linea: payload de 32 bytes.
const goldenPerKey = `
05 10 00 80 00 00 00 00 00 00 00 00 05 00 00 00 00 aa 55 00 00 00 00 00 00 00 00 00 00 00 00 99
14 1c 00 00 00 00 00 01 ff 00 00 02 00 00 28 03 00 00 28 04 00 00 28 05 00 00 28 06 00 00 28 0c
14 1c 01 07 00 00 28 08 00 00 28 09 00 00 28 0a 00 00 28 0b 00 00 28 0c 00 00 28 0d 00 00 28 8f
14 1c 02 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 13 00 00 28 14 00 00 28 a9
14 1c 03 15 00 00 28 16 00 00 28 17 00 00 28 18 00 00 28 19 00 00 28 1a 00 00 28 1b 00 00 28 f3
14 1c 04 1c 00 00 28 1d 00 00 28 1e 00 00 28 1f 00 00 28 20 00 00 28 21 00 00 28 22 00 00 28 25
14 1c 05 00 00 00 00 00 00 00 00 25 00 00 28 26 00 00 28 27 00 00 28 28 00 00 28 29 00 00 28 c0
14 1c 06 2a 00 00 28 2b 00 00 28 2c 00 00 28 2d 00 00 28 2e 00 00 28 2f 00 00 28 30 00 00 28 89
14 1c 07 31 00 00 28 32 00 00 28 33 00 00 28 34 00 00 28 00 00 00 00 00 00 00 00 37 00 00 28 00
14 1c 08 38 00 ff 00 39 00 00 28 3a 00 00 28 3b 00 00 28 3c 00 00 28 3d 00 00 28 3e 00 00 28 c4
14 1c 09 3f 00 00 28 40 00 00 28 41 00 00 28 42 00 00 28 43 00 00 28 44 00 00 28 45 00 00 28 1f
14 1c 0a 46 00 00 28 00 00 00 00 00 00 00 00 49 00 00 28 4a 00 00 28 4b 00 00 28 4c 00 00 28 72
14 1c 0b 4d 00 00 28 4e 00 00 28 4f 00 00 28 50 00 00 28 51 00 00 28 52 00 00 28 53 00 00 28 83
14 1c 0c 54 00 00 28 55 00 00 28 56 00 00 28 57 00 00 28 58 00 00 28 00 00 00 00 00 00 00 00 b2
14 1c 0d 5b 00 00 28 5c 00 00 28 5d 00 00 28 5e 00 00 28 5f 00 00 28 60 00 00 28 61 00 00 28 e7
14 1c 0e 62 00 00 28 63 00 00 28 64 00 00 28 65 00 00 28 66 00 00 28 67 00 00 28 68 00 00 28 19
14 1c 0f 69 00 00 28 6a 00 00 28 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 62
14 1c 10 70 00 00 28 71 00 00 28 00 00 00 00 73 00 00 28 74 00 00 28 75 00 00 28 76 00 00 28 e3
14 1c 11 77 00 00 28 78 00 00 28 79 00 00 28 7a 00 00 28 7b 00 00 28 00 00 00 00 00 00 00 00 66
14 1c 12 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 42
14 1c 13 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 43
14 10 14 00 00 00 00 00 00 00 00 00 00 00 00 00 00 aa 55 00 00 00 00 00 00 00 00 00 00 00 00 37
`

func TestPerKeyPacketsMatchOfficialCapture(t *testing.T) {
	var keys []KeyColor
	for name, idx := range KeyNameToIndex {
		c := KeyColor{LightIndex: idx, R: 0, G: 0, B: 40}
		switch name {
		case "esc":
			c = KeyColor{LightIndex: idx, R: 255, G: 0, B: 0}
		case "a":
			c = KeyColor{LightIndex: idx, R: 0, G: 255, B: 0}
		}
		keys = append(keys, c)
	}
	want := strings.Split(strings.TrimSpace(goldenPerKey), "\n")
	pk := perKeyPackets(keys, 5)
	if len(pk) != len(want) {
		t.Fatalf("paquetes: %d, esperaba %d", len(pk), len(want))
	}
	for i, p := range pk {
		pkt := wirelessPacket(p)
		var sb strings.Builder
		for j, b := range pkt[1:] {
			if j > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(string("0123456789abcdef"[b>>4]))
			sb.WriteString(string("0123456789abcdef"[b&15]))
		}
		if sb.String() != want[i] {
			t.Fatalf("paquete %d distinto\n got: %s\nwant: %s", i, sb.String(), want[i])
		}
	}
}

func TestPerKeyRejectsBadInput(t *testing.T) {
	d := &WirelessDevice{}
	if err := d.SetPerKeyRGB(nil, 6); err == nil {
		t.Error("deberia rechazar brillo 6")
	}
	if err := d.SetPerKeyRGB([]KeyColor{{LightIndex: 200}}, 5); err == nil {
		t.Error("deberia rechazar un indice fuera de rango")
	}
}

func TestBattery(t *testing.T) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK, 0)
	if err != nil {
		t.Skip("socketpair no disponible:", err)
	}
	a, b := os.NewFile(uintptr(fds[0]), "a"), os.NewFile(uintptr(fds[1]), "b")
	defer a.Close()
	defer b.Close()
	go func() {
		pkt := make([]byte, 33)
		if _, err := b.Read(pkt); err != nil {
			// la lectura puede devolver parcial/espera: reintento simple
			for i := 0; i < 50 && err != nil; i++ {
				_, err = b.Read(pkt)
			}
		}
		resp := make([]byte, 32)
		copy(resp, []byte{0x20, 0x01, 0x00, 90})
		b.Write(resp)
	}()
	d := &WirelessDevice{f: a}
	got, err := d.Battery()
	if err != nil || got != 90 {
		t.Fatalf("Battery() = %d, %v; esperaba 90", got, err)
	}
}
