# devhelper

A local web dashboard for scanning and browsing Go project directories.

## What it does

devhelper scans one or more directories for project roots (identified by `.git` or `go.mod`), analyzes each project, and serves an HTML report on `http://localhost:9990`. The page re-scans on every refresh, so it always reflects the current state of your projects.

For each project, the report shows:

- Project path (collapsed to `~/...` where applicable)
- Go module name (from `go.mod`)
- Last modified date based on the most recently changed file
- Git remote origin (read directly from `.git/config`, no credential leakage)
- Lines of code by file extension (top 5, excluding tests, dotfiles, binary files, and `go.mod`/`go.sum`)

Project paths are clickable to open in GoLand or Kiro directly from the browser.

## Usage

```
devhelper [dir1] [dir2] ...
```

If no directories are provided, it defaults to `~/projects` and `~/golang`. At least one of the provided directories must exist.

## Install

```
go install github.com/mlctrez/devhelper@latest
```

## Build from source

```
go build -o devhelper .
```
