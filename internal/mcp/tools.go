package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/fathindos/agit/internal/config"
	"github.com/fathindos/agit/internal/conflicts"
	apperrors "github.com/fathindos/agit/internal/errors"
	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/issuelink"
	"github.com/fathindos/agit/internal/registry"
)

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("could not marshal response: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

// wrapInternalError appends an issue link to internal (non-user) errors
// so that AI agents or their operators can easily report bugs.
// withIssueLink wraps an MCP tool handler to append issue links to internal errors.
func withIssueLink(handler mcpserver.ToolHandlerFunc) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := handler(ctx, request)
		if err != nil && !apperrors.IsUserError(err) && issuelink.Enabled() {
			link := issuelink.ForError(err)
			return result, fmt.Errorf("%w\n\nTo report this bug, open:\n  %s", err, link)
		}
		return result, err
	}
}

func handleListRepos(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repos, err := db.ListRepos()
		if err != nil {
			return nil, fmt.Errorf("could not list repos: %w", err)
		}

		type repoItem struct {
			Name          string `json:"name"`
			Path          string `json:"path"`
			DefaultBranch string `json:"default_branch"`
			RemoteURL     string `json:"remote_url"`
			WorktreeCount int    `json:"worktree_count"`
			TaskCount     int    `json:"task_count"`
		}

		var items []repoItem
		for _, r := range repos {
			stats, _ := db.GetRepoStats(r.ID)
			wc, tc := 0, 0
			if stats != nil {
				wc = stats.ActiveWorktrees
				tc = stats.PendingTasks
			}
			items = append(items, repoItem{
				Name:          r.Name,
				Path:          r.Path,
				DefaultBranch: r.DefaultBranch,
				RemoteURL:     r.RemoteURL,
				WorktreeCount: wc,
				TaskCount:     tc,
			})
		}

		return jsonResult(items)
	}
}

func handleRepoStatus(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repoName, _ := request.Params.Arguments["repo"].(string)
		if repoName == "" {
			return nil, apperrors.NewUserError("repo parameter is required")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		activeStatus := "active"
		worktrees, _ := db.ListWorktrees(repo.ID, &activeStatus)
		tasks, _ := db.ListTasks(repo.ID, nil)
		conflictList, _ := conflicts.Detect(db, repo)

		type wtItem struct {
			ID     string `json:"id"`
			Branch string `json:"branch"`
			Status string `json:"status"`
			Agent  string `json:"agent"`
			Task   string `json:"task"`
		}
		type taskItem struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Status      string `json:"status"`
			Agent       string `json:"agent"`
		}
		type conflictItem struct {
			File      string   `json:"file"`
			Worktrees []string `json:"worktrees"`
		}

		var wts []wtItem
		for _, wt := range worktrees {
			agent := ""
			if wt.AgentID != nil {
				a, err := db.GetAgent(*wt.AgentID)
				if err == nil {
					agent = a.Name
				}
			}
			task := ""
			if wt.TaskDescription != nil {
				task = *wt.TaskDescription
			}
			wts = append(wts, wtItem{wt.ID, wt.Branch, wt.Status, agent, task})
		}

		var tks []taskItem
		for _, t := range tasks {
			agent := ""
			if t.AssignedAgentID != nil {
				a, err := db.GetAgent(*t.AssignedAgentID)
				if err == nil {
					agent = a.Name
				}
			}
			tks = append(tks, taskItem{t.ID, t.Description, t.Status, agent})
		}

		var cls []conflictItem
		for _, c := range conflictList {
			cls = append(cls, conflictItem{c.FilePath, c.Worktrees})
		}

		result := map[string]any{
			"name":           repo.Name,
			"path":           repo.Path,
			"default_branch": repo.DefaultBranch,
			"remote_url":     repo.RemoteURL,
			"worktrees":      wts,
			"tasks":          tks,
			"conflicts":      cls,
		}

		return jsonResult(result)
	}
}

