package aula

import "testing"

func TestParseMacroTarget(t *testing.T) {
	for _, tc := range []struct {
		spec string
		slot [4]byte
	}{
		{"macro:0", [4]byte{6, 0, 0, 0}},
		{"macro:99:times:255", [4]byte{6, 99, 1, 255}},
		{"macro:3:toggle", [4]byte{6, 3, 2, 0}},
		{"macro:4:hold", [4]byte{6, 4, 3, 1}},
	} {
		r, err := ParseRemapTarget(12, tc.spec)
		if err != nil {
			t.Fatalf("%s: %v", tc.spec, err)
		}
		got := [4]byte{byte(r.Action), r.Param1, r.Param2, r.Param3}
		if r.SourceIndex != 12 || got != tc.slot {
			t.Errorf("%s: got %v", tc.spec, got)
		}
	}
	for _, bad := range []string{"macro:", "macro:100", "macro:-1", "macro:1:times:0", "macro:1:times:256", "macro:1:hold:3", "macro:1:toggle:1"} {
		if _, err := ParseRemapTarget(12, bad); err == nil {
			t.Errorf("acepto %q", bad)
		}
	}
}
