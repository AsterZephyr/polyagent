package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/go-services/internal/ai"
	"github.com/polyagent/go-services/internal/models"
	"github.com/polyagent/go-services/internal/storage"
)

// DocumentHandler 文档处理器
type DocumentHandler struct {
	postgres *storage.PostgresStorage
	aiClient *ai.PythonAIClient
}

// NewDocumentHandler 创建文档处理器
func NewDocumentHandler(postgres *storage.PostgresStorage, aiClient *ai.PythonAIClient) *DocumentHandler {
	return &DocumentHandler{
		postgres: postgres,
		aiClient: aiClient,
	}
}

// ListDocuments 获取文档列表
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	_ = userID // 避免unused变量警告

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 简化实现：返回空列表
	documents := []models.Document{}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"total":     len(documents),
		"page":      page,
		"limit":     limit,
	})
}

// UploadDocument 上传文档
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > 32*1024*1024 { // 32MB 限制
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large"})
		return
	}

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// 调用 Python AI 服务处理文档
	response, err := h.aiClient.UploadDocument(userID.(string), header.Filename, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process document"})
		return
	}

	// 创建文档记录
	document := &models.Document{
		ID:       models.NewDocumentID(),
		UserID:   userID.(string),
		Filename: header.Filename,
		Status:   models.DocumentStatusUploaded,
		Metadata: map[string]interface{}{
			"size":      header.Size,
			"mime_type": header.Header.Get("Content-Type"),
		},
	}

	// 保存到数据库（这里应该实现保存逻辑）
	// if err := h.postgres.CreateDocument(document); err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save document"})
	//     return
	// }

	c.JSON(http.StatusCreated, gin.H{
		"document_id": document.ID,
		"filename":    document.Filename,
		"status":      response.Status,
		"message":     response.Message,
	})
}

// GetDocument 获取文档详情
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	documentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 这里应该从数据库获取文档
	// document, err := h.postgres.GetDocument(documentID)
	// if err != nil {
	//     c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
	//     return
	// }

	// 检查权限
	// if document.UserID != userID.(string) {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
	//     return
	// }

	// 简化实现
	c.JSON(http.StatusOK, gin.H{
		"id":       documentID,
		"user_id":  userID,
		"filename": "example.pdf",
		"status":   "processed",
	})
}

// DeleteDocument 删除文档
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	documentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	_ = userID // 避免unused变量警告

	// 这里应该实现删除逻辑
	// 1. 检查文档是否存在
	// 2. 检查权限
	// 3. 从向量数据库删除
	// 4. 从文件存储删除
	// 5. 从数据库删除

	c.JSON(http.StatusOK, gin.H{
		"message":     "Document deleted successfully",
		"document_id": documentID,
	})
}

// IndexDocuments 重建文档索引
func (h *DocumentHandler) IndexDocuments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 这里应该实现索引重建逻辑
	// 1. 获取用户所有文档
	// 2. 调用 Python AI 服务重新处理
	// 3. 更新向量数据库

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Document indexing started",
		"user_id": userID,
	})
}