func handleSpawnWorktree(db *registry.DB, cfg *config.Config) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repoName, _ := request.Params.Arguments["repo"].(string)
		if repoName == "" {
			return nil, apperrors.NewUserError("repo parameter is required")
		}
		task, _ := request.Params.Arguments["task"].(string)
		branch, _ := request.Params.Arguments["branch"].(string)
		agentName, _ := request.Params.Arguments["agent"].(string)

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		shortID := uuid.New().String()[:8]
		if branch == "" {
			if task != "" {
				slug := strings.ToLower(task)
				slug = strings.ReplaceAll(slug, " ", "-")
				cleaned := ""
				for _, c := range slug {
					if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
						cleaned += string(c)
					}
				}
				if len(cleaned) > 40 {
					cleaned = cleaned[:40]
				}
				branch = cfg.Defaults.BranchPrefix + cleaned + "-" + shortID
			} else {
				branch = cfg.Defaults.BranchPrefix + shortID
			}
		}

		worktreePath := filepath.Join(repo.Path, cfg.Defaults.WorktreeDir, "agit-"+shortID)

		if err := gitops.CreateWorktree(repo.Path, worktreePath, branch, repo.DefaultBranch); err != nil {
			return nil, fmt.Errorf("could not create worktree: %w", err)
		}

		var agentID *string
		if agentName != "" {
			agent, err := db.GetAgentByName(agentName)
			if err != nil {
				return nil, err
			}
			if agent == nil {
				agent, err = db.RegisterAgent(agentName, "custom")
				if err != nil {
					return nil, err
				}
			}
			agentID = &agent.ID
		}

		var taskDesc *string
		if task != "" {
			taskDesc = &task
		}

		wt, err := db.CreateWorktree(repo.ID, worktreePath, branch, agentID, taskDesc)
		if err != nil {
			gitops.RemoveWorktree(repo.Path, worktreePath)
			return nil, fmt.Errorf("could not record worktree: %w", err)
		}

		if agentID != nil {
			db.UpdateAgentWorktree(*agentID, &wt.ID)
		}

		return jsonResult(map[string]string{
			"worktree_id": wt.ID,
			"path":        worktreePath,
			"branch":      branch,
		})
	}
}

func handleRemoveWorktree(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repoName, _ := request.Params.Arguments["repo"].(string)
		if repoName == "" {
			return nil, apperrors.NewUserError("repo parameter is required")
		}
		worktreeID, _ := request.Params.Arguments["worktree_id"].(string)
		if worktreeID == "" {
			return nil, apperrors.NewUserError("worktree_id parameter is required")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		wt, err := db.ResolveWorktree(repo.ID, worktreeID)
		if err != nil {
			return nil, err
		}

		gitops.RemoveWorktree(repo.Path, wt.Path)
		db.DeleteWorktree(wt.ID)

		return jsonResult(map[string]any{
			"removed":     true,
			"worktree_id": worktreeID,
		})
	}
}

func handleCheckConflicts(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repoName, _ := request.Params.Arguments["repo"].(string)
		if repoName == "" {
			return nil, apperrors.NewUserError("repo parameter is required")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		conflictList, err := conflicts.Detect(db, repo)
		if err != nil {
			return nil, fmt.Errorf("could not detect conflicts: %w", err)
		}

		activeStatus := "active"
		worktrees, _ := db.ListWorktrees(repo.ID, &activeStatus)

		type conflictItem struct {
			File      string   `json:"file"`
			Worktrees []string `json:"worktrees"`
		}

		var items []conflictItem
		for _, c := range conflictList {
			items = append(items, conflictItem{c.FilePath, c.Worktrees})
		}

		return jsonResult(map[string]any{
			"conflicts":         items,
			"scanned_worktrees": len(worktrees),
		})
	}
}

