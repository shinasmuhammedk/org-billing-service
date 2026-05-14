// internal/stripe/checkout.go

package stripe

import (
	"os"

	stripego "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

func CreateCheckoutSession(userID string, priceID string) (string, error) {
	stripego.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripego.CheckoutSessionParams{
		Mode: stripego.String(string(stripego.CheckoutSessionModeSubscription)),

		LineItems: []*stripego.CheckoutSessionLineItemParams{
			{
				Price:    stripego.String(priceID),
				Quantity: stripego.Int64(1),
			},
		},

		SuccessURL: stripego.String(os.Getenv("STRIPE_SUCCESS_URL")),
		CancelURL:  stripego.String(os.Getenv("STRIPE_CANCEL_URL")),

		Metadata: map[string]string{
			"user_id": userID,
		},

		SubscriptionData: &stripego.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id": userID,
			},
		},
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}

	return s.URL, nil
}