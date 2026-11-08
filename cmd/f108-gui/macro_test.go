package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMacroPayloadValidatesBeforeDevice(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`{"macros":[{"name":"vacia","events":[]}]}`, "eventos"},
		{`{"macros":[{"name":"x","events":["delay:9"]}]}`, "pausa"},
		{`{"macros":[{"name":"x","events":["down:fn"]}]}`, "tecla"},
		{`{"macros":[{"name":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx","events":["tap:a"]}]}`, "nombre"},
		{`{"macros":[{"name":"x","events":["down:a"],"extra":1}]}`, "desconocido"},
		{`{"macros":null}`, "macros"},
	} {
		_, _, err := parseMacroPayload(json.RawMessage(tc.raw))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: esperaba %q, obtuvo %v", tc.raw, tc.want, err)
		}
	}
}

func TestMacroStateUsesConfigDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	state := macroData{Macros: []macroDefinition{{Name: "prueba", Events: []string{"tap:a"}}}}
	if err := macroSave(state); err != nil {
		t.Fatal(err)
	}
	got := macroLoad()
	if len(got.Macros) != 1 || got.Macros[0].Name != "prueba" {
		t.Fatalf("estado: %+v", got)
	}
	p, err := macroPath()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Clean(p))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("permisos: %v, %v", info, err)
	}
}
