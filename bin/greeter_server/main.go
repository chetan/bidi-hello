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

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/chetan/bidi-hello/helloworld"
)

const (
	port = ":50051"
)

func main() {

	grpcServer := grpc.NewServer()
	helloworld.RegisterGreeterServer(grpcServer, helloworld.NewServerImpl())
	reflection.Register(grpcServer)

	// create underlying tcp listener
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	mux := cmux.New(lis)
	grpcL := mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	yamuxL := mux.Match(cmux.Any())

	go grpcServer.Serve(grpcL)
	go runLoop(yamuxL, grpcServer)

	mux.Serve()
}

func runLoop(lis net.Listener, grpcServer *grpc.Server) {
	for {
		// accept a new connection and set up a yamux session on it
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		// server session will be used to multiplex both clients & servers
		session, err := yamux.Server(conn, nil)
		if err != nil {
			panic(err)
		}

		// start grpc server using yamux session (which implements net.Listener)
		go grpcServer.Serve(session)

		// start client conn back to agent
		go func() {
			dialerOpts := grpc.WithDialer(func(addr string, d time.Duration) (net.Conn, error) {
				// custom dialer which opens a new conn via yamux session
				return session.Open()
			})

			gconn, err := grpc.Dial("localhost:50000", grpc.WithInsecure(), dialerOpts)
			if err != nil {
				fmt.Println("failed to create grpc client: ", err)
				return
			}
			defer gconn.Close()

			c := helloworld.NewGreeterClient(gconn)

			for {
				err := helloworld.Greet(c, "client", "bob")
				if err != nil {
					return // stop greeting on this channel
				}
				time.Sleep(helloworld.Timeout)
			}

		}()

	}

}
