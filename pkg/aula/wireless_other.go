//go:build !linux

package aula

import (
	"errors"
	"time"
)

type WirelessDevice struct{}

var errWirelessUnsupported = errors.New("el modo 2.4G solo está soportado en Linux por ahora")

func OpenWireless() (*WirelessDevice, error)                     { return nil, errWirelessUnsupported }
func (d *WirelessDevice) Close() error                           { return nil }
func (d *WirelessDevice) SetKeyRemap(_ []KeyRemap, _ bool) error { return errWirelessUnsupported }

func (d *WirelessDevice) ResetKeys(_ []uint8, _ bool) error { return errWirelessUnsupported }

var ErrNoACK = errors.New("sin ACK del teclado")

func (d *WirelessDevice) SetLighting(_ LightingConfig) error { return errWirelessUnsupported }

func (d *WirelessDevice) SyncClock(_ time.Time) error { return errWirelessUnsupported }

func (d *WirelessDevice) SendMusicFrame(_, _, _ byte, _ [21]byte) error {
	return errWirelessUnsupported
}

func (d *WirelessDevice) Battery() (int, error) { return 0, errWirelessUnsupported }

func (d *WirelessDevice) SetPerKeyRGB(_ []KeyColor, _ uint8) error { return errWirelessUnsupported }

func (d *WirelessDevice) SendMacros(_ [][]MacroEvent) error { return errWirelessUnsupported }
