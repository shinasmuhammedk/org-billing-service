package grpc

import (
	"context"
	"errors"
	"os"

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

	var priceID string

	switch req.Plan {
	case "pro":
		priceID = os.Getenv("STRIPE_PRO_PRICE_ID")

	default:
		return nil, errors.New("invalid plan")
	}

	url, err := stripeCheckout.CreateCheckoutSession(
		req.UserId,
		priceID,
	)

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

	plan, status, err := s.subscriptionService.GetUserSubscription(
		ctx,
		req.UserId,
	)

	if err != nil {
		return nil, err
	}

	return &pb.GetUserSubscriptionResponse{
		Plan:   plan,
		Status: status,
	}, nil
}

func (s *BillingServer) CreatePortalSession(
	ctx context.Context,
	req *pb.CreatePortalSessionRequest,
) (*pb.CreatePortalSessionResponse, error) {

	customerID, err := s.subscriptionService.GetStripeCustomerID(
		ctx,
		req.UserId,
	)

	if err != nil {
		return nil, err
	}

	if customerID == "" {
		return nil, errors.New("stripe customer id not found")
	}

	url, err := stripeCheckout.CreatePortalSession(customerID)
	if err != nil {
		return nil, err
	}

	return &pb.CreatePortalSessionResponse{
		PortalUrl: url,
	}, nil
}