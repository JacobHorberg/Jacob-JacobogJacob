# Ricart-Agrawala algorithm
This is an implementation of the Ricart-Agrawala algorithm in go.
This is for the ITU distributed systems course.

## Generating grpc
1. run `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/proto.proto` in the root of the project

## Running the program
1. Open a terminal and navigate to the server directory (e.g. `cd server`)
1. Start the server by running the appropriate command (e.g. `go run server.go :<port>`).
   Where port is the port the server should listen on
1. Repeat above step for any number of ports you want to listen on
1. Connect the nodes to eachother with the `:<port>` command
   (Connecting is a bit unintuitive: when connecting nodes, they connect to eachother, but they are not added to the network, so at least one of the nodes must initiate a connection)
1. run `request` to request access to the critical zone, and `release` to release access
