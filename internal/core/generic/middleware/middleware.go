package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

type contextKey string

const NamespaceContextKey contextKey = "namespace-slug"
const HubTokenContextKey contextKey = "hub-token"
const UserIDContextKey contextKey = "user-id"

// WSTokenContextKey carries the token a WebSocket subscription stashed
// during its connection_init handshake (see app.go's Websocket InitFunc) --
// graphql-ws has no real HTTP headers to read SampleAuthorization from, so
// SampleAuthDirective falls back to this when opCtx.Headers comes up empty.
const WSTokenContextKey contextKey = "ws-token"

// HubAuthDirective validates that a hub-issued JWT is present and well-formed
// (structurally — signature verification happens at Hub itself when this
// token is later exchanged via getAppToken) and extracts the namespace/user
// claims into context. Used only on getAppToken.
func HubAuthDirective(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	opCtx := graphql.GetOperationContext(ctx)

	raw := ""
	if opCtx != nil {
		raw = strings.TrimSpace(opCtx.Headers.Get("Authorization"))
		if raw == "" {
			raw = strings.TrimSpace(opCtx.Headers.Get("HubAuthorization"))
		}
	}
	if raw == "" {
		if v, ok := ctx.Value(HubTokenContextKey).(string); ok {
			raw = v
		}
	}
	if raw == "" {
		return nil, errors.New("unauthorized: missing Authorization header")
	}

	token := strings.TrimPrefix(raw, "Bearer ")

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("unauthorized: invalid JWT structure")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unauthorized: failed to decode JWT payload: %w", err)
	}

	claims := make(map[string]any)
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unauthorized: failed to parse JWT claims: %w", err)
	}

	namespace, _ := claims["namespace"].(string)
	if namespace == "" && opCtx != nil {
		namespace = opCtx.Headers.Get("Namespace")
	}
	if namespace == "" {
		return nil, errors.New("unauthorized: namespace not found in token")
	}

	userID, _ := claims["sub"].(string)

	newCtx := context.WithValue(ctx, NamespaceContextKey, namespace)
	newCtx = context.WithValue(newCtx, HubTokenContextKey, token)
	if userID != "" {
		newCtx = context.WithValue(newCtx, UserIDContextKey, userID)
	}

	return next(newCtx)
}
