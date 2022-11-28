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
	ctx    context.Context
	stop   chan os.Signal
	cancel func()
	funcs  []func() error
	srvs   map[string]func(context.Context) error
}

type Service interface {
	Register(fn func() error)
	RegisterService(name string, fn func(context.Context) error)
	Wait()
}

func (s *serv) Wait() {
	defer signal.Stop(s.stop)
	for name, fn := range s.srvs {
		log.Printf("boot service %s ...", name)
		safeGo(s.ctx, fn)
	}
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
		os.Exit(0)
	}
}

func (s *serv) Register(fn func() error) {
	s.funcs = append(s.funcs, fn)
}

func (s *serv) RegisterService(name string, fn func(context.Context) error) {
	s.srvs[name] = fn
}

func New(ctx context.Context) (context.Context, Service) {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctxLocal, cancel := context.WithCancel(ctx)
	return ctxLocal, &serv{
		ctx:    ctxLocal,
		stop:   stopChan,
		cancel: cancel,
		srvs:   make(map[string]func(context.Context) error),
	}
}

func safeGo(ctx context.Context, fn func(ctx context.Context) error) {
	go func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				println(err)
			}
		}()
		fn(ctx)
	}(ctx)
}
