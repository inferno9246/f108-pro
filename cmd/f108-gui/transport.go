package main

import (
	"fmt"
	"time"

	"github.com/parsiya/f108-pro/pkg/aula"
)

type wiredController interface {
	SetLighting(aula.LightingConfig) error
	SyncClock(time.Time) error
	SetPerKeyRGB([]aula.KeyColor, uint8) error
	Close()
}

var openWiredDevice = func() (wiredController, error) { return aula.Open() }
var devicePresent = usbPresent

// chooseTransport selecciona el dispositivo para una acción. En automático,
// el USB-C tiene prioridad cuando ambos identificadores están presentes.
func chooseTransport(requested string, usb, dongle bool) (string, error) {
	switch requested {
	case "", "auto":
		if usb {
			return "wired", nil
		}
		if dongle {
			return "wireless", nil
		}
		return "", fmt.Errorf("no se detecta el teclado USB-C ni el dongle 2.4G")
	case "wired":
		if usb {
			return "wired", nil
		}
		return "", fmt.Errorf("teclado USB-C no detectado; conecta el cable y pulsa FN+4")
	case "wireless":
		if dongle {
			return "wireless", nil
		}
		return "", fmt.Errorf("dongle 2.4G no detectado")
	default:
		return "", fmt.Errorf("transporte desconocido: %q", requested)
	}
}

func resolveTransport(requested string) (string, error) {
	return chooseTransport(requested, devicePresent("0c45", "800a"), devicePresent("05ac", "024f"))
}

func wirelessOnly(mode string) error {
	if mode != "wireless" {
		return fmt.Errorf("esta función solo está disponible por 2.4G; selecciona el dongle")
	}
	return nil
}
