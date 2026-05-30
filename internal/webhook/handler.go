package webhook

import (
	"encoding/json"
	"io"
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

		appLogger.Error("failed to read stripe webhook body",
			"error", err.Error(),
		)

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

		appLogger.Error("stripe webhook signature verification failed",
			"error", err.Error(),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	appLogger.Info("stripe event received",
		"event_type", event.Type,
	)

	switch event.Type {

	case "checkout.session.completed":

		var checkoutSession stripe.CheckoutSession

		if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {

			appLogger.Error("failed to parse checkout session",
				"error", err.Error(),
			)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "failed to parse checkout session",
			})
			return
		}

		userID := checkoutSession.Metadata["user_id"]

		if userID == "" {

			appLogger.Error("missing user_id in checkout metadata")

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing user_id in checkout metadata",
			})
			return
		}

		if checkoutSession.Customer == nil || checkoutSession.Customer.ID == "" {

			appLogger.Error("missing customer id in checkout session",
				"user_id", userID,
			)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing customer id",
			})
			return
		}

		if checkoutSession.Subscription == nil || checkoutSession.Subscription.ID == "" {

			appLogger.Error("missing subscription id in checkout session",
				"user_id", userID,
			)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "missing subscription id",
			})
			return
		}

		subscriptionID := checkoutSession.Subscription.ID
		customerID := checkoutSession.Customer.ID

		appLogger.Info("checkout session completed",
			"user_id", userID,
			"customer_id", customerID,
			"subscription_id", subscriptionID,
		)

		sub, err := subscription.Get(subscriptionID, nil)
		if err != nil {

			appLogger.Error("failed to retrieve subscription from stripe",
				"subscription_id", subscriptionID,
				"error", err.Error(),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve subscription",
			})
			return
		}

		if len(sub.Items.Data) == 0 || sub.Items.Data[0].Price == nil {

			appLogger.Error("subscription has no price item",
				"subscription_id", subscriptionID,
			)

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

			appLogger.Error("invalid user id",
				"user_id", userID,
				"error", err.Error(),
			)

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

			appLogger.Error("failed to sync subscription",
				"user_id", userID,
				"customer_id", customerID,
				"subscription_id", subscriptionID,
				"price_id", priceID,
				"status", status,
				"error", err.Error(),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to sync subscription",
			})
			return
		}

		appLogger.Info("subscription synced successfully",
			"user_id", userID,
			"customer_id", customerID,
			"subscription_id", subscriptionID,
			"price_id", priceID,
			"status", status,
		)

	case "customer.subscription.updated":

		appLogger.Info("customer subscription updated event received")

	case "customer.subscription.deleted":

		appLogger.Warn("customer subscription deleted event received")

	case "invoice.payment_failed":

		appLogger.Error("invoice payment failed event received")

	default:

		appLogger.Warn("unhandled stripe event received",
			"event_type", event.Type,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"received": true,
	})
}