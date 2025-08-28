package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/go-services/gateway/handlers"
	"github.com/polyagent/go-services/gateway/middleware"
	"github.com/polyagent/go-services/internal/ai"
	"github.com/polyagent/go-services/internal/config"
	"github.com/polyagent/go-services/internal/scheduler"
	"github.com/polyagent/go-services/internal/storage"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化存储
	postgres, err := storage.NewPostgresStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	defer postgres.Close()

	redis, err := storage.NewRedisStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redis.Close()

	// 初始化Python AI客户端
	aiClient := ai.NewPythonAIClient(cfg)

	// 初始化任务调度器
	taskScheduler := scheduler.NewTaskScheduler(postgres, redis, aiClient)
	if err := taskScheduler.Start(); err != nil {
		log.Fatalf("Failed to start task scheduler: %v", err)
	}
	defer taskScheduler.Stop()

	// 设置Gin模式
	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(redis))

	// 创建处理器
	chatHandler := handlers.NewChatHandler(taskScheduler, postgres, redis, aiClient)
	agentHandler := handlers.NewAgentHandler(postgres, redis)
	documentHandler := handlers.NewDocumentHandler(postgres, aiClient)
	userHandler := handlers.NewUserHandler(postgres, redis)
	healthHandler := handlers.NewHealthHandler(postgres, redis, aiClient)

	// 健康检查路由
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/metrics", healthHandler.Metrics)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 聊天相关
		chat := api.Group("/chat")
		{
			chat.POST("", chatHandler.Chat)
			chat.POST("/stream", chatHandler.StreamChat)
			chat.GET("/sessions", chatHandler.GetSessions)
			chat.GET("/sessions/:session_id/messages", chatHandler.GetSessionMessages)
		}

		// 智能体管理
		agents := api.Group("/agents")
		{
			agents.GET("", agentHandler.ListAgents)
			agents.POST("", agentHandler.CreateAgent)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.PUT("/:id", agentHandler.UpdateAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
		}

		// 文档管理
		documents := api.Group("/documents")
		{
			documents.GET("", documentHandler.ListDocuments)
			documents.POST("/upload", documentHandler.UploadDocument)
			documents.GET("/:id", documentHandler.GetDocument)
			documents.DELETE("/:id", documentHandler.DeleteDocument)
			documents.POST("/index", documentHandler.IndexDocuments)
		}

		// 用户管理
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
			users.GET("/profile", middleware.Auth(), userHandler.GetProfile)
			users.PUT("/profile", middleware.Auth(), userHandler.UpdateProfile)
		}

		// 任务管理
		tasks := api.Group("/tasks")
		tasks.Use(middleware.Auth())
		{
			tasks.GET("", chatHandler.GetTasks)
			tasks.GET("/:id", chatHandler.GetTask)
			tasks.POST("/:id/cancel", chatHandler.CancelTask)
		}

		// 工具管理
		tools := api.Group("/tools")
		{
			tools.GET("", chatHandler.ListTools)
			tools.POST("/execute", middleware.Auth(), chatHandler.ExecuteTool)
		}

		// 统计信息
		stats := api.Group("/stats")
		stats.Use(middleware.Auth())
		{
			stats.GET("/tasks", chatHandler.GetTaskStats)
			stats.GET("/usage", userHandler.GetUsageStats)
		}
	}

	// WebSocket路由
	r.GET("/ws", chatHandler.WebSocketHandler)

	// 静态文件服务（如果需要）
	r.Static("/static", "./static")

	// 启动HTTP服务器
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: int(cfg.Server.MaxBodySize),
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号来优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭，超时时间为30秒
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}