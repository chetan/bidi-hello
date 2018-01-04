package helloworld

import (
	"context"
	"log"
	"time"
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
