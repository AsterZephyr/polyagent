package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/polyagent/eino-polyagent/internal/ai"
	"github.com/polyagent/eino-polyagent/internal/config"
	"github.com/polyagent/eino-polyagent/internal/orchestration"
	"github.com/polyagent/eino-polyagent/pkg/gateway"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	if cfg.IsProduction() {
		logger.SetLevel(logrus.WarnLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	modelRouter := ai.NewModelRouter(cfg, logger)
	defer modelRouter.Stop()

	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	defaultAgentConfig := &orchestration.AgentConfig{
		ID:            "default",
		Name:          "Default Assistant",
		Type:          orchestration.AgentTypeConversational,
		SystemPrompt:  "You are a helpful AI assistant. Respond clearly and concisely to user queries.",
		Model:         cfg.AI.DefaultRoute,
		Temperature:   0.7,
		MaxTokens:     2000,
		ToolsEnabled:  true,
		MemoryEnabled: true,
		SafetyFilters: []string{},
		Metadata:      map[string]interface{}{},
	}

	_, err = orchestrator.CreateAgent(defaultAgentConfig)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create default agent")
	}

	gatewayService := gateway.NewGatewayService(cfg, orchestrator, modelRouter, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := gatewayService.Start(); err != nil {
			logger.WithError(err).Fatal("Gateway service failed to start")
		}
	}()

	logger.Info("PolyAgent service started successfully")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutdown signal received, gracefully shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		cancel()
	}()

	select {
	case <-done:
		logger.Info("Service shutdown completed")
	case <-shutdownCtx.Done():
		logger.Warn("Service shutdown timed out")
	}
}