func handleListTasks(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repoName, _ := request.Params.Arguments["repo"].(string)
		if repoName == "" {
			return nil, apperrors.NewUserError("repo parameter is required")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		var statusFilter *string
		if s, ok := request.Params.Arguments["status"].(string); ok && s != "" {
			statusFilter = &s
		}

		tasks, err := db.ListTasks(repo.ID, statusFilter)
		if err != nil {
			return nil, err
		}

		type taskItem struct {
			ID          string  `json:"id"`
			Description string  `json:"description"`
			Priority    int     `json:"priority"`
			Status      string  `json:"status"`
			Agent       *string `json:"agent"`
			CreatedAt   string  `json:"created_at"`
		}

		var items []taskItem
		for _, t := range tasks {
			var agentName *string
			if t.AssignedAgentID != nil {
				a, err := db.GetAgent(*t.AssignedAgentID)
				if err == nil {
					agentName = &a.Name
				}
			}
			items = append(items, taskItem{
				ID:          t.ID,
				Description: t.Description,
				Priority:    t.Priority,
				Status:      t.Status,
				Agent:       agentName,
				CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		return jsonResult(items)
	}
}

func handleClaimTask(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, _ := request.Params.Arguments["task_id"].(string)
		if taskID == "" {
			return nil, apperrors.NewUserError("task_id parameter is required")
		}
		agentID, _ := request.Params.Arguments["agent_id"].(string)
		if agentID == "" {
			return nil, apperrors.NewUserError("agent_id parameter is required")
		}

		if err := db.ClaimTask(taskID, agentID); err != nil {
			return nil, err
		}

		return jsonResult(map[string]any{
			"claimed":  true,
			"task_id":  taskID,
			"agent_id": agentID,
		})
	}
}

func handleCompleteTask(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, _ := request.Params.Arguments["task_id"].(string)
		if taskID == "" {
			return nil, apperrors.NewUserError("task_id parameter is required")
		}

		var resultPtr *string
		if r, ok := request.Params.Arguments["result"].(string); ok && r != "" {
			resultPtr = &r
		}

		if err := db.CompleteTask(taskID, resultPtr); err != nil {
			return nil, err
		}

		return jsonResult(map[string]any{
			"completed": true,
			"task_id":   taskID,
		})
	}
}

func handleMergeWorktree(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		repoName, _ := request.Params.Arguments["repo"].(string)
		if repoName == "" {
			return nil, apperrors.NewUserError("repo parameter is required")
		}
		worktreeID, _ := request.Params.Arguments["worktree_id"].(string)
		if worktreeID == "" {
			return nil, apperrors.NewUserError("worktree_id parameter is required")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		wt, err := db.ResolveWorktree(repo.ID, worktreeID)
		if err != nil {
			return nil, err
		}

		// Pre-merge conflict check
		canMerge, err := gitops.CanMergeCleanly(repo.Path, wt.Branch)
		if err != nil {
			return nil, fmt.Errorf("could not check merge compatibility: %w", err)
		}
		if !canMerge {
			return jsonResult(map[string]any{
				"error": "merge would result in conflicts",
			})
		}

		// Checkout default branch and merge
		if err := gitops.CheckoutBranch(repo.Path, repo.DefaultBranch); err != nil {
			return nil, fmt.Errorf("could not checkout %s: %w", repo.DefaultBranch, err)
		}

		if err := gitops.MergeBranch(repo.Path, wt.Branch); err != nil {
			return nil, err
		}

		// Auto-cleanup: remove worktree from disk and mark completed
		gitops.RemoveWorktree(repo.Path, wt.Path)
		gitops.DeleteBranch(repo.Path, wt.Branch)
		db.UpdateWorktreeStatus(wt.ID, "completed")

		return jsonResult(map[string]any{
			"merged":           true,
			"branch":           wt.Branch,
			"into":             repo.DefaultBranch,
			"worktree_cleaned": true,
		})
	}
}

func handleRegisterAgent(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, _ := request.Params.Arguments["name"].(string)
		if name == "" {
			return nil, apperrors.NewUserError("name parameter is required")
		}
		agentType, _ := request.Params.Arguments["type"].(string)
		if agentType == "" {
			return nil, apperrors.NewUserError("type parameter is required")
		}

		agent, err := db.RegisterAgent(name, agentType)
		if err != nil {
			return nil, err
		}

		return jsonResult(map[string]string{
			"agent_id": agent.ID,
			"name":     agent.Name,
			"type":     agent.Type,
		})
	}
}

func handleHeartbeat(db *registry.DB) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		agentID, _ := request.Params.Arguments["agent_id"].(string)
		if agentID == "" {
			return nil, apperrors.NewUserError("agent_id parameter is required")
		}

		if err := db.Heartbeat(agentID); err != nil {
			return nil, err
		}

		return jsonResult(map[string]any{
			"ok":       true,
			"agent_id": agentID,
		})
	}
}
