package main

import (
	proto "AquaWallahServer/grpc"
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AquaWallahServer struct {
	proto.UnimplementedAquaWallahServer
	servers []Server
}

type Server struct {
	conn *grpc.ClientConn
	port string
}

func main() {
	if len(os.Args) == 1 {
		panic("No port supplied")
	}

	port := os.Args[1]
	if len(port) == 0 {
		panic("invalid port: " + port)
	}

	srv := grpc.NewServer()
	listerner, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	aws := AquaWallahServer{}
	proto.RegisterAquaWallahServer(srv, &aws)

	go func() {
		if err := srv.Serve(listerner); err != nil {
			panic(err)
		}
	}()

	// TODO: listen for incoming connections
	fmt.Println("usage:")
	fmt.Println("  port <port number>    connect to a port on localhost")
	fmt.Println("  request`              request access to the critical zone")
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		actions := strings.Split(sc.Text(), " ")
		if len(actions) > 2 {
			panic("Too many actions supplied")
		}
		action := actions[0]
		switch action {
		case "port":
			add_server(&aws.servers, actions[1])
			defer aws.servers[len(aws.servers)-1].conn.Close()
		case "request":
			request(&aws.servers)
			fmt.Println("Request not implemented")
		default:
			fmt.Println("Invalid action")
		}
	}
}

func add_server(servers *[]Server, port string) {
	portRegex := "^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])"
	if port == "" {
		fmt.Println("No port specified")
		return
	} else if match, err := regexp.MatchString(portRegex, port); !match {
		fmt.Println("Invalid port:", port)
		return
	} else if err != nil {
		fmt.Println("Failed to connect:", err)
	}

	conn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect:", err)
	}
	fmt.Println("Successfully connected to: localhost:" + port)
	*servers = append(*servers, Server{conn: conn, port: port})
}

// TODO: implement agrawala
func request(servers *[]Server) {

}
