package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseHexColor(t *testing.T) {
	r, g, b, err := parseHexColor("#ff8000")
	if err != nil || r != 255 || g != 128 || b != 0 {
		t.Fatalf("got %d %d %d %v", r, g, b, err)
	}
	for _, bad := range []string{"", "ff8000", "#ff80", "#gg0000", "#ff8000ff", "#+f8000"} {
		if _, _, _, err := parseHexColor(bad); err == nil {
			t.Errorf("deberia rechazar %q", bad)
		}
	}
}

func TestPerkeyApplyValidatesBeforeTouchingTheKeyboard(t *testing.T) {
	cases := map[string]string{
		`{"brightness":6,"keys":{}}`:                      "brillo",
		`{"brightness":5,"keys":{"999":"#ffffff"}}`:       "indice",
		`{"brightness":5,"keys":{"abc":"#ffffff"}}`:       "indice",
		`{"brightness":5,"keys":{"1":"rojo"}}`:            "color",
		`{"brightness":5,"keys":{"1":"#ff0000","2":"x"}}`: "color",
		`no es json`: "payload",
	}
	for body, want := range cases {
		_, err := perkeyApply(json.RawMessage(body))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: esperaba un error con %q, obtuve %v", body, want, err)
		}
	}
}
