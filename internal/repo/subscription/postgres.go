package subscription

import (
	"context"

	"org-billing-service/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type postgresRepository struct {
	q *db.Queries
}

func NewPostgresRepository(q *db.Queries) Repository {
	return &postgresRepository{q: q}
}

func (r *postgresRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (db.Subscription, error) {
	return r.q.GetSubscriptionByUserID(ctx, pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
}

func (r *postgresRepository) UpsertSubscription(
	ctx context.Context,
	params db.UpsertSubscriptionParams,
) (db.Subscription, error) {
	return r.q.UpsertSubscription(ctx, params)
}