package main

import (
	"context"
	"fmt"
	pb "github.com/bigkucha/grpc-go/proto"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/naming"
	"google.golang.org/grpc"
	"log"
	"time"
)

const address = "localhost:50001"
const target = "/services/UserService"

func main() {
	// 连接ETCD
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	// 创建服务解析
	resolver := &naming.GRPCResolver{Client: client}
	b := grpc.RoundRobin(resolver)
	fmt.Println(resolver.Resolve(target))
	conn, err := grpc.Dial(target, grpc.WithBalancer(b), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect %v\n", err)
	}
	defer conn.Close()

	c := pb.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetUserInfo(ctx, &pb.RequestUser{Id: 112})

	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("用户信息: %+v", r)
}
