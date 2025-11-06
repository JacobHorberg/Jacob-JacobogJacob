package main

import (
	proto "AquaWallahServer/grpc"
	"bufio"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	go func() {
		if err := srv.Serve(listerner); err != nil {
			panic(err)
		}
	}()

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		text := sc.Text()
		if text == "" || len(text) > 128 {
			continue
		}
		conn, err := grpc.Dial(text, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println("Failed to connect:", err)
			continue
		}
		defer conn.Close()
	}
}
