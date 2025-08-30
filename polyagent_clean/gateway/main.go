package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// ChatRequest represents an incoming chat request
type ChatRequest struct {
	Message string `json:"message"`
	Context string `json:"context,omitempty"`
	UseTools bool  `json:"use_tools,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
	Model    string `json:"model,omitempty"`
	Cost     float64 `json:"cost,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Models    map[string]string `json:"models,omitempty"`
}

func main() {
	// Get configuration from environment
	port := getEnv("GATEWAY_PORT", "8080")
	host := getEnv("GATEWAY_HOST", "0.0.0.0")

	// Setup routes
	http.HandleFunc("/chat", corsMiddleware(chatHandler))
	http.HandleFunc("/health", corsMiddleware(healthHandler))
	http.HandleFunc("/", corsMiddleware(indexHandler))

	// Setup server
	server := &http.Server{
		Addr:         host + ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("🚀 PolyAgent Gateway starting on http://%s:%s", host, port)
	log.Printf("📍 Endpoints: /chat, /health")
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	log.Println("Server stopped")
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// chatHandler handles chat requests
func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		writeErrorResponse(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Call Python agent
	response, err := callPythonAgent(req)
	if err != nil {
		log.Printf("Agent call failed: %v", err)
		writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// healthHandler handles health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Simple health check - could be enhanced to check Python agent
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// indexHandler serves basic API info
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"name":    "PolyAgent Gateway",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"POST /chat":   "Send chat message",
			"GET /health": "Health check",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// callPythonAgent calls the Python agent
func callPythonAgent(req ChatRequest) (*ChatResponse, error) {
	// Prepare command
	cmd := exec.Command("python3", "main.py")
	cmd.Dir = "../agent"

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("POLYAGENT_TOOLS=%s", strconv.FormatBool(req.UseTools)),
	)

	// Send message via stdin
	cmd.Stdin = nil
	if req.Message != "" {
		// For now, use simple approach - in production, use JSON pipe
		cmdStr := fmt.Sprintf(`echo '%s' | python3 main.py`, req.Message)
		cmd = exec.Command("bash", "-c", cmdStr)
		cmd.Dir = "../agent"
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("POLYAGENT_TOOLS=%s", strconv.FormatBool(req.UseTools)),
		)
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	cmd.Dir = "../agent"

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %v", err)
	}

	return &ChatResponse{
		Response: string(output),
	}, nil
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := ChatResponse{
		Error: message,
	}
	json.NewEncoder(w).Encode(response)
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}