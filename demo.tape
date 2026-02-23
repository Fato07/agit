# agit demo — Infrastructure-aware Git orchestration for AI agents
# Record with: vhs demo.tape

Output docs/demo.gif
Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 650
Set TypingSpeed 50ms
Set Theme "Catppuccin Mocha"

# --- Hidden setup: create isolated temp environment ---
Hide

Type "export PATH=$HOME/go/bin:$PATH"
Enter
Sleep 300ms

Type "export DEMO_DIR=$(mktemp -d)"
Enter
Sleep 500ms

Type "cd $DEMO_DIR"
Enter
Sleep 300ms

Type "git init demo-app"
Enter
Sleep 500ms

Type "cd demo-app"
Enter
Sleep 300ms

Type "git checkout -b main"
Enter
Sleep 300ms

Type "mkdir -p src"
Enter
Sleep 200ms

Type@100ms "echo 'package main' > src/main.go"
Enter
Sleep 200ms

Type@100ms "echo 'package auth' > src/auth.go"
Enter
Sleep 200ms

Type "git add -A"
Enter
Sleep 200ms

Type@100ms "git commit -m 'initial commit'"
Enter
Sleep 500ms

Type "clear"
Enter
Sleep 500ms

Show

# --- Demo starts ---
Sleep 1s
Type "# agit: Infrastructure-aware Git orchestration for AI agents"
Enter
Sleep 1.5s

# Step 1: Initialize agit
Type "agit init"
Enter
Sleep 2s

# Step 2: Register the repository
Type "agit add . --name demo-app"
Enter
Sleep 2s

# Step 3: List registered repos
Type "agit repos"
Enter
Sleep 2s

# Step 4: Spawn worktree for first agent
Type@40ms "agit spawn demo-app --task 'refactor auth' --agent claude-1"
Enter
Sleep 3.5s

# Step 5: Spawn worktree for second agent
Type@40ms "agit spawn demo-app --task 'add tests' --agent cursor-1"
Enter
Sleep 3.5s

# Step 6: Check status
Type "agit status demo-app"
Enter
Sleep 4s

# Step 7: Create overlapping changes in both worktrees (simulating agents working)
Hide

Type "W1=$(ls -d .worktrees/agit-*/ | head -1)"
Enter
Sleep 300ms

Type "W2=$(ls -d .worktrees/agit-*/ | tail -1)"
Enter
Sleep 300ms

Type@100ms "echo 'package auth // refactored' > $W1/src/auth.go"
Enter
Sleep 200ms

Type "cd $W1 && git add -A && git commit -m 'refactor auth' && cd -"
Enter
Sleep 500ms

Type@100ms "echo 'package auth // tested' > $W2/src/auth.go"
Enter
Sleep 200ms

Type "cd $W2 && git add -A && git commit -m 'add tests' && cd -"
Enter
Sleep 500ms

Type "clear"
Enter
Sleep 500ms

Show

Type "# Both agents edited src/auth.go — check for conflicts"
Enter
Sleep 1.5s

# Step 8: Detect conflicts
Type "agit conflicts demo-app"
Enter
Sleep 4s

# Step 9: Create a task with priority
Type "agit tasks demo-app --create 'fix auth bug' --priority 2"
Enter
Sleep 3s

# Step 10: Atomic task dispatch
Type "agit tasks next demo-app --agent claude-1"
Enter
Sleep 3s

# Step 11: List tasks
Type "agit tasks demo-app"
Enter
Sleep 3s

# Step 12: Cleanup
Type "agit cleanup"
Enter
Sleep 3s

# Final
Sleep 1s
Type "# Done! agit keeps your AI agents coordinated."
Enter
Sleep 3s
