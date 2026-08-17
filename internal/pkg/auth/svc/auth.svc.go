package svc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"app/internal/core/cfg"
	hubgtw "app/internal/core/grpc/generated/lotof.hub.gtw/gtw"

	"github.com/golang-jwt/jwt/v5"
	gossiper "github.com/pieceowater-dev/lotof.lib.gossiper/v2"
)

const tokenTTL = 30 * 24 * time.Hour

// AuthService exchanges a valid hub token for this app's own token and
// validates this app's tokens locally (HS256, AppBundleSecret) -- mirrors
// menu/issues/contacts/atrace. Real domains typically extend Auth to also
// resolve a caller's role via their own core service (see
// lotof.issues.gtw's AuthService for the pattern: same shape, plus a
// staffClient.ResolveRole call before signing the token).
type AuthService struct {
	transport gossiper.Transport
	hubClient hubgtw.GatewayServiceClient
	secret    []byte
}

func NewAuthService() *AuthService {
	factory := gossiper.NewTransportFactory()
	grpcTransport := factory.CreateTransport(
		gossiper.GRPC,
		cfg.Inst().LotofHubGatewayGrpcAddress,
	)

	client, err := grpcTransport.CreateClient(hubgtw.NewGatewayServiceClient)
	if err != nil {
		return &AuthService{
			transport: grpcTransport,
			secret:    []byte(cfg.Inst().AppBundleSecret),
		}
	}

	return &AuthService{
		transport: grpcTransport,
		hubClient: client.(hubgtw.GatewayServiceClient),
		secret:    []byte(cfg.Inst().AppBundleSecret),
	}
}

// Auth validates the hub token against the Hub Gateway and issues this
// app's own token, scoped to the given namespace.
func (s *AuthService) Auth(ctx context.Context, hubToken string, namespaceSlug string) (string, error) {
	if hubToken == "" {
		return "", errors.New("hub token is required")
	}
	if namespaceSlug == "" {
		return "", errors.New("namespace is required")
	}
	if s.hubClient == nil {
		return "", errors.New("hub gateway client not configured")
	}

	resp, err := s.hubClient.IsTokenOK(ctx, &hubgtw.GatewayIsTokenOKRequest{Token: hubToken})
	if err != nil {
		return "", err
	}
	if resp == nil || !resp.GetValid() {
		msg := "hub token is invalid"
		if resp != nil && resp.GetMessage() != "" {
			msg = resp.GetMessage()
		}
		return "", errors.New(msg)
	}
	userID := resp.GetUserId()

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":       userID,
		"namespace": namespaceSlug,
		"iat":       now.Unix(),
		"exp":       now.Add(tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

// ValidateToken verifies this app's token locally (signature + expiry) and
// returns its identity claims.
func (s *AuthService) ValidateToken(_ context.Context, token string) (valid bool, userID string, namespace string, err error) {
	if token == "" {
		return false, "", "", errors.New("token is required")
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || parsed == nil || !parsed.Valid {
		return false, "", "", errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return false, "", "", errors.New("invalid token claims")
	}

	userID, _ = claims["sub"].(string)
	namespace, _ = claims["namespace"].(string)

	return true, userID, namespace, nil
}
