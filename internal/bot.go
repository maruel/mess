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
	"sort"
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

	cfg, err := json.Marshal(config{Server: "https://" + host})
	if err != nil {
		panic(err)
	}

	// Create a new zip with config/config.json injected in.
	h := sha256.New()
	r, err := zip.NewReader(bytes.NewReader(botZipRaw[:]), int64(len(botZipRaw)))
	if err != nil {
		panic(err)
	}
	buf := bytes.Buffer{}
	w := zip.NewWriter(&buf)
	names := make([]string, 1, len(r.File)+1)
	names[0] = "config/config.json"
	f, err := w.Create("config/config.json")
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(cfg); err != nil {
		panic(err)
	}
	for _, f := range r.File {
		names = append(names, f.Name)
		if err := w.Copy(f); err != nil {
			panic(err)
		}
	}
	// TODO(maruel): Inject config/bot_config.py.
	if err := w.Close(); err != nil {
		panic(err)
	}

	sort.Strings(names)
	for _, n := range names {
		if n == "config/config.json" {
			hashFile(h, "config/config.json", cfg)
		} else {
			c, err := r.Open(n)
			if err != nil {
				panic(err)
			}
			raw, err := ioutil.ReadAll(c)
			if err != nil {
				panic(err)
			}
			c.Close()
			hashFile(h, n, raw)
		}
	}

	// Zip content and the content's hash.
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
