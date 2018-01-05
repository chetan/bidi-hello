# bidi-hello

Bidirectional gRPC over a single channel (socket) using yamux.

This sample sets up a client which establishes a single connection to a server
and allows RPC calls to be initiated in either direction. Both client and server
run the grpc server and client code. Client and server can be started in any
order and can also be restarted at will. Connections will automatically be
re-established.

## Quickstart

```sh
go get github.com/chetan/bidi-hello/bin/...
greeter_client
<open new tab>
greeter_server
```

## How it works

### helloworld/server.go

This is the grpc server implementation code (as in the original greeter_server)

### helloworld/client.go

This just calls the greet methods using the given client stub (as in the original greeter_client)

### helloworld/yamux_dialer.go

A `grpc.Dialer` implementation which uses a `yamux.Session` for the underlying `net.Con`.

### helloworld/bidi.go

This wraps the setup of both the client and server connections using cmux and
yamux.

### bin/greeter_server/main.go

This starts the "server" end of the bidi channel, i.e., the end which starts a
TCP listener.

### bin/greeter_client/main.go

This starts the "client" end of the bidi channel, i.e., the end which initiates
the TCP connection to the server.
