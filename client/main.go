package main

import (
	"context"
	"fmt"
	pb "github.com/bigkucha/grpc-go/proto"
	"github.com/bigkucha/grpc-go/resolver"
	"github.com/bigkucha/grpc-go/trace"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/balancer"
	"go.etcd.io/etcd/clientv3/balancer/picker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func init() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	resolver.EtcdRegister(client)
	cfg := balancer.Config{
		Policy: picker.RoundrobinBalanced,
		Name:   "",
		Logger: zap.NewExample(),
	}
	balancer.RegisterBuilder(cfg)
}

func main() {
	target := resolver.TargetPrefix("UserService")
	log.Println("target:", target)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	tracer := trace.GetZipkinBasicTracer("MyClient", "localhost:10110")

	var conn, err = grpc.DialContext(
		ctx, target, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithStatsHandler(zipkingrpc.NewClientHandler(tracer)),
	)
	if err != nil {
		log.Fatalf("did not connect %v\n", err)
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	r, err := c.GetUserInfo(ctx, &pb.RequestUser{Id: 112})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("用户信息:%#v\n", r)

	//streamCall(err, c)
	// 预留时间上报zipkin
	time.Sleep(time.Second * 2)
}

func streamCall(err error, c pb.UserServiceClient) {
	infoclient, err := c.StreamUserInfo(context.Background())
	if err != nil {
		panic(err)
	}
	for {
		u, err := infoclient.Recv()
		if err == io.EOF {
			log.Println("接收完毕")
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(u)
	}
	time.Sleep(time.Second)
}
