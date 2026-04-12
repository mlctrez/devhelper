package main

import (
	"os"
	"path/filepath"
	"sort"
)

// findProjects walks each scan directory looking for project roots.
// A .git directory takes priority; otherwise falls back to go.mod.
// Stops descending once either marker is found.
func findProjects(scanDirs []string) []Project {
	seen := make(map[string]bool)
	var projects []Project

	for _, root := range scanDirs {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			hasGit := false
			if info, err := os.Stat(filepath.Join(path, ".git")); err == nil && info.IsDir() {
				hasGit = true
			}
			hasGoMod := false
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				hasGoMod = true
			}
			if hasGit || hasGoMod {
				abs, _ := filepath.Abs(path)
				if !seen[abs] {
					seen[abs] = true
					projects = append(projects, Project{Path: abs})
				}
				return filepath.SkipDir
			}
			return nil
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Path < projects[j].Path
	})
	return projects
}

// sortProjects sorts by most recently modified first.
func sortProjects(projects []Project) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].DaysAgo < projects[j].DaysAgo
	})
}
