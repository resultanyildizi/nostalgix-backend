package auth

import (
	"context"
	"database/sql"
	"time"

	stderr "errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// Service encapsulates the authentication logic.
type Service interface {
	// authenticate authenticates a user using username and password.
	// It returns a JWT token if authentication succeeds. Otherwise, an error is returned.
	LoginUsername(ctx context.Context, username, password string) (entity.AuthTokens, error)
	LoginAnonymous(ctx context.Context, deviceKey string) (entity.AuthTokens, error)
	RefreshTokens(ctx context.Context, refreshToken, deviceKey string) (entity.AuthTokens, error)
}

// Identity represents an authenticated user identity.
type Identity interface {
	// GetID returns the user ID.
	GetID() string
	// GetName returns the user name.
	GetName() string
}

type service struct {
	signingKey      string
	tokenExpiration int
	repo            Repository
	logger          log.Logger
}

// NewService creates a new authentication service.
func NewService(signingKey string, tokenExpiration int, repository Repository, logger log.Logger) Service {
	return service{signingKey, tokenExpiration, repository, logger}
}

// Login authenticates a user and generates a JWT token if authentication succeeds.
// Otherwise, an error is returned.
func (s service) LoginUsername(ctx context.Context, username, password string) (entity.AuthTokens, error) {
	logger := s.logger.With(ctx, "user", username)
	var user entity.User
	if username == "demo" && password == "pass" {
		logger.Infof("authentication successful")
		user = entity.User{ID: "100", Name: "demo"}
	}

	accessToken, err := s.generateJWT(user)

	if err != nil {
		return entity.AuthTokens{}, err
	}
	return entity.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: "",
	}, nil
}

func (s service) LoginAnonymous(ctx context.Context, deviceKey string) (entity.AuthTokens, error) {
	var authTokens entity.AuthTokens
	// check if there is a user with the device key
	user, err := s.repo.GetUserByDeviceKey(ctx, deviceKey)

	if err != nil && stderr.Is(err, sql.ErrNoRows) {
		user, err = s.repo.CreateAnonymousUser(ctx, deviceKey)
		if err != nil {
			s.logger.Errorf("There is an error while getting the user by device key %s %v", deviceKey, err)
			return authTokens, errors.InternalServerError("")
		}
	} else if err != nil {
		s.logger.Errorf("There is an error while getting the user by device key %s %v", deviceKey, err)
		return authTokens, errors.InternalServerError("")
	}

	return s.createAuthTokens(ctx, user, deviceKey)
}

func (s service) RefreshTokens(ctx context.Context, refreshToken, deviceKey string) (entity.AuthTokens, error) {
	var authTokens entity.AuthTokens
	refreshTokenHashed, err := s.hashToken(refreshToken)

	if err != nil {
		return authTokens, err
	}

	userID, err := s.repo.ValidateRefreshToken(ctx, deviceKey, refreshTokenHashed)

	if stderr.Is(err, sql.ErrNoRows) {
		return authTokens, errors.Unauthorized("")
	} else if err != nil {
		return authTokens, errors.InternalServerError("")
	}

	user, err := s.repo.GetUserByUserID(ctx, userID)

	return s.createAuthTokens(ctx, user, deviceKey)
}

func (s service) createAuthTokens(ctx context.Context, user entity.User, deviceKey string) (entity.AuthTokens, error) {
	var authTokens entity.AuthTokens

	accessToken, err := s.generateJWT(user)
	if err != nil {
		return authTokens, errors.Unauthorized("")
	}

	refreshToken := uuid.New().String()
	refreshTokenHashed, err := s.hashToken(refreshToken)

	err = s.repo.CreateNewRefreshToken(ctx, deviceKey, user.ID, refreshTokenHashed)

	if err != nil {
		return authTokens, err
	}

	authTokens.AccessToken = accessToken
	authTokens.RefreshToken = refreshToken

	return authTokens, nil
}

func (s service) hashToken(token string) (string, error) {
	return jwt.New(jwt.SigningMethodHS256).SignedString([]byte(token))

}

// generateJWT generates a JWT that encodes an identity.
func (s service) generateJWT(identity Identity) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   identity.GetID(),
		"name": identity.GetName(),
		"exp":  time.Now().Add(time.Duration(s.tokenExpiration) * time.Minute).Unix(),
	}).SignedString([]byte(s.signingKey))
}
