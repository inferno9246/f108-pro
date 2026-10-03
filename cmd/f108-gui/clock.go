package main

import (
	"errors"
	"time"

	"github.com/parsiya/f108-pro/pkg/aula"
)

func clockSync() (any, error) {
	deviceMu.Lock()
	defer deviceMu.Unlock()
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
