package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"kwatch/runner"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPServer represents the Model Context Protocol server
type MCPServer struct {
	runner    *runner.Runner
	workDir   string
	reader    *bufio.Scanner
	writer    io.Writer
	ctx       context.Context
	cancel    context.CancelFunc
}

// InitializeParams represents MCP initialization parameters
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents the initialization response
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

// Capabilities represents server capabilities
type Capabilities struct {
	Tools   *ToolsCapability   `json:"tools,omitempty"`
	Logging *LoggingCapability `json:"logging,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// LoggingCapability represents logging capability
type LoggingCapability struct{}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema ToolSchema  `json:"inputSchema"`
}

// ToolSchema represents a tool's input schema
type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(workDir string) *MCPServer {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create runner configuration
	config := runner.RunnerConfig{
		DefaultTimeout: 30 * time.Second,
		MaxParallel:    3,
		WorkingDir:     workDir,
	}

	return &MCPServer{
		runner:  runner.NewRunner(config),
		workDir: workDir,
		reader:  bufio.NewScanner(os.Stdin),
		writer:  os.Stdout,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the MCP server
func (s *MCPServer) Start() error {
	for s.reader.Scan() {
		line := s.reader.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if err := s.handleMessage(line); err != nil {
			fmt.Fprintf(os.Stderr, "Error handling message: %v\n", err)
		}
	}

	return s.reader.Err()
}

// Stop stops the MCP server
func (s *MCPServer) Stop() {
	s.cancel()
}

// handleMessage processes incoming JSON-RPC messages
func (s *MCPServer) handleMessage(message string) error {
	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(message), &req); err != nil {
		return s.sendError(nil, -32700, "Parse error", err)
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "notifications/initialized":
		// Client confirms initialization - no response needed
		return nil
	default:
		return s.sendError(req.ID, -32601, "Method not found", nil)
	}
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(req JSONRPCRequest) error {
	var params InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.sendError(req.ID, -32602, "Invalid params", err)
	}

	// Validate protocol version
	if params.ProtocolVersion != "2024-11-05" && params.ProtocolVersion != "2025-03-26" {
		return s.sendError(req.ID, -32602, "Unsupported protocol version", map[string]interface{}{
			"supported": []string{"2024-11-05", "2025-03-26"},
			"requested": params.ProtocolVersion,
		})
	}

	result := InitializeResult{
		ProtocolVersion: "2025-03-26",
		Capabilities: Capabilities{
			Tools: &ToolsCapability{
				ListChanged: true,
			},
			Logging: &LoggingCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "kwatch-mcp",
			Version: "1.0.0",
		},
	}

	return s.sendResponse(req.ID, result)
}

// handleToolsList handles the tools/list request
func (s *MCPServer) handleToolsList(req JSONRPCRequest) error {
	tools := []Tool{
		{
			Name:        "get_build_status",
			Description: "Get the current build status of the monitored project including TypeScript, linting, and test results",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Output format: 'compact' for one-line status, 'detailed' for full JSON",
						"enum":        []string{"compact", "detailed"},
						"default":     "detailed",
					},
				},
			},
		},
		{
			Name:        "run_commands",
			Description: "Execute build commands (TypeScript check, linting, tests) manually",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Specific command to run: 'all', 'tsc', 'lint', or 'test'",
						"enum":        []string{"all", "tsc", "lint", "test"},
						"default":     "all",
					},
				},
			},
		},
		{
			Name:        "get_command_history",
			Description: "Get the history of previously executed commands with results and timestamps",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Maximum number of history entries to return",
						"default":     10,
					},
					"filter": map[string]interface{}{
						"type":        "string",
						"description": "Filter by command type: 'tsc', 'lint', or 'test'",
						"enum":        []string{"tsc", "lint", "test"},
					},
				},
			},
		},
	}

	result := map[string]interface{}{
		"tools": tools,
	}

	return s.sendResponse(req.ID, result)
}

// handleToolsCall handles the tools/call request
func (s *MCPServer) handleToolsCall(req JSONRPCRequest) error {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.sendError(req.ID, -32602, "Invalid params", err)
	}

	switch params.Name {
	case "get_build_status":
		return s.handleGetBuildStatus(req.ID, params.Arguments)
	case "run_commands":
		return s.handleRunCommands(req.ID, params.Arguments)
	case "get_command_history":
		return s.handleGetCommandHistory(req.ID, params.Arguments)
	default:
		return s.sendError(req.ID, -32602, "Unknown tool", map[string]interface{}{
			"tool": params.Name,
		})
	}
}

// handleGetBuildStatus implements the get_build_status tool
func (s *MCPServer) handleGetBuildStatus(id interface{}, args map[string]interface{}) error {
	format := "detailed"
	if f, ok := args["format"].(string); ok {
		format = f
	}

	ctx := context.Background()
	results := s.runner.RunAll(ctx)

	var content string
	var isError bool
	
	if format == "compact" {
		content = runner.FormatCompactStatus(results)
	} else {
		// Format as detailed JSON
		status := map[string]interface{}{
			"directory": s.workDir,
			"timestamp": time.Now().Format(time.RFC3339),
			"commands":  formatCommandResults(results),
		}
		
		jsonBytes, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			content = fmt.Sprintf("Error formatting status: %v", err)
			isError = true
		} else {
			content = string(jsonBytes)
		}
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": content,
			},
		},
		"isError": isError,
	}

	return s.sendResponse(id, result)
}

// handleRunCommands implements the run_commands tool
func (s *MCPServer) handleRunCommands(id interface{}, args map[string]interface{}) error {
	command := "all"
	if c, ok := args["command"].(string); ok {
		command = c
	}

	ctx := context.Background()
	var results map[runner.CommandType]runner.CommandResult
	var isError bool

	switch command {
	case "all":
		results = s.runner.RunAll(ctx)
	case "tsc":
		cmd := runner.Command{
			Type:    runner.TypescriptCheck,
			Command: "npx",
			Args:    []string{"tsc", "--noEmit"},
			Timeout: 30 * time.Second,
		}
		result := s.runner.RunCommand(ctx, cmd)
		results = map[runner.CommandType]runner.CommandResult{
			runner.TypescriptCheck: result,
		}
	case "lint":
		cmd := runner.Command{
			Type:    runner.LintCheck,
			Command: "npx",
			Args:    []string{"eslint", ".", "--ext", ".ts,.tsx,.js,.jsx"},
			Timeout: 30 * time.Second,
		}
		result := s.runner.RunCommand(ctx, cmd)
		results = map[runner.CommandType]runner.CommandResult{
			runner.LintCheck: result,
		}
	case "test":
		cmd := runner.Command{
			Type:    runner.TestRunner,
			Command: "npm",
			Args:    []string{"test"},
			Timeout: 60 * time.Second,
		}
		result := s.runner.RunCommand(ctx, cmd)
		results = map[runner.CommandType]runner.CommandResult{
			runner.TestRunner: result,
		}
	default:
		return s.sendError(id, -32602, "Invalid command", map[string]interface{}{
			"command": command,
		})
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"command":   command,
		"results":   formatCommandResults(results),
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	var content string
	if err != nil {
		content = fmt.Sprintf("Error formatting results: %v", err)
		isError = true
	} else {
		content = string(jsonBytes)
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": content,
			},
		},
		"isError": isError,
	}

	return s.sendResponse(id, result)
}

// handleGetCommandHistory implements the get_command_history tool
func (s *MCPServer) handleGetCommandHistory(id interface{}, args map[string]interface{}) error {
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	filter := ""
	if f, ok := args["filter"].(string); ok {
		filter = f
	}

	history := s.runner.GetHistory()

	// Filter if requested
	if filter != "" {
		var filtered []runner.CommandResult
		for _, entry := range history {
			switch filter {
			case "tsc":
				if strings.Contains(entry.Command, "tsc") {
					filtered = append(filtered, entry)
				}
			case "lint":
				if strings.Contains(entry.Command, "eslint") || strings.Contains(entry.Command, "lint") {
					filtered = append(filtered, entry)
				}
			case "test":
				if strings.Contains(entry.Command, "test") {
					filtered = append(filtered, entry)
				}
			}
		}
		history = filtered
	}

	// Apply limit
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	response := map[string]interface{}{
		"count":   len(history),
		"history": history,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	var content string
	var isError bool
	if err != nil {
		content = fmt.Sprintf("Error formatting history: %v", err)
		isError = true
	} else {
		content = string(jsonBytes)
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": content,
			},
		},
		"isError": isError,
	}

	return s.sendResponse(id, result)
}

// formatCommandResults formats command results for JSON output
func formatCommandResults(results map[runner.CommandType]runner.CommandResult) map[string]interface{} {
	formatted := make(map[string]interface{})
	
	cmdNames := map[runner.CommandType]string{
		runner.TypescriptCheck: "tsc",
		runner.LintCheck:       "lint",
		runner.TestRunner:      "test",
	}

	for cmdType, result := range results {
		name := cmdNames[cmdType]
		if name == "" {
			name = string(cmdType)
		}

		formatted[name] = map[string]interface{}{
			"passed":      result.Passed,
			"issue_count": result.IssueCount,
			"file_count":  result.FileCount,
			"duration":    result.Duration.String(),
			"timestamp":   result.Timestamp.Format(time.RFC3339),
		}
	}

	return formatted
}

// sendResponse sends a JSON-RPC success response
func (s *MCPServer) sendResponse(id interface{}, result interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	return s.writeMessage(response)
}

// sendError sends a JSON-RPC error response
func (s *MCPServer) sendError(id interface{}, code int, message string, data interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	return s.writeMessage(response)
}

// writeMessage writes a JSON-RPC message to stdout
func (s *MCPServer) writeMessage(message interface{}) error {
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "%s\n", string(jsonBytes))
	return err
}