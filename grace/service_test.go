package grace

import (
	"context"
	"log"
	"syscall"
	"testing"
	"time"
)

func TestWait(t *testing.T) {
	_, service := New(context.Background())
	service.Register(func() error {
		log.Println("exit 1")
		return nil
	})
	service.Register(func() error {
		log.Println("exit 2")
		return nil
	})
	go func() {
		time.Sleep(10 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()
	service.Wait()
}
