package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/polyagent/go-services/internal/ai"
	"github.com/polyagent/go-services/internal/config"
	"github.com/polyagent/go-services/internal/scheduler"
	"github.com/polyagent/go-services/internal/storage"
)

func main() {
	log.Println("Starting PolyAgent Task Scheduler...")

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
	
	// 启动调度器
	if err := taskScheduler.Start(); err != nil {
		log.Fatalf("Failed to start task scheduler: %v", err)
	}

	log.Println("Task scheduler started successfully")

	// 等待中断信号来优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down task scheduler...")

	// 创建一个带超时的上下文来优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = ctx // 避免unused变量警告

	// 停止调度器
	taskScheduler.Stop()

	log.Println("Task scheduler stopped")
}