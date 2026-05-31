package domain

import "testing"

func TestCouponMinimumCartPriceMoney(t *testing.T) {
	c := Coupon{MinimumCartPrice: INR(50000)}
	if c.MinimumCartPrice.AmountMinor != 50000 {
		t.Fatalf("got %d", c.MinimumCartPrice.AmountMinor)
	}
}

func TestSubscriptionPlanPriceMoney(t *testing.T) {
	p := SubscriptionPlan{PriceMonthly: INR(19900)}
	if p.PriceMonthly.AmountMinor != 19900 {
		t.Fatalf("got %d", p.PriceMonthly.AmountMinor)
	}
}

func TestSubscriptionOrderConsolidatedPrice(t *testing.T) {
	o := SubscriptionOrder{Price: INR(19900)}
	if o.Price.AmountMinor != 19900 || o.Price.Currency != CurrencyINR {
		t.Fatalf("got %v", o.Price)
	}
}
