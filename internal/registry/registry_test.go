package registry

import (
	"fmt"
	"testing"
	"time"
)

func mustOpenMemory(t *testing.T) *DB {
	t.Helper()
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// --- Repos ---

func TestAddAndGetRepo(t *testing.T) {
	db := mustOpenMemory(t)

	repo, err := db.AddRepo("myrepo", "/tmp/myrepo", "https://github.com/test/repo", "main")
	if err != nil {
		t.Fatalf("AddRepo: %v", err)
	}
	if repo.Name != "myrepo" {
		t.Errorf("expected name myrepo, got %s", repo.Name)
	}

	got, err := db.GetRepo("myrepo")
	if err != nil {
		t.Fatalf("GetRepo: %v", err)
	}
	if got.ID != repo.ID {
		t.Errorf("ID mismatch: %s vs %s", got.ID, repo.ID)
	}
	if got.Path != "/tmp/myrepo" {
		t.Errorf("path mismatch: %s", got.Path)
	}
}

func TestGetRepoNotFound(t *testing.T) {
	db := mustOpenMemory(t)

	_, err := db.GetRepo("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
}

func TestListRepos(t *testing.T) {
	db := mustOpenMemory(t)

	db.AddRepo("alpha", "/tmp/a", "", "main")
	db.AddRepo("beta", "/tmp/b", "", "main")

	repos, err := db.ListRepos()
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	// Should be ordered by name
	if repos[0].Name != "alpha" {
		t.Errorf("expected first repo alpha, got %s", repos[0].Name)
	}
}

func TestRemoveRepo(t *testing.T) {
	db := mustOpenMemory(t)

	db.AddRepo("todelete", "/tmp/td", "", "main")

	if err := db.RemoveRepo("todelete"); err != nil {
		t.Fatalf("RemoveRepo: %v", err)
	}

	_, err := db.GetRepo("todelete")
	if err == nil {
		t.Fatal("repo should be gone")
	}
}

func TestRemoveRepoNotFound(t *testing.T) {
	db := mustOpenMemory(t)

	if err := db.RemoveRepo("nope"); err == nil {
		t.Fatal("expected error for missing repo")
	}
}

func TestGetRepoStats(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("stats", "/tmp/s", "", "main")
	db.CreateWorktree(repo.ID, "/tmp/wt1", "branch1", nil, nil)
	db.CreateTask(repo.ID, "task1", 0)

	stats, err := db.GetRepoStats(repo.ID)
	if err != nil {
		t.Fatalf("GetRepoStats: %v", err)
	}
	if stats.ActiveWorktrees != 1 {
		t.Errorf("expected 1 active worktree, got %d", stats.ActiveWorktrees)
	}
	if stats.PendingTasks != 1 {
		t.Errorf("expected 1 pending task, got %d", stats.PendingTasks)
	}
}

// --- Worktrees ---

func TestCreateAndGetWorktree(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("wr", "/tmp/wr", "", "main")

	taskDesc := "test task"
	wt, err := db.CreateWorktree(repo.ID, "/tmp/wt", "feat-1", nil, &taskDesc)
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if wt.Status != "active" {
		t.Errorf("expected active status, got %s", wt.Status)
	}

	got, err := db.GetWorktree(wt.ID)
	if err != nil {
		t.Fatalf("GetWorktree: %v", err)
	}
	if got.Branch != "feat-1" {
		t.Errorf("expected branch feat-1, got %s", got.Branch)
	}
}

func TestListWorktrees(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("lw", "/tmp/lw", "", "main")
	db.CreateWorktree(repo.ID, "/tmp/lw1", "b1", nil, nil)
	db.CreateWorktree(repo.ID, "/tmp/lw2", "b2", nil, nil)

	wts, err := db.ListWorktrees(repo.ID, nil)
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}

	// Filter by status
	active := "active"
	wts, _ = db.ListWorktrees(repo.ID, &active)
	if len(wts) != 2 {
		t.Fatalf("expected 2 active worktrees, got %d", len(wts))
	}

	completed := "completed"
	wts, _ = db.ListWorktrees(repo.ID, &completed)
	if len(wts) != 0 {
		t.Fatalf("expected 0 completed worktrees, got %d", len(wts))
	}
}

func TestUpdateWorktreeStatus(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("us", "/tmp/us", "", "main")
	wt, _ := db.CreateWorktree(repo.ID, "/tmp/us1", "b1", nil, nil)

	if err := db.UpdateWorktreeStatus(wt.ID, "completed"); err != nil {
		t.Fatalf("UpdateWorktreeStatus: %v", err)
	}

	got, _ := db.GetWorktree(wt.ID)
	if got.Status != "completed" {
		t.Errorf("expected completed, got %s", got.Status)
	}
}

func TestDeleteWorktree(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("dw", "/tmp/dw", "", "main")
	wt, _ := db.CreateWorktree(repo.ID, "/tmp/dw1", "b1", nil, nil)

	if err := db.DeleteWorktree(wt.ID); err != nil {
		t.Fatalf("DeleteWorktree: %v", err)
	}

	_, err := db.GetWorktree(wt.ID)
	if err == nil {
		t.Fatal("worktree should be deleted")
	}
}

func TestPruneOrphanedWorktrees(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("po", "/tmp/po", "", "main")
	// Create worktree pointing to a non-existent directory
	db.CreateWorktree(repo.ID, "/tmp/nonexistent-path-xyz", "b1", nil, nil)

	count, err := db.PruneOrphanedWorktrees(repo.ID)
	if err != nil {
		t.Fatalf("PruneOrphanedWorktrees: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 pruned, got %d", count)
	}
}

func TestFindWorktreeByPrefix(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("fp", "/tmp/fp", "", "main")
	wt, _ := db.CreateWorktree(repo.ID, "/tmp/fp1", "b1", nil, nil)

	// Use first 8 chars of ID
	prefix := wt.ID[:8]
	got, err := db.FindWorktreeByPrefix(repo.ID, prefix)
	if err != nil {
		t.Fatalf("FindWorktreeByPrefix: %v", err)
	}
	if got.ID != wt.ID {
		t.Errorf("ID mismatch")
	}

	// Too short prefix
	_, err = db.FindWorktreeByPrefix(repo.ID, "ab")
	if err == nil {
		t.Fatal("expected error for short prefix")
	}
}

func TestResolveWorktree(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("rw", "/tmp/rw", "", "main")
	wt, _ := db.CreateWorktree(repo.ID, "/tmp/rw1", "b1", nil, nil)

	// Exact ID
	got, err := db.ResolveWorktree(repo.ID, wt.ID)
	if err != nil {
		t.Fatalf("ResolveWorktree exact: %v", err)
	}
	if got.ID != wt.ID {
		t.Errorf("ID mismatch")
	}

	// Prefix
	got, err = db.ResolveWorktree(repo.ID, wt.ID[:8])
	if err != nil {
		t.Fatalf("ResolveWorktree prefix: %v", err)
	}
	if got.ID != wt.ID {
		t.Errorf("ID mismatch")
	}
}

// --- Agents ---

func TestRegisterAndGetAgent(t *testing.T) {
	db := mustOpenMemory(t)

	agent, err := db.RegisterAgent("claude-1", "claude")
	if err != nil {
		t.Fatalf("RegisterAgent: %v", err)
	}
	if agent.Name != "claude-1" {
		t.Errorf("expected name claude-1, got %s", agent.Name)
	}
	if agent.Status != "active" {
		t.Errorf("expected active status, got %s", agent.Status)
	}

	got, err := db.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("GetAgent: %v", err)
	}
	if got.Type != "claude" {
		t.Errorf("expected type claude, got %s", got.Type)
	}
}

