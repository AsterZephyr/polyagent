package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/polyagent/go-services/internal/config"
	"github.com/polyagent/go-services/internal/storage"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
		} else {
			log.Printf("Panic recovered: %v", recovered)
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		c.Abort()
	})
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 允许的来源列表
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://polyagent.dev",
		}
		
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Auth JWT认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenString := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		// 验证JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 确认签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.AppConfig.Auth.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// 检查token是否有效
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 检查过期时间
			if exp, ok := claims["exp"].(float64); ok {
				if int64(exp) < time.Now().Unix() {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "Token expired",
						"code":  "TOKEN_EXPIRED",
					})
					c.Abort()
					return
				}
			}

			// 设置用户信息到context
			if userID, ok := claims["user_id"].(string); ok {
				c.Set("user_id", userID)
			}
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}

			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
				"code":  "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}
	}
}

// RateLimiter 限流中间件
func RateLimiter(redis *storage.RedisStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端标识
		identifier := getClientIdentifier(c)
		
		// 检查限流
		allowed, err := redis.CheckRateLimit(identifier, 100, time.Minute) // 每分钟100次请求
		if err != nil {
			log.Printf("Rate limit check error: %v", err)
			c.Next() // 出错时不限流
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIdentifier 获取客户端标识
func getClientIdentifier(c *gin.Context) string {
	// 优先使用用户ID
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%s", userID)
	}

	// 使用IP地址
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// RequestSizeLimit 请求大小限制中间件
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
				"code":  "BODY_TOO_LARGE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 使用通道来处理超时
		done := make(chan bool)
		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
				"code":  "REQUEST_TIMEOUT",
			})
			c.Abort()
		}
	}
}

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}

// APIKey API密钥认证中间件
func APIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"code":  "MISSING_API_KEY",
			})
			c.Abort()
			return
		}

		// 这里应该验证API密钥的有效性
		// 可以从数据库或配置文件中查询
		if !isValidAPIKey(apiKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isValidAPIKey 验证API密钥
func isValidAPIKey(apiKey string) bool {
	// 这里实现API密钥验证逻辑
	// 可以连接数据库查询或使用其他验证方式
	return true // 简化实现
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 使用时间戳和随机数生成请求ID
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}