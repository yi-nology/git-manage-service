package rag

import (
	"strings"
	"unicode"
)

type Chunk struct {
	FilePath  string
	Content   string
	StartLine int
	EndLine   int
}

func ChunkFile(filePath, content string, maxChunkSize int) []*Chunk {
	if maxChunkSize <= 0 {
		maxChunkSize = 800
	}

	if isBinaryFile(filePath) {
		return nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}

	if len(content) <= maxChunkSize {
		return []*Chunk{{
			FilePath:  filePath,
			Content:   content,
			StartLine: 1,
			EndLine:   len(lines),
		}}
	}

	var chunks []*Chunk
	currentLines := make([]string, 0, 20)
	currentStart := 1
	currentSize := 0

	flush := func() {
		if len(currentLines) == 0 {
			return
		}
		text := strings.Join(currentLines, "\n")
		chunks = append(chunks, &Chunk{
			FilePath:  filePath,
			Content:   text,
			StartLine: currentStart,
			EndLine:   currentStart + len(currentLines) - 1,
		})
		currentLines = currentLines[:0]
		currentSize = 0
	}

	for i, line := range lines {
		isBoundary := isChunkBoundary(line)
		lineLen := len(line) + 1

		if (currentSize+lineLen > maxChunkSize && len(currentLines) > 0) ||
			(isBoundary && currentSize > maxChunkSize/3 && len(currentLines) > 3) {
			flush()
			currentStart = i + 1
		}

		currentLines = append(currentLines, line)
		currentSize += lineLen
	}

	flush()
	return chunks
}

func isChunkBoundary(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	if strings.HasPrefix(trimmed, "func ") ||
		strings.HasPrefix(trimmed, "function ") ||
		strings.HasPrefix(trimmed, "class ") ||
		strings.HasPrefix(trimmed, "type ") ||
		strings.HasPrefix(trimmed, "interface ") ||
		strings.HasPrefix(trimmed, "struct ") ||
		strings.HasPrefix(trimmed, "export ") ||
		strings.HasPrefix(trimmed, "export default ") ||
		strings.HasPrefix(trimmed, "const ") ||
		strings.HasPrefix(trimmed, "public ") ||
		strings.HasPrefix(trimmed, "private ") ||
		strings.HasPrefix(trimmed, "protected ") ||
		strings.HasPrefix(trimmed, "package ") ||
		strings.HasPrefix(trimmed, "import ") ||
		strings.HasPrefix(trimmed, "@") {
		return true
	}

	if strings.HasPrefix(trimmed, "def ") && (len(trimmed) < 4 || unicode.IsSpace(rune(trimmed[4])) || trimmed[4] == '(') {
		return true
	}

	if strings.HasPrefix(trimmed, "//") && strings.Contains(trimmed, "===") {
		return true
	}

	return false
}

func isBinaryFile(path string) bool {
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true,
		".ico": true, ".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".pdf": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true, ".o": true,
		".pyc": true, ".class": true, ".jar": true, ".wasm": true,
	}
	for ext := range binaryExts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}
	return false
}
