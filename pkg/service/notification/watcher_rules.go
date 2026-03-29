package notification

import "fmt"

// DefaultOrderRule returns a WatchRule that fires when an order's status or
// cancellation reason changes, notifying both the customer and the seller.
//
// Expected Firestore document structure (orders/{orderId}):
//
//	{
//	  "status":             "confirmed" | "shipped" | "delivered" | "cancelled",
//	  "cancellationReason": "...",
//	  "userId":             "<customer-uid>",
//	  "shopId":             "<seller-uid>",
//	  ...
//	}
func DefaultOrderRule() WatchRule {
	return WatchRule{
		Collection:      "orders",
		MonitoredFields: []string{"status", "cancellationReason"},
		NotifyUser:      true,
		NotifySeller:    true,
		UserIDField:     "userId",
		SellerIDField:   "shopId",
		EventType:       "order_status_changed",
		MessageBuilder:  orderMessageBuilder,
	}
}

// DefaultProductRule fires when a product's price, stock, or approval status
// changes, notifying only the seller.
//
// Expected fields: "price", "stockCount", "approvalStatus", "shopId"
func DefaultProductRule() WatchRule {
	return WatchRule{
		Collection:      "products",
		MonitoredFields: []string{"price", "stockCount", "approvalStatus", "status"},
		NotifyUser:      false,
		NotifySeller:    true,
		SellerIDField:   "shopId",
		EventType:       "product_updated",
		MessageBuilder:  productMessageBuilder,
	}
}

// DefaultShopRule fires when a shop's verification status changes, notifying
// the seller.
//
// Expected fields: "verificationStatus", "isActive", "sellerId"
func DefaultShopRule() WatchRule {
	return WatchRule{
		Collection:      "shops",
		MonitoredFields: []string{"verificationStatus", "isActive", "status"},
		NotifyUser:      false,
		NotifySeller:    true,
		SellerIDField:   "sellerId",
		EventType:       "shop_status_changed",
		MessageBuilder:  shopMessageBuilder,
	}
}

// DefaultEnquiryRule fires when an enquiry/negotiation document in the
// "enquiry" collection changes.
//
// Actual Firestore document fields monitored:
//
//	status                 — overall enquiry state (e.g. "completed_accepted")
//	finalStatus            — negotiation outcome  ("accepted" | "rejected")
//	acceptedBy             — who finalised the deal ("seller" | "client")
//	acceptedPrice          — agreed price (string)
//	customerNegotiatedPrice — customer's latest counter-price
//	customerFinalResponse  — customer's final price
//	sellerFinalPrice       — seller's final asking price
//	sellerInitialPrice     — seller's opening price
//	availability           — product availability flag
//
// Recipient routing is driven by the document status:
//   - Seller updated  → notify user  (pending_seller_price, pending_seller_final, seller_final_update)
//   - Customer updated → notify seller (pending_customer_price, pending_customer_final, seller_final_update)
//   - Deal finalised   → notify the other party based on acceptedBy field
func DefaultEnquiryRule() WatchRule {
	return WatchRule{
		Collection: "enquiry",
		MonitoredFields: []string{
			"status",
			"finalStatus",
			"acceptedBy",
			"acceptedPrice",
			"customerNegotiatedPrice",
			"customerFinalResponse",
			"sellerFinalPrice",
			"sellerInitialPrice",
			"availability",
		},
		// NotifyUser / NotifySeller are overridden at runtime by RecipientResolver.
		NotifyUser:        true,
		NotifySeller:      true,
		UserIDField:       "userId",
		SellerIDField:     "sellerId",
		EventType:         "enquiry_updated",
		NotifyOnCreate:    true,
		MessageBuilder:    enquiryMessageBuilder,
		RecipientResolver: enquiryRecipientResolver,
	}
}

// enquiryRecipientResolver decides who receives a notification for an enquiry
// document update based on its current status and acceptedBy field.
//
// Routing rules:
//
//	Seller updated  → notify user:
//	  pending_seller_price | pending_seller_final | seller_final_update
//	  completed_accepted/rejected + acceptedBy == "seller"
//
//	Customer updated → notify seller:
//	  pending_customer_price | pending_customer_final | seller_final_update
//	  completed_accepted/rejected + acceptedBy == "client"
func enquiryRecipientResolver(docData map[string]interface{}) (notifyUser bool, notifySeller bool) {
	status, _ := docData["status"].(string)
	acceptedBy, _ := docData["acceptedBy"].(string)

	// Statuses that indicate the SELLER just acted → notify the USER.
	sellerActedStatuses := map[string]bool{
		"pending_seller_price": true,
		"pending_seller_final": true,
		"seller_final_update":  true,
	}

	// Statuses that indicate the CUSTOMER just acted → notify the SELLER.
	customerActedStatuses := map[string]bool{
		"pending_customer_price": true,
		"pending_customer_final": true,
		"seller_final_update":    true,
	}

	if sellerActedStatuses[status] {
		notifyUser = true
	}
	if customerActedStatuses[status] {
		notifySeller = true
	}

	// For completed/rejected deals, route to the OTHER party based on who finalised.
	if status == "completed_accepted" || status == "completed_rejected" {
		switch acceptedBy {
		case "seller":
			notifyUser = true
		case "client":
			notifySeller = true
		}
	}

	return
}

