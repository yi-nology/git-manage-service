package codereview

import (
	"strings"
)

type DiffHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Content  string
	Lines    []DiffLine
}

type DiffLine struct {
	Type     string
	OldLine  int
	NewLine  int
	Content  string
	FilePath string
}

type FileDiff struct {
	OldPath   string
	NewPath   string
	IsNew     bool
	IsDeleted bool
	IsRenamed bool
	Hunks     []*DiffHunk
	RawDiff   string
	Additions int
	Deletions int
}

func ParseDiff(raw string) []*FileDiff {
	if raw == "" {
		return nil
	}

	var files []*FileDiff
	var current *FileDiff
	var currentHunk *DiffHunk
	oldLine, newLine := 0, 0

	lines := strings.Split(raw, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if current != nil {
				finalizeFileDiff(current)
				files = append(files, current)
			}
			current = &FileDiff{}
			currentHunk = nil
			oldLine, newLine = 0, 0
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "--- ") {
			current.OldPath = stripPathPrefix(line[4:])
			continue
		}
		if strings.HasPrefix(line, "+++ ") {
			current.NewPath = stripPathPrefix(line[4:])
			continue
		}
		if strings.HasPrefix(line, "new file") {
			current.IsNew = true
			continue
		}
		if strings.HasPrefix(line, "deleted file") {
			current.IsDeleted = true
			continue
		}
		if strings.HasPrefix(line, "rename from ") {
			current.IsRenamed = true
			current.OldPath = line[12:]
			continue
		}
		if strings.HasPrefix(line, "rename to ") {
			current.NewPath = line[10:]
			continue
		}

		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil {
				current.Hunks = append(current.Hunks, currentHunk)
			}
			os, oc, ns, nc := parseHunkHeader(line)
			currentHunk = &DiffHunk{
				OldStart: os, OldCount: oc,
				NewStart: ns, NewCount: nc,
			}
			oldLine = os
			newLine = ns
			continue
		}

		if currentHunk == nil {
			current.RawDiff += line + "\n"
			continue
		}

		current.RawDiff += line + "\n"

		dl := DiffLine{FilePath: current.NewPath}
		switch {
		case strings.HasPrefix(line, "+"):
			newLine++
			dl.Type = "add"
			dl.NewLine = newLine
			dl.Content = line[1:]
			current.Additions++
		case strings.HasPrefix(line, "-"):
			oldLine++
			dl.Type = "del"
			dl.OldLine = oldLine
			dl.Content = line[1:]
			current.Deletions++
		default:
			if line != "" && line != `\ No newline at end of file` {
				oldLine++
				newLine++
				dl.Type = "ctx"
				dl.OldLine = oldLine
				dl.NewLine = newLine
				dl.Content = line
			}
		}
		currentHunk.Lines = append(currentHunk.Lines, dl)
	}

	if current != nil {
		if currentHunk != nil {
			current.Hunks = append(current.Hunks, currentHunk)
		}
		finalizeFileDiff(current)
		files = append(files, current)
	}

	return files
}

func finalizeFileDiff(f *FileDiff) {
	if f.NewPath == "" {
		f.NewPath = f.OldPath
	}
	if f.OldPath == "" {
		f.OldPath = f.NewPath
	}
}

func stripPathPrefix(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "a/")
	p = strings.TrimPrefix(p, "b/")
	if len(p) > 2 && p[0] == '"' && p[len(p)-1] == '"' {
		p = p[1 : len(p)-1]
	}
	return p
}

func parseHunkHeader(line string) (oldStart, oldCount, newStart, newCount int) {
	s := line
	at := strings.Index(s, "@@")
	if at < 0 {
		return
	}
	s = s[at+2:]
	at2 := strings.Index(s, "@@")
	if at2 >= 0 {
		s = s[:at2]
	}
	s = strings.TrimSpace(s)

	parts := strings.SplitN(s, " ", 2)
	parseRange(parts[0], &oldStart, &oldCount)
	if len(parts) > 1 {
		parseRange(parts[1], &newStart, &newCount)
	}
	return
}

func parseRange(s string, start, count *int) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	parts := strings.SplitN(s, ",", 2)
	if len(parts) == 1 {
		fmtSScanf(parts[0], start)
		*count = 1
	} else {
		fmtSScanf(parts[0], start)
		fmtSScanf(parts[1], count)
	}
}

func fmtSScanf(s string, v *int) {
	s = strings.TrimSpace(s)
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return
		}
	}
	if len(s) > 0 {
		n := 0
		for _, c := range s {
			n = n*10 + int(c-'0')
		}
		*v = n
	}
}
