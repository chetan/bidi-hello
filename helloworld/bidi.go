package helloworld

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

// Connect to the server and establish a yamux channel for bidi grpc
func Connect(addr string, grpcServer *grpc.Server) *grpc.ClientConn {
	// create client
	yDialer := NewYamuxDialer()
	gconn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithDialer(yDialer.Dial))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	go func() {
		for {
			// connect and loop forever to keep retrying in the case that the connection
			// drops for any reason
			connect(addr, grpcServer, yDialer)
			time.Sleep(1 * time.Second)
		}
	}()

	return gconn
}

// connect to server
//
// dials out to the target server and then sets up a yamux channel for
// multiplexing both grpc client and server on the same underlying tcp socket
//
// this is separate from the run-loop for easy resource cleanup via defer
func connect(addr string, grpcServer *grpc.Server, yDialer *YamuxDialer) {
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
	yDialer.SetSession(session)

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

// Listen starts the server side of the yamux channel for bidi grpc.
//
// Standard unidirectional grpc without yamux is still accepted as normal for
// clients which don't support or need bidi communication.
func Listen(addr string, grpcServer *grpc.Server) *grpc.ClientConn {
	// create client
	yDialer := NewYamuxDialer()
	gconn, err := grpc.Dial("localhost:50000", grpc.WithInsecure(), grpc.WithDialer(yDialer.Dial))
	if err != nil {
		fmt.Println("failed to create grpc client: ", err)
		return nil
	}

	// create underlying tcp listener
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// use cmux to look for plain http2 clients
	// this allows us to support non-yamux enabled clients (such as non-golang impls)
	mux := cmux.New(lis)
	grpcL := mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	yamuxL := mux.Match(cmux.Any())

	// start servers for both plain-grpc and yamux bidi grpc
	go grpcServer.Serve(grpcL)
	go listenLoop(yamuxL, grpcServer, yDialer)

	go mux.Serve()

	return gconn
}

func listenLoop(lis net.Listener, grpcServer *grpc.Server, dialer *YamuxDialer) {
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

		// pass session to grpc client
		dialer.SetSession(session)
	}

}
