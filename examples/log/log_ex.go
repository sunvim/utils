package main

import (
	"fmt"

	"github.com/sunvim/utils/log"
)

func main() {
	log.SetPrefix("Example")

	fmt.Println("debug level")
	log.SetLevel(log.LevelDebug)
	log.Debug("hello debug")
	log.Info("hello info")
	log.Error("hello error")

	fmt.Println("info level")
	log.SetLevel(log.LevelInfo)
	log.Debug("hello debug")
	log.Info("hello info")
	log.Error("hello error")

	fmt.Println("error level")
	log.SetLevel(log.LevelError)
	log.Debug("hello debug")
	log.Info("hello info")
	log.Error("hello error")

}
