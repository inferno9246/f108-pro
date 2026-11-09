package aula

import (
	"bytes"
	"os"
	"testing"
)

func TestSettingsPacketsMatchOfficialCapture(t *testing.T) {
	golden, err := os.ReadFile("testdata/settings-seventeen.bin")
	if err != nil {
		t.Fatal(err)
	}
	// Orden de los 17 envíos del software oficial en todo.pcapng.
	cases := []struct {
		sleep, response byte
		flags           [3]byte
	}{
		{2, 1, [3]byte{}}, {2, 2, [3]byte{}}, {2, 3, [3]byte{}},
		{2, 4, [3]byte{}}, {2, 5, [3]byte{}}, {2, 1, [3]byte{}},
		{0, 1, [3]byte{}}, {1, 1, [3]byte{}}, {2, 1, [3]byte{}},
		{3, 1, [3]byte{}}, {2, 1, [3]byte{}},
		{2, 1, [3]byte{0, 0, 1}}, {2, 1, [3]byte{}},
		{2, 1, [3]byte{0, 1, 0}}, {2, 1, [3]byte{}},
		{2, 1, [3]byte{1, 0, 0}}, {2, 1, [3]byte{}},
	}
	if len(golden) != len(cases)*33 {
		t.Fatalf("longitud golden: %d", len(golden))
	}
	for i, c := range cases {
		p, err := settingsPacket(c.sleep, c.response, c.flags)
		if err != nil {
			t.Fatalf("paquete %d: %v", i, err)
		}
		if !bytes.Equal(p[:], golden[i*33:(i+1)*33]) {
			t.Fatalf("paquete %d\n got % x\nwant % x", i, p, golden[i*33:(i+1)*33])
		}
	}
}

func TestSettingsRejectInvalidBeforeIO(t *testing.T) {
	for _, c := range []struct{ sleep, response byte }{{4, 1}, {2, 0}, {2, 6}} {
		if _, err := settingsPacket(c.sleep, c.response, [3]byte{}); err == nil {
			t.Errorf("aceptó %d, %d", c.sleep, c.response)
		}
		if err := (&WirelessDevice{}).SetSettings(c.sleep, c.response); err == nil {
			t.Errorf("SetSettings aceptó %d, %d", c.sleep, c.response)
		}
	}
	if _, err := settingsPacket(2, 1, [3]byte{2}); err == nil {
		t.Fatal("aceptó casilla fuera de rango")
	}
}