func TestListAgents(t *testing.T) {
	db := mustOpenMemory(t)

	db.RegisterAgent("agent-a", "custom")
	db.RegisterAgent("agent-b", "custom")

	agents, err := db.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
}

func TestHeartbeat(t *testing.T) {
	db := mustOpenMemory(t)

	agent, _ := db.RegisterAgent("hb", "custom")
	time.Sleep(10 * time.Millisecond)

	if err := db.Heartbeat(agent.ID); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}

	got, _ := db.GetAgent(agent.ID)
	if got.LastSeen.Before(agent.LastSeen) {
		t.Error("last_seen should be updated")
	}
}

func TestHeartbeatNotFound(t *testing.T) {
	db := mustOpenMemory(t)

	if err := db.Heartbeat("nonexistent"); err == nil {
		t.Fatal("expected error for missing agent")
	}
}

func TestSweepStaleAgents(t *testing.T) {
	db := mustOpenMemory(t)

	db.RegisterAgent("stale", "custom")
	// Manually backdate last_seen
	db.conn.Exec("UPDATE agents SET last_seen = ? WHERE name = 'stale'",
		time.Now().Add(-10*time.Minute))

	count, err := db.SweepStaleAgents(5 * time.Minute)
	if err != nil {
		t.Fatalf("SweepStaleAgents: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 swept, got %d", count)
	}

	agent, _ := db.GetAgentByName("stale")
	if agent.Status != "disconnected" {
		t.Errorf("expected disconnected, got %s", agent.Status)
	}
}

