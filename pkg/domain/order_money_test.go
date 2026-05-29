package domain

import "testing"

func TestShopOrderMoneyFields(t *testing.T) {
	o := ShopOrder{OrderTotal: INR(12345), Discount: INR(500)}
	net, err := o.OrderTotal.Sub(o.Discount)
	if err != nil || net.AmountMinor != 11845 {
		t.Fatalf("got %v err=%v", net, err)
	}
}

func TestOrderLineMoneyField(t *testing.T) {
	acc := INR(0)
	lines := []OrderLine{{Price: INR(100)}, {Price: INR(250)}}
	for _, l := range lines {
		var err error
		acc, err = acc.Add(l.Price)
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}
	}
	if acc.AmountMinor != 350 {
		t.Fatalf("got %d want 350", acc.AmountMinor)
	}
}

func TestOrderReturnAndPaymentMethodMoneyFields(t *testing.T) {
	r := OrderReturn{RefundAmount: INR(999)}
	if r.RefundAmount.AmountMinor != 999 {
		t.Fatalf("refund got %d", r.RefundAmount.AmountMinor)
	}
	p := PaymentMethod{MaximumAmount: INR(50000)}
	if p.MaximumAmount.AmountMinor != 50000 {
		t.Fatalf("max got %d", p.MaximumAmount.AmountMinor)
	}
}
