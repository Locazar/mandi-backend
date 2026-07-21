# Products & Shop Details APIs

Base URL (local dev): `http://localhost:3000/api`

All endpoints below require a Bearer token — the `/api` group has
`AuthenticateUser()` middleware applied ([pkg/api/routes/user.go:57](../pkg/api/routes/user.go#L57))
before both the `/products` and `/shop` groups are registered.

```
-H 'Authorization: Bearer <token>'
```

---

## 1. Search / list products

`GET /api/products/search`
(same handler is also mounted at `GET /api/search/`)

Handler: `SearchProducts` — [pkg/api/handler/product.go:1191](../pkg/api/handler/product.go#L1191)

All query params are optional.

| Param | Type | Default | Filters by |
|---|---|---|---|
| `q` | string | — | Search keyword |
| `name` | string | — | Alternate search keyword (used only if `q` is empty) |
| `category_id` | string | — | Category (e.g. `cat_00003`) |
| `department_id` | string | — | Department |
| `brand_id` | string | — | Brand |
| `location_id` | string | — | Location |
| `shop_id` | string | — | A specific shop's products |
| `lat` | float | 0 | Latitude, for geo/radius search |
| `lng` (or `long`) | float | 0 | Longitude, for geo/radius search |
| `radius` | float | 0 | Radius in km around `lat`/`lng` |
| `pincode` | uint | — | Pincode |
| `limit` | int | `20` | Page size |
| `offset` | int | `0` | Pagination offset |

No price-range, sort, in-stock, or rating filters exist on this endpoint today.

Response shape:
```json
{ "products": [ { "product_item_id": "...", "product_name": "...", "..." : "..." } ] }
```
Full field list on each item (`pkg/api/handler/response/product.go:88-111`):
`product_item_id, product_name, category_id, department_id, sub_category_id, category_name, main_category_name, sub_category_image_url, product_item_images, dynamic_fields, description, highlights, offer_products, discount_rate, shop_id, shop_name, created_at, updated_at, view_count, distance_km, stock, is_subscribed`

### curl examples

Plain keyword search:
```bash
curl --location 'http://localhost:3000/api/products/search?q=running%20shoes&limit=20&offset=0' \
--header 'Authorization: Bearer <token>'
```

Filtered by department + category, paginated:
```bash
curl --location 'http://localhost:3000/api/products/search?department_id=dept_00001&category_id=cat_00003&limit=10&offset=20' \
--header 'Authorization: Bearer <token>'
```

Geo/radius search near a location:
```bash
curl --location 'http://localhost:3000/api/products/search?lat=28.6139&lng=77.2090&radius=5&limit=20' \
--header 'Authorization: Bearer <token>'
```

By shop + pincode:
```bash
curl --location 'http://localhost:3000/api/products/search?shop_id=shop_00042&pincode=110001' \
--header 'Authorization: Bearer <token>'
```

All filters combined:
```bash
curl --location 'http://localhost:3000/api/products/search?q=shoes&department_id=dept_00001&category_id=cat_00003&brand_id=brand_00007&location_id=loc_00002&shop_id=shop_00042&lat=28.6139&lng=77.2090&radius=5&pincode=110001&limit=20&offset=0' \
--header 'Authorization: Bearer <token>'
```

---

## 2. Get shop details (single shop)

`GET /api/shop/:shop_id`

Handler: `GetShopByID` — [pkg/api/handler/user.go:817](../pkg/api/handler/user.go#L817)

| Param | Type | Required | Description |
|---|---|---|---|
| `shop_id` | string (path) | yes | Shop to fetch |

No query params. Personalized fields (`is_following`, `user_rating`, etc.) are derived from the caller's auth token, not a query param.

Response shape (`pkg/api/handler/response/shop.go:13-45`):
`shop_id, shop_name, email, phone, address_line1, address_line2, city, state, country, pincode, shop_type, shop_verification_status, shop_image_url, latitude, longitude, follower_count, following_count, like_count, rating_count, review_count, average_rating, is_following, is_liked, is_subscribed, user_rating, user_review, distance_km, is_open, created_at, updated_at, reviews`

### curl example
```bash
curl --location 'http://localhost:3000/api/shop/shop_00042' \
--header 'Authorization: Bearer <token>'
```

---

## 3. Search / list shops (with filters)

`GET /api/shop/search`

Handler: `SearchShopList` — [pkg/api/handler/user.go:574](../pkg/api/handler/user.go#L574)
("Unified endpoint supporting: name search, geolocation, and pincode filtering")

| Param | Type | Default | Filters by |
|---|---|---|---|
| `q` | string | — | Shop name search |
| `lat` | float | 0 | Latitude, for geo/radius search |
| `lng` (or `long`) | float | 0 | Longitude, for geo/radius search |
| `radius` | float | 0 | Radius in km around `lat`/`lng` |
| `pincode` | uint | — | Pincode |
| `department_id` | string | — | Department |
| `category_id` | string | — | Category |
| `limit` | int | `25` | Page size |
| `offset` | int | `0` | Pagination offset |

Response: `{ "shops": [ ...same shop fields as GetShopByID, minus reviews... ] }` — empty array with `204 No Content` if nothing matches.

### curl examples

Name search:
```bash
curl --location 'http://localhost:3000/api/shop/search?q=fresh%20mart&limit=10' \
--header 'Authorization: Bearer <token>'
```

Geo/radius search:
```bash
curl --location 'http://localhost:3000/api/shop/search?lat=28.6139&lng=77.2090&radius=5&limit=25&offset=0' \
--header 'Authorization: Bearer <token>'
```

By pincode + department:
```bash
curl --location 'http://localhost:3000/api/shop/search?pincode=110001&department_id=dept_00001' \
--header 'Authorization: Bearer <token>'
```

All filters combined:
```bash
curl --location 'http://localhost:3000/api/shop/search?q=fresh&lat=28.6139&lng=77.2090&radius=10&pincode=110001&department_id=dept_00001&category_id=cat_00003&limit=25&offset=0' \
--header 'Authorization: Bearer <token>'
```

---

## 4. Get shop social details (bonus, related)

`GET /api/shop/:shop_id/social`

Handler: `GetShopSocialDetails` — [pkg/api/handler/user.go:863](../pkg/api/handler/user.go#L863)
Same `shop_id` path param, returns just the social summary (follower/like/rating counts) used to enrich `GetShopByID`'s response.

```bash
curl --location 'http://localhost:3000/api/shop/shop_00042/social' \
--header 'Authorization: Bearer <token>'
```
