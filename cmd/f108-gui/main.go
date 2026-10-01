package main

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sort"

	"github.com/parsiya/f108-pro/pkg/aula"
)

//go:embed web
var webFS embed.FS

type keyInfo struct {
	Name  string `json:"name"`
	Index uint8  `json:"index"`
	HID   byte   `json:"hid"`
}

func main() {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		log.Fatal(err)
	}
	token := hex.EncodeToString(raw)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	addr := ln.Addr().String()

	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/api/keys", func(w http.ResponseWriter, r *http.Request) {
		var keys []keyInfo
		for name, idx := range aula.KeyNameToIndex {
			keys = append(keys, keyInfo{Name: name, Index: idx, HID: aula.KeyNameToHID[name]})
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].Index < keys[j].Index })
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	})

	// Protección contra DNS rebinding y CSRF: host exacto y token en /api.
	guarded := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != addr {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
			if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Token")), []byte(token)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})

	fmt.Printf("Abre: http://%s/?t=%s\n", addr, token)
	log.Fatal(http.Serve(ln, guarded))
}
