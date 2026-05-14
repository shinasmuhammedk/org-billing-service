package app

import (
	"org-billing-service/internal/handler"
	webhookHandler "org-billing-service/internal/webhook"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.POST("/billing/checkout", handler.CreateCheckout)
	r.POST("/webhooks/stripe", webhookHandler.StripeWebhook)
}