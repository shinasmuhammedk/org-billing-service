package grpc

import (
	"context"

	subscriptionService "org-billing-service/internal/service/subscription"
	stripeCheckout "org-billing-service/internal/stripe"
	pb "org-billing-service/proto"
)

type BillingServer struct {
	pb.UnimplementedBillingServiceServer
	subscriptionService *subscriptionService.Service
}

func NewBillingServer(subscriptionService *subscriptionService.Service) *BillingServer {
	return &BillingServer{
		subscriptionService: subscriptionService,
	}
}

func (s *BillingServer) CreateCheckoutSession(
	ctx context.Context,
	req *pb.CreateCheckoutSessionRequest,
) (*pb.CreateCheckoutSessionResponse, error) {

	url, err := stripeCheckout.CreateCheckoutSession(req.UserId, req.PriceId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateCheckoutSessionResponse{
		CheckoutUrl: url,
	}, nil
}

func (s *BillingServer) GetUserSubscription(
	ctx context.Context,
	req *pb.GetUserSubscriptionRequest,
) (*pb.GetUserSubscriptionResponse, error) {

	plan, status, err := s.subscriptionService.GetUserSubscription(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserSubscriptionResponse{
		Plan:   plan,
		Status: status,
	}, nil
}
