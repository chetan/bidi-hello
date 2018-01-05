package helloworld

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
)

type YamuxDialer struct {
	session *yamux.Session
	mu      *sync.Mutex
}

func NewYamuxDialer() *YamuxDialer {
	return &YamuxDialer{mu: &sync.Mutex{}}
}

// SetSession updates the session used by the dialer
func (y *YamuxDialer) SetSession(sess *yamux.Session) {
	y.mu.Lock()
	defer y.mu.Unlock()
	y.session = sess
}

// Dial a connection using a yamux session
func (y *YamuxDialer) Dial(addr string, d time.Duration) (net.Conn, error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if y.session == nil || y.session.IsClosed() {
		return nil, errors.New("not connected")
	}
	return y.session.Open()
}
