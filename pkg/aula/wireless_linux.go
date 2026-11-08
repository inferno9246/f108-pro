//go:build linux

package aula

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type WirelessDevice struct{ f *os.File }

// OpenWireless busca la interfaz vendor (usage page 0xFF60) del dongle 05ac:024f.
func OpenWireless() (*WirelessDevice, error) {
	nodes, _ := filepath.Glob("/sys/class/hidraw/hidraw*")
	for _, n := range nodes {
		ue, err := os.ReadFile(filepath.Join(n, "device/uevent"))
		if err != nil || !strings.Contains(string(ue), "HID_ID=0003:000005AC:0000024F") {
			continue
		}
		desc, err := os.ReadFile(filepath.Join(n, "device/report_descriptor"))
		if err != nil || !bytes.HasPrefix(desc, []byte{0x06, 0x60, 0xFF}) {
			continue
		}
		f, err := os.OpenFile("/dev/"+filepath.Base(n), os.O_RDWR, 0)
		if err != nil {
			return nil, err
		}
		return &WirelessDevice{f: f}, nil
	}
	return nil, errors.New("dongle 2.4G no encontrado")
}

func (d *WirelessDevice) Close() error { return d.f.Close() }

// SetKeyRemap envía un paquete por tecla. A diferencia del USB, las teclas
// que no están en la lista NO se tocan.
func (d *WirelessDevice) SetKeyRemap(remaps []KeyRemap, fn bool) error {
	cmd := byte(0x03)
	if fn {
		cmd = 0x0D
	}
	for _, r := range remaps {
		p := []byte{cmd, 0x04, r.SourceIndex, byte(r.Action), r.Param1, r.Param2, r.Param3}
		if err := d.send(p); err != nil {
			return fmt.Errorf("tecla %d: %w", r.SourceIndex, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// defaultNormalSlot es el slot "sin remap" de la capa normal: la tecla a sí misma.
func defaultNormalSlot(hid byte) [4]byte {
	if hid >= 0xE0 && hid <= 0xE7 { // modificadores: bit en vez de usage
		return [4]byte{0x02, 1 << (hid - 0xE0), 0x00, 0x00}
	}
	return [4]byte{0x02, 0x00, hid, 0x00}
}

// ResetKeys restaura teclas a su valor por defecto. Con indices vacío restaura
// todas las teclas conocidas, salvo key_index 86, que el driver trata aparte.
// Capa normal: la tecla a sí misma. Capa FN: slot en blanco (sin confirmar).
func (d *WirelessDevice) ResetKeys(indices []uint8, fn bool) error {
	hidByIndex := map[uint8]byte{}
	for name, idx := range KeyNameToIndex {
		if hid, ok := KeyNameToHID[name]; ok {
			hidByIndex[idx] = hid
		}
	}
	if len(indices) == 0 {
		for idx := range hidByIndex {
			if idx != 86 {
				indices = append(indices, idx)
			}
		}
		sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })
	}
	cmd := byte(0x03)
	if fn {
		cmd = 0x0D
	}
	for _, idx := range indices {
		var slot [4]byte
		if !fn {
			hid, ok := hidByIndex[idx]
			if !ok {
				return fmt.Errorf("la tecla %d no tiene código HID conocido", idx)
			}
			slot = defaultNormalSlot(hid)
		}
		p := []byte{cmd, 0x04, idx, slot[0], slot[1], slot[2], slot[3]}
		if err := d.send(p); err != nil {
			return fmt.Errorf("tecla %d: %w", idx, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// ErrNoACK indica que el paquete se escribió pero el teclado no lo confirmó a tiempo.
// El comando puede haberse aplicado igualmente: un teclado 2.4G en reposo responde tarde.
var ErrNoACK = errors.New("sin ACK del teclado")

func (d *WirelessDevice) send(payload []byte) error {
	var pkt [33]byte // [0]=report ID, [1..31]=payload, [32]=checksum
	copy(pkt[1:32], payload)
	var sum byte
	for _, b := range pkt[:32] {
		sum += b
	}
	pkt[32] = sum

	buf := make([]byte, 64)
	for { // vaciar reportes pendientes (una sola vez, antes del primer envío)
		d.f.SetReadDeadline(time.Now().Add(time.Millisecond))
		if _, err := d.f.Read(buf); err != nil {
			break
		}
	}
	for attempt := 0; attempt < 3; attempt++ {
		var err error
		for try := 0; try < 2; try++ { // el driver original reintenta una vez
			if _, err = d.f.Write(pkt[:]); err == nil {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err != nil {
			return err
		}
		// Un ACK tardío del envío anterior también vale: lleva la misma cabecera.
		for end := time.Now().Add(700 * time.Millisecond); time.Now().Before(end); {
			d.f.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
			if n, err := d.f.Read(buf); err == nil && n >= 3 && bytes.Equal(buf[:3], pkt[1:4]) {
				return nil
			}
		}
	}
	return ErrNoACK
}

// SetLighting envía el paquete de iluminación global (05 10) por el dongle 2.4G.
// Offsets del payload (el offset 0 del paquete es el report ID):
// 3 modo, 4-6 RGB, 11 multicolor, 12 brillo, 13 velocidad, 14 dirección, 17-18 AA 55.
func (d *WirelessDevice) SetLighting(cfg LightingConfig) error {
	if cfg.Mode < 0 || cfg.Mode > 19 {
		return fmt.Errorf("invalid mode %d (valid range: 0-19)", cfg.Mode)
	}
	if cfg.Brightness > 5 || cfg.Speed > 5 || cfg.Direction > 1 {
		return fmt.Errorf("brightness and speed must be 0-5, direction 0-1")
	}
	colorful := byte(0)
	if cfg.Colorful {
		colorful = 1
	}
	return d.send([]byte{
		0x05, 0x10, 0x00, byte(cfg.Mode), cfg.R, cfg.G, cfg.B,
		0, 0, 0, 0, colorful, cfg.Brightness, cfg.Speed, cfg.Direction,
		0, 0, 0xAA, 0x55,
	})
}

// SyncClock sincroniza el reloj de la pantalla por el dongle 2.4G.
// Formato reconstruido de FUN_00423c40 (offsets del payload; el offset 0 del
// paquete es el report ID): 0 = 0x0C, 1 = 0x10, 4 = perfil, 5 = 0x5A, 6 = año-2000,
// 7 = mes, 8 = día, 9 = hora, 10 = minuto, 11 = segundo, 13 = día de la semana
// (0 = domingo), 17-18 = AA 55.
func (d *WirelessDevice) SyncClock(t time.Time) error {
	return d.send([]byte{
		0x0C, 0x10, 0x00, 0x00, 0x01, 0x5A,
		byte(t.Year() % 2000), byte(t.Month()), byte(t.Day()),
		byte(t.Hour()), byte(t.Minute()), byte(t.Second()),
		0x00, byte(t.Weekday()), 0, 0, 0, 0xAA, 0x55,
	})
}

// SendMusicFrame envia un fotograma del modo musica (0B 1C) por el dongle 2.4G.
// Sin reintentos y con una espera corta de ACK: se manda ~30 veces por segundo y un
// fotograma perdido no importa. Formato (payload; el offset 0 del paquete es el report ID):
// 0-2 = 0B 1C 00, 3 = efecto, 4 = fondo, 5 = brillo del fondo (0-100), 6-26 = 21 niveles.
func (d *WirelessDevice) SendMusicFrame(fore, back, bright byte, levels [21]byte) error {
	var pkt [33]byte
	pkt[1], pkt[2] = 0x0B, 0x1C
	pkt[4], pkt[5], pkt[6] = fore, back, bright
	copy(pkt[7:28], levels[:])
	var sum byte
	for _, b := range pkt[:32] {
		sum += b
	}
	pkt[32] = sum

	buf := make([]byte, 64)
	for { // vaciar reportes pendientes
		d.f.SetReadDeadline(time.Now().Add(time.Millisecond))
		if _, err := d.f.Read(buf); err != nil {
			break
		}
	}
	if _, err := d.f.Write(pkt[:]); err != nil {
		return err
	}
	for end := time.Now().Add(60 * time.Millisecond); time.Now().Before(end); {
		d.f.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		if n, err := d.f.Read(buf); err == nil && n >= 3 && bytes.Equal(buf[:3], pkt[1:4]) {
			return nil
		}
	}
	return ErrNoACK
}

// wirelessPacket arma el paquete de 33 bytes: [0] report ID, [1..31] payload, [32] checksum.
func wirelessPacket(payload []byte) [33]byte {
	var pkt [33]byte
	copy(pkt[1:32], payload)
	var sum byte
	for _, b := range pkt[:32] {
		sum += b
	}
	pkt[32] = sum
	return pkt
}

// request envia un paquete y devuelve la respuesta cuya cabecera (3 bytes) coincide.
func (d *WirelessDevice) request(payload []byte) ([]byte, error) {
	pkt := wirelessPacket(payload)
	buf := make([]byte, 64)
	for { // vaciar reportes pendientes
		d.f.SetReadDeadline(time.Now().Add(time.Millisecond))
		if _, err := d.f.Read(buf); err != nil {
			break
		}
	}
	if _, err := d.f.Write(pkt[:]); err != nil {
		return nil, err
	}
	for end := time.Now().Add(400 * time.Millisecond); time.Now().Before(end); {
		d.f.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
		if n, err := d.f.Read(buf); err == nil && n >= 4 && bytes.Equal(buf[:3], pkt[1:4]) {
			return append([]byte(nil), buf[:n]...), nil
		}
	}
	return nil, ErrNoACK
}

// Battery pregunta el nivel de bateria (comando 20 01; el byte 3 de la respuesta es el
// porcentaje, igual que en las capturas del software oficial).
func (d *WirelessDevice) Battery() (int, error) {
	resp, err := d.request([]byte{0x20, 0x01})
	if err != nil {
		return 0, err
	}
	if resp[3] > 100 {
		return 0, fmt.Errorf("respuesta de bateria inesperada: %d", resp[3])
	}
	return int(resp[3]), nil
}

// perKeyPackets arma la secuencia del software oficial (reconstruida de una captura):
// 05 10 con modo 0x80 (luz de usuario) y brillo 0-5, luego 20 trozos 14 1C de 28 bytes
// (560 bytes = ranuras de 4 bytes [indice, R, G, B] en el offset indice*4) y 14 10 14.
func perKeyPackets(keys []KeyColor, brightness uint8) [][]byte {
	mode := make([]byte, 19)
	copy(mode, []byte{0x05, 0x10, 0x00, 0x80})
	mode[12] = brightness
	mode[17], mode[18] = 0xAA, 0x55

	var buf [560]byte
	for _, k := range keys {
		i := int(k.LightIndex) * 4
		buf[i], buf[i+1], buf[i+2], buf[i+3] = k.LightIndex, k.R, k.G, k.B
	}
	out := [][]byte{mode}
	for c := 0; c < 20; c++ {
		p := append([]byte{0x14, 0x1C, byte(c)}, buf[c*28:(c+1)*28]...)
		out = append(out, p)
	}
	fin := make([]byte, 19)
	copy(fin, []byte{0x14, 0x10, 0x14})
	fin[17], fin[18] = 0xAA, 0x55
	return append(out, fin)
}

// SetPerKeyRGB envia el color de cada tecla por el dongle 2.4G (luz de usuario).
// Las teclas que no se indican quedan apagadas.
func (d *WirelessDevice) SetPerKeyRGB(keys []KeyColor, brightness uint8) error {
	if brightness > 5 {
		return fmt.Errorf("invalid brightness %d (valid range: 0-5)", brightness)
	}
	for _, k := range keys {
		if int(k.LightIndex)*4+3 >= 560 {
			return fmt.Errorf("light index %d out of range", k.LightIndex)
		}
	}
	for n, p := range perKeyPackets(keys, brightness) {
		if err := d.send(p); err != nil {
			return fmt.Errorf("paquete %d: %w (si el teclado estaba en reposo, pulsa una tecla y reintenta)", n+1, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// SendMacros sustituye la tabla completa. Valida todos los eventos antes del primer envío.
func (d *WirelessDevice) SendMacros(macros [][]MacroEvent) error {
	table, err := BuildMacroTable(macros)
	if err != nil {
		return err
	}
	packets, err := MacroPackets(table)
	if err != nil {
		return err
	}
	for i, p := range packets {
		if err := d.send(p); err != nil {
			return fmt.Errorf("trozo de macros %d/%d: %w", i+1, len(packets), err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}
