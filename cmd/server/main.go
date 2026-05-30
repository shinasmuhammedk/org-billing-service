package main

import (
	"net"

	"org-billing-service/internal/app"
	"org-billing-service/internal/db"
	grpcserver "org-billing-service/internal/grpc"
	"org-billing-service/internal/logger"
	subscriptionRepo "org-billing-service/internal/repo/subscription"
	subscriptionService "org-billing-service/internal/service/subscription"
	"org-billing-service/internal/webhook"

	pb "org-billing-service/proto"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	appLogger := logger.New()

	if err := godotenv.Load(); err != nil {
		appLogger.Warn("no .env file found")
	}

	db.Connect()
	appLogger.Info("database connected")

	appLogger.Info("billing service starting", "service", "billing-service")

	subRepo := subscriptionRepo.NewPostgresRepository(db.QueriesInstance)

	subService := subscriptionService.NewService(
		subRepo,
        appLogger,
		// later pass logger here also
	)

	webhook.Init(subService,appLogger)

	go func() {
		r := gin.Default()

		app.RegisterRoutes(r)

		appLogger.Info("billing HTTP server running", "port", "8081")

		if err := r.Run(":8081"); err != nil {
			appLogger.Error("billing HTTP server failed", "error", err.Error())
		}
	}()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		appLogger.Error("failed to listen for gRPC", "error", err.Error())
		return
	}

	server := grpc.NewServer()

	pb.RegisterBillingServiceServer(
		server,
		grpcserver.NewBillingServer(subService,appLogger),
	)

	appLogger.Info("billing gRPC server running", "port", "50052")

	if err := server.Serve(lis); err != nil {
		appLogger.Error("billing gRPC server failed", "error", err.Error())
	}
}