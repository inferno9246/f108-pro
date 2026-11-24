//go:build linux

package aula

import (
	"bytes"
	"io"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestWirelessClockMatchesOfficialCapture(t *testing.T) {
	want, err := os.ReadFile("testdata/clock-one.bin")
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != 33 {
		t.Fatalf("longitud golden: %d", len(want))
	}
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
		ack := make([]byte, 32)
		copy(ack, pkt[1:4])
		_, _ = b.Write(ack)
	}()
	stamp := time.Date(2026, time.October, 6, 20, 31, 31, 0, time.UTC)
	if err := (&WirelessDevice{f: a}).SyncClock(stamp); err != nil {
		t.Fatal(err)
	}
	if pkt := <-got; !bytes.Equal(pkt[:], want) {
		t.Fatalf("paquete diferente\n got % x\nwant % x", pkt, want)
	}
}
