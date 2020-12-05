package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"etcd-grpc/ip"
	"etcd-grpc/naming"
	pb "etcd-grpc/proto"

	"google.golang.org/grpc"
)

var (
	schema = "hello"

	serv = flag.String("service", "greeter_service", "service name")
	port = flag.Int("port", 50001, "listening port")
	reg  = flag.String("reg", "127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", ip.InternalIP(), *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	defer s.GracefulStop()
	log.Printf("starting hello service at %d", *port)

	err = naming.Register(*reg, *serv, addr, schema, 15)
	if err != nil {
		panic(err)
	}
	fmt.Println("regiester at:", *reg)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("receive signal '%v'", s)
		if err := naming.UnRegister(*serv, addr, schema); err != nil {
			fmt.Println("unregister err:", err)
		}
		os.Exit(1)
	}()

	s.Serve(lis)
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("%v: Receive is %s\n", time.Now(), in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}
