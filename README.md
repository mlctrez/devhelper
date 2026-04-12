# devhelper

A local web dashboard that lets you quickly open any of your Go projects in GoLand or Kiro with a single click.

It scans your project directories, presents them in a sortable table at `http://localhost:9990`, and each project links directly to launch your IDE of choice. No more hunting through file dialogs or recent project lists.

## What it shows

- Project path (collapsed to `~/...` where applicable)
- Go module name (from `go.mod`)
- Last modified date based on the most recently changed file
- Git remote origin (read directly from `.git/config`, no credential leakage)
- Lines of code by file extension (top 5, excluding tests, dotfiles, binary files, and `go.mod`/`go.sum`)

Projects are detected by the presence of `.git` or `go.mod`. The page re-scans on every refresh.

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
