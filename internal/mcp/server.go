package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/fathindos/agit/internal/config"
	"github.com/fathindos/agit/internal/registry"
)

// NewServer creates a configured MCP server with all tools and resources registered.
func NewServer(db *registry.DB, cfg *config.Config) *server.MCPServer {
	s := server.NewMCPServer(
		"agit",
		"0.1.0",
		server.WithResourceCapabilities(true, true),
	)

	registerTools(s, db, cfg)
	registerResources(s, db)

	return s
}

func registerTools(s *server.MCPServer, db *registry.DB, cfg *config.Config) {
	s.AddTool(
		mcp.NewTool("agit_list_repos",
			mcp.WithDescription("List all registered repositories"),
		),
		handleListRepos(db),
	)

	s.AddTool(
		mcp.NewTool("agit_repo_status",
			mcp.WithDescription("Get detailed status for a specific repository"),
			mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
		),
		handleRepoStatus(db),
	)

	s.AddTool(
		mcp.NewTool("agit_spawn_worktree",
			mcp.WithDescription("Create an isolated worktree for an agent"),
			mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
			mcp.WithString("task", mcp.Description("Task description")),
			mcp.WithString("branch", mcp.Description("Custom branch name (auto-generated if omitted)")),
			mcp.WithString("agent", mcp.Description("Agent name to assign")),
		),
		handleSpawnWorktree(db, cfg),
	)

	s.AddTool(
		mcp.NewTool("agit_remove_worktree",
			mcp.WithDescription("Remove a worktree from disk and registry"),
			mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
			mcp.WithString("worktree_id", mcp.Required(), mcp.Description("Worktree ID to remove")),
		),
		handleRemoveWorktree(db),
	)

	s.AddTool(
		mcp.NewTool("agit_check_conflicts",
			mcp.WithDescription("Scan for file conflicts across active worktrees"),
			mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
		),
		handleCheckConflicts(db),
	)

	s.AddTool(
		mcp.NewTool("agit_list_tasks",
			mcp.WithDescription("List tasks for a repository"),
			mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
			mcp.WithString("status", mcp.Description("Filter by status (pending/claimed/in_progress/completed/failed)")),
		),
		handleListTasks(db),
	)

	s.AddTool(
		mcp.NewTool("agit_claim_task",
			mcp.WithDescription("Atomically claim a pending task for an agent"),
			mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to claim")),
			mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID claiming the task")),
		),
		handleClaimTask(db),
	)

	s.AddTool(
		mcp.NewTool("agit_complete_task",
			mcp.WithDescription("Mark a task as completed with optional result"),
			mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to complete")),
			mcp.WithString("result", mcp.Description("Result description")),
		),
		handleCompleteTask(db),
	)

	s.AddTool(
		mcp.NewTool("agit_merge_worktree",
			mcp.WithDescription("Merge a worktree branch into the default branch, then auto-cleanup"),
			mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
			mcp.WithString("worktree_id", mcp.Required(), mcp.Description("Worktree ID to merge")),
		),
		handleMergeWorktree(db),
	)

	s.AddTool(
		mcp.NewTool("agit_register_agent",
			mcp.WithDescription("Register a new AI agent"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Agent name")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Agent type (e.g., claude, custom)")),
		),
		handleRegisterAgent(db),
	)

	s.AddTool(
		mcp.NewTool("agit_heartbeat",
			mcp.WithDescription("Update agent heartbeat timestamp"),
			mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent ID")),
		),
		handleHeartbeat(db),
	)
}

func registerResources(s *server.MCPServer, db *registry.DB) {
	// Static resources
	s.AddResource(
		mcp.NewResource("agit://repos", "Registered repositories",
			mcp.WithMIMEType("application/json"),
		),
		handleReposResource(db),
	)

	s.AddResource(
		mcp.NewResource("agit://agents", "Registered agents",
			mcp.WithMIMEType("application/json"),
		),
		handleAgentsResource(db),
	)

	// Template resources
	s.AddResourceTemplate(
		mcp.NewResourceTemplate("agit://repos/{name}", "Repository details",
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleRepoDetailResource(db),
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate("agit://repos/{name}/conflicts", "Repository conflicts",
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleRepoConflictsResource(db),
	)

	s.AddResourceTemplate(
		mcp.NewResourceTemplate("agit://repos/{name}/tasks", "Repository tasks",
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleRepoTasksResource(db),
	)
}
