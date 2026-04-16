package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed icon.svg
var iconSVG []byte

var desktopTemplate = template.Must(template.New("desktop").Parse(`[Desktop Entry]
Name=DevHelper
Comment=Go project dashboard and IDE launcher
Exec={{.Exec}}
Icon={{.Icon}}
Terminal=false
Type=Application
Categories=Development;
`))

func runInstall() {
	// Terminate any existing instances before reinstalling
	killExistingInstances()

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to determine home directory: %v\n", err)
		os.Exit(1)
	}

	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(home, ".local", "share")
	}

	// Resolve the binary path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to determine executable path: %v\n", err)
		os.Exit(1)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to resolve executable path: %v\n", err)
		os.Exit(1)
	}

	// Write icon
	iconDir := filepath.Join(dataDir, "icons", "hicolor", "scalable", "apps")
	if err := os.MkdirAll(iconDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating icon directory: %v\n", err)
		os.Exit(1)
	}
	iconPath := filepath.Join(iconDir, "devhelper.svg")
	if err := os.WriteFile(iconPath, iconSVG, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing icon: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Installed icon to %s\n", iconPath)

	// Write desktop entry
	appsDir := filepath.Join(dataDir, "applications")
	if err := os.MkdirAll(appsDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating applications directory: %v\n", err)
		os.Exit(1)
	}
	desktopPath := filepath.Join(appsDir, "devhelper.desktop")
	f, err := os.Create(desktopPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating desktop file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	err = desktopTemplate.Execute(f, struct {
		Exec string
		Icon string
	}{
		Exec: execPath,
		Icon: "devhelper",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing desktop file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Installed desktop entry to %s\n", desktopPath)
	fmt.Println("You may need to log out and back in for the icon to appear in your application menu.")
}

func runUninstall() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to determine home directory: %v\n", err)
		os.Exit(1)
	}

	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(home, ".local", "share")
	}

	iconPath := filepath.Join(dataDir, "icons", "hicolor", "scalable", "apps", "devhelper.svg")
	desktopPath := filepath.Join(dataDir, "applications", "devhelper.desktop")

	for _, path := range []string{iconPath, desktopPath} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", path, err)
		} else if err == nil {
			fmt.Printf("Removed %s\n", path)
		}
	}
}
