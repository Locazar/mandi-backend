package repository

// ShopActivePredicate is a self-contained SQL fragment (no bind parameters)
// that restricts rows to products/shops whose owning shop is fully approved
// and live. shop_status is the single source of truth for the seller
// onboarding -> admin review -> customer visibility lifecycle
// (under_review -> active | rejected; active <-> suspended) — a shop in any
// other state must never surface to customers. References the `sd.shop_status`
// column, so it can be spliced verbatim into any query that exposes a
// `shop_details sd` alias: ... WHERE <existing> AND <predicate>.
//
// Unlike the subscription gate this is not feature-flagged: visibility must
// always follow shop_status.
const ShopActivePredicate = `sd.shop_status = 'active'`
