package aula

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// MacroEvent es un evento de cuatro bytes del formato 2.4G.
type MacroEvent [4]byte

func MacroKeyDown(hid byte) MacroEvent { return MacroEvent{0, 0, hid, 0xB0} }
func MacroKeyUp(hid byte) MacroEvent   { return MacroEvent{0, 0, hid, 0x30} }

func MacroDelay(ms int) (MacroEvent, error) {
	if ms < 10 || ms > 65535 {
		return MacroEvent{}, fmt.Errorf("la pausa debe estar entre 10 y 65535 ms")
	}
	return MacroEvent{byte(ms), byte(ms >> 8), 0, 0x50}, nil
}

// ParseMacroEvents acepta down:a, up:a, delay:30, tap:a[:ms] y text:abc.
// text solo acepta caracteres presentes literalmente en el mapa HID, sin modificadores.
func ParseMacroEvents(items []string) ([]MacroEvent, error) {
	var events []MacroEvent
	key := func(name string) (byte, error) {
		hid, ok := KeyNameToHID[strings.ToLower(name)]
		if !ok || hid == 0xAF {
			return 0, fmt.Errorf("tecla desconocida en la macro: %q", name)
		}
		return hid, nil
	}
	for _, item := range items {
		kind, arg, ok := strings.Cut(item, ":")
		if !ok {
			return nil, fmt.Errorf("evento de macro inválido: %q", item)
		}
		kind = strings.ToLower(kind)
		switch kind {
		case "down", "up":
			hid, err := key(arg)
			if err != nil {
				return nil, err
			}
			if kind == "down" {
				events = append(events, MacroKeyDown(hid))
			} else {
				events = append(events, MacroKeyUp(hid))
			}
		case "delay":
			ms, err := strconv.Atoi(arg)
			if err != nil {
				return nil, fmt.Errorf("pausa inválida: %q", arg)
			}
			ev, err := MacroDelay(ms)
			if err != nil {
				return nil, err
			}
			events = append(events, ev)
		case "tap":
			name, duration, hasDuration := strings.Cut(arg, ":")
			hid, err := key(name)
			if err != nil {
				return nil, err
			}
			ms := 10
			if hasDuration {
				ms, err = strconv.Atoi(duration)
				if err != nil {
					return nil, fmt.Errorf("pausa inválida: %q", duration)
				}
			}
			pause, err := MacroDelay(ms)
			if err != nil {
				return nil, err
			}
			events = append(events, MacroKeyDown(hid), pause, MacroKeyUp(hid), pause)
		case "text":
			for _, ch := range arg {
				if _, ok := KeyNameToHID[string(ch)]; !ok {
					return nil, fmt.Errorf("carácter sin tecla directa en la macro: %q", ch)
				}
				hid, err := key(string(ch))
				if err != nil {
					return nil, err
				}
				pause, _ := MacroDelay(10)
				events = append(events, MacroKeyDown(hid), pause, MacroKeyUp(hid), pause)
			}
		default:
			return nil, fmt.Errorf("evento de macro desconocido: %q", kind)
		}
	}
	return events, nil
}

// BuildMacroTable construye la tabla completa. Una lista vacía borra todas las macros.
func BuildMacroTable(macros [][]MacroEvent) ([]byte, error) {
	if len(macros) > 100 {
		return nil, fmt.Errorf("máximo 100 macros")
	}
	totalEvents := 0
	for i, events := range macros {
		if len(events) == 0 {
			return nil, fmt.Errorf("macro %d sin eventos", i)
		}
		totalEvents += len(events)
		for j, ev := range events {
			switch ev[3] {
			case 0xB0, 0x30:
				known := false
				for _, hid := range KeyNameToHID {
					if hid == ev[2] && hid != 0xAF {
						known = true
						break
					}
				}
				if ev[0] != 0 || ev[1] != 0 || !known {
					return nil, fmt.Errorf("evento %d de macro %d inválido", j, i)
				}
			case 0x50:
				ms := int(binary.LittleEndian.Uint16(ev[:2]))
				if ev[2] != 0 || ms < 10 {
					return nil, fmt.Errorf("pausa %d de macro %d inválida", j, i)
				}
			default:
				return nil, fmt.Errorf("evento %d de macro %d no verificado", j, i)
			}
		}
	}
	if (len(macros)+totalEvents)*8 > 3600 || 400+len(macros)*8+totalEvents*4 > 3584 {
		return nil, fmt.Errorf("tabla de macros demasiado grande")
	}
	table := make([]byte, 400, 400+len(macros)*8+totalEvents*4)
	for i, events := range macros {
		binary.LittleEndian.PutUint32(table[i*4:], uint32(len(table)))
		block := make([]byte, 8+4*len(events))
		binary.LittleEndian.PutUint16(block, uint16(len(events)))
		for j, ev := range events {
			copy(block[8+4*j:], ev[:])
		}
		table = append(table, block...)
	}
	return table, nil
}

// MacroPackets devuelve los payloads 09 1C y el cierre 09 <útiles> <trozos>.
func MacroPackets(table []byte) ([][]byte, error) {
	if len(table) < 400 || len(table) > 3584 {
		return nil, fmt.Errorf("longitud de tabla de macros inválida: %d", len(table))
	}
	n := (len(table) + 27) / 28
	if n > 128 {
		return nil, fmt.Errorf("demasiados trozos de macro")
	}
	out := make([][]byte, 0, n+1)
	for i := 0; i < n; i++ {
		p := make([]byte, 31)
		p[0], p[1], p[2] = 0x09, 0x1C, byte(i)
		copy(p[3:], table[i*28:])
		out = append(out, p)
	}
	fin := make([]byte, 31)
	fin[0], fin[1], fin[2] = 0x09, byte(len(table)-28*(n-1)), byte(n)
	fin[29], fin[30] = 0xAA, 0x55
	return append(out, fin), nil
}
