package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/fathindos/agit/internal/config"
	"github.com/fathindos/agit/internal/registry"
)

func mustDB(t *testing.T) *registry.DB {
	t.Helper()
	db, err := registry.OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func callTool(t *testing.T, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) map[string]any {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// Extract text content from result
	if len(result.Content) == 0 {
		t.Fatal("empty result content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &out); err != nil {
		// Try as array
		var arr []any
		if err2 := json.Unmarshal([]byte(textContent.Text), &arr); err2 != nil {
			t.Fatalf("could not unmarshal result: %v (text: %s)", err, textContent.Text)
		}
		// Wrap array in a map for consistent return
		return map[string]any{"_array": arr}
	}
	return out
}

func callToolExpectError(t *testing.T, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) error {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args

	_, err := handler(context.Background(), req)
	return err
}

func TestHandleListRepos(t *testing.T) {
	db := mustDB(t)
	db.AddRepo("test-repo", "/tmp/test", "https://github.com/test/repo", "main")

	handler := handleListRepos(db)
	result := callTool(t, handler, nil)

	arr, ok := result["_array"].([]any)
	if !ok {
		t.Fatal("expected array result")
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(arr))
	}
}

func TestHandleRepoStatus(t *testing.T) {
	db := mustDB(t)
	db.AddRepo("status-repo", "/tmp/status", "", "main")

	handler := handleRepoStatus(db)

	// Missing repo param
	err := callToolExpectError(t, handler, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing repo param")
	}

	result := callTool(t, handler, map[string]any{"repo": "status-repo"})
	if result["name"] != "status-repo" {
		t.Errorf("expected name status-repo, got %v", result["name"])
	}
}

func TestHandleClaimTask(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("claim-repo", "/tmp/claim", "", "main")
	task, _ := db.CreateTask(repo.ID, "test task", 0)
	agent, _ := db.RegisterAgent("claimer", "custom")

	handler := handleClaimTask(db)
	result := callTool(t, handler, map[string]any{
		"task_id":  task.ID,
		"agent_id": agent.ID,
	})
	if result["claimed"] != true {
		t.Error("expected claimed true")
	}
}

func TestHandleCompleteTask(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("comp-repo", "/tmp/comp", "", "main")
	task, _ := db.CreateTask(repo.ID, "test task", 0)

	handler := handleCompleteTask(db)
	result := callTool(t, handler, map[string]any{
		"task_id": task.ID,
		"result":  "done",
	})
	if result["completed"] != true {
		t.Error("expected completed true")
	}
}

func TestHandleRegisterAgent(t *testing.T) {
	db := mustDB(t)

	handler := handleRegisterAgent(db)
	result := callTool(t, handler, map[string]any{
		"name": "test-agent",
		"type": "claude",
	})
	if result["name"] != "test-agent" {
		t.Errorf("expected name test-agent, got %v", result["name"])
	}
}

func TestHandleHeartbeat(t *testing.T) {
	db := mustDB(t)
	agent, _ := db.RegisterAgent("hb-agent", "custom")

	handler := handleHeartbeat(db)
	result := callTool(t, handler, map[string]any{
		"agent_id": agent.ID,
	})
	if result["ok"] != true {
		t.Error("expected ok true")
	}
}

// --- Tests for new handlers ---

func TestHandleCreateTask(t *testing.T) {
	db := mustDB(t)
	db.AddRepo("ct-repo", "/tmp/ct", "", "main")

	handler := handleCreateTask(db)
	result := callTool(t, handler, map[string]any{
		"repo":        "ct-repo",
		"description": "implement feature X",
		"priority":    float64(5),
	})
	if result["description"] != "implement feature X" {
		t.Errorf("expected description, got %v", result["description"])
	}
	if result["status"] != "pending" {
		t.Errorf("expected pending status, got %v", result["status"])
	}
}

func TestHandleCreateTaskMissingParams(t *testing.T) {
	db := mustDB(t)

	handler := handleCreateTask(db)
	if err := callToolExpectError(t, handler, map[string]any{}); err == nil {
		t.Fatal("expected error for missing repo")
	}
	if err := callToolExpectError(t, handler, map[string]any{"repo": "x"}); err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestHandleFailTask(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("ft-repo", "/tmp/ft", "", "main")
	task, _ := db.CreateTask(repo.ID, "will fail", 0)

	handler := handleFailTask(db)
	result := callTool(t, handler, map[string]any{
		"task_id": task.ID,
		"result":  "broken",
	})
	if result["failed"] != true {
		t.Error("expected failed true")
	}
}

func TestHandleStartTask(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("st-repo", "/tmp/st", "", "main")
	task, _ := db.CreateTask(repo.ID, "start me", 0)
	agent, _ := db.RegisterAgent("starter", "custom")
	db.ClaimTask(task.ID, agent.ID)
	wt, _ := db.CreateWorktree(repo.ID, "/tmp/st-wt", "b1", nil, nil)

	handler := handleStartTask(db)
	result := callTool(t, handler, map[string]any{
		"task_id":     task.ID,
		"worktree_id": wt.ID,
	})
	if result["started"] != true {
		t.Error("expected started true")
	}
}

func TestHandleListAgents(t *testing.T) {
	db := mustDB(t)
	db.RegisterAgent("agent-1", "claude")
	db.RegisterAgent("agent-2", "custom")

	handler := handleListAgents(db)
	result := callTool(t, handler, nil)

	arr, ok := result["_array"].([]any)
	if !ok {
		t.Fatal("expected array result")
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(arr))
	}
}

func TestHandleListWorktrees(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("lw-repo", "/tmp/lw", "", "main")
	db.CreateWorktree(repo.ID, "/tmp/lw1", "b1", nil, nil)
	db.CreateWorktree(repo.ID, "/tmp/lw2", "b2", nil, nil)

	handler := handleListWorktrees(db)
	result := callTool(t, handler, map[string]any{"repo": "lw-repo"})

	arr, ok := result["_array"].([]any)
	if !ok {
		t.Fatal("expected array result")
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(arr))
	}
}

