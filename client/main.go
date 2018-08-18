package main

import (
	pb "github.com/bigkucha/grpc-go/proto"
	"log"

	"context"
	"time"

	"google.golang.org/grpc"
)

const address = "localhost:50001"

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalln("did not connect %v", err)
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
