//go:build !linux

package main

import (
	"encoding/json"
	"errors"
)

var errMusicUnsupported = errors.New("el modo musica solo esta soportado en Linux por ahora")

func musicStartPayload(_ json.RawMessage) (any, error) { return nil, errMusicUnsupported }
func musicStop() (any, error)                          { return nil, errMusicUnsupported }
func musicStatus() (any, error)                        { return nil, errMusicUnsupported }
