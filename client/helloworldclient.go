package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	"etcd-grpc/naming"
	pb "etcd-grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
)

const (
	authority = "helloclient"
)

var (
	serv = flag.String("service", "greeter_service", "service name")
	reg  = flag.String("reg", "127.0.0.1:2379", "register etcd address")
)

func main() {
	flag.Parse()
	fmt.Println("serv", *serv)

	//var logger = grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
	//grpclog.SetLoggerV2(logger)

	// 解析etcd服务地址
	r := naming.NewResolver(*reg, naming.Schema)
	resolver.Register(r)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx,
		fmt.Sprintf("%s://%s/%s", naming.Schema, authority, *serv),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("conn...")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {
		client := pb.NewGreeterClient(conn)
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "world " + strconv.Itoa(t.Second())})
		if err == nil {
			fmt.Printf("%v: Reply is %s\n", t, resp.Message)
		} else {
			fmt.Println(err)
		}
	}
}
