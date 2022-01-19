package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	"github.com/oxtoacart/bpool"
	"github.com/rcrowley/go-metrics"
	"github.com/sunvim/utils/netpoll"
)

func main() {
	setLimit()
	go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	var handler = &netpoll.DataHandler{
		Pool: bpool.NewBytePool(1024, 12*1024),
		HandlerFunc: func(req []byte) (res []byte) {
			res = req
			return
		},
	}
	if err := netpoll.ListenAndServe("tcp", ":9999", handler); err != nil {
		panic(err)
	}
}

func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	log.Printf("set cur limit: %d", rLimit.Cur)
}