// NewCustomRule is a convenience builder for ad-hoc watch rules.
//
//	rule := NewCustomRule("returns", []string{"status", "refundAmount"}).
//	    NotifyBoth("userId", "shopId").
//	    WithEventType("return_updated").
//	    WithTemplates("Return Updated", "Your return request has a new status").
//	    Build()
func NewCustomRule(collection string, monitoredFields []string) *customRuleBuilder {
	return &customRuleBuilder{
		rule: WatchRule{
			Collection:      collection,
			MonitoredFields: monitoredFields,
			EventType:       collection + "_updated",
		},
	}
}

// customRuleBuilder provides a fluent API for constructing a WatchRule.
type customRuleBuilder struct {
	rule WatchRule
}

func (b *customRuleBuilder) NotifyUser(userIDField string) *customRuleBuilder {
	b.rule.NotifyUser = true
	b.rule.UserIDField = userIDField
	return b
}

func (b *customRuleBuilder) NotifySeller(sellerIDField string) *customRuleBuilder {
	b.rule.NotifySeller = true
	b.rule.SellerIDField = sellerIDField
	return b
}

func (b *customRuleBuilder) NotifyBoth(userIDField, sellerIDField string) *customRuleBuilder {
	b.rule.NotifyUser = true
	b.rule.UserIDField = userIDField
	b.rule.NotifySeller = true
	b.rule.SellerIDField = sellerIDField
	return b
}

func (b *customRuleBuilder) WithEventType(et string) *customRuleBuilder {
	b.rule.EventType = et
	return b
}

func (b *customRuleBuilder) WithTemplates(title, body string) *customRuleBuilder {
	b.rule.TitleTemplate = title
	b.rule.BodyTemplate = body
	return b
}

func (b *customRuleBuilder) WithMessageBuilder(fn MessageBuilder) *customRuleBuilder {
	b.rule.MessageBuilder = fn
	return b
}

func (b *customRuleBuilder) Build() WatchRule {
	return b.rule
}

// ---------------------------------------------------------------------------
// Built-in message builders
// ---------------------------------------------------------------------------

func orderMessageBuilder(doc map[string]interface{}, changes []WatchFieldChange) (title, body string) {
	status, _ := doc["status"].(string)
	orderID := resolveID(doc, "orderId", "id", "order_id")

	switch status {
	case "confirmed":
		return "Order Confirmed!", fmt.Sprintf("Order #%s has been confirmed.", orderID)
	case "packed":
		return "Order Packed", fmt.Sprintf("Order #%s is packed and ready for pickup.", orderID)
	case "shipped":
		return "Order Shipped!", fmt.Sprintf("Order #%s is on its way.", orderID)
	case "out_for_delivery":
		return "Out for Delivery", fmt.Sprintf("Order #%s will be delivered today.", orderID)
	case "delivered":
		return "Order Delivered!", fmt.Sprintf("Order #%s has been delivered. Rate your experience.", orderID)
	case "cancelled":
		reason, _ := doc["cancellationReason"].(string)
		if reason != "" {
			return "Order Cancelled", fmt.Sprintf("Order #%s was cancelled: %s", orderID, reason)
		}
		return "Order Cancelled", fmt.Sprintf("Order #%s has been cancelled.", orderID)
	case "return_requested":
		return "Return Requested", fmt.Sprintf("A return has been requested for Order #%s.", orderID)
	case "refunded":
		return "Refund Processed", fmt.Sprintf("Your refund for Order #%s has been processed.", orderID)
	default:
		return "Order Updated", fmt.Sprintf("Order #%s status: %s", orderID, status)
	}
}

func productMessageBuilder(doc map[string]interface{}, changes []WatchFieldChange) (title, body string) {
	productName, _ := doc["name"].(string)
	if productName == "" {
		productName = "Your product"
	}

	for _, c := range changes {
		switch c.Field {
		case "approvalStatus":
			status, _ := c.NewValue.(string)
			switch status {
			case "approved":
				return "Product Approved!", fmt.Sprintf("'%s' is now live on the marketplace.", productName)
			case "rejected":
				return "Product Rejected", fmt.Sprintf("'%s' was not approved. Please review and resubmit.", productName)
			}
		case "price":
			return "Price Updated", fmt.Sprintf("Price for '%s' has been updated.", productName)
		case "stockCount":
			count, _ := c.NewValue.(int64)
			if count == 0 {
				return "Out of Stock", fmt.Sprintf("'%s' is now out of stock.", productName)
			}
			return "Stock Updated", fmt.Sprintf("Stock for '%s' has been updated.", productName)
		}
	}
	return "Product Updated", fmt.Sprintf("'%s' has been updated.", productName)
}

