package aula

import "fmt"

// settingsPayload representa el comando 07 10 del software oficial.
// flags son las tres casillas cuyo significado aún no está confirmado.
func settingsPayload(sleep, response byte, flags [3]byte) ([]byte, error) {
	if sleep > 3 {
		return nil, fmt.Errorf("el reposo debe estar entre 0 y 3")
	}
	if response < 1 || response > 5 {
		return nil, fmt.Errorf("el tiempo de respuesta debe estar entre 1 y 5")
	}
	for _, flag := range flags {
		if flag > 1 {
			return nil, fmt.Errorf("las casillas deben ser 0 o 1")
		}
	}
	return []byte{
		0x07, 0x10, 0, 0, 1, flags[0], flags[1], flags[2],
		0, sleep, 0, response, 0, 0, 0, 0, 0, 0xAA, 0x55,
	}, nil
}

func settingsPacket(sleep, response byte, flags [3]byte) ([33]byte, error) {
	var pkt [33]byte
	payload, err := settingsPayload(sleep, response, flags)
	if err != nil {
		return pkt, err
	}
	copy(pkt[1:], payload)
	for _, b := range pkt[:32] {
		pkt[32] += b
	}
	return pkt, nil
}
