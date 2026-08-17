package ctrl

import (
	"context"
	"errors"
	"fmt"

	"app/internal/core/generic/middleware"
	"app/internal/core/graph/model"
	"app/internal/pkg/auth/svc"
)

type AuthController struct {
	authService *svc.AuthService
}

func NewAuthController(service *svc.AuthService) *AuthController {
	return &AuthController{authService: service}
}

// GetAppToken exchanges the hub token (validated by @hubAuth, present in
// context) for this app's own token.
func (c *AuthController) GetAppToken(ctx context.Context) (*model.AppToken, error) {
	hubToken, ok := ctx.Value(middleware.HubTokenContextKey).(string)
	if !ok || hubToken == "" {
		return nil, fmt.Errorf("hub token not found in context")
	}

	namespaceSlug, ok := ctx.Value(middleware.NamespaceContextKey).(string)
	if !ok || namespaceSlug == "" {
		return nil, fmt.Errorf("namespace not found in context")
	}

	token, err := c.authService.Auth(ctx, hubToken, namespaceSlug)
	if err != nil {
		return nil, err
	}

	return &model.AppToken{Token: token}, nil
}

func (c *AuthController) ValidateToken(ctx context.Context, sampleToken string) (bool, string, string, error) {
	return c.authService.ValidateToken(ctx, sampleToken)
}

func (c *AuthController) MustValidateToken(ctx context.Context, sampleToken string) (bool, string, string, error) {
	valid, userID, namespace, err := c.ValidateToken(ctx, sampleToken)
	if err != nil {
		return false, "", "", err
	}
	if !valid {
		return false, "", "", errors.New("invalid token")
	}
	return true, userID, namespace, nil
}
