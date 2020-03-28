package main

import (
	"context"
	"fmt"
	pb "github.com/bigkucha/grpc-go/proto"
	"go.etcd.io/etcd/clientv3"
	etcdnaming "go.etcd.io/etcd/clientv3/naming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

const ip = "127.0.0.1"
const port = "50001"
const target = "/services/UserService"

// server is used to implement
type server struct{}

func (s *server) GetUserInfo(ctx context.Context, in *pb.RequestUser) (*pb.User, error) {
	log.Printf("请求查看用户 %d 的信息", in.Id)
	return &pb.User{ID: in.Id, Name: "张三", Mobile: "18898987765", Age: 12}, nil
}

func (s *server) Create(ctx context.Context, in *pb.User) (*pb.User, error) {
	log.Printf("创建用户，%+v", in)
	return &pb.User{ID: 999, Name: in.Name, Mobile: in.Mobile, Age: in.Age}, nil
}

func main() {
	addr := net.JoinHostPort(ip, port)
	// 服务注册
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	lease, err := client.Lease.Grant(context.TODO(), 10)
	if err != nil {
		panic(err)
	}
	resolver := etcdnaming.GRPCResolver{Client: client}
	err = resolver.Update(context.TODO(), target, naming.Update{
		Op:       naming.Add,
		Addr:     addr,
		Metadata: "metadata-=",
	}, clientv3.WithLease(lease.ID))
	_, _ = client.Lease.KeepAlive(context.TODO(), lease.ID)
	fmt.Println("----", err)
	if err != nil {
		fmt.Println("===", err)
	}
	fmt.Println(addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
