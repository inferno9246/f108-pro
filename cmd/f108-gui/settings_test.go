package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSettingsPayloadRejectsBeforeHardware(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`{"sleep":4,"response":1}`, "reposo"},
		{`{"sleep":2,"response":0}`, "respuesta"},
		{`{"sleep":2,"response":6}`, "respuesta"},
		{`{"sleep":2}`, "respuesta"},
		{`{"response":1}`, "reposo"},
		{`{"sleep":2,"response":1,"flag":1}`, "desconocido"},
	} {
		_, err := parseSettingsPayload(json.RawMessage(tc.raw))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: esperaba %q, obtuvo %v", tc.raw, tc.want, err)
		}
	}
}

func TestSettingsStateSavedLocally(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	initial, err := settingsState()
	if err != nil {
		t.Fatal(err)
	}
	if initial.Known || initial.Sleep != 2 || initial.Response != 1 {
		t.Fatalf("estado inicial: %+v", initial)
	}
	want := settingsData{Sleep: 3, Response: 4, Known: true}
	if err := settingsSave(want); err != nil {
		t.Fatal(err)
	}
	got, err := settingsState()
	if err != nil || got != want {
		t.Fatalf("estado guardado: %+v, %v", got, err)
	}
	p, err := settingsPath()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("permisos: %v, %v", info, err)
	}
}
