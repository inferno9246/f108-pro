package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parsiya/f108-pro/pkg/aula"
)

type lightPayload struct {
	Mode       int   `json:"mode"`
	R          uint8 `json:"r"`
	G          uint8 `json:"g"`
	B          uint8 `json:"b"`
	Brightness uint8 `json:"brightness"`
	Speed      uint8 `json:"speed"`
	Direction  uint8 `json:"direction"`
	Colorful   bool  `json:"colorful"`
}

func lightApplyPayload(raw json.RawMessage, transport string) (any, error) {
	var p lightPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}
	return lightApply(aula.LightingConfig{
		Mode: aula.LightingMode(p.Mode), R: p.R, G: p.G, B: p.B,
		Brightness: p.Brightness, Speed: p.Speed, Direction: p.Direction, Colorful: p.Colorful,
	}, transport)
}

func lightApply(cfg aula.LightingConfig, transport string) (any, error) {
	if cfg.Mode < 0 || cfg.Mode > 19 || cfg.Brightness > 5 || cfg.Speed > 5 || cfg.Direction > 1 {
		return nil, fmt.Errorf("iluminación fuera de rango")
	}
	deviceMu.Lock()
	defer deviceMu.Unlock()
	if transport == "wired" {
		dev, err := openWiredDevice()
		if err != nil {
			return nil, err
		}
		defer dev.Close()
		if err := dev.SetLighting(cfg); err != nil {
			return nil, err
		}
		return map[string]bool{"ok": true, "confirmed": true}, nil
	}
	if transport != "wireless" {
		return nil, fmt.Errorf("transporte no soportado: %q", transport)
	}
	wdev, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	defer wdev.Close()

	confirmed := true
	if err := wdev.SetLighting(cfg); err != nil {
		if !errors.Is(err, aula.ErrNoACK) {
			return nil, err
		}
		confirmed = false
	}
	return map[string]bool{"ok": true, "confirmed": confirmed}, nil
}
