package domain

import "testing"

func TestOrderStatusTypeIsValid(t *testing.T) {
	if !StatusOrderPlaced.IsValid() || !StatusOrderReturned.IsValid() {
		t.Error("known order statuses should be valid")
	}
	if OrderStatusType("shipped").IsValid() {
		t.Error("unknown order status should be invalid")
	}
}

func TestPaymentTypeIsValid(t *testing.T) {
	if !RazopayPayment.IsValid() || !CodPayment.IsValid() || !StripePayment.IsValid() {
		t.Error("known payment types should be valid")
	}
	if PaymentType("bitcoin").IsValid() {
		t.Error("unknown payment type should be invalid")
	}
}

func TestTransactionTypeIsValid(t *testing.T) {
	if !Debit.IsValid() || !Credit.IsValid() {
		t.Error("DEBIT/CREDIT should be valid")
	}
	if TransactionType("REVERSAL").IsValid() {
		t.Error("unknown transaction type should be invalid")
	}
}

func TestEnquiryStatusIsValid(t *testing.T) {
	if !StatusNew.IsValid() || !StatusDisputeResolved.IsValid() {
		t.Error("known enquiry statuses should be valid")
	}
	if EnquiryStatus("paused").IsValid() {
		t.Error("unknown enquiry status should be invalid")
	}
}
