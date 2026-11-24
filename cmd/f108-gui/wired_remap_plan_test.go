package main

import (
	"testing"

	"github.com/parsiya/f108-pro/pkg/aula"
)

func TestPlanWiredRemapPreservesInsertMacro(t *testing.T) {
	state := guiState{Normal: map[string]string{"116": "macro:0"}, FN: map[string]string{}}
	got, err := planWiredRemap(state, "normal", aula.KeyNameToIndex["capslock"], "esc", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("se esperaban Insert y CapsLock; recibidos %d", len(got))
	}
	for _, r := range got {
		if r.SourceIndex == aula.KeyNameToIndex["insert"] && (r.Action != aula.RemapMacro || r.Param1 != 0) {
			t.Fatalf("Insert perdió la macro W: %+v", r)
		}
	}
}

func TestPlanWiredRemapRejectsCorruptState(t *testing.T) {
	state := guiState{Normal: map[string]string{"116": "unknown:thing"}, FN: map[string]string{}}
	if _, err := planWiredRemap(state, "normal", aula.KeyNameToIndex["capslock"], "esc", false); err == nil {
		t.Fatal("estado desconocido aceptado para una tabla completa")
	}
}
