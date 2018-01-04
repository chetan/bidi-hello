package helloworld

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
)

// Timeout between greetings
const Timeout = 1 * time.Second

// Greet using the given client
func Greet(c GreeterClient, dest string, name string) error {
	r, err := c.SayHello(context.Background(), &HelloRequest{Name: name})
	if err != nil {
		log.Printf("could not greet: %v", err)
		return err
	}
	log.Printf("Greeting from %s: %s", dest, r.Message)

	r, err = c.SayHelloAgain(context.Background(), &HelloRequest{Name: name})
	if err != nil {
		log.Printf("could not greet: %v", err)
		return err
	}
	log.Printf("Greeting from %s: %s", dest, r.Message)
	return nil
}

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
