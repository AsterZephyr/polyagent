package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/polyagent/go-services/internal/config"
	"github.com/polyagent/go-services/internal/models"
	"github.com/polyagent/go-services/internal/storage"
)

// UserHandler 用户处理器
type UserHandler struct {
	postgres *storage.PostgresStorage
	redis    *storage.RedisStorage
}

// NewUserHandler 创建用户处理器
func NewUserHandler(postgres *storage.PostgresStorage, redis *storage.RedisStorage) *UserHandler {
	return &UserHandler{
		postgres: postgres,
		redis:    redis,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	// 这里应该实现数据库查询
	// existingUser, err := h.postgres.GetUserByUsername(req.Username)
	// if err == nil {
	//     c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
	//     return
	// }

	// 检查邮箱是否已存在
	// existingUser, err = h.postgres.GetUserByEmail(req.Email)
	// if err == nil {
	//     c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
	//     return
	// }

	// 创建新用户
	user := &models.User{
		ID:       models.NewUserID(),
		Username: req.Username,
		Email:    req.Email,
		Status:   models.UserStatusActive,
		Config:   make(map[string]interface{}),
	}

	// 密码加密（这里应该使用 bcrypt）
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
	//     return
	// }

	// 保存到数据库
	// if err := h.postgres.CreateUser(user); err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
	//     return
	// }

	// 生成JWT token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(config.AppConfig.Auth.TokenExpiry) * time.Second),
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 这里应该实现用户验证逻辑
	// 1. 根据用户名查询用户
	// 2. 验证密码
	// 3. 检查用户状态

	// 简化实现：创建模拟用户
	user := &models.User{
		ID:       models.NewUserID(),
		Username: req.Username,
		Email:    req.Username + "@example.com",
		Status:   models.UserStatusActive,
	}

	// 生成JWT token
	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 更新最后登录时间
	// user.LastLogin = &time.Now()
	// h.postgres.UpdateUser(user)

	c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(config.AppConfig.Auth.TokenExpiry) * time.Second),
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 从数据库获取用户信息
	// user, err := h.postgres.GetUser(userID.(string))
	// if err != nil {
	//     c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	//     return
	// }

	// 简化实现
	c.JSON(http.StatusOK, UserInfo{
		ID:       userID.(string),
		Username: "example_user",
		Email:    "user@example.com",
	})
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Username string                 `json:"username"`
		Email    string                 `json:"email"`
		Config   map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 这里应该实现更新逻辑
	// 1. 验证数据
	// 2. 检查用户名/邮箱是否重复
	// 3. 更新数据库

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user_id": userID,
	})
}

// GetUsageStats 获取使用统计
func (h *UserHandler) GetUsageStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 从Redis获取统计数据
	today := time.Now().Format("2006-01-02")
	
	// 获取今日请求数
	requestKey := "stats:user:requests:" + userID.(string) + ":" + today
	requestCount, _ := h.redis.GetCounter(requestKey)

	// 获取今日消息数
	messageKey := "stats:user:messages:" + userID.(string) + ":" + today
	messageCount, _ := h.redis.GetCounter(messageKey)

	c.JSON(http.StatusOK, gin.H{
		"today": gin.H{
			"requests": requestCount,
			"messages": messageCount,
		},
		"user_id": userID,
		"date":    today,
	})
}

// generateToken 生成JWT token
func (h *UserHandler) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Duration(config.AppConfig.Auth.TokenExpiry) * time.Second).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.Auth.JWTSecret))
}