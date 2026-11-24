package main

import "testing"

func TestChooseTransport(t *testing.T) {
	for _, tc := range []struct {
		name, requested string
		usb, dongle     bool
		want            string
		fails           bool
	}{
		{"auto prefers USB", "auto", true, true, "wired", false},
		{"empty means auto", "", false, true, "wireless", false},
		{"explicit dongle", "wireless", true, true, "wireless", false},
		{"explicit cable", "wired", true, true, "wired", false},
		{"missing cable", "wired", false, true, "", true},
		{"missing dongle", "wireless", true, false, "", true},
		{"nothing connected", "auto", false, false, "", true},
		{"invalid choice", "bluetooth", true, true, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := chooseTransport(tc.requested, tc.usb, tc.dongle)
			if got != tc.want || (err != nil) != tc.fails {
				t.Fatalf("chooseTransport(%q, %v, %v) = %q, %v", tc.requested, tc.usb, tc.dongle, got, err)
			}
		})
	}
}
