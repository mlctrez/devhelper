package main

import (
	"fmt"
	"html"
	"io"
	"net/url"
	"strings"
)

func writeHTML(w io.Writer, projects []Project, homeDir string) {
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>DevHelper - Project Report</title>
<style>
  *, *::before, *::after { box-sizing: border-box; }
  body {
    background: #1e1e2e; color: #cdd6f4; font-family: 'Segoe UI', system-ui, sans-serif;
    margin: 0; padding: 24px;
  }
  h1 { color: #89b4fa; margin-bottom: 20px; }
  table { width: 100%; border-collapse: collapse; }
  th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #313244; }
  th { background: #181825; color: #89b4fa; position: sticky; top: 0; }
  tr:hover { background: #28283d; }
  .ext { color: #a6adc8; }
  .lines { color: #f9e2af; }
  .remote { color: #94e2d5; word-break: break-all; }
  .no-git { color: #585b70; }
  .days { color: #fab387; }
  .path { color: #cdd6f4; font-family: monospace; font-size: 0.9em; }
</style>
</head>
<body>
<h1>DevHelper - Project Report</h1>
`)

	fmt.Fprintf(w, "<p>%d projects found</p>\n", len(projects))
	fmt.Fprint(w, `<table>
<thead>
<tr><th>Project</th><th>IDE</th><th>Module</th><th>Last Modified</th><th>Git Remote</th><th>Lines of Code</th></tr>
</thead>
<tbody>
`)

	for _, p := range projects {
		fmt.Fprint(w, "<tr>")

		// Project path (clickable to open folder)
		escapedPath := url.PathEscape(p.Path)
		openFolder := "/open?ide=xdg-open&path=" + escapedPath
		displayPath := shortenPath(p.Path, homeDir)
		fmt.Fprintf(w, `<td class="path"><a href="%s" style="color:inherit" title="Open folder">%s</a></td>`,
			openFolder, html.EscapeString(displayPath))

		// IDE launchers
		fmt.Fprint(w, `<td style="white-space:nowrap">`)
		ides := []struct {
			Label string
			Cmd   string
			Title string
		}{
			{"G", "goland", "GoLand"},
			{"K", "kiro", "Kiro"},
			{"C", "clion", "CLion"},
			{"W", "webstorm", "WebStorm"},
			{"P", "pycharm", "PyCharm"},
		}
		for i, ide := range ides {
			if i > 0 {
				fmt.Fprint(w, " ")
			}
			link := "/open?ide=" + ide.Cmd + "&path=" + escapedPath
			fmt.Fprintf(w, `<a href="%s" title="%s" style="color:#89b4fa;text-decoration:none;font-weight:bold;font-size:0.85em">%s</a>`,
				link, ide.Title, ide.Label)
		}
		fmt.Fprint(w, "</td>")

		// Module name
		fmt.Fprintf(w, `<td class="path">%s</td>`, html.EscapeString(p.ModuleName))

		// Days ago
		fmt.Fprintf(w, `<td class="days">%s</td>`, formatDaysAgo(&p))

		// Git remote
		if !p.IsGit {
			fmt.Fprint(w, `<td class="no-git">not a git repo</td>`)
		} else if p.RemoteOrigin == "" {
			fmt.Fprint(w, `<td class="no-git">no remote</td>`)
		} else {
			href := gitRemoteToURL(p.RemoteOrigin)
			fmt.Fprintf(w, `<td class="remote"><a href="%s" target="_blank" style="color:inherit">%s</a></td>`,
				html.EscapeString(href), html.EscapeString(p.RemoteOrigin))
		}

		// LOC by extension (inline, single line)
		fmt.Fprint(w, `<td style="white-space:nowrap">`)
		entries := p.LOCSorted()
		if len(entries) == 0 {
			fmt.Fprint(w, `<span class="ext">no source files</span>`)
		}
		for i, e := range entries {
			ext := e.Ext
			if len(ext) > 5 {
				ext = ext[:5] + "…"
			}
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			fmt.Fprintf(w, `<span class="ext">%s</span>&nbsp;<span class="lines">%d</span>`,
				html.EscapeString(ext), e.Lines)
		}
		fmt.Fprint(w, "</td>")

		fmt.Fprint(w, "</tr>\n")
	}

	fmt.Fprint(w, "</tbody>\n</table>\n</body>\n</html>\n")
}

func formatDaysAgo(p *Project) string {
	if p.DaysAgo <= 90 {
		return fmt.Sprintf("%d days", p.DaysAgo)
	}
	return p.OldestMod.Format("Jan 06")
}

func shortenPath(path, homeDir string) string {
	if homeDir == "" {
		return path
	}
	projectsDir := homeDir + "/projects/"
	golangDir := homeDir + "/golang/"

	switch {
	case strings.HasPrefix(path, projectsDir):
		return "P " + path[len(projectsDir):]
	case strings.HasPrefix(path, golangDir):
		return "G " + path[len(golangDir):]
	case strings.HasPrefix(path, homeDir):
		return "~" + path[len(homeDir):]
	default:
		return path
	}
}
