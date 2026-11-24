//go:build linux

package aula

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"
)

// Solo para una prueba guiada: el usuario debe confirmar que el estado inicial
// es reposo 1, respuesta 1 y casillas 0. El retorno se intenta siempre.
func TestHW_WirelessNoSleepRoundTrip(t *testing.T) {
	if os.Getenv("AULA_TEST_NO_SLEEP") != "1" {
		t.Skip("requiere AULA_TEST_NO_SLEEP=1 y estado inicial confirmado")
	}
	golden, err := os.ReadFile("testdata/settings-seventeen.bin")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		sleep byte
		index int
	}{{0, 6}, {1, 7}} {
		pkt, err := settingsPacket(c.sleep, 1, [3]byte{})
		if err != nil || len(golden) < (c.index+1)*33 || !bytes.Equal(pkt[:], golden[c.index*33:(c.index+1)*33]) {
			t.Fatalf("reposo %d no coincide con la captura: %v", c.sleep, err)
		}
	}
	dev, err := OpenWireless()
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()
	defer func() {
		for attempt := 1; attempt <= 3; attempt++ {
			if err := dev.SetSettings(1, 1); err == nil {
				fmt.Println("RETORNO: reposo 1 minuto y respuesta 1 confirmados por ACK")
				return
			} else if attempt == 3 {
				t.Errorf("RETORNO SIN ACK tras %d intentos: %v", attempt, err)
			}
		}
	}()
	if err := dev.SetSettings(0, 1); err != nil {
		t.Fatalf("sin ACK al configurar Sin reposo: %v", err)
	}
	fmt.Println("INICIO OBSERVACIÓN: Sin reposo con ACK; no pulses teclas durante 100 segundos")
	time.Sleep(100 * time.Second)
}
