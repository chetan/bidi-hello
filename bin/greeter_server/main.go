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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/chetan/bidi-hello/helloworld"
)

const (
	port = ":50051"
)

func main() {

	// create server
	grpcServer := grpc.NewServer()
	helloworld.RegisterGreeterServer(grpcServer, helloworld.NewServerImpl())
	reflection.Register(grpcServer)

	gconn := helloworld.Listen(port, grpcServer)
	defer gconn.Close()
	grpcClient := helloworld.NewGreeterClient(gconn)

	// start client conn back to agent
	for {
		helloworld.Greet(grpcClient, "client", "bob")
		time.Sleep(helloworld.Timeout)
	}

}