func TestRemoveAgent(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("ra", "/tmp/ra", "", "main")
	agent, _ := db.RegisterAgent("to-remove", "custom")

	// Create task and claim it
	task, _ := db.CreateTask(repo.ID, "test task", 0)
	db.ClaimTask(task.ID, agent.ID)

	if err := db.RemoveAgent("to-remove"); err != nil {
		t.Fatalf("RemoveAgent: %v", err)
	}

	// Task should be unclaimed
	got, _ := db.GetTask(task.ID)
	if got.Status != "pending" {
		t.Errorf("expected pending after agent removal, got %s", got.Status)
	}

	// Agent should be gone
	a, _ := db.GetAgentByName("to-remove")
	if a != nil {
		t.Error("agent should be deleted")
	}
}

// --- Tasks ---

func TestCreateAndGetTask(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("tt", "/tmp/tt", "", "main")

	task, err := db.CreateTask(repo.ID, "implement feature", 5)
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if task.Status != "pending" {
		t.Errorf("expected pending, got %s", task.Status)
	}
	if task.Priority != 5 {
		t.Errorf("expected priority 5, got %d", task.Priority)
	}

	got, err := db.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Description != "implement feature" {
		t.Errorf("expected 'implement feature', got %s", got.Description)
	}
}

func TestClaimTask(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("ct", "/tmp/ct", "", "main")
	agent, _ := db.RegisterAgent("claimer", "custom")
	task, _ := db.CreateTask(repo.ID, "claim me", 0)

	if err := db.ClaimTask(task.ID, agent.ID); err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}

	got, _ := db.GetTask(task.ID)
	if got.Status != "claimed" {
		t.Errorf("expected claimed, got %s", got.Status)
	}

	// Double claim should fail
	if err := db.ClaimTask(task.ID, agent.ID); err == nil {
		t.Fatal("expected error for double claim")
	}
}

func TestStartTask(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("st", "/tmp/st", "", "main")
	agent, _ := db.RegisterAgent("starter", "custom")
	task, _ := db.CreateTask(repo.ID, "start me", 0)
	wt, _ := db.CreateWorktree(repo.ID, "/tmp/st1", "b1", nil, nil)

	db.ClaimTask(task.ID, agent.ID)

	if err := db.StartTask(task.ID, wt.ID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}

	got, _ := db.GetTask(task.ID)
	if got.Status != "in_progress" {
		t.Errorf("expected in_progress, got %s", got.Status)
	}
}

func TestCompleteTask(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("cot", "/tmp/cot", "", "main")
	task, _ := db.CreateTask(repo.ID, "complete me", 0)

	result := "all done"
	if err := db.CompleteTask(task.ID, &result); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	got, _ := db.GetTask(task.ID)
	if got.Status != "completed" {
		t.Errorf("expected completed, got %s", got.Status)
	}
	if got.Result == nil || *got.Result != "all done" {
		t.Error("result not saved")
	}
}

