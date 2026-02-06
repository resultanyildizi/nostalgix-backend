package auth

import (
	"net/http"

	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(rg *routing.RouteGroup, service Service, authHandler routing.Handler, logger log.Logger) {
	r := resource{
		service: service,
		logger:  logger,
	}

	rg.Post("/auth/login/username", r.loginUsername)
	rg.Post("/auth/login/anonymous", r.loginAnonymous)
	rg.Post("/auth/refresh", r.refreshTokens)

	rg.Use(authHandler)
	rg.Get("/auth/user", r.getUser)
	rg.Post("/auth/logout", r.logout)
}

type resource struct {
	service Service
	logger  log.Logger
}

// loginUsername returns a handler that handles user loginUsername request.
func (r resource) loginUsername(c *routing.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Read(&req); err != nil {
		r.logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
		return errors.BadRequest("", "")
	}
	authTokens, err := r.service.LoginUsername(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		return err
	}
	return c.WriteWithStatus(authTokens, http.StatusOK)
}

func (r resource) loginAnonymous(c *routing.Context) error {
	var req struct {
		DeviceKey string `json:"device_key" validate:"required"`
	}

	if err := c.Read(&req); err != nil {
		r.logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
		return errors.BadRequest("", "")
	}

	if req.DeviceKey == "" {
		r.logger.With(c.Request.Context()).Errorf("invalid request")
		return errors.BadRequest("Device key is required", "")
	}

	authTokens, err := r.service.LoginAnonymous(c.Request.Context(), req.DeviceKey)
	if err != nil {
		return err
	}
	return c.WriteWithStatus(authTokens, http.StatusOK)
}

func (r resource) refreshTokens(c *routing.Context) error {
	var req struct {
		DeviceKey    string `json:"device_key"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Read(&req); err != nil {
		r.logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
		return errors.BadRequest("", "")
	}

	if req.DeviceKey == "" || req.RefreshToken == "" {
		r.logger.With(c.Request.Context()).Errorf("invalid request")
		return errors.BadRequest("Device key is required", "")
	}

	authTokens, err := r.service.RefreshTokens(c.Request.Context(), req.RefreshToken, req.DeviceKey)
	if err != nil {
		return err
	}
	return c.WriteWithStatus(authTokens, http.StatusOK)
}

func (r resource) getUser(c *routing.Context) error {
	ctx := c.Request.Context()
	user, err := r.service.GetUser(ctx, CurrentUser(ctx).GetID())

	if err != nil {
		return err
	}

	return c.WriteWithStatus(user, http.StatusOK)
}
func (r resource) logout(c *routing.Context) error {
	ctx := c.Request.Context()

	var req struct {
		DeviceKey string `json:"device_key"`
	}
	if err := c.Read(&req); err != nil {
		r.logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
		return errors.BadRequest("", "")
	}

	if req.DeviceKey == "" {
		r.logger.With(c.Request.Context()).Errorf("invalid request")
		return errors.BadRequest("Device key is required", "")
	}

	err := r.service.Logout(ctx, req.DeviceKey)

	if err != nil {
		return err
	}

	return c.WriteWithStatus("success", http.StatusOK)
}
