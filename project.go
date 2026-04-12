package main

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Project holds analysis results for a single Go project.
type Project struct {
	Path         string
	ModuleName   string
	OldestMod    time.Time
	DaysAgo      int
	IsGit        bool
	RemoteOrigin string
	LOC          map[string]int // extension -> line count
}

// LOCEntry represents a single extension line count.
type LOCEntry struct {
	Ext   string
	Lines int
}

// LOCSorted returns the top 5 LOC entries sorted by line count descending.
func (p *Project) LOCSorted() []LOCEntry {
	var entries []LOCEntry
	for ext, lines := range p.LOC {
		entries = append(entries, LOCEntry{ext, lines})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Lines > entries[j].Lines
	})
	if len(entries) > 5 {
		entries = entries[:5]
	}
	return entries
}

func analyzeProject(p *Project) {
	p.LOC = make(map[string]int)
	now := time.Now()
	var oldest time.Time

	_ = filepath.WalkDir(p.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()

		// Skip hidden directories and common non-source dirs
		if d.IsDir() && strings.HasPrefix(name, ".") {
			return filepath.SkipDir
		}
		if d.IsDir() && (name == "vendor" || name == "node_modules") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Track most recent modification time
		mod := info.ModTime()
		if oldest.IsZero() || mod.After(oldest) {
			oldest = mod
		}

		// Skip test files, dotfiles, and Go module files for LOC counting
		if strings.HasSuffix(name, "_test.go") {
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}
		if name == "go.mod" || name == "go.sum" {
			return nil
		}

		ext := filepath.Ext(name)
		if ext == "" {
			return nil
		}

		if !isTextFile(path) {
			return nil
		}

		lines, err := countLines(path)
		if err != nil {
			return nil
		}
		p.LOC[ext] += lines

		return nil
	})

	if !oldest.IsZero() {
		p.OldestMod = oldest
		p.DaysAgo = int(math.Floor(now.Sub(oldest).Hours() / 24))
	}

	// Check for git
	gitDir := filepath.Join(p.Path, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		p.IsGit = true
		p.RemoteOrigin = gitRemoteOrigin(p.Path)
	}

	// Read module name from go.mod
	p.ModuleName = goModuleName(p.Path)
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

// isTextFile reads the first 512 bytes and returns false if any null bytes are found.
func isTextFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return false
		}
	}
	return n > 0
}

func goModuleName(dir string) string {
	f, err := os.Open(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}
	return ""
}



func gitRemoteOrigin(dir string) string {
	configPath := filepath.Join(dir, ".git", "config")
	f, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	inOrigin := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == `[remote "origin"]` {
			inOrigin = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inOrigin = false
			continue
		}
		if inOrigin && strings.HasPrefix(line, "url = ") {
			return strings.TrimPrefix(line, "url = ")
		}
	}
	return ""
}

// gitRemoteToURL converts a git remote URL to a clickable HTTPS URL.
// SSH-style URLs like git@github.com:user/repo.git are translated.
func gitRemoteToURL(remote string) string {
	// Already an HTTP(S) URL
	if strings.HasPrefix(remote, "https://") || strings.HasPrefix(remote, "http://") {
		return strings.TrimSuffix(remote, ".git")
	}
	// SSH style: git@github.com:user/repo.git
	if strings.Contains(remote, "@") && strings.Contains(remote, ":") {
		// Split at @, then at :
		parts := strings.SplitN(remote, "@", 2)
		if len(parts) == 2 {
			hostAndPath := parts[1]
			colonIdx := strings.Index(hostAndPath, ":")
			if colonIdx > 0 {
				host := hostAndPath[:colonIdx]
				path := hostAndPath[colonIdx+1:]
				path = strings.TrimSuffix(path, ".git")
				return "https://" + host + "/" + path
			}
		}
	}
	return remote
}

