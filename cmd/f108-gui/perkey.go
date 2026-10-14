package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/parsiya/f108-pro/pkg/aula"
)

// Luz por tecla por el dongle 2.4G. El teclado no se puede leer, asi que los colores
// aplicados se guardan en ~/.config/f108-pro/gui-perkey.json (permisos 0600) para
// restaurar la pantalla al volver a abrir la interfaz.

type perkeyData struct {
	Brightness int               `json:"brightness"`
	Keys       map[string]string `json:"keys"` // indice de luz -> "#rrggbb"
}

var perkeyMu sync.Mutex

func perkeyPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "f108-pro", "gui-perkey.json"), nil
}

func perkeyLoad() perkeyData {
	st := perkeyData{Brightness: 5, Keys: map[string]string{}}
	if p, err := perkeyPath(); err == nil {
		if data, err := os.ReadFile(p); err == nil {
			_ = json.Unmarshal(data, &st)
		}
	}
	if st.Keys == nil {
		st.Keys = map[string]string{}
	}
	if st.Brightness < 0 || st.Brightness > 5 {
		st.Brightness = 5
	}
	return st
}

func perkeySave(st perkeyData) error {
	p, err := perkeyPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func parseHexColor(s string) (r, g, b uint8, err error) {
	if len(s) != 7 || s[0] != '#' {
		return 0, 0, 0, fmt.Errorf("color invalido %q (use #rrggbb)", s)
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("color invalido %q (use #rrggbb)", s)
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v), nil
}

func perkeyStateHandler() (any, error) {
	perkeyMu.Lock()
	defer perkeyMu.Unlock()
	return perkeyLoad(), nil
}

func perkeyApply(raw json.RawMessage) (any, error) {
	var p perkeyData
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}
	if p.Brightness < 0 || p.Brightness > 5 {
		return nil, fmt.Errorf("el brillo debe estar entre 0 y 5")
	}
	keys := make([]aula.KeyColor, 0, len(p.Keys))
	clean := make(map[string]string, len(p.Keys))
	for k, hex := range p.Keys {
		n, err := strconv.Atoi(k)
		if err != nil || n < 0 || n > 255 || !validKeyIndex(uint8(n)) {
			return nil, fmt.Errorf("indice de tecla desconocido: %q", k)
		}
		r, g, b, err := parseHexColor(hex)
		if err != nil {
			return nil, err
		}
		keys = append(keys, aula.KeyColor{LightIndex: uint8(n), R: r, G: g, B: b})
		clean[strconv.Itoa(n)] = strings.ToLower(hex)
	}

	deviceMu.Lock()
	wdev, err := aula.OpenWireless()
	if err != nil {
		deviceMu.Unlock()
		return nil, err
	}
	err = wdev.SetPerKeyRGB(keys, uint8(p.Brightness))
	wdev.Close()
	deviceMu.Unlock()
	if err != nil {
		return nil, err
	}

	perkeyMu.Lock()
	defer perkeyMu.Unlock()
	_ = perkeySave(perkeyData{Brightness: p.Brightness, Keys: clean})
	return map[string]any{"ok": true, "keys": len(keys)}, nil
}

func batteryStatus() (any, error) {
	deviceMu.Lock()
	defer deviceMu.Unlock()
	wdev, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	defer wdev.Close()
	pct, err := wdev.Battery()
	if err != nil {
		return nil, err
	}
	return map[string]int{"percent": pct}, nil
}