func shopMessageBuilder(doc map[string]interface{}, changes []WatchFieldChange) (title, body string) {
	shopName, _ := doc["name"].(string)
	if shopName == "" {
		shopName = "Your shop"
	}

	for _, c := range changes {
		switch c.Field {
		case "verificationStatus":
			status, _ := c.NewValue.(string)
			switch status {
			case "verified":
				return "Shop Verified!", fmt.Sprintf("'%s' has been verified and is now active.", shopName)
			case "rejected":
				return "Verification Failed", fmt.Sprintf("'%s' could not be verified. Please check your documents.", shopName)
			case "pending":
				return "Verification Pending", fmt.Sprintf("'%s' verification is under review.", shopName)
			}
		case "isActive":
			active, _ := c.NewValue.(bool)
			if active {
				return "Shop Activated", fmt.Sprintf("'%s' is now active.", shopName)
			}
			return "Shop Deactivated", fmt.Sprintf("'%s' has been deactivated.", shopName)
		}
	}
	return "Shop Updated", fmt.Sprintf("'%s' has been updated.", shopName)
}

// enquiryMessageBuilder produces condition-based notification text for every
// monitored field in the "enquiry" collection.
//
// Priority order: the first matched field change wins (most important first).
func enquiryMessageBuilder(doc map[string]interface{}, changes []WatchFieldChange) (title, body string) {
	enquiryID, _ := doc["enquiryId"].(string)
	if enquiryID == "" {
		enquiryID = resolveID(doc, "enquiryId", "id")
	}

	// enquiryRef / enquiryRefCap are used inline so that when no document ID is
	// available the copy still reads naturally without a bare "#".
	enquiryRef := "your enquiry"
	enquiryRefCap := "Your enquiry"
	// if enquiryID != "" {
	// 	enquiryRef = "enquiry #" + enquiryID
	// 	enquiryRefCap = "Enquiry #" + enquiryID
	// }

	// Helper — safely extract a string value from a change.
	strVal := func(v interface{}) string {
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}

	// Helper — format a numeric price coming as int64, float64, or string.
	formatPrice := func(v interface{}) string {
		switch p := v.(type) {
		case string:
			return p
		case int64:
			return fmt.Sprintf("%d", p)
		case float64:
			if p == float64(int64(p)) {
				return fmt.Sprintf("%.0f", p)
			}
			return fmt.Sprintf("%.2f", p)
		}
		return ""
	}

	// Evaluate fields in descending priority.
	for _, c := range changes {
		switch c.Field {

		// ── Overall status ─────────────────────────────────────────────────────
		case "status":
			switch strVal(c.NewValue) {
			case "completed_accepted":
				price := formatPrice(doc["acceptedPrice"])
				if price != "" {
					return "Deal Confirmed! 🎉",
						fmt.Sprintf("%s has been successfully finalised at ₹%s. Thank you for choosing us!", enquiryRefCap, price)
				}
				return "Deal Confirmed! 🎉",
					fmt.Sprintf("Great news — %s has been accepted and completed successfully.", enquiryRef)
			case "completed_rejected":
				return "Enquiry Closed",
					fmt.Sprintf("%s has been closed without a confirmed deal. We hope to match you better next time.", enquiryRefCap)
			case "pending_customer_final", "pending_seller_final", "pending_customer_price":
				return "Negotiation Underway 🤝",
					fmt.Sprintf("%s has entered the negotiation stage. Stay tuned — we'll notify you of every update.", enquiryRefCap)
			case "seller_final_update":
				return "Enquiry Under Review ⏳",
					fmt.Sprintf("%s is currently being reviewed by the seller. Expect a response shortly.", enquiryRefCap)
			case "resolved":
				return "Enquiry Resolved ✅",
					fmt.Sprintf("%s has been resolved. Thank you for your patience throughout the process.", enquiryRefCap)
			case "cancelled":
				return "Enquiry Cancelled",
					fmt.Sprintf("%s has been cancelled. If this was unexpected, please contact our support team for assistance.", enquiryRefCap)
			default:
				if strVal(c.NewValue) != "" {
					return "New Enquiry Received 🛍️",
						fmt.Sprintf("A customer is interested in this product and is asking about its availability. Open the app to review and respond promptly.")
				}
			}

		// ── Negotiation outcome ────────────────────────────────────────────────
		case "finalStatus":
			switch strVal(c.NewValue) {
			case "completed_accepted":
				price := formatPrice(doc["acceptedPrice"])
				if price != "" {
					return "Offer Accepted! ✅",
						fmt.Sprintf("Congratulations! The offer of ₹%s on %s has been accepted. Proceed to confirm your order.", price, enquiryRef)
				}
				return "Offer Accepted! ✅",
					fmt.Sprintf("Congratulations! Your offer on %s has been accepted. Please proceed to finalise the order.", enquiryRef)
			case "completed_rejected":
				return "Offer Not Accepted",
					fmt.Sprintf("Your offer on %s was not accepted this time. You may revise your offer or explore other available options.", enquiryRef)
			case "counter":
				return "Counter Offer Received 💬",
					fmt.Sprintf("A counter offer has been made on %s. Review it now and respond to keep the negotiation going.", enquiryRef)
			default:
				if strVal(c.NewValue) != "" {
					return "Negotiation Update",
						fmt.Sprintf("The negotiation on %s has been updated. Open the app to review the latest terms.", enquiryRef)
				}
			}

		// ── Deal finalisation ──────────────────────────────────────────────────
		case "acceptedBy":
			who := strVal(c.NewValue)
			price := formatPrice(doc["acceptedPrice"])
			displayWho := who
			if who == "seller" {
				displayWho = "Seller"
			} else if who == "customer" {
				displayWho = "Buyer"
			}
			if price != "" {
				return "Deal Finalised ✅",
					fmt.Sprintf("%s has been finalised by the %s at ₹%s. All parties have agreed — it's a deal!", enquiryRefCap, displayWho, price)
			}
			return "Deal Finalised ✅",
				fmt.Sprintf("%s has been confirmed by the %s. The deal is now complete.", enquiryRefCap, displayWho)

		case "acceptedPrice":
			price := formatPrice(c.NewValue)
			if price != "" {
				return "Price Agreed ✅",
					fmt.Sprintf("Both parties have agreed on ₹%s for %s. The deal is nearly done — confirm to proceed.", price, enquiryRef)
			}

		// ── Buyer price moves ──────────────────────────────────────────────────
		case "customerFinalResponse":
			price := formatPrice(c.NewValue)
			if price != "" {
				return "Buyer's Final Offer Submitted",
					fmt.Sprintf("The buyer has submitted a final offer of ₹%s on %s. This is their best price — respond now.", price, enquiryRef)
			}
			return "Buyer Responded",
				fmt.Sprintf("The buyer has submitted their final response on %s. Open the app to review and reply.", enquiryRef)

		case "customerNegotiatedPrice":
			price := formatPrice(c.NewValue)
			if price != "" {
				return "New Offer from Buyer 💬",
					fmt.Sprintf("The buyer has placed a new offer of ₹%s on %s. Review it and send your response.", price, enquiryRef)
			}
			return "Buyer Updated Their Offer",
				fmt.Sprintf("The buyer has revised their offer on %s. Check the latest price and keep negotiating.", enquiryRef)

		// ── Seller price moves ─────────────────────────────────────────────────
		case "sellerFinalPrice":
			price := formatPrice(c.NewValue)
			if price != "" {
				return "Seller's Final Price 🏷️",
					fmt.Sprintf("The seller's final asking price for %s is ₹%s. This is their best offer — accept or make a counter offer.", enquiryRef, price)
			}
			return "Seller Has Updated Their Price",
				fmt.Sprintf("The seller has revised their final price on %s. Open the app to review the latest offer.", enquiryRef)

		case "sellerInitialPrice":
			price := formatPrice(c.NewValue)
			if price != "" {
				return "Seller Responded 🏪",
					fmt.Sprintf("The seller has quoted ₹%s for %s. You can accept this price or start negotiating.", price, enquiryRef)
			}
			return "Seller Has Responded",
				fmt.Sprintf("The seller has replied to %s. Open the app to see their offer and respond.", enquiryRef)

		// ── Product availability ───────────────────────────────────────────────
		case "availability":
			switch strVal(c.NewValue) {
			case "available":
				return "Good News — Product Available! 🎉",
					fmt.Sprintf("The product in %s is now available. Act quickly and lock in your deal before it runs out.", enquiryRef)
			case "unavailable", "out_of_stock":
				return "Product Currently Unavailable",
					fmt.Sprintf("We're sorry — the product in %s is temporarily unavailable. We'll notify you as soon as stock is restored.", enquiryRef)
			default:
				if strVal(c.NewValue) != "" {
					return "Availability Update",
						fmt.Sprintf("The availability status for %s has changed. Open the app for the latest details.", enquiryRef)
				}
			}
		}
	}

	return "Enquiry Update",
		fmt.Sprintf("There's a new activity on %s. Open the app to stay up to date.", enquiryRef)
}
