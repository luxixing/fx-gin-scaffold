package main

import (
	"os"

	_ "github.com/luxixing/fx-gin-scaffold/docs/swagger" // swagger docs
	"github.com/luxixing/fx-gin-scaffold/internal/bootstrap"
	"go.uber.org/fx"
)

func main() {
	// Build FX options
	options := []fx.Option{
		bootstrap.GetModule(),
		fx.Invoke(bootstrap.RegisterHooks),
	}

	// Disable FX logs only in production
	env := os.Getenv("APP_ENV")
	if env == "production" {
		options = append([]fx.Option{fx.NopLogger}, options...)
	}

	app := fx.New(options...)
	app.Run()
}
