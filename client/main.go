package main

import (
	"context"
	pb "github.com/bigkucha/grpc-go/proto"
	"github.com/bigkucha/grpc-go/resolver"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/balancer"
	"go.etcd.io/etcd/clientv3/balancer/picker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock())
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
