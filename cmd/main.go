package main

import (
	"go-grpc/cmd/config"
	"go-grpc/cmd/services"
	productPb "go-grpc/pb/product"
	"log"
	"net"

	"google.golang.org/grpc"
)

const (
	port = ":5000"
)

func main() {
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen %v", err.Error())
	}

	db, err := config.GetDB()
	if err != nil {
		log.Fatalf("Failed to connect database %v", err.Error())
	}

	grpcServer := grpc.NewServer()
	productService := services.ProductService{DB: db}
	productPb.RegisterProductServiceServer(grpcServer, &productService)

	log.Printf("Server started at %v", listen.Addr())

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve %v", err.Error())
	}
}
