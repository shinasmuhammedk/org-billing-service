package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	billingstripe "org-billing-service/internal/stripe"
)

var appLogger *slog.Logger

func InitLogger(logger *slog.Logger) {
	appLogger = logger
}

func CreateCheckout(c *gin.Context) {
	userID := c.GetString("user_id")

	if userID == "" {
		appLogger.Warn("unauthorized checkout request")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var req struct {
		PriceID string `json:"price_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {

		appLogger.Error("invalid checkout request body",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if req.PriceID == "" {

		appLogger.Warn("missing price id in checkout request",
			slog.String("user_id", userID),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "price_id is required",
		})
		return
	}

	appLogger.Info("creating stripe checkout session",
		slog.String("user_id", userID),
		slog.String("price_id", req.PriceID),
	)

	checkoutURL, err := billingstripe.CreateCheckoutSession(userID, req.PriceID)
	if err != nil {

		appLogger.Error("failed to create stripe checkout session",
			slog.String("user_id", userID),
			slog.String("price_id", req.PriceID),
			slog.String("error", err.Error()),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	appLogger.Info("stripe checkout session created successfully",
		slog.String("user_id", userID),
		slog.String("price_id", req.PriceID),
	)

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": checkoutURL,
	})
}