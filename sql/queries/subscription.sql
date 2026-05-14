-- name: GetSubscriptionByUserID :one
SELECT *
FROM subscriptions
WHERE user_id = $1
LIMIT 1;

-- name: CreateSubscription :one
INSERT INTO subscriptions (
    id,
    user_id,
    stripe_customer_id,
    stripe_subscription_id,
    plan,
    status,
    current_period_end
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpsertSubscription :one
INSERT INTO subscriptions (
    id,
    user_id,
    stripe_customer_id,
    stripe_subscription_id,
    plan,
    status,
    current_period_end
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (user_id)
DO UPDATE SET
    stripe_customer_id = EXCLUDED.stripe_customer_id,
    stripe_subscription_id = EXCLUDED.stripe_subscription_id,
    plan = EXCLUDED.plan,
    status = EXCLUDED.status,
    current_period_end = EXCLUDED.current_period_end
RETURNING *;