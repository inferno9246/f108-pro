package aula

import (
	"fmt"
	"sort"
	"strings"
)

// ParseRemapTarget converts a destination specifier into a KeyRemap.
// Formats: key name ("a"), "media:play", "mouse:lclick", "combo:ctrl+c".
func ParseRemapTarget(srcIdx uint8, spec string) (KeyRemap, error) {
	spec = strings.ToLower(strings.TrimSpace(spec))

	if name, ok := strings.CutPrefix(spec, "media:"); ok {
		code, ok := ConsumerNameToCode[name]
		if !ok {
			return KeyRemap{}, fmt.Errorf("unknown media action '%s' (options: %s)", name, strings.Join(sortedKeys(ConsumerNameToCode), ", "))
		}
		return NewConsumerRemap(srcIdx, code), nil
	}

	if name, ok := strings.CutPrefix(spec, "mouse:"); ok {
		params, ok := MouseNameToParams[name]
		if !ok {
			return KeyRemap{}, fmt.Errorf("unknown mouse action '%s' (options: %s)", name, strings.Join(sortedKeys(MouseNameToParams), ", "))
		}
		return NewMouseRemap(srcIdx, params[0], params[1], params[2]), nil
	}

	if expr, ok := strings.CutPrefix(spec, "combo:"); ok {
		return parseComboExpr(srcIdx, expr)
	}

	dstHID, ok := KeyNameToHID[spec]
	if !ok {
		return KeyRemap{}, fmt.Errorf("unknown key '%s' (run 'f108-pro keys' to list)", spec)
	}
	return NewKeySwap(srcIdx, dstHID), nil
}

// parseComboExpr parses "ctrl+c" or "win+shift+d".
func parseComboExpr(srcIdx uint8, expr string) (KeyRemap, error) {
	modMap := map[string]byte{
		"ctrl": 0x01, "shift": 0x02, "alt": 0x04, "win": 0x08,
		"rctrl": 0x10, "rshift": 0x20, "ralt": 0x40, "rwin": 0x80,
	}

	parts := strings.Split(expr, "+")
	if len(parts) < 2 {
		return KeyRemap{}, fmt.Errorf("combo must be modifier+key (e.g., ctrl+c)")
	}

	var modBits byte
	for _, part := range parts[:len(parts)-1] {
		bit, ok := modMap[part]
		if !ok {
			return KeyRemap{}, fmt.Errorf("unknown modifier '%s' (options: ctrl, shift, alt, win, rctrl, rshift, ralt, rwin)", part)
		}
		modBits |= bit
	}

	keyName := parts[len(parts)-1]
	keyHID, ok := KeyNameToHID[keyName]
	if !ok {
		return KeyRemap{}, fmt.Errorf("unknown key '%s' in combo (run 'f108-pro keys' to list)", keyName)
	}
	return NewKeyCombo(srcIdx, modBits, keyHID), nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
