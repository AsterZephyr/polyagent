package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/polyagent/go-services/internal/models"
	"github.com/polyagent/go-services/internal/storage"
)

// TaskScheduler 任务调度器
type TaskScheduler struct {
	postgres    *storage.PostgresStorage
	redis       *storage.RedisStorage
	pythonAI    PythonAIClient
	workers     int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	taskChannel chan *models.AgentTask
}

// PythonAIClient Python AI服务客户端接口
type PythonAIClient interface {
	ExecuteTask(task *models.AgentTask) (*models.AgentResponse, error)
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(postgres *storage.PostgresStorage, redis *storage.RedisStorage, pythonAI PythonAIClient) *TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &TaskScheduler{
		postgres:    postgres,
		redis:       redis,
		pythonAI:    pythonAI,
		workers:     10, // 默认10个工作协程
		ctx:         ctx,
		cancel:      cancel,
		taskChannel: make(chan *models.AgentTask, 100),
	}
}

// Start 启动调度器
func (ts *TaskScheduler) Start() error {
	log.Println("Starting task scheduler...")

	// 启动工作协程
	for i := 0; i < ts.workers; i++ {
		ts.wg.Add(1)
		go ts.worker(i)
	}

	// 启动任务队列监听器
	ts.wg.Add(1)
	go ts.queueListener()

	// 启动定期清理器
	ts.wg.Add(1)
	go ts.cleaner()

	// 启动健康检查器
	ts.wg.Add(1)
	go ts.healthChecker()

	log.Printf("Task scheduler started with %d workers", ts.workers)
	return nil
}

// Stop 停止调度器
func (ts *TaskScheduler) Stop() {
	log.Println("Stopping task scheduler...")
	
	ts.cancel()
	close(ts.taskChannel)
	ts.wg.Wait()
	
	log.Println("Task scheduler stopped")
}

// SubmitTask 提交任务
func (ts *TaskScheduler) SubmitTask(task *models.AgentTask) error {
	// 设置任务初始状态
	task.Status = models.TaskStatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// 保存到数据库
	if err := ts.postgres.CreateTask(task); err != nil {
		return fmt.Errorf("failed to save task to database: %w", err)
	}

	// 加入Redis队列
	if err := ts.redis.EnqueueTask(task); err != nil {
		log.Printf("Warning: failed to enqueue task to Redis: %v", err)
		// 即使Redis失败也继续，任务已保存到数据库
	}

	log.Printf("Task %s submitted successfully", task.TaskID)
	return nil
}

// queueListener 队列监听器
func (ts *TaskScheduler) queueListener() {
	defer ts.wg.Done()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			// 从Redis队列获取任务
			task, err := ts.redis.DequeueTask()
			if err != nil {
				// 如果Redis队列为空，从数据库获取待处理任务
				tasks, dbErr := ts.postgres.GetPendingTasks(10)
				if dbErr != nil {
					log.Printf("Error getting pending tasks from database: %v", dbErr)
					continue
				}

				for _, dbTask := range tasks {
					select {
					case ts.taskChannel <- dbTask:
					case <-ts.ctx.Done():
						return
					default:
						// 通道满了，稍后再试
						log.Printf("Task channel full, skipping task %s", dbTask.TaskID)
					}
				}
				continue
			}

			// 将任务发送给工作协程
			select {
			case ts.taskChannel <- task:
			case <-ts.ctx.Done():
				return
			default:
				// 通道满了，重新入队
				if err := ts.redis.EnqueueTask(task); err != nil {
					log.Printf("Error re-enqueueing task: %v", err)
				}
			}
		}
	}
}

// worker 工作协程
func (ts *TaskScheduler) worker(id int) {
	defer ts.wg.Done()
	
	log.Printf("Worker %d started", id)
	
	for {
		select {
		case <-ts.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		case task, ok := <-ts.taskChannel:
			if !ok {
				log.Printf("Worker %d: task channel closed", id)
				return
			}
			
			ts.processTask(id, task)
		}
	}
}

