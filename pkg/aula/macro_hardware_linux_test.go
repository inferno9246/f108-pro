//go:build linux

package aula

import (
	"os"
	"testing"
)

// Solo se ejecuta con autorización explícita: sustituye todas las macros del teclado.
// La secuencia enviada coincide byte a byte con macro-one.bin de todo.pcapng.
func TestHW_WirelessMacroOne(t *testing.T) {
	if os.Getenv("AULA_TEST_MACRO") != "1" {
		t.Skip("requiere AULA_TEST_MACRO=1; sustituye todas las macros del teclado")
	}
	delay, err := MacroDelay(10)
	if err != nil {
		t.Fatal(err)
	}
	dev, err := OpenWireless()
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()
	if err := dev.SendMacros([][]MacroEvent{{MacroKeyDown(0x1A), delay, MacroKeyUp(0x1A), delay}}); err != nil {
		t.Fatal(err)
	}
}
