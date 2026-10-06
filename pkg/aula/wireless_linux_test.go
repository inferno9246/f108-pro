//go:build linux

package aula

import (
	"bytes"
	"io"
	"os"
	"syscall"
	"testing"
)

// El paquete debe coincidir byte a byte con uno capturado del software oficial:
// 0b 1c 00 02 00 64 | 00 00 01 00... | checksum 8e.
func TestSendMusicFrameMatchesCapture(t *testing.T) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK, 0)
	if err != nil {
		t.Skip("socketpair no disponible:", err)
	}
	a := os.NewFile(uintptr(fds[0]), "a")
	b := os.NewFile(uintptr(fds[1]), "b")
	defer a.Close()
	defer b.Close()

	got := make(chan [33]byte, 1)
	go func() {
		var pkt [33]byte
		if _, err := io.ReadFull(b, pkt[:]); err != nil {
			return
		}
		got <- pkt
		echo := make([]byte, 32)
		copy(echo, pkt[1:4])
		b.Write(echo)
	}()

	d := &WirelessDevice{f: a}
	var levels [21]byte
	levels[2] = 1
	if err := d.SendMusicFrame(2, 0, 100, levels); err != nil {
		t.Fatalf("SendMusicFrame: %v", err)
	}
	pkt := <-got

	want := []byte{0x00, 0x0b, 0x1c, 0x00, 0x02, 0x00, 0x64, 0x00, 0x00, 0x01}
	want = append(want, make([]byte, 33-len(want))...)
	want[32] = 0x8e
	if !bytes.Equal(pkt[:], want) {
		t.Fatalf("paquete distinto\n got: % x\nwant: % x", pkt, want)
	}
}

func TestSendMusicFrameNoACK(t *testing.T) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK, 0)
	if err != nil {
		t.Skip("socketpair no disponible:", err)
	}
	a := os.NewFile(uintptr(fds[0]), "a")
	b := os.NewFile(uintptr(fds[1]), "b")
	defer a.Close()
	defer b.Close()
	d := &WirelessDevice{f: a}
	if err := d.SendMusicFrame(2, 0, 100, [21]byte{}); err != ErrNoACK {
		t.Fatalf("esperaba ErrNoACK, obtuve %v", err)
	}
}
