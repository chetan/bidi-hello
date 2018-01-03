package helloworld

import (
	"fmt"
	"net"
	"sync"
)

// left over from the first attempt. not currently used w/ yamux impl

type fakeListener struct {
	conn     net.Conn
	returned bool
	cond     *sync.Cond
}

func NewFakeListener() *fakeListener {
	return &fakeListener{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (fl *fakeListener) SetConn(conn net.Conn) {
	fl.cond.L.Lock()
	fl.conn = conn
	fl.returned = false
	fl.cond.Broadcast()
	fl.cond.L.Unlock()
}

func (fl *fakeListener) Accept() (net.Conn, error) {
	if !fl.returned {
		fl.returned = true
		return fl.conn, nil
	}
	fmt.Println("done accepting, will block forever now..")
	fl.cond.L.Lock()
	for fl.returned == true {
		fl.cond.Wait()
	}
	defer fl.cond.L.Unlock()
	return fl.conn, nil
}

func (fl *fakeListener) Close() error {
	fmt.Println("Close called..")
	return nil
	// return fl.conn.Close()
}

func (fl *fakeListener) Addr() net.Addr {
	return fl.conn.RemoteAddr()
}
