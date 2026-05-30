package subscription

import (
	"context"
	"log/slog"
	"org-billing-service/internal/db"
	"org-billing-service/internal/helper"
	"time"

	repo "org-billing-service/internal/repo/subscription"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	repo repo.Repository
    logger *slog.Logger
}

func NewService(r repo.Repository, logger *slog.Logger) *Service {
	return &Service{
		repo: r,
        logger: logger,
	}
}

func (s *Service) GetUserSubscription(
	ctx context.Context,
	userID string,
) (string, string, error) {

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return "", "", err
	}

	subscription, err := s.repo.GetByUserID(
		ctx,
		parsedUserID,
	)

	if err != nil {
		return "free", "active", nil
	}

	return subscription.Plan, subscription.Status, nil
}

func (s *Service) SyncSubscription(
	ctx context.Context,
	userID uuid.UUID,
	customerID string,
	subscriptionID string,
	priceID string,
	status string,
	currentPeriodEnd time.Time,
) error {
	_, err := s.repo.UpsertSubscription(ctx, db.UpsertSubscriptionParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},

		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},

		StripeCustomerID: pgtype.Text{
			String: customerID,
			Valid:  true,
		},

		StripeSubscriptionID: pgtype.Text{
			String: subscriptionID,
			Valid:  true,
		},

		Plan: helper.GetPlanFromPriceID(priceID), 
        Status: status,

		CurrentPeriodEnd: pgtype.Timestamp{
			Time:  currentPeriodEnd,
			Valid: true,
		},
	})

	return err
}

func (s *Service) GetStripeCustomerID(
	ctx context.Context,
	userID string,
) (string, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}

	subscription, err := s.repo.GetByUserID(ctx, parsedUserID)
	if err != nil {
		return "", err
	}

	return subscription.StripeCustomerID.String, nil
}
