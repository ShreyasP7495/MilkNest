package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init configures the global zerolog logger. In development we use a pretty
// console writer; in production we emit JSON to stdout for the platform log
// agent to scrape.
func Init(env string) {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if env == "production" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		return
	}

	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()
}

// L returns the global logger so call-sites do not have to import zerolog/log
// directly.
func L() *zerolog.Logger {
	return &log.Logger
}
