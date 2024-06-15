package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kjone1/imageElevator/config"
	"github.com/Kjone1/imageElevator/docker"
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

	registryConfig := config.RegistryConfig()
	runnerConfig := config.RunnerConfig()

	registryAdapter := docker.NewRegistry(&registryConfig)

	runner := runner.NewRunner(ctx, registryAdapter, &runnerConfig, &registryConfig)
	handler := handler.NewHandler(runner)

	httpServer := serveHttp(server, handler)

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Msgf("Server forced to shutdown: %s", err)
	}

	log.Info().Msg("Server exiting")
}

func serveHttp(ginEngine *gin.Engine, handler *handler.Handler) *http.Server {

	v1 := ginEngine.Group("/v1")
	v1.GET("/ping", handler.Health)
	v1.GET("/sync", handler.Sync)

	port := config.ReadEnvWithDefault("PORT", "8080")
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: ginEngine,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("failed to start server: %s", err)
		}
	}()

	return httpServer
}
