package git

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

func (s *GitService) generateDiff(from, to, filePath string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	buf.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

	fromLines := splitLines(from)
	toLines := splitLines(to)

	ops := s.computeDiffOps(fromLines, toLines)
	lineNum := 1
	for _, op := range ops {
		switch op.kind {
		case ' ':
			buf.WriteString(fmt.Sprintf(" %s\n", op.line))
			lineNum++
		case '+':
			buf.WriteString(fmt.Sprintf("+%s\n", op.line))
		case '-':
			buf.WriteString(fmt.Sprintf("-%s\n", op.line))
			lineNum++
		}
	}

	return buf.String()
}

type diffOp struct {
	kind byte
	line string
}

func (s *GitService) computeDiffOps(from, to []string) []diffOp {
	var ops []diffOp
	fi, ti := 0, 0
	for fi < len(from) && ti < len(to) {
		if from[fi] == to[ti] {
			ops = append(ops, diffOp{' ', from[fi]})
			fi++
			ti++
		} else {
			ops = append(ops, diffOp{'-', from[fi]})
			fi++
		}
	}
	for fi < len(from) {
		ops = append(ops, diffOp{'-', from[fi]})
		fi++
	}
	for ti < len(to) {
		ops = append(ops, diffOp{'+', to[ti]})
		ti++
	}
	return ops
}

func countDiffStats(diff string) (additions, deletions int) {
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			if line[0] == '+' && !strings.HasPrefix(line, "+++") {
				additions++
			} else if line[0] == '-' && !strings.HasPrefix(line, "---") {
				deletions++
			}
		}
	}
	return
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
		".ico": true, ".svg": true, ".pdf": true, ".zip": true, ".tar": true,
		".gz": true, ".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true, ".mp3": true,
		".mp4": true, ".avi": true, ".mov": true, ".wasm": true,
	}
	return binaryExts[ext]
}
