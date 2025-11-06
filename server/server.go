package main

import (
	proto "AquaWallahServer/grpc"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
)

type AquaWallahServer struct {
	proto.UnimplementedAquaWallahServer
}

func main() {
	port := os.Args[1]
	fmt.Println(port)
	srv := grpc.NewServer()
	listerner, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	proto.RegisterAquaWallahServer(srv, &AquaWallahServer{})
	err = srv.Serve(listerner)
	if err != nil {
		panic(err)
	}
}
