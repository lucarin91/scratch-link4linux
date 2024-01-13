package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"tinygo.org/x/bluetooth"

	"github.com/lucarin91/scratch-link4linux/scratchlink"
)

func setLogger(debug bool) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	setLogger(*debug)

	var adapter = bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		log.Fatal().Err(err).Msgf("BLE cannot be enabled")
	}

	http.Handle("/scratch/ble", scratchlink.GetHandler(adapter))

	server := &http.Server{
		Addr:              ":20111",
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Info().Msgf("Starting scratch server on %q", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
