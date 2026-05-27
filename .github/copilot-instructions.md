<!-- Copilot / AI agent instructions for mandi-backend -->

Purpose
- Help an AI coding assistant quickly become productive in this repository (Go backend for an e-commerce/marketplace).

Big picture
- This is a Go monorepo implementing a REST API using Gin (`pkg/api/server.go`) with a Clean Architecture layering: `repository` -> `usecase` -> `api/handler` -> `api/server`.
- Entry point: `cmd/api/main.go` loads config (`pkg/config`) and calls the DI initializer `pkg/di.InitializeApi` which wires DB, services and handlers together.
- Dependency injection is implemented with Google Wire: see `pkg/di/wire.go` (wire injector) and generated `pkg/di/wire_gen.go`.
- DB connection and schema seeds/triggers live under `pkg/db` and Docker init scripts in `docker/postgres/initdb/` (e.g. `01-init.sql`).

Key directories & files (actionable)
- `cmd/api/main.go` — program entry. Use `make run` or `go run ./cmd/api/main.go` to start.
- `pkg/api` — HTTP layer and request routing. `server.go` creates Gin engine and registers routes in `pkg/api/routes`.
- `pkg/api/handler` — HTTP handlers; follow existing handler signatures in `handler/*.go` and `handler/interfaces`.
- `pkg/usecase` — application business logic. Use injected repository interfaces from `pkg/repository/interfaces` when adding features.
- `pkg/repository` — database access; return domain models defined in `pkg/domain`.
- `pkg/service` — auxiliary services (token, otp, cloud). Example: `pkg/service/otp` (Twilio implementation) and `pkg/service/token` (JWT).
- `pkg/di` — Wire setup. Modify `wire.go` to add new constructors and run `make wire` to regenerate `wire_gen.go`.
- `pkg/config/config.go` — central configuration loader used at startup.
- `pkg/db/connection.go` — DB initialization used by DI. DB access uses `database/sql` or `gorm` patterns (inspect repository implementations).

Developer workflows / useful commands
- Start locally: `make run` (uses `go run ./cmd/api/main.go`) or `make build && ./build/bin/api`.
- Run tests: `make test` (runs `go test ./... -cover`). For coverage: `make test-coverage`.
- Generate Wire DI bindings: `make wire` (runs `wire` in `pkg/di`). Regenerate `wire_gen.go` after changing `wire.go`.
- Swagger docs: `make swag` then browse `/swagger/index.html` when server runs. To install swag CLI: `make swagger`.
- Dockerized DB: `docker-compose up` or look at `docker/postgres` Dockerfile and `docker-compose.yml` for local Postgres setup and `initdb` SQL files.
- Linting: `golangci-lint run` via `make check`.

Patterns & conventions specific to this repo
- **Clean Architecture layers**: Repository → UseCase → Handler → Server. Repositories return domain models; usecases contain business logic; handlers call usecases and format responses.
- **Constructor naming**: `NewXxxHandler`, `NewXxxUseCase`, `NewXxxRepository` (all referenced in `pkg/di/wire.go`).
- **Handler pattern**: Each domain (auth, user, product, promotion, etc.) has its own handler file with methods matching the interface in `pkg/api/handler/interfaces/xxx.go`. Example: `PromotionHandler` struct with field `promotionUseCase usecaseInterface.PromotionUseCase`.
- **Response wrapping**: Use `response.SuccessResponse(ctx, statusCode, "message", data)` and `response.ErrorResponse(ctx, statusCode, "message", err, data)` for consistent JSON responses with `{success, message, error, data}` structure.
- **Pagination**: Extract via `pagination := request.GetPagination(ctx)` (defaults: limit=25, offset=0); pass to usecase methods accepting `pagination` param.
- **Route grouping**: `pkg/api/routes/user.go` and `pkg/api/routes/admin.go` register routes via handler methods; called from `server.go` which uses `engine.Group("/api")`.
- **Static assets**: `views/*.html` served from root, `uploads/` directory for media. Icon path generation pattern: `fmt.Sprintf("/uploads/promotions/%s/%s.png", categoryName, typeName)` (replace spaces with underscores).
- **Error handling**: Wrap errors with context; use `pkg/utils/error.go` utilities where needed. Pass actual error to `ErrorResponse()` for logging.

Integration points & external deps
- **Postgres**: Configured in `pkg/db/connection.go` (uses `database/sql` patterns); initialized via `docker-compose up` with SQL migrations in `docker/postgres/initdb/01-init.sql`.
- **Authentication**: JWT tokens via `pkg/service/token/token.go`; Google OAuth via `pkg/api/handler/auth_google.go`. Admin and user tokens use separate keys (`ADMIN_AUTH_KEY`, `USER_AUTH_KEY`).
- **Payment gateways**: Razorpay (`pkg/api/handler/payment_razorpay.go`) and Stripe (`pkg/api/handler/payment_stripe.go`) — separate implementations for different payment providers.
- **OTP/SMS**: Twilio integration in `pkg/service/otp/` — implements OTP service interface.
- **Cloud storage**: AWS S3 via `pkg/service/cloud/` — handles uploads for products, profiles, promotions.
- **Search**: Elasticsearch client in `pkg/service/elasticsearch/` for product search optimization.
- **Image processing**: Graphics service in `pkg/service/graphics/` for thumbnail generation and image manipulation.
- **Mocking for tests**: `mockgen` generates mocks from interfaces; stored in `pkg/mock/mockrepo/`, `pkg/mock/mockservice/`, `pkg/mock/mockusecase/`.

