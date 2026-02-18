package main

import (
	"net/http"

	_ "github.com/joho/godotenv/autoload"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/imidll/selfshop.ms-catalog/internal/entry/httpentry"
	"github.com/imidll/selfshop.ms-catalog/internal/platform/config"
	"github.com/imidll/selfshop.ms-catalog/internal/platform/logger"
)

func main() {
	fx.New(
		fx.Provide(func( /* params */ ) *config.T { return config.MustNew(data) }),
		fx.Provide(func(conf *config.T) *logger.T {
			return logger.MustNew(conf).With(zap.String("version", version), zap.String("comhash", comhash), zap.String("buildAt", buildAt))
		}),
		fx.WithLogger(logger.Transform),
		fx.Provide(httpentry.NewRouter),
		fx.Provide(httpentry.NewServer), fx.Invoke(func(*http.Server) {}),
	).Run()
}