func TestFailTask(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("ft", "/tmp/ft", "", "main")
	task, _ := db.CreateTask(repo.ID, "fail me", 0)

	reason := "broken"
	if err := db.FailTask(task.ID, &reason); err != nil {
		t.Fatalf("FailTask: %v", err)
	}

	got, _ := db.GetTask(task.ID)
	if got.Status != "failed" {
		t.Errorf("expected failed, got %s", got.Status)
	}
}

func TestListTasks(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("lt", "/tmp/lt", "", "main")
	db.CreateTask(repo.ID, "task 1", 0)
	db.CreateTask(repo.ID, "task 2", 5)

	tasks, err := db.ListTasks(repo.ID, nil)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	// Filter by status
	pending := "pending"
	tasks, _ = db.ListTasks(repo.ID, &pending)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", len(tasks))
	}
}

// --- File Touches & Conflicts ---

func TestRecordFileTouchesAndFindConflicts(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("fc", "/tmp/fc", "", "main")
	wt1, _ := db.CreateWorktree(repo.ID, "/tmp/fc1", "b1", nil, nil)
	wt2, _ := db.CreateWorktree(repo.ID, "/tmp/fc2", "b2", nil, nil)

	// Both worktrees touch the same file
	touches1 := []FileTouch{
		{FilePath: "main.go", ChangeType: "modified"},
		{FilePath: "readme.md", ChangeType: "modified"},
	}
	touches2 := []FileTouch{
		{FilePath: "main.go", ChangeType: "modified"},
		{FilePath: "other.go", ChangeType: "added"},
	}

	if err := db.RecordFileTouches(repo.ID, wt1.ID, touches1); err != nil {
		t.Fatalf("RecordFileTouches wt1: %v", err)
	}
	if err := db.RecordFileTouches(repo.ID, wt2.ID, touches2); err != nil {
		t.Fatalf("RecordFileTouches wt2: %v", err)
	}

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}

	// Only main.go should conflict
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].FilePath != "main.go" {
		t.Errorf("expected main.go conflict, got %s", conflicts[0].FilePath)
	}
	if len(conflicts[0].Worktrees) != 2 {
		t.Errorf("expected 2 worktrees in conflict, got %d", len(conflicts[0].Worktrees))
	}
}

