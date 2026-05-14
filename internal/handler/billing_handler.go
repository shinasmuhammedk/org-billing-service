// internal/handler/billing_handler.go

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	billingstripe "org-billing-service/internal/stripe"
)

func CreateCheckout(c *gin.Context) {
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var req struct {
		PriceID string `json:"price_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if req.PriceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "price_id is required",
		})
		return
	}

	checkoutURL, err := billingstripe.CreateCheckoutSession(userID, req.PriceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": checkoutURL,
	})
}