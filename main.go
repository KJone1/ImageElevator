package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kjone1/imageElevator/config"
	"github.com/Kjone1/imageElevator/handler"
	"github.com/Kjone1/imageElevator/runner"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	if os.Getenv("GIN_MODE") != "release" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}
	config.LoadConfig()
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGPIPE,
	)

	defer func() {
		log.Debug().Msg("Shutting down gracefully...")
		cancel()
	}()

	server := gin.Default()

	v1 := server.Group("/v1")
	runner := runner.NewRunner(ctx)
	handler := handler.NewHandler(runner)

	v1.GET("/ping", handler.Health)
	v1.GET("/sync", handler.Sync)

	httpServer := serverHttp(server)
	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Msgf("Server forced to shutdown: %s", err)
	}

	log.Info().Msg("Server exiting")
}

func serverHttp(handler http.Handler) *http.Server {

	port := config.ReadEnvWithDefault("PORT", "8080")
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("failed to start server: %s", err)
		}
	}()

	return httpServer
}
