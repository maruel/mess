//go:generate go run genbot.go

package internal

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// GetBotVersion return the swarming_bot.zip's hashed content.
func GetBotVersion(ctx context.Context, url string) string {
	mu.Lock()
	v := botVersion[url]
	mu.Unlock()
	// Was already cached, quick return.
	if v != "" {
		return v
	}
	GetBotZIP(ctx, url)
	mu.Lock()
	v = botVersion[url]
	mu.Unlock()
	return v
}

// GetBotZIP return the swarming_bot.zip bytes.
func GetBotZIP(ctx context.Context, url string) []byte {
	mu.Lock()
	b := botCode[url]
	mu.Unlock()
	// Was already cached, quick return.
	if b != nil {
		return b
	}

	s := time.Now()
	cfg, err := json.Marshal(config{Server: url})
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

	race := false
	mu.Lock()
	if b2 := botCode[url]; b2 != nil {
		// Discard our version.
		race = true
		b = b2
	} else {
		botCode[url] = b
		botVersion[url] = v
	}
	mu.Unlock()

	log.Ctx(ctx).Info().Str("url", url).Str("hash", v).
		Int("size", len(b)).Bool("race", race).
		Dur("ms", time.Since(s).Round(time.Millisecond/10)).
		Msg("GetBotZIP")
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
