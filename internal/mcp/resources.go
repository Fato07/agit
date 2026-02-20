package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/fathindos/agit/internal/conflicts"
	"github.com/fathindos/agit/internal/registry"
)

func textResource(uri, mimeType string, v any) ([]mcp.ResourceContents, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("could not marshal resource: %w", err)
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: mimeType,
			Text:     string(data),
		},
	}, nil
}

func handleReposResource(db *registry.DB) func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		repos, err := db.ListRepos()
		if err != nil {
			return nil, err
		}

		type repoSummary struct {
			Name            string `json:"name"`
			Path            string `json:"path"`
			DefaultBranch   string `json:"default_branch"`
			RemoteURL       string `json:"remote_url"`
			ActiveWorktrees int    `json:"active_worktrees"`
			PendingTasks    int    `json:"pending_tasks"`
			ActiveAgents    int    `json:"active_agents"`
		}

		var items []repoSummary
		for _, r := range repos {
			stats, _ := db.GetRepoStats(r.ID)
			wc, tc := 0, 0
			if stats != nil {
				wc = stats.ActiveWorktrees
				tc = stats.PendingTasks
			}
			// Count active agents with worktrees in this repo
			activeStatus := "active"
			worktrees, _ := db.ListWorktrees(r.ID, &activeStatus)
			agentSet := make(map[string]bool)
			for _, wt := range worktrees {
				if wt.AgentID != nil {
					agentSet[*wt.AgentID] = true
				}
			}
			items = append(items, repoSummary{
				Name:            r.Name,
				Path:            r.Path,
				DefaultBranch:   r.DefaultBranch,
				RemoteURL:       r.RemoteURL,
				ActiveWorktrees: wc,
				PendingTasks:    tc,
				ActiveAgents:    len(agentSet),
			})
		}

		return textResource("agit://repos", "application/json", items)
	}
}

func extractRepoName(uri string) string {
	// URI format: agit://repos/{name} or agit://repos/{name}/conflicts etc.
	trimmed := strings.TrimPrefix(uri, "agit://repos/")
	parts := strings.SplitN(trimmed, "/", 2)
	return parts[0]
}

func handleRepoDetailResource(db *registry.DB) func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		repoName := extractRepoName(request.Params.URI)
		if repoName == "" {
			return nil, fmt.Errorf("repo name is required in URI")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		activeStatus := "active"
		worktrees, _ := db.ListWorktrees(repo.ID, &activeStatus)
		tasks, _ := db.ListTasks(repo.ID, nil)

		type wtSummary struct {
			ID        string `json:"id"`
			Branch    string `json:"branch"`
			Status    string `json:"status"`
			AgentName string `json:"agent_name"`
			Task      string `json:"task"`
			CreatedAt string `json:"created_at"`
		}

		var wts []wtSummary
		agentSet := make(map[string]bool)
		for _, wt := range worktrees {
			agent := ""
			if wt.AgentID != nil {
				agentSet[*wt.AgentID] = true
				a, err := db.GetAgent(*wt.AgentID)
				if err == nil {
					agent = a.Name
				}
			}
			task := ""
			if wt.TaskDescription != nil {
				task = *wt.TaskDescription
			}
			wts = append(wts, wtSummary{wt.ID, wt.Branch, wt.Status, agent, task, wt.CreatedAt.Format("2006-01-02T15:04:05Z")})
		}

		// Task summary by status
		taskSummary := map[string]int{
			"pending": 0, "claimed": 0, "in_progress": 0, "completed": 0, "failed": 0,
		}
		for _, t := range tasks {
			taskSummary[t.Status]++
		}

		result := map[string]any{
			"name":           repo.Name,
			"path":           repo.Path,
			"default_branch": repo.DefaultBranch,
			"remote_url":     repo.RemoteURL,
			"worktrees":      wts,
			"task_summary":   taskSummary,
			"active_agents":  len(agentSet),
		}

		return textResource(request.Params.URI, "application/json", result)
	}
}

