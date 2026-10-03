package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/parsiya/f108-pro/pkg/aula"
)

type invokeReq struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type invokeResp struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type remapPayload struct {
	Transport string `json:"transport"` // por ahora solo "wireless"
	Layer     string `json:"layer"`     // "normal" | "fn"
	KeyIndex  uint8  `json:"key_index"`
	Target    string `json:"target"` // "a", "media:play", "mouse:lclick", "combo:ctrl+c"
}

var deviceMu sync.Mutex // un solo acceso al teclado a la vez

func handleInvoke(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	var req invokeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeResp(w, nil, fmt.Errorf("bad request: %w", err))
		return
	}
	var data any
	var err error
	switch req.Action {
	case "remap_apply":
		data, err = remapApply(req.Payload)
	case "remap_reset_key":
		data, err = remapResetKey(req.Payload)
	case "remap_state":
		data, err = remapState()
	case "remap_reset_all":
		data, err = remapResetAll()
	case "light_off":
		data, err = lightApply(aula.LightingConfig{Mode: aula.ModeOff})
	case "light_default":
		data, err = lightApply(aula.LightingConfig{Mode: aula.ModeBreath, R: 147, G: 51, B: 234, Brightness: 5, Speed: 3})
	case "light_apply":
		data, err = lightApplyPayload(req.Payload)
	case "sync_clock":
		data, err = clockSync()
	case "get_device_status":
		data = map[string]bool{
			"usb_connected":    usbPresent("0c45", "800a"),
			"dongle_connected": usbPresent("05ac", "024f"),
		}
	case "get_app_info":
		data = map[string]string{"cli_path": "(integrado)"}
	default:
		err = fmt.Errorf("unknown action %q", req.Action)
	}
	writeResp(w, data, err)
}

func writeResp(w http.ResponseWriter, data any, err error) {
	resp := invokeResp{Success: err == nil, Data: data}
	if err != nil {
		resp.Error = err.Error()
	}
	json.NewEncoder(w).Encode(resp)
}

func validKeyIndex(idx uint8) bool {
	for _, v := range aula.KeyNameToIndex {
		if v == idx {
			return true
		}
	}
	return false
}

func parseRemapPayload(raw json.RawMessage) (remapPayload, error) {
	var p remapPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("invalid payload: %w", err)
	}
	if p.Transport != "wireless" {
		return p, fmt.Errorf("transport %q not supported yet (only \"wireless\")", p.Transport)
	}
	if p.Layer != "normal" && p.Layer != "fn" {
		return p, fmt.Errorf("layer must be \"normal\" or \"fn\"")
	}
	if !validKeyIndex(p.KeyIndex) {
		return p, fmt.Errorf("unknown key_index %d", p.KeyIndex)
	}
	return p, nil
}

func remapApply(raw json.RawMessage) (any, error) {
	p, err := parseRemapPayload(raw)
	if err != nil {
		return nil, err
	}
	if p.KeyIndex == aula.KeyNameToIndex["fn"] {
		return nil, fmt.Errorf("remapear la tecla FN esta bloqueado por seguridad")
	}
	remap, err := aula.ParseRemapTarget(p.KeyIndex, p.Target)
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

	confirmed := true
	if err := wdev.SetKeyRemap([]aula.KeyRemap{remap}, p.Layer == "fn"); err != nil {
		if !errors.Is(err, aula.ErrNoACK) {
			return nil, err
		}
		confirmed = false // el paquete se envió; el teclado respondió tarde o no respondió
	}
	stateSet(p.Layer, p.KeyIndex, p.Target)
	return map[string]bool{"ok": true, "confirmed": confirmed}, nil
}

func remapResetKey(raw json.RawMessage) (any, error) {
	p, err := parseRemapPayload(raw)
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

	confirmed := true
	if err := wdev.ResetKeys([]uint8{p.KeyIndex}, p.Layer == "fn"); err != nil {
		if !errors.Is(err, aula.ErrNoACK) {
			return nil, err
		}
		confirmed = false
	}
	stateDelete(p.Layer, p.KeyIndex)
	return map[string]bool{"ok": true, "confirmed": confirmed}, nil
}

func usbPresent(vid, pid string) bool {
	dirs, _ := filepath.Glob("/sys/bus/usb/devices/*")
	for _, d := range dirs {
		v, err1 := os.ReadFile(filepath.Join(d, "idVendor"))
		p, err2 := os.ReadFile(filepath.Join(d, "idProduct"))
		if err1 == nil && err2 == nil &&
			strings.TrimSpace(string(v)) == vid && strings.TrimSpace(string(p)) == pid {
			return true
		}
	}
	return false
}
