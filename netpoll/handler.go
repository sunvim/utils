// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package netpoll

import (
	"errors"
	"net"
	"sync"
)

// ErrHandlerFunc is the error when the HandlerFunc is nil
var ErrHandlerFunc = errors.New("HandlerFunc must be not nil")

// ErrUpgradeFunc is the error when the Upgrade func is nil
var ErrUpgradeFunc = errors.New("Upgrade function must be not nil")

// ErrServeFunc is the error when the Serve func is nil
var ErrServeFunc = errors.New("Serve function must be not nil")

// Context is returned by Upgrade for serving.
type Context interface{}

// Handler responds to a single request.
type Handler interface {
	// Upgrade upgrades the net.Conn to a Context.
	Upgrade(net.Conn) (Context, error)
	// Serve should serve a single request with the Context.
	Serve(Context) error
}

// NewHandler returns a new Handler.
func NewHandler(upgrade func(net.Conn) (Context, error), serve func(Context) error) Handler {
	return &ConnHandler{upgrade: upgrade, serve: serve}
}

// ConnHandler implements the Handler interface.
type ConnHandler struct {
	upgrade func(net.Conn) (Context, error)
	serve   func(Context) error
}

// SetUpgrade sets the Upgrade function for upgrading the net.Conn.
func (h *ConnHandler) SetUpgrade(upgrade func(net.Conn) (Context, error)) *ConnHandler {
	h.upgrade = upgrade
	return h
}

// SetServe sets the Serve function for once serving.
func (h *ConnHandler) SetServe(serve func(Context) error) *ConnHandler {
	h.serve = serve
	return h
}

// Upgrade implements the Handler Upgrade method.
func (h *ConnHandler) Upgrade(conn net.Conn) (Context, error) {
	if h.upgrade == nil {
		return nil, ErrUpgradeFunc
	}
	return h.upgrade(conn)
}

// Serve implements the Handler Serve method.
func (h *ConnHandler) Serve(ctx Context) error {
	if h.serve == nil {
		return ErrServeFunc
	}
	return h.serve(ctx)
}

type BytePool interface {
	Get() []byte
	Put(b []byte)
	NumPooled() int
	Width() int
}

// DataHandler implements the Handler interface.
type DataHandler struct {
	Pool    BytePool
	upgrade func(net.Conn) (net.Conn, error)
	// HandlerFunc is the data Serve function.
	HandlerFunc func(req []byte) (res []byte)
}

type context struct {
	reading sync.Mutex
	writing sync.Mutex
	upgrade bool
	conn    net.Conn
	pool    BytePool
	buffer  []byte
}

// SetUpgrade sets the Upgrade function for upgrading the net.Conn.
func (h *DataHandler) SetUpgrade(upgrade func(net.Conn) (net.Conn, error)) {
	h.upgrade = upgrade
}

// Upgrade sets the net.Conn to a Context.
func (h *DataHandler) Upgrade(conn net.Conn) (Context, error) {
	if h.HandlerFunc == nil {
		return nil, ErrHandlerFunc
	}
	var upgrade bool
	if h.upgrade != nil {
		c, err := h.upgrade(conn)
		if err != nil {
			return nil, err
		} else if c != nil && c != conn {
			upgrade = true
			conn = c
		}
	}
	var ctx = &context{upgrade: upgrade, conn: conn, pool: h.Pool}
	return ctx, nil
}

// Serve should serve a single request with the Context ctx.
func (h *DataHandler) Serve(ctx Context) error {
	c := ctx.(*context)
	var conn = c.conn
	var n int
	var err error
	var buf []byte
	var req []byte
	buf = c.pool.Get()
	defer c.pool.Put(buf)
	if c.upgrade {
		c.reading.Lock()
	}
	n, err = conn.Read(buf)
	if c.upgrade {
		c.reading.Unlock()
	}
	if err != nil {
		return err
	}
	req = buf[:n]
	res := h.HandlerFunc(req)
	if c.upgrade {
		c.writing.Lock()
	}
	_, err = conn.Write(res)
	if c.upgrade {
		c.writing.Unlock()
	}
	c.pool.Put(res)
	return err
}
