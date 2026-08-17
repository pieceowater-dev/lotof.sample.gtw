package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

// SampleAuthDirective protects operations with this app's own token (issued
// by getAppToken, signed with AppBundleSecret). The token is read from the
// SampleAuthorization header (fallback: Authorization) and validated
// locally. On success the namespace and user id are injected into the
// context.
//
// validate is the app's local token verifier (see auth/svc.ValidateToken) --
// passed in rather than imported directly to avoid an import cycle between
// this package and the auth module.
//
// Real domains typically extend this with a role check once they have a
// role model (see lotof.issues.gtw's IssuesAuthDirective for the pattern:
// same shape, plus a `roles []model.SomeRoleEnum` parameter).
func SampleAuthDirective(
	ctx context.Context,
	next graphql.Resolver,
	validate func(ctx context.Context, token string) (bool, string, string, error),
) (any, error) {
	opCtx := graphql.GetOperationContext(ctx)

	raw := ""
	if opCtx != nil {
		raw = strings.TrimSpace(opCtx.Headers.Get("SampleAuthorization"))
		if raw == "" {
			raw = strings.TrimSpace(opCtx.Headers.Get("Authorization"))
		}
	}
	if raw == "" {
		if v, ok := ctx.Value(WSTokenContextKey).(string); ok {
			raw = strings.TrimSpace(v)
		}
	}
	if raw == "" {
		return nil, errors.New("unauthorized: token is missing for key \"SampleAuthorization\"")
	}

	token := strings.TrimPrefix(raw, "Bearer ")

	valid, userID, namespace, err := validate(ctx, token)
	if err != nil || !valid {
		return nil, errors.New("unauthorized: SampleAuthorization token is invalid")
	}
	if namespace == "" {
		return nil, errors.New("unauthorized: namespace not found in token")
	}

	newCtx := context.WithValue(ctx, NamespaceContextKey, namespace)
	if userID != "" {
		newCtx = context.WithValue(newCtx, UserIDContextKey, userID)
	}

	return next(newCtx)
}
