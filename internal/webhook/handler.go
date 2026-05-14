package webhook

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

func StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		MaxBodyBytes,
	)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read body",
		})
		return
	}

	signatureHeader := c.GetHeader("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(
		payload,
		signatureHeader,
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)

	if err != nil {
		log.Println("Stripe webhook signature error:", err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	switch event.Type {

	case "checkout.session.completed":

		var session stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "failed to parse checkout session",
			})
			return
		}

		userID := session.Metadata["user_id"]

		log.Println("===== CHECKOUT COMPLETED =====")
		log.Println("User ID:", userID)
		log.Println("Customer ID:", session.Customer.ID)
		log.Println("Subscription ID:", session.Subscription.ID)

		subscriptionID := session.Subscription.ID
		customerID := session.Customer.ID

		sub, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			log.Println("failed to retrieve subscription:", err)
			return
		}

		if len(sub.Items.Data) == 0 {
			log.Println("subscription has no price item")
			return
		}

		priceID := sub.Items.Data[0].Price.ID
		status := string(sub.Status)

		currentPeriodEnd := time.Unix(
			sub.Items.Data[0].CurrentPeriodEnd,
			0,
		)

		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			log.Println("invalid user id:", err)
			return
		}

		err = SubscriptionService.SyncSubscription(
			c.Request.Context(),
			parsedUserID,
			customerID,
			subscriptionID,
			priceID,
			status,
			currentPeriodEnd,
		)

		if err != nil {
			log.Println("failed to sync subscription:", err)
			return
		}

		log.Println("subscription saved successfully")

		log.Println("===== STRIPE SUBSCRIPTION DETAILS =====")
		log.Println("Customer ID:", customerID)
		log.Println("Subscription ID:", subscriptionID)
		log.Println("Price ID:", priceID)
		log.Println("Status:", status)

	case "customer.subscription.updated":
		log.Println("subscription updated")

	case "customer.subscription.deleted":
		log.Println("subscription deleted")

	case "invoice.payment_failed":
		log.Println("payment failed")
	}

	c.JSON(http.StatusOK, gin.H{
		"received": true,
	})
}
