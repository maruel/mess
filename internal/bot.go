//go:generate go run genbot.go

package internal

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"sync"
)

// GetBotZIP return the swarming_bot.zip's hashed content.
func GetBotVersion(host string) string {
	mu.Lock()
	v := botVersion[host]
	mu.Unlock()
	if v != "" {
		return v
	}
	GetBotZIP(host)
	mu.Lock()
	v = botVersion[host]
	mu.Unlock()
	return v
}

// GetBotZIP return the swarming_bot.zip bytes.
func GetBotZIP(host string) []byte {
	log.Printf("GetBotZIP(%s)", host)
	mu.Lock()
	b := botCode[host]
	mu.Unlock()
	if b != nil {
		return b
	}

	h := sha256.New()
	// Create a new zip with config.json injected in.
	r, err := zip.NewReader(bytes.NewReader(botZipRaw[:]), int64(len(botZipRaw)))
	if err != nil {
		panic(err)
	}
	buf := bytes.Buffer{}
	w := zip.NewWriter(&buf)
	for _, f := range r.File {
		c, err := f.Open()
		if err != nil {
			panic(err)
		}
		raw, err := ioutil.ReadAll(c)
		if err != nil {
			panic(err)
		}
		c.Close()
		hashFile(h, f.Name, raw)
		if err := w.Copy(f); err != nil {
			panic(err)
		}
	}

	f, err := w.Create("config/config.json")
	if err != nil {
		panic(err)
	}
	raw, err := json.Marshal(config{Server: "https://" + host})
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(raw); err != nil {
		panic(err)
	}
	hashFile(h, "config/config.json", raw)
	if err := w.Close(); err != nil {
		panic(err)
	}

	// TODO(maruel): bot_config.py

	b = buf.Bytes()
	v := hex.EncodeToString(h.Sum(nil))

	mu.Lock()
	if b2 := botCode[host]; b2 != nil {
		// Discard our version.
		b = b2
	} else {
		botCode[host] = b
		botVersion[host] = v
	}
	mu.Unlock()
	return b
}

func hashFile(h io.Writer, name string, raw []byte) {
	_, _ = h.Write([]byte(strconv.Itoa(len(name))))
	_, _ = h.Write([]byte(name))
	_, _ = h.Write([]byte(strconv.Itoa(len(raw))))
	_, _ = h.Write(raw)
}

type config struct {
	Server             string `json:"server"`
	ServerVersion      string `json:"server_version"`
	EnableTSMonitoring bool   `json:"enable_ts_monitoring"`
}

var (
	mu         sync.Mutex
	botCode    = map[string][]byte{}
	botVersion = map[string]string{}
)
