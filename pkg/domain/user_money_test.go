package domain

import "testing"

func TestCartMoneyFields(t *testing.T) {
	c := Cart{TotalPrice: INR(10000), DiscountAmount: INR(1500)}
	net, err := c.TotalPrice.Sub(c.DiscountAmount)
	if err != nil || net.AmountMinor != 8500 {
		t.Fatalf("got %v err=%v", net, err)
	}
}

func TestWalletAndTransactionMoneyFields(t *testing.T) {
	w := Wallet{TotalAmount: INR(2000)}
	credit := Transaction{Amount: INR(500), TransactionType: Credit}
	updated, err := w.TotalAmount.Add(credit.Amount)
	if err != nil || updated.AmountMinor != 2500 {
		t.Fatalf("got %v err=%v", updated, err)
	}
}
