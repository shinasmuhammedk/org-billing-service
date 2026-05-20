package main

import (
	"log"
	"net"

	"org-billing-service/internal/app"
	"org-billing-service/internal/db"
	grpcserver "org-billing-service/internal/grpc"
	subscriptionRepo "org-billing-service/internal/repo/subscription"
	subscriptionService "org-billing-service/internal/service/subscription"
	"org-billing-service/internal/webhook"

	pb "org-billing-service/proto"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	db.Connect()

	go func() {
		r := gin.Default()

		app.RegisterRoutes(r)

		log.Println("Billing HTTP server running on :8081")

		if err := r.Run(":8081"); err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()

	subRepo := subscriptionRepo.NewPostgresRepository(
		db.QueriesInstance,
	)

	subService := subscriptionService.NewService(
		subRepo,
	)

	webhook.Init(subService)
        
	pb.RegisterBillingServiceServer(
		server,
		grpcserver.NewBillingServer(subService),
	)

	log.Println("Billing gRPC server running on :50052")

	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
