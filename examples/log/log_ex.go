package main

import "github.com/sunvim/utils/log"

func main() {
	log.SetPrefix("Hello")

	log.Info("info hello log")

	log.Error("error log")

}
