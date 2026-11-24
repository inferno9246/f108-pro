package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/parsiya/f108-pro/pkg/aula"
)

type fakeWiredDevice struct {
	lighting *aula.LightingConfig
	clock    bool
	keys     []aula.KeyColor
	bright   uint8
	closed   bool
}

func TestWiredAPIUsesCableAndBlocksUnsupportedRemap(t *testing.T) {
	previousOpen, previousPresent := openWiredDevice, devicePresent
	defer func() { openWiredDevice, devicePresent = previousOpen, previousPresent }()
	devicePresent = func(vid, pid string) bool { return vid == "0c45" && pid == "800a" }
	d := new(fakeWiredDevice)
	openWiredDevice = func() (wiredController, error) { return d, nil }

	call := func(body string) invokeResp {
		t.Helper()
		req := httptest.NewRequest("POST", "/api/invoke", strings.NewReader(body))
		res := httptest.NewRecorder()
		handleInvoke(res, req)
		var got invokeResp
		if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		return got
	}
	if result := call(`{"action":"light_default","transport":"wired"}`); !result.Success || d.lighting == nil {
		t.Fatalf("iluminación USB: %+v", result)
	}
	if result := call(`{"action":"remap_apply","transport":"wired","payload":{"layer":"normal","key_index":116,"target":"a"}}`); result.Success || !strings.Contains(result.Error, "2.4G") {
		t.Fatalf("remapeo USB debería estar bloqueado: %+v", result)
	}
	if result := call(`{"action":"light_default","transport":"bluetooth"}`); result.Success || !strings.Contains(result.Error, "desconocido") {
		t.Fatalf("transporte inválido debería rechazarse: %+v", result)
	}
}

func (d *fakeWiredDevice) SetLighting(cfg aula.LightingConfig) error {
	d.lighting = &cfg
	return nil
}
func (d *fakeWiredDevice) SyncClock(time.Time) error { d.clock = true; return nil }
func (d *fakeWiredDevice) SetPerKeyRGB(keys []aula.KeyColor, brightness uint8) error {
	d.keys, d.bright = keys, brightness
	return nil
}
func (d *fakeWiredDevice) Close() { d.closed = true }

func TestWiredGUIRoutesKnownActions(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	previous := openWiredDevice
	defer func() { openWiredDevice = previous }()
	var opened []*fakeWiredDevice
	openWiredDevice = func() (wiredController, error) {
		d := new(fakeWiredDevice)
		opened = append(opened, d)
		return d, nil
	}

	if _, err := lightApply(aula.DefaultLightingConfig(), "wired"); err != nil {
		t.Fatal(err)
	}
	if _, err := clockSync("wired"); err != nil {
		t.Fatal(err)
	}
	if _, err := perkeyApply(json.RawMessage(`{"brightness":3,"keys":{"1":"#112233"}}`), "wired"); err != nil {
		t.Fatal(err)
	}
	if len(opened) != 3 || !opened[0].closed || !opened[1].closed || !opened[2].closed {
		t.Fatalf("aperturas/cierres: %+v", opened)
	}
	if opened[0].lighting == nil || *opened[0].lighting != aula.DefaultLightingConfig() {
		t.Fatal("iluminación no enviada al transporte cableado")
	}
	if !opened[1].clock || len(opened[2].keys) != 1 || opened[2].keys[0].LightIndex != 1 || opened[2].bright != 3 {
		t.Fatal("reloj o luz por tecla no enviados al transporte cableado")
	}
}

func TestWiredGUIRejectsInvalidInputBeforeOpeningDevice(t *testing.T) {
	previous := openWiredDevice
	defer func() { openWiredDevice = previous }()
	opened := false
	openWiredDevice = func() (wiredController, error) {
		opened = true
		return new(fakeWiredDevice), nil
	}
	if _, err := lightApplyPayload(json.RawMessage(`{"mode":20}`), "wired"); err == nil {
		t.Fatal("modo inválido aceptado")
	}
	if _, err := perkeyApply(json.RawMessage(`{"brightness":6,"keys":{}}`), "wired"); err == nil {
		t.Fatal("brillo inválido aceptado")
	}
	if opened {
		t.Fatal("se abrió el teclado con datos inválidos")
	}
}
