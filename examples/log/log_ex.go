package main

import (
	"github.com/sunvim/utils/log"
)

func main() {
	logger := log.NewLogger()

	logger.Info().Str("hello", "world").Msg("main")

}
