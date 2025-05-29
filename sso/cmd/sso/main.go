package main

import (
	"log/slog"
	"os"
	"os/signal"
	"protos/sso/internal/app"
	"protos/sso/internal/config"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	// инициализировать приложение (app)
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	// запустить gRPC-сервер приложения (команда запуска сервера тоже была блокирующая, поэтому её теперь нужно выполнять асинхронно)
	go func() {
		application.GRPCServer.MustRun()
	}()

	// Graceful shutdown
	// Создаём канал для передачи информации о сигналах
	stop := make(chan os.Signal, 1)
	// "Слушаем" перечисленные сигналы
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGINT (pkill -2) or SIGTERM (Ждём данных из канала)
	<-stop

	// initiate graceful shutdown
	application.GRPCServer.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
	log.Info("Gracefully stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
