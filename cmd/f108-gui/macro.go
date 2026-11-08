package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/parsiya/f108-pro/pkg/aula"
)

type macroDefinition struct {
	Name   string   `json:"name"`
	Events []string `json:"events"`
}

type macroData struct {
	Macros []macroDefinition `json:"macros"`
}

var macroMu sync.Mutex

func macroPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "f108-pro", "gui-macros.json"), nil
}

func macroLoad() macroData {
	st := macroData{Macros: []macroDefinition{}}
	if p, err := macroPath(); err == nil {
		if b, err := os.ReadFile(p); err == nil {
			_ = json.Unmarshal(b, &st)
		}
	}
	if st.Macros == nil {
		st.Macros = []macroDefinition{}
	}
	return st
}

func macroSave(st macroData) error {
	p, err := macroPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func parseMacroPayload(raw json.RawMessage) (macroData, [][]aula.MacroEvent, error) {
	var st macroData
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&st); err != nil {
		if strings.Contains(err.Error(), "unknown field") {
			return st, nil, fmt.Errorf("campo desconocido: %w", err)
		}
		return st, nil, fmt.Errorf("payload de macros inválido: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return st, nil, fmt.Errorf("payload de macros inválido: datos adicionales")
	}
	if st.Macros == nil {
		return st, nil, fmt.Errorf("se requiere una lista de macros")
	}
	macros := make([][]aula.MacroEvent, len(st.Macros))
	for i, m := range st.Macros {
		if utf8.RuneCountInString(m.Name) > 40 {
			return st, nil, fmt.Errorf("nombre de macro %d supera 40 caracteres", i)
		}
		if len(m.Events) == 0 {
			return st, nil, fmt.Errorf("macro %d sin eventos", i)
		}
		events, err := aula.ParseMacroEvents(m.Events)
		if err != nil {
			return st, nil, fmt.Errorf("macro %d: %w", i, err)
		}
		macros[i] = events
	}
	if _, err := aula.BuildMacroTable(macros); err != nil {
		return st, nil, err
	}
	return st, macros, nil
}

func macroState() (any, error) {
	macroMu.Lock()
	defer macroMu.Unlock()
	return macroLoad(), nil
}

func macroSend(raw json.RawMessage) (any, error) {
	st, macros, err := parseMacroPayload(raw)
	if err != nil {
		return nil, err
	}
	deviceMu.Lock()
	defer deviceMu.Unlock()
	wdev, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	defer wdev.Close()
	if err := wdev.SendMacros(macros); err != nil {
		return nil, err
	}
	macroMu.Lock()
	defer macroMu.Unlock()
	if err := macroSave(st); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "macros": len(macros)}, nil
}

func macroClear() (any, error) {
	deviceMu.Lock()
	defer deviceMu.Unlock()
	wdev, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	defer wdev.Close()
	if err := wdev.SendMacros(nil); err != nil {
		return nil, err
	}
	macroMu.Lock()
	defer macroMu.Unlock()
	if err := macroSave(macroData{Macros: []macroDefinition{}}); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
