package auth

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/go-ozzo/ozzo-routing/v2/auth"
	"github.com/qiangxue/go-rest-api/internal/entity"
)

// Handler returns a JWT-based authentication middleware.
func Handler(verificationKey string, service Service) routing.Handler {
	return auth.JWT(verificationKey, auth.JWTOptions{
		TokenHandler: func(c *routing.Context, token *jwt.Token) error {
			return handleToken(c, token, service)
		},
	})
}

// handleToken stores the user identity in the request context so that it can be accessed elsewhere.
func handleToken(c *routing.Context, token *jwt.Token, service Service) error {
	ctx := c.Request.Context()
	userID := token.Claims.(jwt.MapClaims)["id"].(string)
	user, err := service.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	ctx = WithUser(ctx, user)
	c.Request = c.Request.WithContext(ctx)
	return nil
}

type contextKey int

const (
	userKey contextKey = iota
)

// WithUser returns a context that contains the user identity from the given JWT.
func WithUser(ctx context.Context, user entity.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// CurrentUser returns the user identity from the given context.
// Nil is returned if no user identity is found in the context.
func CurrentUser(ctx context.Context) *entity.User {
	if user, ok := ctx.Value(userKey).(entity.User); ok {
		return &user
	}
	return nil
}
