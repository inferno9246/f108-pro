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

func lightApplyPayload(raw json.RawMessage) (any, error) {
	var p lightPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}
	return lightApply(aula.LightingConfig{
		Mode: aula.LightingMode(p.Mode), R: p.R, G: p.G, B: p.B,
		Brightness: p.Brightness, Speed: p.Speed, Direction: p.Direction, Colorful: p.Colorful,
	})
}

func lightApply(cfg aula.LightingConfig) (any, error) {
	deviceMu.Lock()
	defer deviceMu.Unlock()
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
