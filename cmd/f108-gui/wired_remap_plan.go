package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/parsiya/f108-pro/pkg/aula"
)

// planWiredRemap builds the complete replacement layer without touching hardware.
// A USB remap must preserve every assignment known to the GUI. This plan still
// cannot prove that the keyboard has no assignments created outside the GUI.
func planWiredRemap(st guiState, layer string, index uint8, target string, reset bool) ([]aula.KeyRemap, error) {
	if layer != "normal" && layer != "fn" {
		return nil, fmt.Errorf("capa desconocida: %q", layer)
	}
	if !validKeyIndex(index) || index == aula.KeyNameToIndex["fn"] {
		return nil, fmt.Errorf("tecla no permitida: %d", index)
	}
	entries := st.Normal
	if layer == "fn" {
		entries = st.FN
	}
	remaps := make([]aula.KeyRemap, 0, len(entries)+1)
	seen := make(map[uint8]bool, len(entries))
	for key, spec := range entries {
		n, err := strconv.Atoi(key)
		if err != nil || n < 0 || n > 255 || !validKeyIndex(uint8(n)) || uint8(n) == aula.KeyNameToIndex["fn"] {
			return nil, fmt.Errorf("índice inválido en el estado de remapeo: %q", key)
		}
		idx := uint8(n)
		if seen[idx] {
			return nil, fmt.Errorf("índice duplicado en el estado de remapeo: %d", idx)
		}
		seen[idx] = true
		if idx == index {
			continue
		}
		r, err := aula.ParseRemapTarget(idx, spec)
		if err != nil {
			return nil, fmt.Errorf("asignación actual de la tecla %d: %w", idx, err)
		}
		remaps = append(remaps, r)
	}
	if !reset {
		r, err := aula.ParseRemapTarget(index, target)
		if err != nil {
			return nil, err
		}
		remaps = append(remaps, r)
	}
	sort.Slice(remaps, func(i, j int) bool { return remaps[i].SourceIndex < remaps[j].SourceIndex })
	return remaps, nil
}
