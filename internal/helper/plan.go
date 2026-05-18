package helper

import "os"

func GetPlanFromPriceID(priceID string) string {
	switch priceID {
	case os.Getenv("STRIPE_PRO_PRICE_ID"):
		return "pro"
	default:
		return "free"
	}
}