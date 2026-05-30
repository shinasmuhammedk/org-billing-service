package webhook

import (
	"log/slog"

	subscriptionService "org-billing-service/internal/service/subscription"
)

var (
	SubscriptionService *subscriptionService.Service
	appLogger           *slog.Logger
)

func Init(service *subscriptionService.Service, logger *slog.Logger) {
	SubscriptionService = service
	appLogger = logger
}