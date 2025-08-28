# PolyAgent API 接口设计

## 核心接口定义

### 1. Go ↔ Python 服务间通信

#### 任务执行接口
```go
// Go 发送给 Python
type AgentTask struct {
    TaskID      string            `json:"task_id"`
    AgentType   string            `json:"agent_type"`   // "chat", "rag", "tool"
    Input       string            `json:"input"`
    Context     map[string]any    `json:"context"`
    Tools       []string          `json:"tools"`        // 可用工具列表
    Memory      *ConversationMemory `json:"memory"`
}

// Python 返回给 Go  
type AgentResponse struct {
    TaskID      string            `json:"task_id"`
    Status      string            `json:"status"`       // "success", "error", "streaming"
    Output      string            `json:"output"`
    ToolCalls   []ToolCall        `json:"tool_calls"`
    Memory      *ConversationMemory `json:"memory"`
    Metadata    map[string]any    `json:"metadata"`
}
```

#### RAG 检索接口
```go
type RAGQuery struct {
    UserID      string   `json:"user_id"`
    Query       string   `json:"query"`
    TopK        int      `json:"top_k"`
    Filters     map[string]any `json:"filters"`
}

type RAGResult struct {
    Documents   []Document `json:"documents"`
    Scores      []float64  `json:"scores"`
    Context     string     `json:"context"`
}
```

### 2. 客户端 API 接口

#### 对话接口
```
POST /api/v1/chat
Content-Type: application/json

{
    "message": "用户输入",
    "session_id": "会话ID", 
    "agent_type": "general|rag|code",
    "tools": ["search", "calculator"],
    "stream": true
}
```

#### Agent 管理接口
```
GET    /api/v1/agents           # 获取智能体列表
POST   /api/v1/agents           # 创建智能体
PUT    /api/v1/agents/:id       # 更新智能体
DELETE /api/v1/agents/:id       # 删除智能体
```

#### RAG 文档管理
```
POST   /api/v1/documents/upload # 上传文档
GET    /api/v1/documents        # 文档列表
DELETE /api/v1/documents/:id    # 删除文档
POST   /api/v1/documents/index  # 重建索引
```

### 3. 工具调用接口

#### 工具注册
```go
type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]Parameter   `json:"parameters"`
    Handler     string                 `json:"handler"`    // 处理函数路径
}

type Parameter struct {
    Type        string `json:"type"`
    Description string `json:"description"`
    Required    bool   `json:"required"`
}
```

#### 工具执行
```
POST /api/v1/tools/execute
{
    "tool_name": "web_search",
    "parameters": {
        "query": "搜索内容",
        "max_results": 5
    }
}
```

### 4. 流式响应格式
```json
{
    "type": "text|tool_call|error|done",
    "content": "响应内容",
    "metadata": {
        "tool_name": "工具名",
        "step": "当前步骤"
    }
}
```