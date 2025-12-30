package auth

import (
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// RegisterHandlers registers handlers for different HTTP requests.
func RegisterHandlers(rg *routing.RouteGroup, service Service, logger log.Logger) {
	rg.Post("/login/username", loginUsername(service, logger))
	rg.Post("/login/anonymous", loginAnonymous(service, logger))
}

// loginUsername returns a handler that handles user loginUsername request.
func loginUsername(service Service, logger log.Logger) routing.Handler {
	return func(c *routing.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.Read(&req); err != nil {
			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		token, err := service.LoginUsername(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			return err
		}
		return c.Write(struct {
			Token string `json:"token"`
		}{token})
	}
}

func loginAnonymous(service Service, logger log.Logger) routing.Handler {
	return func(c *routing.Context) error {
		var req struct {
			DeviceKey string `json:"device_key" validate:"required"`
		}

		if err := c.Read(&req); err != nil {
			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
			return errors.BadRequest("")
		}

		if req.DeviceKey == "" {
			logger.With(c.Request.Context()).Errorf("invalid request")
			return errors.BadRequest("Device key is required ")
		}

		token, err := service.LoginAnonymous(c.Request.Context(), req.DeviceKey)
		if err != nil {
			return err
		}
		return c.Write(struct {
			Token string `json:"token"`
		}{token})
	}
}
