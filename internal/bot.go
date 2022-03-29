//go:generate go run genbot.go

package internal

import (
	"archive/zip"
	"bytes"
)

func GetBotZIP(config []byte) []byte {
	// Create a new zip with config.json injected in.
	r, err := zip.NewReader(bytes.NewReader(botZipRaw[:]), int64(len(botZipRaw)))
	if err != nil {
		// This shouldn't fail.
		panic(err)
	}
	b := bytes.Buffer{}
	w := zip.NewWriter(&b)
	for _, i := range r.File {
		if err := w.Copy(i); err != nil {
			// This shouldn't fail.
			panic(err)
		}
	}
	f, err := w.Create("config.json")
	if err != nil {
		// This shouldn't fail.
		panic(err)
	}
	if _, err = f.Write(config); err != nil {
		// This shouldn't fail.
		panic(err)
	}
	if err := w.Close(); err != nil {
		// This shouldn't fail.
		panic(err)
	}
	return b.Bytes()
}