func TestHandleGetTask(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("gt-repo", "/tmp/gt", "", "main")
	task, _ := db.CreateTask(repo.ID, "get me", 3)

	handler := handleGetTask(db)
	result := callTool(t, handler, map[string]any{"task_id": task.ID})

	if result["description"] != "get me" {
		t.Errorf("expected 'get me', got %v", result["description"])
	}
	if result["status"] != "pending" {
		t.Errorf("expected pending, got %v", result["status"])
	}
}

func TestHandleAddRepo(t *testing.T) {
	db := mustDB(t)

	// Create a temp git repo
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(gitDir, 0755)

	handler := handleAddRepo(db)
	result := callTool(t, handler, map[string]any{
		"path": dir,
		"name": "added-repo",
	})
	if result["name"] != "added-repo" {
		t.Errorf("expected name added-repo, got %v", result["name"])
	}
}

func TestHandleAddRepoNotGit(t *testing.T) {
	db := mustDB(t)

	dir := t.TempDir()

	handler := handleAddRepo(db)
	if err := callToolExpectError(t, handler, map[string]any{"path": dir}); err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestHandleCleanupWorktrees(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("cl-repo", "/tmp/cl", "", "main")
	// Create worktree pointing to non-existent path
	db.CreateWorktree(repo.ID, "/tmp/nonexistent-cleanup-test", "b1", nil, nil)

	handler := handleCleanupWorktrees(db)
	result := callTool(t, handler, map[string]any{"repo": "cl-repo"})

	pruned, ok := result["pruned"].(float64)
	if !ok || pruned != 1 {
		t.Errorf("expected 1 pruned, got %v", result["pruned"])
	}
}

func TestWithIssueLink(t *testing.T) {
	// Verify wrapper doesn't interfere with normal operation
	db := mustDB(t)
	db.AddRepo("wrap-repo", "/tmp/wrap", "", "main")

	handler := withIssueLink(handleListRepos(db))
	req := mcp.CallToolRequest{}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("withIssueLink should not add error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHandleNextTask(t *testing.T) {
	db := mustDB(t)
	repo, _ := db.AddRepo("nt-repo", "/tmp/nt", "", "main")
	agent, _ := db.RegisterAgent("nt-agent", "custom")

	// Create tasks with different priorities
	db.CreateTask(repo.ID, "low prio", 1)
	db.CreateTask(repo.ID, "high prio", 10)
	db.CreateTask(repo.ID, "med prio", 5)

	handler := handleNextTask(db)
	result := callTool(t, handler, map[string]any{
		"repo":     "nt-repo",
		"agent_id": agent.ID,
	})

	task, ok := result["task"].(map[string]any)
	if !ok {
		t.Fatalf("expected task map, got %v", result["task"])
	}
	if task["description"] != "high prio" {
		t.Errorf("expected highest priority task, got %v", task["description"])
	}
}

func TestHandleNextTaskNoPending(t *testing.T) {
	db := mustDB(t)
	db.AddRepo("nt2-repo", "/tmp/nt2", "", "main")
	agent, _ := db.RegisterAgent("nt2-agent", "custom")

	handler := handleNextTask(db)
	result := callTool(t, handler, map[string]any{
		"repo":     "nt2-repo",
		"agent_id": agent.ID,
	})

	if result["task"] != nil {
		t.Errorf("expected nil task, got %v", result["task"])
	}
	if result["message"] != "no pending tasks" {
		t.Errorf("expected 'no pending tasks' message, got %v", result["message"])
	}
}

func TestHandleNextTaskMissingParams(t *testing.T) {
	db := mustDB(t)

	handler := handleNextTask(db)

	if err := callToolExpectError(t, handler, map[string]any{}); err == nil {
		t.Fatal("expected error for missing repo")
	}
	if err := callToolExpectError(t, handler, map[string]any{"repo": "x"}); err == nil {
		t.Fatal("expected error for missing agent_id")
	}
}

// Verify NewServer creates server with all tools
func TestNewServer(t *testing.T) {
	db := mustDB(t)
	cfg := config.DefaultConfig()

	s := NewServer(db, cfg)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}
