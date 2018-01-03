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

### bin/greeter_server/main.go

This starts the main listener endpoint on port 50051 and establishes a
`yamux.Server` session to create a channel over which both grpc client and
server can operate. yamux allows us to initiate the connection in either
direction and still provide both net.Conn and Net.Listener interfaces.

### bin/greeter_client/main.go

The client opens a single connection to the server on port 50051 and establishes
a `yamux.Client` session over which both grpc client and server are created, in
much the same as the server, above.
