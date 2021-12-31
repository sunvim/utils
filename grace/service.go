package grace

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type serv struct {
	stop   chan os.Signal
	cancel func()
	funcs  []func() error
}

type Service interface {
	Register(fn func() error)
	Wait()
}

func (s *serv) Wait() {
	defer signal.Stop(s.stop)
	select {
	case <-s.stop:
		s.cancel()
		for _, fn := range s.funcs {
			if err := fn(); err != nil {
				log.Printf("err: %v \n", err)
			}
		}
		time.Sleep(100 * time.Millisecond)
		log.Println("all services exited totally.")
	}
}

func (s *serv) Register(fn func() error) {
	s.funcs = append(s.funcs, fn)
}

func New(ctx context.Context) (context.Context, Service) {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctxLocal, cancel := context.WithCancel(ctx)
	return ctxLocal, &serv{
		stop:   stopChan,
		cancel: cancel,
	}
}
