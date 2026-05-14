package subscription

import (
	"context"

	"org-billing-service/internal/db"

	"github.com/google/uuid"
)

type Repository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (db.Subscription, error)

	UpsertSubscription(ctx context.Context, params db.UpsertSubscriptionParams) (db.Subscription, error)
}
