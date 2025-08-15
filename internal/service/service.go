package service

import (
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"go.uber.org/fx"
)

// GetModule returns the fx.Option for the service module
func GetModule() fx.Option {
	return fx.Options(
		// Provide services
		fx.Provide(
			fx.Annotate(
				NewAuthService,
				fx.As(new(domain.AuthService)),
			),
		),
		fx.Provide(
			fx.Annotate(
				NewUserService,
				fx.As(new(domain.UserService)),
			),
		),
	)
}