When changing DI or adding a constructor
- Add constructor (NewXxx) in the concrete package.
- Add the constructor to `pkg/di/wire.go` provider list and run `make wire` to regenerate `wire_gen.go`.
- If adding a new handler: create handler struct in `pkg/api/handler/xxx.go`, define interface in `pkg/api/handler/interfaces/xxx.go`, add constructor to `wire.go`, then register routes in `pkg/api/routes/` and pass handler to `NewServerHTTP()`.
- If adding a new usecase: create usecase struct in `pkg/usecase/xxx.go`, define interface in `pkg/usecase/interfaces/xxx.go`, add constructor to `wire.go`.

Tests & mocks
- Tests live alongside packages (`*_test.go`). Use `make test` for full run.
- To create mocks for interfaces, use `mockgen` as shown in `makefile:mockgen` targets. Generated mocks are kept in `pkg/mock/...`.

Small examples (copyable)
- Start API locally: `make run` or `go run ./cmd/api/main.go`.
- Regenerate DI after adding providers: `make wire` (then `git add pkg/di/wire_gen.go`).
- Generate swagger: `make swag` and then `make run` to view `/swagger/index.html`.

What to avoid / repository rules
- Do not edit `pkg/di/wire_gen.go` manually — it's generated by Wire.
- Keep handler signatures and interfaces stable; add new methods to interfaces only when also updating `pkg/di` and implementations.

If unclear
- Inspect these files first: `cmd/api/main.go`, `pkg/di/wire.go`, `pkg/api/server.go`, `pkg/db/connection.go`, `makefile`.
- Ask for runtime environment (local Postgres credentials) if you need to run integration tests or start the app locally.

Feedback
- Tell me if you'd like more examples (e.g., add-new-endpoint checklist) or stricter lint/test rules.

---

## ai-service (Go — port 3001)

**What it is:** A lightweight Go microservice that wraps Claude Vision and Embeddings. mandi-backend calls it via `pkg/service/ai/client.go`.

**Architecture:** `cmd/api/main.go` → `pkg/di.InitializeAPI()` → `pkg/api.NewServerHTTP()` → `pkg/api/routes/ai.go` → `pkg/api/handlers/` → `pkg/service/`

**Handler pattern:**
```go
func (h *XHandler) Method(ctx *gin.Context) {
    var req domain.XRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        utils.BadRequestResponse(ctx, "Invalid request format", err)
        return
    }
    result, err := h.service.DoThing(req.Field)
    if err != nil {
        utils.InternalErrorResponse(ctx, "Failed to do thing", err)
        return
    }
    utils.SuccessResponse(ctx, http.StatusOK, "completed", result)
}
```

**Response helpers:** `utils.SuccessResponse`, `utils.BadRequestResponse`, `utils.InternalErrorResponse` — NOT `response.*` (that's mandi-backend).

**Domain models:** All request/response structs go in `pkg/domain/models.go`.

**Known pitfalls:**
- All image inputs must be JPEG — `media_type` is hardcoded to `image/jpeg`. PNG/WebP inputs fail silently at the Claude API.
- Embeddings use a mock (`generateMockEmbedding`). `callEmbeddingAPI` exists but is not wired.
- No auth middleware is implemented despite `AI_SERVICE_SECRET` being defined.

**Adding a new endpoint:**
1. Add request/response structs to `pkg/domain/models.go`.
2. Add service method to `pkg/service/` (call Anthropic SDK or ai-service logic).
3. Add handler method to `pkg/api/handlers/`.
4. Add route in `pkg/api/routes/ai.go`.
5. Update Wire: `pkg/di/wire.go` → `make wire`.

---

## locazar-admin-portal (Next.js 14 — port 4000)

**What it is:** An admin dashboard for managing the Locazar marketplace. Calls mandi-backend's `/api/admin/` routes.

**Key files:**
- `src/lib/api.ts` — typed fetch wrapper. Always use `api.get()`, `api.post()`, etc., never raw `fetch()`.
- `src/lib/auth.tsx` — `AuthContext` providing `{ token, user, login, logout }`. Get token via `const { token } = useAuth()`.
- `src/app/layout.tsx` — wraps all routes in `AuthProvider`.

**Auth flow:**
1. `POST /api/admin/auth/signin/` with `{email, password}`.
2. Token is at `response.data.tokens.access_token`.
3. Token stored in `localStorage` as `admin_token`.
4. All API calls inject `Authorization: Bearer <token>` automatically via `api.ts`.

**Adding a new admin page:**
1. Create `src/app/dashboard/<section>/page.tsx` — no shared components dir, co-locate everything.
2. Use `const { token } = useAuth()` for the auth token.
3. Call `await api.get('/admin/endpoint', token)` or `await api.post(...)`.
4. Use Tailwind with `brand-*` (primary actions, indigo) and `surface-*` (panels, gray).

**Design tokens (tailwind.config.ts):**
- `brand-500` — primary button color (indigo)
- `surface-100/200/300` — card and panel backgrounds (gray scale)
