package cmd

import (
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/registry"
)

// completeRepoNames provides dynamic completion for repository names.
func completeRepoNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	db, err := registry.Open()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer db.Close()

	repos, err := db.ListRepos()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, r := range repos {
		names = append(names, r.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeWorktreeIDs provides dynamic completion for worktree short IDs.
func completeWorktreeIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	db, err := registry.Open()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer db.Close()

	worktrees, err := db.ListAllActiveWorktrees()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var ids []string
	for _, wt := range worktrees {
		id := wt.ID
		if len(id) > 12 {
			id = id[:12]
		}
		ids = append(ids, id)
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

// completeAgentNames provides dynamic completion for agent names.
func completeAgentNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	db, err := registry.Open()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer db.Close()

	agents, err := db.ListAgents()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, a := range agents {
		names = append(names, a.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
