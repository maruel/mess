package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func configureLog() {
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		return short + ":" + strconv.Itoa(line)
	}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	//log.Logger = log.Logger.With().Caller().Logger()
}

var started = time.Now().UTC()

type logOut struct {
	id  int
	buf bytes.Buffer
}

func (l *logOut) Write(b []byte) (int, error) {
	l.buf.Write(b)
	for bytes.IndexByte(l.buf.Bytes(), '\n') != -1 {
		line, _ := l.buf.ReadString('\n')
		log.Info().Int("bot", l.id).Str("l", line[:len(line)-1]).Msg("")
	}
	return len(b), nil
}

func mainImpl() error {
	configureLog()
	s := flag.String("S", "http://localhost:7899", "mess server to connect to")
	num := flag.Int("num", 1, "number of bots to start")
	root := flag.String("R", "loadtest", "root directory to start bots in")
	keep := flag.Bool("k", true, "keep previous files")

	flag.Parse()

	if *num < 1 || *num >= 50000 {
		return errors.New("specify a reasonable argument for -num")
	}

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	wg.Add(1)
	go func() {
		<-c
		cancel()
		wg.Done()
	}()

	if stat, err := os.Stat(*root); os.IsNotExist(err) {
		if err := os.Mkdir(*root, 0o700); err != nil {
			return err
		}
	} else if !stat.IsDir() || !*keep {
		if err := os.Remove(*root); err != nil {
			return err
		}
		log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Str("root", *root).Msg("Deleted")
		if err := os.Mkdir(*root, 0o700); err != nil {
			return err
		}
	}

	for i := 0; i < *num; i++ {
		d := filepath.Join(*root, fmt.Sprintf("bot%d", i+1))
		if err := os.Mkdir(d, 0o700); err != nil && !(os.IsExist(err) && *keep) {
			return err
		}
		// TODO(maruel): Copy the bot code instead of downloading N times but I'm lazy.
		resp, err := http.Get(*s + "/bot_code")
		if err != nil {
			return err
		}
		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("got HTTP %d", resp.StatusCode)
		}
		if err = os.WriteFile(filepath.Join(d, "swarming_bot.zip"), b, 0o600); err != nil {
			return err
		}
	}
	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).
		Str("s", *s).Msg("Downloaded")

	host, err := os.Hostname()
	if err != nil {
		return err
	}

	procs := make([]*exec.Cmd, 0, *num)
	for i := 0; i < *num; i++ {
		c := exec.Command("python3", "swarming_bot.zip", "start_bot")
		lo := &logOut{id: i + 1}
		c.Stderr = lo
		c.Stdout = lo
		c.Dir = filepath.Join(*root, fmt.Sprintf("bot%d", i+1))
		c.Env = append(os.Environ(), fmt.Sprintf("SWARMING_BOT_ID=%s--%d", host, i+1))
		if err := c.Start(); err != nil {
			return err
		}
		procs = append(procs, c)
	}

	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Msg("started")

	<-ctx.Done()
	stopping := time.Now()
	log.Info().Dur("ms", time.Since(started).Round(time.Millisecond)/10).Msg("Terminating")
	wg.Wait()
	for _, p := range procs {
		// os.Interrupt on Windows
		p.Process.Signal(syscall.SIGTERM)
	}
	// TODO(maruel): 30s grace period, then SIGKILL.
	for _, p := range procs {
		p.Wait()
	}
	log.Info().Dur("ms", time.Since(stopping).Round(time.Millisecond)/10).Msg("Terminated")
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "loadtestmess: %v\n", err)
		os.Exit(1)
	}
}
