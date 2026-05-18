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
	stripewebhook "github.com/stripe/stripe-go/v82/webhook"
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

	event, err := stripewebhook.ConstructEventWithOptions(
		payload,
		signatureHeader,
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
		stripewebhook.ConstructEventOptions{
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

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	log.Println("Stripe event received:", event.Type)

	switch event.Type {

	case "checkout.session.completed":
		var checkoutSession stripe.CheckoutSession

		if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "failed to parse checkout session",
			})
			return
		}

		userID := checkoutSession.Metadata["user_id"]

		if userID == "" {
			log.Println("missing user_id in checkout metadata")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing user_id in checkout metadata",
			})
			return
		}

		if checkoutSession.Customer == nil || checkoutSession.Customer.ID == "" {
			log.Println("missing customer id in checkout session")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing customer id",
			})
			return
		}

		if checkoutSession.Subscription == nil || checkoutSession.Subscription.ID == "" {
			log.Println("missing subscription id in checkout session")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing subscription id",
			})
			return
		}

		subscriptionID := checkoutSession.Subscription.ID
		customerID := checkoutSession.Customer.ID

		log.Println("===== CHECKOUT COMPLETED =====")
		log.Println("User ID:", userID)
		log.Println("Customer ID:", customerID)
		log.Println("Subscription ID:", subscriptionID)

		sub, err := subscription.Get(subscriptionID, nil)
		if err != nil {
			log.Println("failed to retrieve subscription:", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve subscription",
			})
			return
		}

		if len(sub.Items.Data) == 0 || sub.Items.Data[0].Price == nil {
			log.Println("subscription has no price item")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "subscription has no price item",
			})
			return
		}

		priceID := sub.Items.Data[0].Price.ID
		status := string(sub.Status)
		currentPeriodEnd := time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0)

		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			log.Println("invalid user id:", err)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user id",
			})
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

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to sync subscription",
			})
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

	default:
		log.Println("unhandled event:", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{
		"received": true,
	})
}