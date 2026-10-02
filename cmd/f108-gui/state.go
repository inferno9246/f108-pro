package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"github.com/parsiya/f108-pro/pkg/aula"
)

// guiState recuerda qué teclas se remaparon desde esta interfaz (el teclado no se
// puede leer). Se guarda en ~/.config/f108-pro/gui-state.json con permisos 0600.
type guiState struct {
	Normal map[string]string `json:"normal"`
	FN     map[string]string `json:"fn"`
}

var stateMu sync.Mutex

func statePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "f108-pro", "gui-state.json"), nil
}

func loadState() guiState {
	st := guiState{}
	if p, err := statePath(); err == nil {
		if data, err := os.ReadFile(p); err == nil {
			_ = json.Unmarshal(data, &st)
		}
	}
	if st.Normal == nil {
		st.Normal = map[string]string{}
	}
	if st.FN == nil {
		st.FN = map[string]string{}
	}
	return st
}

func saveState(st guiState) {
	p, err := statePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmp, p)
}

func layerMap(st *guiState, layer string) map[string]string {
	if layer == "fn" {
		return st.FN
	}
	return st.Normal
}

func stateSet(layer string, idx uint8, target string) {
	stateMu.Lock()
	defer stateMu.Unlock()
	st := loadState()
	layerMap(&st, layer)[strconv.Itoa(int(idx))] = target
	saveState(st)
}

func stateDelete(layer string, idx uint8) {
	stateMu.Lock()
	defer stateMu.Unlock()
	st := loadState()
	delete(layerMap(&st, layer), strconv.Itoa(int(idx)))
	saveState(st)
}

func remapState() (any, error) {
	stateMu.Lock()
	defer stateMu.Unlock()
	return loadState(), nil
}

// remapResetAll restaura solo las teclas que esta interfaz modificó, una a una.
// Una tecla sin ACK se cuenta como "sin confirmar" y no detiene el proceso.
func remapResetAll() (any, error) {
	deviceMu.Lock()
	defer deviceMu.Unlock()
	stateMu.Lock()
	defer stateMu.Unlock()

	st := loadState()
	if len(st.Normal)+len(st.FN) == 0 {
		return map[string]int{"restored": 0, "unconfirmed": 0}, nil
	}
	wdev, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	defer wdev.Close()

	restored, unconfirmed := 0, 0
	for _, layer := range []string{"normal", "fn"} {
		m := layerMap(&st, layer)
		var idxs []uint8
		for k := range m {
			n, err := strconv.Atoi(k)
			if err != nil || n < 0 || n > 255 || !validKeyIndex(uint8(n)) {
				delete(m, k) // entrada inválida: se descarta
				continue
			}
			idxs = append(idxs, uint8(n))
		}
		sort.Slice(idxs, func(i, j int) bool { return idxs[i] < idxs[j] })
		for _, idx := range idxs {
			if err := wdev.ResetKeys([]uint8{idx}, layer == "fn"); err != nil {
				if !errors.Is(err, aula.ErrNoACK) {
					saveState(st)
					return nil, err
				}
				unconfirmed++
			}
			delete(m, strconv.Itoa(int(idx)))
			restored++
		}
	}
	saveState(st)
	return map[string]int{"restored": restored, "unconfirmed": unconfirmed}, nil
}
