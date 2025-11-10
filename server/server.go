package main

import (
	proto "AquaWallahServer/grpc"
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AquaWallahServer struct {
	proto.UnimplementedAquaWallahServer
	port      string
	servers   []Server
	timestamp int64
	mutex     sync.Mutex
}

type Server struct {
	conn *grpc.ClientConn
	port string
}

func main() {
	if len(os.Args) == 1 {
		panic("No port supplied")
	}

	aws := AquaWallahServer{}
	aws.port = os.Args[1]
	portRegex := "^:([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])"
	if match, err := regexp.MatchString(portRegex, aws.port); !match {
		panic("Invalid port:" + aws.port)
	} else if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	listerner, err := net.Listen("tcp", aws.port)
	if err != nil {
		panic(err)
	}
	proto.RegisterAquaWallahServer(srv, &aws)

	go func() {
		if err := srv.Serve(listerner); err != nil {
			panic(err)
		}
	}()

	// TODO: listen for incoming connections
	fmt.Println("usage:")
	fmt.Println("  :<port number>    connect to a port on localhost")
	fmt.Println("  request`          request access to the critical zone")
	fmt.Println("to quit press ctrl+shitf+d")
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		actions := strings.Split(sc.Text(), " ")
		for _, action := range actions {
			if action[0] == ':' {
				if action == aws.port {
					fmt.Println("Refusing to connect to self")
					continue
				}
				aws.mutex.Lock()
				err := add_server(&aws, action)
				if err != nil {
					fmt.Println("Failed to add server:", err)
					continue
				}
				conn := aws.servers[len(aws.servers)-1].conn
				c := proto.NewAquaWallahClient(conn)
				ctx, cancel := context.WithCancel(context.Background())
				aws.timestamp++
				_, err = c.JoinNetwork(ctx, &proto.Server{Timestamp: aws.timestamp, Port: aws.port})
				if err != nil {
					fmt.Println("Failed to connect to", action+": ", err)
					conn.Close()
					aws.mutex.Unlock()
					cancel()
					continue
				}
				fmt.Println("Successfully connected to: " + action)
				cancel()
				aws.mutex.Unlock()
			} else if action == "request" {
				for _, srv := range aws.servers {
					conn := srv.conn
					c := proto.NewAquaWallahClient(conn)
					ctx, cancel := context.WithCancel(context.Background())
					aws.timestamp++
					c.SendRequest(ctx, &proto.Request{})
					cancel()
				}
			} else {
				fmt.Println("Invalid action:", action)
			}
			fmt.Println(aws.timestamp)
		}
	}

	fmt.Println("Quitting...")
	for _, srv := range aws.servers {
		c := proto.NewAquaWallahClient(srv.conn)
		ctx, cancel := context.WithCancel(context.Background())
		aws.timestamp++
		_, err = c.LeaveNetwork(ctx, &proto.Server{Timestamp: aws.timestamp, Port: aws.port})
		cancel()
		srv.conn.Close()
	}
}

func add_server(aws *AquaWallahServer, port string) error {
	portRegex := "^:([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])"
	if match, err := regexp.MatchString(portRegex, port); !match {
		return errors.New("Invalid port:" + port)
	} else if err != nil {
		return err
	}

	conn, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	(*aws).servers = append((*aws).servers, Server{conn: conn, port: port})
	return nil
}

func (aws *AquaWallahServer) SendRequest(ctx context.Context, in *proto.Request) (*proto.Reply, error) {
	aws.mutex.Lock()
	defer aws.mutex.Unlock()
	fmt.Println("Stubbed out")
	return &proto.Reply{}, nil
}

func (aws *AquaWallahServer) JoinNetwork(ctx context.Context, server *proto.Server) (*proto.Empty, error) {
	aws.mutex.Lock()
	defer aws.mutex.Unlock()
	aws.timestamp = max(server.Timestamp, aws.timestamp) + 1
	for _, srv := range aws.servers {
		if srv.port == server.Port {
			fmt.Println("Refused to connect to", server.Port, "(already connected)")
			return &proto.Empty{}, errors.New("Already connected to server on " + server.Port)
		}
	}
	err := add_server(aws, string(server.Port))
	if err != nil {
		fmt.Println("Failed to add server:", err)
		return &proto.Empty{}, err
	}
	aws.servers = append(aws.servers, Server{conn: aws.servers[len(aws.servers)-1].conn, port: string(server.Port)})
	fmt.Println("Server on port", server.Port, "successfully connected")

	return &proto.Empty{}, nil
}

func (aws *AquaWallahServer) LeaveNetwork(ctx context.Context, server *proto.Server) (*proto.Empty, error) {
	aws.mutex.Lock()
	defer aws.mutex.Unlock()
	aws.timestamp = max(server.Timestamp, aws.timestamp) + 1
	for i, srv := range aws.servers {
		if srv.port == server.Port {
			srv.conn.Close()
			fmt.Println("Disconnected from: ", server.Port)
			aws.servers = append(aws.servers[:i], aws.servers[i+1:]...)
			return &proto.Empty{}, nil
		}
	}
	return &proto.Empty{}, errors.New(("Not connected to server on port " + server.Port))
}
