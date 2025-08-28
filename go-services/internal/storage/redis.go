package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/polyagent/go-services/internal/config"
	"github.com/polyagent/go-services/internal/models"
)

type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStorage 创建 Redis 存储实例
func NewRedisStorage(cfg *config.Config) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConn,
	})

	ctx := context.Background()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisStorage{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// 缓存键前缀
const (
	TaskCachePrefix    = "task:"
	SessionCachePrefix = "session:"
	UserCachePrefix    = "user:"
	AgentCachePrefix   = "agent:"
	ToolCachePrefix    = "tool:"
	QueuePrefix        = "queue:"
)

// Task队列相关

// EnqueueTask 将任务加入队列
func (rs *RedisStorage) EnqueueTask(task *models.AgentTask) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 根据优先级选择队列
	queueName := fmt.Sprintf("%shigh", QueuePrefix)
	if task.Priority < 5 {
		queueName = fmt.Sprintf("%slow", QueuePrefix)
	}

	// 将任务加入队列
	err = rs.client.LPush(rs.ctx, queueName, taskData).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// 缓存任务信息
	taskKey := fmt.Sprintf("%s%s", TaskCachePrefix, task.TaskID)
	err = rs.client.Set(rs.ctx, taskKey, taskData, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache task: %w", err)
	}

	return nil
}

// DequeueTask 从队列中取出任务
func (rs *RedisStorage) DequeueTask() (*models.AgentTask, error) {
	// 优先处理高优先级队列
	queues := []string{
		fmt.Sprintf("%shigh", QueuePrefix),
		fmt.Sprintf("%slow", QueuePrefix),
	}

	for _, queue := range queues {
		result, err := rs.client.BRPop(rs.ctx, time.Second*1, queue).Result()
		if err != nil {
			if err == redis.Nil {
				continue // 队列为空，尝试下一个
			}
			return nil, fmt.Errorf("failed to dequeue task: %w", err)
		}

		if len(result) < 2 {
			continue
		}

		var task models.AgentTask
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task: %w", err)
		}

		return &task, nil
	}

	return nil, fmt.Errorf("no tasks available")
}

// GetQueueLength 获取队列长度
func (rs *RedisStorage) GetQueueLength() (int64, error) {
	highQueue := fmt.Sprintf("%shigh", QueuePrefix)
	lowQueue := fmt.Sprintf("%slow", QueuePrefix)

	highLen := rs.client.LLen(rs.ctx, highQueue).Val()
	lowLen := rs.client.LLen(rs.ctx, lowQueue).Val()

	return highLen + lowLen, nil
}

// Task缓存相关

// CacheTask 缓存任务
func (rs *RedisStorage) CacheTask(task *models.AgentTask, ttl time.Duration) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	key := fmt.Sprintf("%s%s", TaskCachePrefix, task.TaskID)
	err = rs.client.Set(rs.ctx, key, taskData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache task: %w", err)
	}

	return nil
}

// GetCachedTask 获取缓存的任务
func (rs *RedisStorage) GetCachedTask(taskID string) (*models.AgentTask, error) {
	key := fmt.Sprintf("%s%s", TaskCachePrefix, taskID)
	
	data, err := rs.client.Get(rs.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("task not found in cache")
		}
		return nil, fmt.Errorf("failed to get cached task: %w", err)
	}

	var task models.AgentTask
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// DeleteTaskCache 删除任务缓存
func (rs *RedisStorage) DeleteTaskCache(taskID string) error {
	key := fmt.Sprintf("%s%s", TaskCachePrefix, taskID)
	return rs.client.Del(rs.ctx, key).Err()
}

// Session相关

// CacheSession 缓存会话信息
func (rs *RedisStorage) CacheSession(sessionID string, data interface{}, ttl time.Duration) error {
	sessionData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := fmt.Sprintf("%s%s", SessionCachePrefix, sessionID)
	err = rs.client.Set(rs.ctx, key, sessionData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache session: %w", err)
	}

	return nil
}

// GetCachedSession 获取缓存的会话
func (rs *RedisStorage) GetCachedSession(sessionID string, dest interface{}) error {
	key := fmt.Sprintf("%s%s", SessionCachePrefix, sessionID)
	
	data, err := rs.client.Get(rs.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("session not found in cache")
		}
		return fmt.Errorf("failed to get cached session: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return nil
}

// ExtendSessionTTL 延长会话TTL
func (rs *RedisStorage) ExtendSessionTTL(sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf("%s%s", SessionCachePrefix, sessionID)
	return rs.client.Expire(rs.ctx, key, ttl).Err()
}

// Rate Limiting 相关

// CheckRateLimit 检查速率限制
func (rs *RedisStorage) CheckRateLimit(identifier string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	
	// 使用滑动窗口算法
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	pipe := rs.client.Pipeline()
	
	// 移除过期的记录
	pipe.ZRemRangeByScore(rs.ctx, key, "0", fmt.Sprintf("%d", windowStart))
	
	// 添加当前请求
	pipe.ZAdd(rs.ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})
	
	// 计算当前窗口内的请求数
	pipe.ZCard(rs.ctx, key)
	
	// 设置过期时间
	pipe.Expire(rs.ctx, key, window)
	
	results, err := pipe.Exec(rs.ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	
	// 获取请求数
	count := results[2].(*redis.IntCmd).Val()
	
	return count <= int64(limit), nil
}

// 分布式锁相关

// AcquireLock 获取分布式锁
func (rs *RedisStorage) AcquireLock(lockKey string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:%s", lockKey)
	
	result, err := rs.client.SetNX(rs.ctx, key, "locked", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	
	return result, nil
}

// ReleaseLock 释放分布式锁
func (rs *RedisStorage) ReleaseLock(lockKey string) error {
	key := fmt.Sprintf("lock:%s", lockKey)
	return rs.client.Del(rs.ctx, key).Err()
}

// ExtendLock 延长锁时间
func (rs *RedisStorage) ExtendLock(lockKey string, ttl time.Duration) error {
	key := fmt.Sprintf("lock:%s", lockKey)
	return rs.client.Expire(rs.ctx, key, ttl).Err()
}

// 统计相关

// IncrementCounter 增加计数器
func (rs *RedisStorage) IncrementCounter(key string, ttl time.Duration) (int64, error) {
	pipe := rs.client.Pipeline()
	pipe.Incr(rs.ctx, key)
	pipe.Expire(rs.ctx, key, ttl)
	
	results, err := pipe.Exec(rs.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}
	
	return results[0].(*redis.IntCmd).Val(), nil
}

// GetCounter 获取计数器值
func (rs *RedisStorage) GetCounter(key string) (int64, error) {
	result, err := rs.client.Get(rs.ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}
	return result, nil
}

// Pub/Sub相关

// PublishMessage 发布消息
func (rs *RedisStorage) PublishMessage(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	return rs.client.Publish(rs.ctx, channel, data).Err()
}

// SubscribeChannel 订阅频道
func (rs *RedisStorage) SubscribeChannel(channel string) *redis.PubSub {
	return rs.client.Subscribe(rs.ctx, channel)
}

// Close 关闭Redis连接
func (rs *RedisStorage) Close() error {
	return rs.client.Close()
}