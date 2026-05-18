package stripe

import (
	"os"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billingportal/session"
)

func CreatePortalSession(customerID string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(os.Getenv("STRIPE_PORTAL_RETURN_URL")),
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}

	return s.URL, nil
}