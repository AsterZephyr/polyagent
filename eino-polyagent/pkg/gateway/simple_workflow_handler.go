package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/eino-polyagent/internal/orchestration"
)

// SimpleWorkflowRequest represents a simplified workflow request
type SimpleWorkflowRequest struct {
	WorkflowType string `json:"workflow_type" binding:"required"`
	AgentID      string `json:"agent_id"`
}

// SimpleWorkflowResponse represents a simplified workflow response
type SimpleWorkflowResponse struct {
	WorkflowID   string                        `json:"workflow_id"`
	WorkflowName string                        `json:"workflow_name"`
	Status       string                        `json:"status"`
	Steps        int                           `json:"steps"`
	StepsDetail  []*orchestration.WorkflowStep `json:"steps_detail"`
	CreatedAt    string                        `json:"created_at"`
}

// Add simplified workflow endpoints to the gateway service
func (s *GatewayService) setupSimpleWorkflowRoutes() {
	api := s.router.Group("/api/v1")
	
	api.POST("/workflows/:type/execute", s.executeSimpleWorkflow)
	api.GET("/workflows/templates", s.getWorkflowTemplates)
}

func (s *GatewayService) executeSimpleWorkflow(c *gin.Context) {
	workflowType := c.Param("type")
	
	var req SimpleWorkflowRequest
	req.WorkflowType = workflowType
	
	// Parse JSON body if present
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			// Ignore JSON parsing errors for simple GET requests
		}
	}
	
	// Use default agent if not specified
	if req.AgentID == "" {
		req.AgentID = "default"
	}

	var workflow *orchestration.SimpleWorkflow
	
	switch workflowType {
	case "coding":
		workflow = orchestration.NewCodingWorkflow(req.AgentID)
	case "research":
		workflow = orchestration.NewResearchWorkflow(req.AgentID)
	case "problem-solving":
		workflow = orchestration.NewProblemSolvingWorkflow(req.AgentID)
	case "content-creation":
		workflow = orchestration.NewContentCreationWorkflow(req.AgentID)
	case "data-analysis":
		workflow = orchestration.NewDataAnalysisWorkflow(req.AgentID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown workflow type"})
		return
	}

	response := SimpleWorkflowResponse{
		WorkflowID:   workflow.ID,
		WorkflowName: workflow.Name,
		Status:       string(workflow.Status),
		Steps:        len(workflow.Steps),
		StepsDetail:  workflow.Steps,
		CreatedAt:    workflow.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

func (s *GatewayService) getWorkflowTemplates(c *gin.Context) {
	templates := []map[string]interface{}{
		{
			"id":          "coding",
			"name":        "编程工作流",
			"description": "分析需求、生成代码、验证代码",
			"steps":       3,
		},
		{
			"id":          "research",
			"name":        "研究工作流", 
			"description": "分析主题、收集信息、综合发现",
			"steps":       3,
		},
		{
			"id":          "problem-solving",
			"name":        "问题解决",
			"description": "定义问题、生成方案、评估选项",
			"steps":       3,
		},
		{
			"id":          "content-creation",
			"name":        "内容创作",
			"description": "研究主题、创建草稿、审核内容",
			"steps":       3,
		},
		{
			"id":          "data-analysis",
			"name":        "数据分析",
			"description": "探索数据、处理数据、生成洞察",
			"steps":       3,
		},
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}