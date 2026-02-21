package conflicts

import (
	"testing"

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

func TestFindConflictsViaDB(t *testing.T) {
	db := mustDB(t)

	repo, _ := db.AddRepo("conflict-repo", "/tmp/cr", "", "main")
	wt1, _ := db.CreateWorktree(repo.ID, "/tmp/cr1", "b1", nil, nil)
	wt2, _ := db.CreateWorktree(repo.ID, "/tmp/cr2", "b2", nil, nil)

	// Both touch the same file
	db.RecordFileTouches(repo.ID, wt1.ID, []registry.FileTouch{
		{FilePath: "shared.go", ChangeType: "modified"},
		{FilePath: "unique1.go", ChangeType: "added"},
	})
	db.RecordFileTouches(repo.ID, wt2.ID, []registry.FileTouch{
		{FilePath: "shared.go", ChangeType: "modified"},
		{FilePath: "unique2.go", ChangeType: "added"},
	})

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}

	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].FilePath != "shared.go" {
		t.Errorf("expected shared.go, got %s", conflicts[0].FilePath)
	}
}

func TestFindConflictsNoOverlap(t *testing.T) {
	db := mustDB(t)

	repo, _ := db.AddRepo("no-conflict", "/tmp/nc", "", "main")
	wt1, _ := db.CreateWorktree(repo.ID, "/tmp/nc1", "b1", nil, nil)
	wt2, _ := db.CreateWorktree(repo.ID, "/tmp/nc2", "b2", nil, nil)

	db.RecordFileTouches(repo.ID, wt1.ID, []registry.FileTouch{
		{FilePath: "a.go", ChangeType: "modified"},
	})
	db.RecordFileTouches(repo.ID, wt2.ID, []registry.FileTouch{
		{FilePath: "b.go", ChangeType: "modified"},
	})

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestFindConflictsEmpty(t *testing.T) {
	db := mustDB(t)

	repo, _ := db.AddRepo("empty-repo", "/tmp/er", "", "main")

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
}

func TestFindConflictsCompletedWorktreeExcluded(t *testing.T) {
	db := mustDB(t)

	repo, _ := db.AddRepo("excl-repo", "/tmp/ex", "", "main")
	wt1, _ := db.CreateWorktree(repo.ID, "/tmp/ex1", "b1", nil, nil)
	wt2, _ := db.CreateWorktree(repo.ID, "/tmp/ex2", "b2", nil, nil)

	// Both touch same file
	db.RecordFileTouches(repo.ID, wt1.ID, []registry.FileTouch{
		{FilePath: "file.go", ChangeType: "modified"},
	})
	db.RecordFileTouches(repo.ID, wt2.ID, []registry.FileTouch{
		{FilePath: "file.go", ChangeType: "modified"},
	})

	// Mark one as completed
	db.UpdateWorktreeStatus(wt2.ID, "completed")

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}
	// Only active worktrees should be considered
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts (one worktree completed), got %d", len(conflicts))
	}
}

func TestFindConflictsMultipleFiles(t *testing.T) {
	db := mustDB(t)

	repo, _ := db.AddRepo("multi-repo", "/tmp/mr", "", "main")
	wt1, _ := db.CreateWorktree(repo.ID, "/tmp/mr1", "b1", nil, nil)
	wt2, _ := db.CreateWorktree(repo.ID, "/tmp/mr2", "b2", nil, nil)

	db.RecordFileTouches(repo.ID, wt1.ID, []registry.FileTouch{
		{FilePath: "a.go", ChangeType: "modified"},
		{FilePath: "b.go", ChangeType: "modified"},
	})
	db.RecordFileTouches(repo.ID, wt2.ID, []registry.FileTouch{
		{FilePath: "a.go", ChangeType: "modified"},
		{FilePath: "b.go", ChangeType: "added"},
	})

	conflicts, err := db.FindConflicts(repo.ID)
	if err != nil {
		t.Fatalf("FindConflicts: %v", err)
	}
	if len(conflicts) != 2 {
		t.Errorf("expected 2 conflicts, got %d", len(conflicts))
	}
}
