package main

import (
	"context"
	"github.com/bigkucha/grpc-go/pbs"
	pb "github.com/bigkucha/grpc-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

const port = ":50001"

// server is used to implement
type server struct{}
type studentServer struct{}

func (s *server) GetUserInfo(ctx context.Context, in *pb.RequestUser) (*pb.User, error) {
	log.Printf("请求查看用户 %d 的信息", in.Id)
	return &pb.User{ID: in.Id, Name: "张三", Mobile: "18898987765", Age: 12}, nil
}

func (s *server) Create(ctx context.Context, in *pb.User) (*pb.User, error) {
	log.Printf("创建用户，%+v", in)
	return &pb.User{ID: 999, Name: in.Name, Mobile: in.Mobile, Age: in.Age}, nil
}

func (ss *studentServer) Get(ctx context.Context, in *pbs.RequestStudent) (*pbs.ResponseStudent, error) {
	log.Printf("请求查看学生 %d 到信息", in.Id)
	return &pbs.ResponseStudent{Id: in.Id, Name: "张三", Type: pbs.ResponseStudent_Female}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})
	pbs.RegisterStudentServer(s, &studentServer{})
	//pb.RegisterOrderServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
