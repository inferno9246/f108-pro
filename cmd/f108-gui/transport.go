package main

import "fmt"

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
	return chooseTransport(requested, usbPresent("0c45", "800a"), usbPresent("05ac", "024f"))
}
