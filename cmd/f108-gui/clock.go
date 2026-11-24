package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/parsiya/f108-pro/pkg/aula"
)

func clockSync(transport string) (any, error) {
	deviceMu.Lock()
	defer deviceMu.Unlock()
	if transport == "wired" {
		dev, err := openWiredDevice()
		if err != nil {
			return nil, err
		}
		defer dev.Close()
		if err := dev.SyncClock(time.Now()); err != nil {
			return nil, err
		}
		return map[string]bool{"ok": true, "confirmed": true}, nil
	}
	if transport != "wireless" {
		return nil, fmt.Errorf("transporte no soportado: %q", transport)
	}
	wdev, err := aula.OpenWireless()
	if err != nil {
		return nil, err
	}
	defer wdev.Close()

	confirmed := true
	if err := wdev.SyncClock(time.Now()); err != nil {
		if !errors.Is(err, aula.ErrNoACK) {
			return nil, err
		}
		confirmed = false
	}
	return map[string]bool{"ok": true, "confirmed": confirmed}, nil
}
