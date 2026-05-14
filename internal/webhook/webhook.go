package webhook

import subscriptionService "org-billing-service/internal/service/subscription"

var SubscriptionService *subscriptionService.Service

func Init(service *subscriptionService.Service) {
	SubscriptionService = service
}