// processTask 处理任务
func (ts *TaskScheduler) processTask(workerID int, task *models.AgentTask) {
	log.Printf("Worker %d processing task %s", workerID, task.TaskID)
	
	start := time.Now()
	
	// 更新任务状态为运行中
	task.Status = models.TaskStatusRunning
	if err := ts.postgres.UpdateTaskStatus(task.TaskID, models.TaskStatusRunning); err != nil {
		log.Printf("Error updating task status: %v", err)
	}

	// 获取分布式锁，防止重复处理
	lockKey := fmt.Sprintf("task:%s", task.TaskID)
	locked, err := ts.redis.AcquireLock(lockKey, 5*time.Minute)
	if err != nil {
		log.Printf("Error acquiring lock for task %s: %v", task.TaskID, err)
		return
	}
	if !locked {
		log.Printf("Task %s is already being processed", task.TaskID)
		return
	}
	defer func() {
		if err := ts.redis.ReleaseLock(lockKey); err != nil {
			log.Printf("Error releasing lock for task %s: %v", task.TaskID, err)
		}
	}()

	// 执行任务
	response, err := ts.pythonAI.ExecuteTask(task)
	if err != nil {
		log.Printf("Error executing task %s: %v", task.TaskID, err)
		
		// 更新任务状态为失败
		if dbErr := ts.postgres.UpdateTaskStatus(task.TaskID, models.TaskStatusFailed); dbErr != nil {
			log.Printf("Error updating failed task status: %v", dbErr)
		}
		
		// 发布任务失败事件
		ts.publishTaskEvent("task.failed", task.TaskID, map[string]interface{}{
			"error":    err.Error(),
			"duration": time.Since(start),
		})
		return
	}

	// 处理响应
	if response.Status == "success" {
		// 完成任务
		if err := ts.postgres.CompleteTask(task.TaskID); err != nil {
			log.Printf("Error completing task: %v", err)
		}

		// 保存对话记忆
		if response.Memory != nil {
			if err := ts.postgres.SaveConversationMemory(response.Memory); err != nil {
				log.Printf("Error saving conversation memory: %v", err)
			}
		}

		log.Printf("Worker %d completed task %s in %v", workerID, task.TaskID, time.Since(start))
		
		// 发布任务完成事件
		ts.publishTaskEvent("task.completed", task.TaskID, map[string]interface{}{
			"output":   response.Output,
			"duration": time.Since(start),
		})
	} else {
		// 任务执行失败
		if err := ts.postgres.UpdateTaskStatus(task.TaskID, models.TaskStatusFailed); err != nil {
			log.Printf("Error updating failed task status: %v", err)
		}
		
		log.Printf("Worker %d: task %s failed: %s", workerID, task.TaskID, response.Error)
		
		// 发布任务失败事件
		ts.publishTaskEvent("task.failed", task.TaskID, map[string]interface{}{
			"error":    response.Error,
			"duration": time.Since(start),
		})
	}

	// 删除任务缓存
	if err := ts.redis.DeleteTaskCache(task.TaskID); err != nil {
		log.Printf("Error deleting task cache: %v", err)
	}

	// 统计
	ts.updateTaskStats(response.Status == "success")
}

// cleaner 定期清理器
func (ts *TaskScheduler) cleaner() {
	defer ts.wg.Done()
	
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			ts.cleanupExpiredTasks()
		}
	}
}

// cleanupExpiredTasks 清理过期任务
func (ts *TaskScheduler) cleanupExpiredTasks() {
	// 这里可以实现清理逻辑，比如：
	// 1. 清理长时间未完成的任务
	// 2. 清理过期的缓存
	// 3. 清理旧的对话记忆
	log.Println("Running periodic cleanup...")
	
	// 获取队列长度用于监控
	queueLen, err := ts.redis.GetQueueLength()
	if err != nil {
		log.Printf("Error getting queue length: %v", err)
	} else {
		log.Printf("Current queue length: %d", queueLen)
	}
}

// healthChecker 健康检查器
func (ts *TaskScheduler) healthChecker() {
	defer ts.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ts.ctx.Done():
			return
		case <-ticker.C:
			ts.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (ts *TaskScheduler) performHealthCheck() {
	// 检查数据库连接
	// 检查Redis连接
	// 检查Python AI服务连接
	// 检查工作协程状态
	// 这里可以实现具体的健康检查逻辑
}

// publishTaskEvent 发布任务事件
func (ts *TaskScheduler) publishTaskEvent(eventType, taskID string, data map[string]interface{}) {
	event := map[string]interface{}{
		"type":      eventType,
		"task_id":   taskID,
		"timestamp": time.Now(),
		"data":      data,
	}

	if err := ts.redis.PublishMessage("task_events", event); err != nil {
		log.Printf("Error publishing task event: %v", err)
	}
}

// updateTaskStats 更新任务统计
func (ts *TaskScheduler) updateTaskStats(success bool) {
	today := time.Now().Format("2006-01-02")
	
	totalKey := fmt.Sprintf("stats:tasks:total:%s", today)
	if _, err := ts.redis.IncrementCounter(totalKey, 24*time.Hour); err != nil {
		log.Printf("Error updating total task stats: %v", err)
	}

	if success {
		successKey := fmt.Sprintf("stats:tasks:success:%s", today)
		if _, err := ts.redis.IncrementCounter(successKey, 24*time.Hour); err != nil {
			log.Printf("Error updating success task stats: %v", err)
		}
	} else {
		failureKey := fmt.Sprintf("stats:tasks:failure:%s", today)
		if _, err := ts.redis.IncrementCounter(failureKey, 24*time.Hour); err != nil {
			log.Printf("Error updating failure task stats: %v", err)
		}
	}
}

// GetTaskStats 获取任务统计
func (ts *TaskScheduler) GetTaskStats(date string) (map[string]int64, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	stats := make(map[string]int64)
	
	totalKey := fmt.Sprintf("stats:tasks:total:%s", date)
	total, err := ts.redis.GetCounter(totalKey)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	successKey := fmt.Sprintf("stats:tasks:success:%s", date)
	success, err := ts.redis.GetCounter(successKey)
	if err != nil {
		return nil, err
	}
	stats["success"] = success

	failureKey := fmt.Sprintf("stats:tasks:failure:%s", date)
	failure, err := ts.redis.GetCounter(failureKey)
	if err != nil {
		return nil, err
	}
	stats["failure"] = failure

	return stats, nil
}