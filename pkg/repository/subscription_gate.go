package repository

// SubscribedShopPredicate is a self-contained SQL fragment (no bind parameters)
// that restricts rows to products whose owning shop has an ACTIVE subscription.
// A trial counts as active (trial rows carry is_active=true). It references the
// `sd.admin_id` column, so it can be spliced verbatim into any query that
// exposes a `shop_details sd` alias:  ... WHERE <existing> AND <predicate>.
//
// Sellers subscribe through the same flow as users (migration 000009 dropped the
// users FK), so user_subscriptions.user_id == shop_details.admin_id for sellers.
const SubscribedShopPredicate = `EXISTS (
	SELECT 1 FROM user_subscriptions us
	WHERE us.user_id = sd.admin_id
	  AND us.is_active = true
	  AND us.end_date > NOW()
)`

// IsSubscribedColumn is the SubscribedShopPredicate wrapped as a boolean output
// column, for shop queries that expose is_subscribed without filtering rows.
const IsSubscribedColumn = `CASE WHEN ` + SubscribedShopPredicate + ` THEN true ELSE false END AS is_subscribed`

// subscriptionGateEnabled is the process-wide toggle for customer-facing
// product-visibility gating. It is set once at startup from config
// (SUBSCRIPTION_GATE_ENABLED) and defaults to false so the feature ships dark.
var subscriptionGateEnabled bool

// SetSubscriptionGateEnabled configures the global product-visibility gate.
// Called at startup from config; may be toggled by integration tests.
func SetSubscriptionGateEnabled(enabled bool) { subscriptionGateEnabled = enabled }

// SubscriptionGateEnabled reports whether the product-visibility gate is active.
func SubscriptionGateEnabled() bool { return subscriptionGateEnabled }
