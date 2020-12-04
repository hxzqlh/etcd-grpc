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

	"etcd-grpc/naming"
	pb "etcd-grpc/proto"

	"google.golang.org/grpc"
)

var (
	serv = flag.String("service", "greeter_service", "service name")
	port = flag.Int("port", 50001, "listening port")
	reg  = flag.String("reg", "127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		panic(err)
	}

	// TODO: find my ip
	err = naming.Register(*reg, *serv, fmt.Sprintf("%s:%v", "127.0.0.1", *port), naming.Schema, 15)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		log.Printf("receive signal '%v'", s)
		naming.UnRegister(*serv, fmt.Sprintf("%s:%v", "127.0.0.1", *port), naming.Schema)
		os.Exit(1)
	}()

	log.Printf("starting hello service at %d", *port)
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	defer s.GracefulStop()
	s.Serve(lis)
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("%v: Receive is %s\n", time.Now(), in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}
