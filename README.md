# lotof.sample.gtw

Template GraphQL gateway for bootstrapping a new LOTOF domain. It ships with
the full auth/tenant scaffolding every current domain gateway (menu, issues,
contacts, atrace) already runs — hub-token → app-token exchange, namespace
propagation to the backing svc, pooled gRPC connections, and the inbound
`tenants.GatewayService` Hub calls to provision a namespace — plus one worked
example module (`domainItem`) so the wiring is visible end to end.

Pairs with [lotof.sample.proto](https://github.com/pieceowater-dev/lotof.sample.proto)
(the gRPC contracts) and [lotof.sample.svc](https://github.com/pieceowater-dev/lotof.sample.svc)
(the microservice this gateway talks to).

## What's generic vs what's the example

Keep as-is — this is shared auth/tenant plumbing, not domain logic:

```
internal/core/grpcpool/pool.go        # pooled gRPC client to the svc (dilutes an x/net HPACK panic risk under load)
internal/core/generic/middleware/
  middleware.go            # context keys + HubAuthDirective (hub token -> namespace/user claims)
  sampleauth.go            # SampleAuthDirective -- this app's own token, validated locally
  namespace.client.go      # propagates namespace + user-id to the svc over outgoing gRPC metadata
  cors.go                  # CORS + "/api-<service>" path-prefix stripping for ALB ingress
internal/pkg/
  auth/                    # hub token -> this app's own token (HS256, AppBundleSecret)
  gateway/                 # inbound tenants.GatewayService -- Hub calls this to provision a namespace, fans out to the svc's AppTenantsService
```

Delete or rename — this is the disposable example:

```
internal/pkg/domainItem/   # schema + resolvers + svc/ctrl calling the svc's SampleDomainItemService
```

## Bootstrapping a new service

1. Fork this repo as `lotof.<domain>.gtw`.
2. Repoint the proto dependency: `go get github.com/pieceowater-dev/lotof.<domain>.proto@latest`.
3. Rename `internal/pkg/domainItem/` to your first real entity (schema/resolvers/svc/ctrl), update `internal/pkg/router.go` and `internal/pkg/_resolvers/resolver.go` to reference it instead.
4. `internal/pkg/auth` mirrors menu/issues/contacts/atrace but skips role resolution (no role model here). If your domain needs roles, extend `AuthService.Auth` to resolve one via your svc (see `lotof.issues.gtw`'s `AuthService` for the pattern) and add a `roles` argument to `@sampleAuth` (see `lotof.issues.gtw`'s `@issuesAuth`).
5. Update `.env` / `cfg.go` defaults: `APP_BUNDLE_NAME` (must match the svc template's), `LOTOF_SAMPLE_SVC_GRPC_ADDRESS`.
6. `make setup && make generate && make build`.

## Multi-tenancy model

Hub only knows this gateway's address (from `namespace_apps`), not the
backing svc's — so tenant provisioning always routes through here first.
`internal/pkg/gateway` implements the inbound `tenants.GatewayService`
(`AddNamespaceTenant`/`GetNamespaceHealth`) and fans each call out to every
backing microservice's `AppTenantsService` (today just one: this domain's
own svc). Every outbound call to the svc carries the tenant namespace (and
calling user id, for the svc's on-demand tenant provisioning fallback) via
`NamespaceClientInterceptor`/`UserIDClientInterceptor` — see `app.go`'s
pooled client setup.

Auth is two-layered: a Hub-issued JWT (validated structurally by
`@hubAuth`, used only on `getAppToken`) is exchanged for this app's own
namespace-scoped token (HS256, `AppBundleSecret`), which `@sampleAuth`
then validates locally on every other field. WebSocket subscriptions can't
send custom headers, so the token is instead passed via the
`connection_init` payload and stashed in context by `app.go`'s `InitFunc`.

## Prerequisites

- [Go](https://golang.org/doc/install) 1.25+
- [Docker](https://docs.docker.com/get-docker/)
- `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc` (`make setup` installs the Go generators; `protoc` itself must already be on PATH)

## Common tasks

```
make setup     # go get the latest proto package + go mod tidy
make generate  # regenerate gRPC stubs (own proto + Hub's) and gqlgen code
make build     # compile to bin/lotof.sample.gtw
make run       # build + run
make test      # go build ./... (no test suite yet)
```

## Notes

- Customize `Dockerfile` as needed for your project's specific requirements.
- The server entry point is at `./cmd/server/main.go`.
- `internal/core/graph/` (gqlgen output) is committed, unlike `internal/core/grpc/generated/` (gitignored) — `make gql-gen` stages it automatically via `git add -A`.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author
![PCWT Dev Logo](https://avatars.githubusercontent.com/u/168465239?s=50)
### [PCWT Dev](https://github.com/pieceowater-dev)
