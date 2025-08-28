package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/polyagent/go-services/internal/config"
	"github.com/polyagent/go-services/internal/models"
)

// PythonAIClient Python AI服务客户端
type PythonAIClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// NewPythonAIClient 创建Python AI客户端
func NewPythonAIClient(cfg *config.Config) *PythonAIClient {
	timeout := time.Duration(cfg.PythonAI.Timeout) * time.Second
	
	return &PythonAIClient{
		baseURL: cfg.PythonAI.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// ExecuteTask 执行任务
func (client *PythonAIClient) ExecuteTask(task *models.AgentTask) (*models.AgentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tasks/execute", client.baseURL)
	
	// 准备请求体
	requestBody, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PolyAgent-Go-Client/1.0")

	// 发送请求
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service returned status %d", resp.StatusCode)
	}

	// 解析响应
	var response models.AgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ExecuteRAGQuery 执行RAG查询
func (client *PythonAIClient) ExecuteRAGQuery(query *RAGQuery) (*RAGResult, error) {
	url := fmt.Sprintf("%s/api/v1/rag/query", client.baseURL)
	
	requestBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PolyAgent-Go-Client/1.0")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RAG service returned status %d", resp.StatusCode)
	}

	var result RAGResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UploadDocument 上传文档
func (client *PythonAIClient) UploadDocument(userID, filename string, content []byte) (*DocumentUploadResponse, error) {
	url := fmt.Sprintf("%s/api/v1/documents/upload", client.baseURL)
	
	uploadReq := DocumentUploadRequest{
		UserID:   userID,
		Filename: filename,
		Content:  string(content),
	}

	requestBody, err := json.Marshal(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal upload request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PolyAgent-Go-Client/1.0")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("document service returned status %d", resp.StatusCode)
	}

	var response DocumentUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// GetAvailableTools 获取可用工具列表
func (client *PythonAIClient) GetAvailableTools() ([]*ToolInfo, error) {
	url := fmt.Sprintf("%s/api/v1/tools/list", client.baseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "PolyAgent-Go-Client/1.0")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tools service returned status %d", resp.StatusCode)
	}

	var response ToolListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Tools, nil
}

// ExecuteTool 执行工具
func (client *PythonAIClient) ExecuteTool(toolCall *models.ToolCall) (*ToolExecutionResult, error) {
	url := fmt.Sprintf("%s/api/v1/tools/execute", client.baseURL)
	
	execReq := ToolExecutionRequest{
		Name:       toolCall.Name,
		Parameters: toolCall.Parameters,
	}

	requestBody, err := json.Marshal(execReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execution request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PolyAgent-Go-Client/1.0")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tool execution service returned status %d", resp.StatusCode)
	}

	var result ToolExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HealthCheck 健康检查
func (client *PythonAIClient) HealthCheck() error {
	url := fmt.Sprintf("%s/health", client.baseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// 请求和响应类型定义

// RAGQuery RAG查询请求
type RAGQuery struct {
	UserID  string                 `json:"user_id"`
	Query   string                 `json:"query"`
	TopK    int                    `json:"top_k"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// RAGResult RAG查询结果
type RAGResult struct {
	Documents []RAGDocument `json:"documents"`
	Scores    []float64     `json:"scores"`
	Context   string        `json:"context"`
}

// RAGDocument RAG文档
type RAGDocument struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Score    float64                `json:"score"`
}

// DocumentUploadRequest 文档上传请求
type DocumentUploadRequest struct {
	UserID   string `json:"user_id"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// DocumentUploadResponse 文档上传响应
type DocumentUploadResponse struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Parameters  map[string]ToolParameter     `json:"parameters"`
	Category    string                       `json:"category"`
}

// ToolParameter 工具参数
type ToolParameter struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// ToolListResponse 工具列表响应
type ToolListResponse struct {
	Tools []*ToolInfo `json:"tools"`
}

// ToolExecutionRequest 工具执行请求
type ToolExecutionRequest struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolExecutionResult 工具执行结果
type ToolExecutionResult struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error,omitempty"`
}