func TestFindConflictsNoOverlap(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("nc", "/tmp/nc", "", "main")
	wt1, _ := db.CreateWorktree(repo.ID, "/tmp/nc1", "b1", nil, nil)
	wt2, _ := db.CreateWorktree(repo.ID, "/tmp/nc2", "b2", nil, nil)

	db.RecordFileTouches(repo.ID, wt1.ID, []FileTouch{{FilePath: "a.go", ChangeType: "modified"}})
	db.RecordFileTouches(repo.ID, wt2.ID, []FileTouch{{FilePath: "b.go", ChangeType: "modified"}})

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestFindConflictsEmpty(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("ec", "/tmp/ec", "", "main")

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

// --- NextTask ---

func TestNextTaskHighestPriority(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("nt", "/tmp/nt", "", "main")
	agent, _ := db.RegisterAgent("nt-agent", "custom")

	// Create tasks with different priorities
	db.CreateTask(repo.ID, "low priority", 1)
	db.CreateTask(repo.ID, "high priority", 10)
	db.CreateTask(repo.ID, "medium priority", 5)

	task, err := db.NextTask(repo.ID, agent.ID)
	if err != nil {
		t.Fatalf("NextTask: %v", err)
	}
	if task == nil {
		t.Fatal("expected a task, got nil")
	}
	if task.Priority != 10 {
		t.Errorf("expected priority 10 (highest), got %d", task.Priority)
	}
	if task.Description != "high priority" {
		t.Errorf("expected 'high priority', got %s", task.Description)
	}
	if task.Status != "claimed" {
		t.Errorf("expected claimed status, got %s", task.Status)
	}
}

func TestNextTaskFIFOTiebreak(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("fifo", "/tmp/fifo", "", "main")
	agent, _ := db.RegisterAgent("fifo-agent", "custom")

	// Create tasks with same priority — FIFO by created_at
	first, _ := db.CreateTask(repo.ID, "first created", 5)
	db.CreateTask(repo.ID, "second created", 5)

	task, err := db.NextTask(repo.ID, agent.ID)
	if err != nil {
		t.Fatalf("NextTask: %v", err)
	}
	if task == nil {
		t.Fatal("expected a task")
	}
	if task.ID != first.ID {
		t.Errorf("expected first-created task %s, got %s", first.ID, task.ID)
	}
}

func TestNextTaskNoPending(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("np", "/tmp/np", "", "main")
	agent, _ := db.RegisterAgent("np-agent", "custom")

	task, err := db.NextTask(repo.ID, agent.ID)
	if err != nil {
		t.Fatalf("NextTask: %v", err)
	}
	if task != nil {
		t.Errorf("expected nil task, got %+v", task)
	}
}

func TestNextTaskConcurrentClaim(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("cc", "/tmp/cc", "", "main")
	agent1, _ := db.RegisterAgent("agent-1", "custom")
	agent2, _ := db.RegisterAgent("agent-2", "custom")

	// Create 2 tasks
	db.CreateTask(repo.ID, "task A", 10)
	db.CreateTask(repo.ID, "task B", 5)

	// Both agents claim — should get different tasks
	t1, err := db.NextTask(repo.ID, agent1.ID)
	if err != nil {
		t.Fatalf("NextTask agent1: %v", err)
	}
	t2, err := db.NextTask(repo.ID, agent2.ID)
	if err != nil {
		t.Fatalf("NextTask agent2: %v", err)
	}

	if t1 == nil || t2 == nil {
		t.Fatal("both agents should get a task")
	}
	if t1.ID == t2.ID {
		t.Errorf("agents should claim different tasks, both got %s", t1.ID)
	}
}

func TestRecordFileTouchesReplace(t *testing.T) {
	db := mustOpenMemory(t)

	repo, _ := db.AddRepo("rft", "/tmp/rft", "", "main")
	wt, _ := db.CreateWorktree(repo.ID, "/tmp/rft1", "b1", nil, nil)

	// Record initial touches
	db.RecordFileTouches(repo.ID, wt.ID, []FileTouch{
		{FilePath: "old.go", ChangeType: "modified"},
	})

	// Replace with new touches
	db.RecordFileTouches(repo.ID, wt.ID, []FileTouch{
		{FilePath: "new.go", ChangeType: "added"},
	})

	// Create another worktree touching old.go - should not conflict
	wt2, _ := db.CreateWorktree(repo.ID, "/tmp/rft2", "b2", nil, nil)
	db.RecordFileTouches(repo.ID, wt2.ID, []FileTouch{
		{FilePath: "old.go", ChangeType: "modified"},
	})

	conflicts, _ := db.FindConflicts(repo.ID)
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts after replace, got %d", len(conflicts))
	}
}

func BenchmarkNextTask(b *testing.B) {
	db, err := OpenMemory()
	if err != nil {
		b.Fatalf("OpenMemory: %v", err)
	}
	defer db.Close()

	repo, _ := db.AddRepo("bench-repo", "/tmp/bench", "", "main")
	agent, _ := db.RegisterAgent("bench-agent", "custom")

	// Insert 1,000 tasks with varying priorities
	for i := 0; i < 1000; i++ {
		priority := i % 10
		db.CreateTask(repo.ID, fmt.Sprintf("task-%d", i), priority)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Re-create pending tasks for each iteration if depleted
		task, _ := db.NextTask(repo.ID, agent.ID)
		if task == nil {
			// All claimed; reset by creating more
			b.StopTimer()
			for j := 0; j < 100; j++ {
				db.CreateTask(repo.ID, fmt.Sprintf("refill-%d-%d", i, j), j%10)
			}
			b.StartTimer()
		}
	}
}
