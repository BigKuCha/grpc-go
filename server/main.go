package main

import (
	"context"
	"encoding/json"
	pb "github.com/bigkucha/grpc-go/proto"
	etcdresolver "github.com/bigkucha/grpc-go/resolver"
	"github.com/bigkucha/grpc-go/trace"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/resolver"
	"log"
	"net"
	"time"
)

const ip = "127.0.0.1"
const port = "50001"
const serviceName = "UserService"

// server is used to implement
type server struct{}

func (s *server) StreamUserInfo(infoServer pb.UserService_StreamUserInfoServer) error {
	var err error
	for i := 1; i < 4; i++ {
		err = infoServer.Send(&pb.User{
			ID:     int32(i),
			Name:   "stream",
			Mobile: "199898989898",
			Age:    int32(i) * 10,
		})
	}
	log.Println("stream send finished")
	return err
}

func (s *server) GetUserInfo(ctx context.Context, in *pb.RequestUser) (*pb.User, error) {
	log.Printf("请求查看用户 %d 的信息", in.Id)
	return &pb.User{ID: in.Id, Name: "张三", Mobile: "18898987765", Age: 12}, nil
}

func (s *server) Create(ctx context.Context, in *pb.User) (*pb.User, error) {
	log.Printf("创建用户，%+v", in)
	time.Sleep(time.Second)
	return &pb.User{ID: 999, Name: in.Name, Mobile: in.Mobile, Age: in.Age}, nil
}

func registerSrv() {
	// 连接etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	// 声明租约
	lease, err := client.Lease.Grant(context.TODO(), 10)
	if err != nil {
		panic(err)
	}
	addr := net.JoinHostPort(ip, port)
	target := etcdresolver.Target(serviceName, addr)
	address := resolver.Address{
		Addr:       addr,
		Type:       0,
		ServerName: "",
		Metadata:   nil,
	}
	addressStr, _ := json.Marshal(address)
	// 写入etcd
	_, err = client.Put(context.TODO(), target, string(addressStr), clientv3.WithLease(lease.ID))
	if err != nil {
		panic(err)
	}
	ch, _ := client.KeepAlive(context.TODO(), lease.ID)
	go func() {
		for response := range ch {
			_ = response
			//fmt.Printf("keep: %#v\n", response)
		}
	}()
}

func main() {
	// 注册服务
	addr := net.JoinHostPort(ip, port)
	registerSrv()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//tracer := trace.GetZipkinBasicTracer("GRPCServer", addr)
	tracer := trace.GetZipkinTracer("GRPCServer", addr)
	//s := grpc.NewServer(grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer)))
	s := grpc.NewServer(
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(tracer)),
				streamServerInterceptor(),
			)),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(tracer)),
				UnaryServerInterceptor(),
			),
		),
	)
	pb.RegisterUserServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		log.Println("haha, unary interceptor")
		return handler(ctx, req)
	}
}
func streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Println("middleware")
		return handler(srv, ss)
	}
}
