package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/lib/logger/handlers/slogpretty"
	"syscall"
)

const (
	logLocal = "local"
	logDev   = "dev"
	logProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env, setupPrettySlog)
	log.Info("starting application", slog.Any("config", cfg))

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop

	application.GRPCSrv.Stop()

	log.Info("stopping application", slog.String("signal", sign.String()))
}

func setupLogger(env string, setupFunc func(slog.Leveler) *slog.Logger) *slog.Logger {
	var log *slog.Logger
	switch env {
	case logLocal:
		log = setupFunc(slog.LevelDebug)
	case logDev:
		log = setupFunc(slog.LevelDebug)
	case logProd:
		log = setupFunc(slog.LevelInfo)
	}
	return log
}

func setupPrettySlog(level slog.Leveler) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: level,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
