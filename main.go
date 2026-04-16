package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: devhelper [command] [dir1] [dir2] ...\n")
	fmt.Fprintf(os.Stderr, "  Starts an HTTP server on port 9990\n")
	fmt.Fprintf(os.Stderr, "  If no directories are provided, defaults to ~/projects and ~/golang\n")
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  install     Install desktop entry and icon\n")
	fmt.Fprintf(os.Stderr, "  uninstall   Remove desktop entry and icon\n")
	os.Exit(1)
}

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "install":
			runInstall()
			return
		case "uninstall":
			runUninstall()
			return
		}
	}

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

	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "shutting down")
		go func() {
			time.Sleep(100 * time.Millisecond)
			os.Exit(0)
		}()
	})

	// Check if another instance is already running
	if isAlreadyRunning() {
		fmt.Println("Another instance is already running, opening browser.")
		_ = exec.Command("xdg-open", "http://localhost:9990").Start()
		return
	}

	fmt.Println("Listening on http://localhost:9990")
	go func() {
		_ = exec.Command("xdg-open", "http://localhost:9990").Start()
	}()
	if err := http.ListenAndServe(":9990", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func isAlreadyRunning() bool {
	conn, err := net.DialTimeout("tcp", "localhost:9990", 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func killExistingInstances() {
	if !isAlreadyRunning() {
		return
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:9990/quit")
	if err != nil {
		return
	}
	resp.Body.Close()
	// Wait for the old instance to release the port
	for i := 0; i < 10; i++ {
		time.Sleep(300 * time.Millisecond)
		if !isAlreadyRunning() {
			return
		}
	}
}
