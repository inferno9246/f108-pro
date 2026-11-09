package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/parsiya/f108-pro/pkg/aula"
)

type settingsData struct {
	Sleep    int  `json:"sleep"`
	Response int  `json:"response"`
	Known    bool `json:"known"` // solo indica que esta GUI guardó un envío anterior
}

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "f108-pro", "gui-settings.json"), nil
}

func settingsState() (settingsData, error) {
	st := settingsData{Sleep: 2, Response: 1}
	p, err := settingsPath()
	if err != nil {
		return st, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return st, fmt.Errorf("estado de ajustes inválido: %w", err)
	}
	if st.Sleep < 0 || st.Sleep > 3 || st.Response < 1 || st.Response > 5 {
		return st, fmt.Errorf("estado de ajustes fuera de rango")
	}
	return st, nil
}

func settingsSave(st settingsData) error {
	p, err := settingsPath()
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

func parseSettingsPayload(raw json.RawMessage) (settingsData, error) {
	var req struct {
		Sleep    *int `json:"sleep"`
		Response *int `json:"response"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		if strings.Contains(err.Error(), "unknown field") {
			return settingsData{}, fmt.Errorf("campo desconocido: %w", err)
		}
		return settingsData{}, fmt.Errorf("ajustes inválidos: %w", err)
	}
	if req.Sleep == nil || *req.Sleep < 0 || *req.Sleep > 3 {
		return settingsData{}, fmt.Errorf("el reposo debe estar entre 0 y 3")
	}
	if req.Response == nil || *req.Response < 1 || *req.Response > 5 {
		return settingsData{}, fmt.Errorf("el tiempo de respuesta debe estar entre 1 y 5")
	}
	return settingsData{Sleep: *req.Sleep, Response: *req.Response, Known: true}, nil
}

func settingsApply(raw json.RawMessage) (any, error) {
	st, err := parseSettingsPayload(raw)
	if err != nil {
		return nil, err
	}
	deviceMu.Lock()
	defer deviceMu.Unlock()
	d, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	defer d.Close()
	confirmed := true
	if err := d.SetSettings(byte(st.Sleep), byte(st.Response)); err != nil {
		if !errors.Is(err, aula.ErrNoACK) {
			return nil, err
		}
		confirmed = false
	}
	if err := settingsSave(st); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true, "confirmed": confirmed}, nil
}
