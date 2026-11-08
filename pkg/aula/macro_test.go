package aula

import (
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"testing"
)

func TestMacroPacketsMatchOfficialCaptures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		macros [][]MacroEvent
	}{
		{"macro-one.bin", [][]MacroEvent{{MacroKeyDown(0x1a), mustDelay(t, 10), MacroKeyUp(0x1a), mustDelay(t, 10)}}},
		{"macro-three.bin", capturedThreeMacros(t)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			golden, err := os.ReadFile("testdata/" + tc.name)
			if err != nil {
				t.Fatal(err)
			}
			table, err := BuildMacroTable(tc.macros)
			if err != nil {
				t.Fatal(err)
			}
			packets, err := MacroPackets(table)
			if err != nil {
				t.Fatal(err)
			}
			if len(golden) != len(packets)*33 {
				t.Fatalf("longitud: %d != %d", len(golden), len(packets)*33)
			}
			for i, payload := range packets {
				got := testWirelessPacket(payload)
				want := golden[i*33 : (i+1)*33]
				if !bytes.Equal(got[:], want) {
					t.Fatalf("paquete %d\n got % x\nwant % x", i, got, want)
				}
			}
		})
	}
}

// Reconstruye los eventos de la segunda captura sin reutilizar el serializador probado.
func capturedThreeMacros(t *testing.T) [][]MacroEvent {
	t.Helper()
	b, err := os.ReadFile("testdata/macro-three.bin")
	if err != nil {
		t.Fatal(err)
	}
	var table []byte
	for i := 0; i < 22; i++ {
		table = append(table, b[i*33+4:i*33+32]...)
	}
	var out [][]MacroEvent
	for i := 0; i < 3; i++ {
		off := int(binary.LittleEndian.Uint32(table[i*4:]))
		n := int(binary.LittleEndian.Uint16(table[off:]))
		events := make([]MacroEvent, n)
		for j := range events {
			copy(events[j][:], table[off+8+j*4:off+12+j*4])
		}
		out = append(out, events)
	}
	return out
}

func testWirelessPacket(payload []byte) [33]byte {
	var pkt [33]byte
	copy(pkt[1:], payload)
	for _, b := range pkt[:32] {
		pkt[32] += b
	}
	return pkt
}

func mustDelay(t *testing.T, ms int) MacroEvent {
	t.Helper()
	e, err := MacroDelay(ms)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestParseMacroEventsAndLimits(t *testing.T) {
	events, err := ParseMacroEvents([]string{"down:a", "up:a", "delay:30", "tap:b:20", "text:ab"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3+4+8 {
		t.Fatalf("eventos: %d", len(events))
	}
	if events[0] != MacroKeyDown(4) || events[1] != MacroKeyUp(4) || events[2] != mustDelay(t, 30) {
		t.Fatal(events[:3])
	}
	for _, bad := range [][]string{{"delay:9"}, {"delay:65536"}, {"down:fn"}, {"tap:a:0"}, {"text:A"}, {"unknown:a"}} {
		if _, err := ParseMacroEvents(bad); err == nil {
			t.Errorf("acepto %v", bad)
		}
	}
	if _, err := BuildMacroTable([][]MacroEvent{{}}); err == nil {
		t.Error("acepto macro vacia")
	}
	if _, err := BuildMacroTable(make([][]MacroEvent, 101)); err == nil {
		t.Error("acepto 101 macros")
	}
	many := make([]MacroEvent, 800)
	for i := range many {
		many[i] = MacroKeyDown(4)
	}
	if _, err := BuildMacroTable([][]MacroEvent{many}); err == nil {
		t.Error("acepto tabla grande")
	}
	if _, err := BuildMacroTable([][]MacroEvent{{{0, 0, 4, 0x90}}}); err == nil {
		t.Error("acepto raton sin verificar")
	}
	if _, err := BuildMacroTable([][]MacroEvent{{MacroKeyDown(0xff)}}); err == nil {
		t.Error("acepto HID desconocido")
	}
	if _, err := MacroPackets([]byte(strings.Repeat("x", 3585))); err == nil {
		t.Error("acepto 3585 bytes")
	}
}

func TestSendMacrosRejectsBeforeDeviceIO(t *testing.T) {
	d := &WirelessDevice{}
	if err := d.SendMacros([][]MacroEvent{{}}); err == nil {
		t.Fatal("la macro vacía debería fallar antes de abrir hidraw")
	}
	if err := d.SendMacros(nil); err == nil {
		t.Fatal("el borrado total sin captura debería fallar antes de abrir hidraw")
	}
}
