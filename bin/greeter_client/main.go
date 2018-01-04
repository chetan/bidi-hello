/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/hashicorp/yamux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/chetan/bidi-hello/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// args
	name := defaultName
	if len(os.Args) > 1 {
		if os.Args[1] == "--plain" {
			doPlainClient(address, name)
			return
		}
		name = os.Args[1]
	}

	// create reusable grpc server
	grpcServer := grpc.NewServer()
	helloworld.RegisterGreeterServer(grpcServer, helloworld.NewServerImpl())
	reflection.Register(grpcServer)

	for {
		// connect and loop forever to keep retrying in the case that the connection
		// drops for any reason
		connect(address, name, grpcServer)
		time.Sleep(1 * time.Second)
	}
}

// doPlainClient creates a client-only connection to the server w/o the use of
// yamux. This is to show that a non-wrapped client can still talk to the server.
func doPlainClient(addr string, name string) {
	gconn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer gconn.Close()

	grpcClient := helloworld.NewGreeterClient(gconn)
	for {
		err := helloworld.Greet(grpcClient, "server", name)
		if err != nil {
			log.Printf("greet err: %s", err)
			break
		}
		time.Sleep(helloworld.Timeout)
	}
}

// connect to server
//
// dials out to the target server and then sets up a yamux channel for
// multiplexing both grpc client and server on the same underlying tcp socket
//
// this is separate from the run-loop for easy resource cleanup via defer
func connect(addr string, name string, grpcServer *grpc.Server) {
	// dial underlying tcp connection
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", addr)
	if err != nil {
		log.Printf("Failed to connect %s", err)
		return
	}
	defer conn.Close()

	session, err := yamux.Client(conn, nil)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// now that we have a connection, create both clients & servers

	// setup client
	go func() {
		dialerOpts := grpc.WithDialer(func(addr string, d time.Duration) (net.Conn, error) {
			// custom dialer which opens a new conn via yamux session
			return session.Open()
		})
		gconn, err := grpc.Dial(address, grpc.WithInsecure(), dialerOpts)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer gconn.Close()

		grpcClient := helloworld.NewGreeterClient(gconn)
		for {
			err := helloworld.Greet(grpcClient, "server", name)
			if err != nil {
				log.Printf("greet err: %s", err)
				break
			}
			time.Sleep(helloworld.Timeout)
		}
	}()

	// setup server
	go func() {
		// start grpc server using yamux session which implements the net.Listener interface
		if err := grpcServer.Serve(session); err != nil {
			// err will be returned when the server exits (underlying connection closes / client goes away)
			log.Printf("grpc serve err: %v", err)
		}
	}()

	// reading from a channel blocks until some data is avail
	// no data will ever be sent on this channel, it will simple close
	// when the session (connection) closes or any reason, thus unblocking
	// and continuing program execution.
	//
	// alternatively we could have used a sync.WaitGroup to look for our two
	// goroutines to exit, but this is simpler as it watches the underlying
	// resource.
	<-session.CloseChan()
}
