package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: devhelper [dir1] [dir2] ...\n")
	fmt.Fprintf(os.Stderr, "  Starts an HTTP server on port 9990\n")
	fmt.Fprintf(os.Stderr, "  If no directories are provided, defaults to ~/projects and ~/golang\n")
	os.Exit(1)
}

func main() {
	args := os.Args[1:]

	var scanDirs []string
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			usage()
		}
		scanDirs = append(scanDirs, arg)
	}

	// Default to ~/projects and ~/golang if no directories provided
	if len(scanDirs) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: unable to determine home directory: %v\n", err)
			os.Exit(1)
		}
		scanDirs = []string{
			filepath.Join(home, "projects"),
			filepath.Join(home, "golang"),
		}
	}

	// Resolve paths and filter to directories that exist
	var validDirs []string
	for _, dir := range scanDirs {
		abs, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot resolve %q: %v\n", dir, err)
			continue
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Warning: skipping %q (not a valid directory)\n", dir)
			continue
		}
		validDirs = append(validDirs, abs)
	}

	if len(validDirs) == 0 {
		fmt.Fprintf(os.Stderr, "Error: none of the provided scan directories exist\n")
		os.Exit(1)
	}

	home, _ := os.UserHomeDir()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		projects := findProjects(validDirs)
		for i := range projects {
			analyzeProject(&projects[i])
		}
		sortProjects(projects)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		writeHTML(w, projects, home)
	})

	http.HandleFunc("/open", func(w http.ResponseWriter, r *http.Request) {
		projectPath := r.URL.Query().Get("path")
		if projectPath == "" {
			http.Error(w, "missing path parameter", http.StatusBadRequest)
			return
		}
		info, err := os.Stat(projectPath)
		if err != nil || !info.IsDir() {
			http.Error(w, "invalid project path", http.StatusBadRequest)
			return
		}
		ide := r.URL.Query().Get("ide")
		if ide == "" {
			ide = "goland"
		}
		if err := exec.Command(ide, projectPath).Start(); err != nil {
			http.Error(w, "failed to launch "+ide+": "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	fmt.Println("Listening on http://localhost:9990")
	if err := http.ListenAndServe(":9990", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
