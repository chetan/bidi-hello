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
	"os"
	"time"

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

	// open channel and create client
	gconn := helloworld.Connect(address, grpcServer)
	defer gconn.Close()
	grpcClient := helloworld.NewGreeterClient(gconn)

	// run client loop
	for {
		helloworld.Greet(grpcClient, "server", name)
		time.Sleep(helloworld.Timeout)
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