func handleRepoConflictsResource(db *registry.DB) func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		repoName := extractRepoName(request.Params.URI)
		if repoName == "" {
			return nil, fmt.Errorf("repo name is required in URI")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		conflictList, err := conflicts.Detect(db, repo)
		if err != nil {
			return nil, err
		}

		type wtRef struct {
			ID     string `json:"id"`
			Branch string `json:"branch"`
		}
		type conflictDetail struct {
			File      string  `json:"file"`
			Worktrees []wtRef `json:"worktrees"`
		}

		var items []conflictDetail
		for _, c := range conflictList {
			var refs []wtRef
			for _, wtID := range c.Worktrees {
				wt, err := db.GetWorktree(wtID)
				if err == nil {
					refs = append(refs, wtRef{wt.ID, wt.Branch})
				}
			}
			items = append(items, conflictDetail{c.FilePath, refs})
		}

		result := map[string]any{
			"repo":            repoName,
			"conflicts":       items,
			"total_conflicts": len(items),
		}

		return textResource(request.Params.URI, "application/json", result)
	}
}

func handleRepoTasksResource(db *registry.DB) func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		repoName := extractRepoName(request.Params.URI)
		if repoName == "" {
			return nil, fmt.Errorf("repo name is required in URI")
		}

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return nil, err
		}

		tasks, err := db.ListTasks(repo.ID, nil)
		if err != nil {
			return nil, err
		}

		type taskDetail struct {
			ID             string  `json:"id"`
			Description    string  `json:"description"`
			Status         string  `json:"status"`
			AssignedAgent  *string `json:"assigned_agent"`
			WorktreeBranch *string `json:"worktree_branch"`
			CreatedAt      string  `json:"created_at"`
			CompletedAt    *string `json:"completed_at"`
			Result         *string `json:"result"`
		}

		var items []taskDetail
		for _, t := range tasks {
			var agentName *string
			if t.AssignedAgentID != nil {
				a, err := db.GetAgent(*t.AssignedAgentID)
				if err == nil {
					agentName = &a.Name
				}
			}
			var wtBranch *string
			if t.WorktreeID != nil {
				wt, err := db.GetWorktree(*t.WorktreeID)
				if err == nil {
					wtBranch = &wt.Branch
				}
			}
			var completedAt *string
			if t.CompletedAt != nil {
				s := t.CompletedAt.Format("2006-01-02T15:04:05Z")
				completedAt = &s
			}
			items = append(items, taskDetail{
				ID:             t.ID,
				Description:    t.Description,
				Status:         t.Status,
				AssignedAgent:  agentName,
				WorktreeBranch: wtBranch,
				CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z"),
				CompletedAt:    completedAt,
				Result:         t.Result,
			})
		}

		result := map[string]any{
			"repo":  repoName,
			"tasks": items,
		}

		return textResource(request.Params.URI, "application/json", result)
	}
}

func handleAgentsResource(db *registry.DB) func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		agents, err := db.ListAgents()
		if err != nil {
			return nil, err
		}

		type wtInfo struct {
			ID     string `json:"id"`
			Branch string `json:"branch"`
			Repo   string `json:"repo"`
		}
		type agentDetail struct {
			ID              string  `json:"id"`
			Name            string  `json:"name"`
			Type            string  `json:"type"`
			Status          string  `json:"status"`
			LastSeen        string  `json:"last_seen"`
			CurrentWorktree *wtInfo `json:"current_worktree"`
		}

		var items []agentDetail
		for _, a := range agents {
			var cwt *wtInfo
			if a.CurrentWorktreeID != nil {
				wt, err := db.GetWorktree(*a.CurrentWorktreeID)
				if err == nil {
					repo, err := db.GetRepoByID(wt.RepoID)
					repoName := ""
					if err == nil {
						repoName = repo.Name
					}
					cwt = &wtInfo{wt.ID, wt.Branch, repoName}
				}
			}
			items = append(items, agentDetail{
				ID:              a.ID,
				Name:            a.Name,
				Type:            a.Type,
				Status:          a.Status,
				LastSeen:        a.LastSeen.Format("2006-01-02T15:04:05Z"),
				CurrentWorktree: cwt,
			})
		}

		return textResource("agit://agents", "application/json", items)
	}
}
