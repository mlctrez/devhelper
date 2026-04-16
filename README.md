# devhelper

A local web dashboard that lets you quickly open any of your Go projects in GoLand or Kiro with a single click.

It scans your project directories, presents them in a dark-mode table at `http://localhost:9990`, and each project links directly to launch your IDE of choice. No more hunting through file dialogs or recent project lists.

## What it shows

- Project path (collapsed to `~/...`), clickable to open in GoLand or Kiro
- Go module name (from `go.mod`)
- Last modified date (days for recent, month/year for older)
- Git remote origin (read from `.git/config`, clickable, no credential leakage)
- Lines of code by file extension (top 5, excluding tests, dotfiles, binaries, and `go.mod`/`go.sum`)

Projects are detected by `.git` directory first, falling back to `go.mod`. The page re-scans on every refresh.

## Install

```
go install github.com/mlctrez/devhelper@latest
```

Register the desktop entry and application icon:

```
devhelper install
```

This writes a `.desktop` file and SVG icon to `~/.local/share` (respects `XDG_DATA_HOME`). If a previous instance is running, it will be gracefully terminated first.

To remove the desktop entry and icon:

```
devhelper uninstall
```

## Usage

```
devhelper [dir1] [dir2] ...
```

If no directories are provided, it defaults to `~/projects` and `~/golang`. At least one directory must exist.

On launch, devhelper starts an HTTP server on port 9990 and opens your browser automatically. If another instance is already running, it opens the browser to the existing instance instead.

## API Endpoints

- `/` — project dashboard (re-scans on each request)
- `/open?path=<dir>&ide=<goland|kiro>` — launch an IDE for the given project
- `/quit` — gracefully shut down the running instance

## Build from source

```
go build -o devhelper .
```
