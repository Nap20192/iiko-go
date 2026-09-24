# iiko-go

Go clients for iiko, the restaurant management system.

Two independent clients in one module. Both are **stdlib-only** — importing either adds no transitive dependencies to your binary.

| Package | API | Covers |
|---|---|---|
| `iikoserver` | iikoServer back office, `https://host/resto/api` | 168 documented endpoints across 11 domains: OLAP reports, stock, documents, recipes, staff, cash shifts, nomenclature, pricing, EDI, events, suppliers |
| `iikocloud` | iikoCloud / iikoTransport, `https://api-ru.iiko.services` | 305 methods across 8 domains: orders, deliveries, menu, customers, employees, inventory, finance, webhooks |

The two share nothing on purpose: opposite auth, formats, session models and rate semantics.

## Install

The repository is private, so tell Go not to use the public proxy:

```bash
go env -w GOPRIVATE=github.com/Nap20192/*
go get github.com/Nap20192/iiko-go@latest
```

`git` must be able to reach the repo — SSH key or `gh auth setup-git`.

## iikoServer

```go
srv := iikoserver.New(iikoserver.Config{
    BaseURL:  "https://your-stand.iiko.it/resto", // must end in /resto
    Login:    os.Getenv("IIKO_LOGIN"),
    Password: os.Getenv("IIKO_PASSWORD"),         // plaintext; SHA1-hashed on the wire
})
defer srv.Close(ctx) // releases the licence seat — see below

deps, err := srv.Corporation.ListEntities(ctx, corporation.KindDepartments, false)
card, err := srv.Recipes.ChartPrepared(ctx, day, productID, "")
shifts, err := srv.Cashshifts.ListCashShifts(ctx, from, to, cashshifts.ShiftAny, "", "")
```

Domains hang off the facade: `Corporation`, `Nomenclature`, `Recipes`, `Reports`, `Documents`, `Cashshifts`, `Staff`, `Suppliers`, `Pricing`, `Events`, `EDI`.

### One session is one licence seat

Every successful `/resto/api/auth` occupies an iiko **licence seat**, not a rate-limit slot. A customer with a single-seat API licence can hold exactly one token; a leaked one keeps a cashier locked out of iikoOffice until the ~1h idle timeout.

`Client` holds exactly one session for its lifetime and re-authenticates only on 401. **Always `defer Close(ctx)`.** Do not run two clients against the same credentials.

Requests are also serialized: iiko's documentation requires strictly sequential calls, so `Client` holds a mutex across every request. This is a contract, not a bottleneck to optimise away.

## iikoCloud

```go
cloud, err := iikocloud.New(iikocloud.Config{
    APIKey:       os.Getenv("IIKO_API_KEY"),
    AppID:        os.Getenv("IIKO_APP_ID"),
    ClientSecret: os.Getenv("IIKO_CLIENT_SECRET"),
})

cities, err := cloud.Organizations.GetCities(ctx, gen.CitiesRequest{OrganizationIDs: ids})
```

The token refreshes itself 5 minutes before expiry and retries once on a 401. A 429 surfaces as `*rest.RateLimitError` with `RetryAfter` and is **not** retried automatically — iiko's docs treat repeated identical requests as grounds for blocking the API login.

The v1 report endpoints have no documented response schema, so `reports.RawReport` hands back the XML document rather than guessing at a shape.

## Generated code

`iikoserver/*/types.go` and `iikocloud/gen/` are generated and carry a `DO NOT EDIT` header.

```bash
make gen          # both
make gen-server   # research/dtos.yaml   -> iikoserver/*/types.go
make gen-cloud    # live OpenAPI spec    -> iikocloud/gen/
```

`research/` holds the machine-readable catalog the server client is built from: `endpoints.yaml` (201 records / 168 documented distinct paths) and `dtos.yaml` (186 DTOs, 1354 fields, 71 enums), extracted from iiko's own documentation. A test regenerates everything in memory and compares byte-for-byte, so a hand edit or an unregenerated catalog change fails the build.

`gen-cloud` downloads the spec by URL and verifies `tools/gen/cloud/spec.sha256`; upstream drift fails the build instead of being silently patched over. Five overlay files repair known defects in iiko's published spec (missing `servers` and `securitySchemes`, pseudo-types, CLR-qualified component names).

## Status

**Not verified against a live iiko server.** Every behaviour is tested against fixtures taken from iiko's documentation — 485 tests, 25 packages, race-clean. Field names, enum values and six of the HTTP verbs are inferred from the docs, which are internally inconsistent in places the catalog records. Verify on a demo stand before trusting a value in production.

## Versioning

Semantic import versioning. Tags are `vMAJOR.MINOR.PATCH`; `v0.x` means the API may still move.

## License

MIT
