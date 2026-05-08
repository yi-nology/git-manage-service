package rag

import (
	"strings"
	"testing"
)

func TestChunkFile_SmallFile(t *testing.T) {
	content := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	chunks := ChunkFile("main.go", content, 800)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].StartLine != 1 || chunks[0].EndLine != 6 {
		t.Errorf("unexpected line range: %d-%d", chunks[0].StartLine, chunks[0].EndLine)
	}
}

func TestChunkFile_LargeFile(t *testing.T) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, strings.Repeat("x", 50))
	}
	content := strings.Join(lines, "\n")

	chunks := ChunkFile("test.go", content, 500)
	if len(chunks) <= 1 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	totalLines := 0
	for _, c := range chunks {
		lineCount := strings.Count(c.Content, "\n") + 1
		totalLines += lineCount
		if lineCount > 30 {
			t.Errorf("chunk has too many lines: %d", lineCount)
		}
	}
	if totalLines < 95 {
		t.Errorf("lost lines: only %d of 100 accounted", totalLines)
	}
}

func TestChunkFile_BinaryFile(t *testing.T) {
	chunks := ChunkFile("image.png", "binary data", 800)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for binary file, got %d", len(chunks))
	}
}

func TestChunkFile_FuncBoundary(t *testing.T) {
	var b strings.Builder
	b.WriteString("package main\n\n")
	for i := 0; i < 3; i++ {
		b.WriteString("func func" + strings.Repeat("x", 60) + "() {\n")
		for j := 0; j < 10; j++ {
			b.WriteString("\tline := \"" + strings.Repeat("y", 40) + "\"\n")
		}
		b.WriteString("}\n\n")
	}

	content := b.String()
	chunks := ChunkFile("main.go", content, 400)
	if len(chunks) < 2 {
		t.Fatalf("expected chunking at func boundaries, got %d chunks", len(chunks))
	}
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		path   string
		binary bool
	}{
		{"main.go", false},
		{"image.png", true},
		{"photo.jpg", true},
		{"archive.zip", true},
		{"video.mp4", true},
		{"README.md", false},
		{"config.yaml", false},
	}
	for _, tt := range tests {
		got := isBinaryFile(tt.path)
		if got != tt.binary {
			t.Errorf("isBinaryFile(%q) = %v, want %v", tt.path, got, tt.binary)
		}
	}
}

func TestVectorStore_IndexAndSearch(t *testing.T) {
	store := NewVectorStore(100)

	vectors := []*Vector{
		{
			ID: "file1:1", FilePath: "file1.go", StartLine: 1, EndLine: 10,
			Content: "package main",
			Values:  []float64{1.0, 0.0, 0.0},
		},
		{
			ID: "file2:1", FilePath: "file2.go", StartLine: 1, EndLine: 10,
			Content: "package util",
			Values:  []float64{0.0, 1.0, 0.0},
		},
		{
			ID: "file3:1", FilePath: "file3.go", StartLine: 1, EndLine: 10,
			Content: "package service",
			Values:  []float64{0.9, 0.1, 0.0},
		},
	}

	store.Index("test-repo", vectors)

	query := []float64{1.0, 0.0, 0.0}
	results := store.Search("test-repo", query, 2, 0.5)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Vector.FilePath != "file1.go" {
		t.Errorf("top result should be file1.go, got %s", results[0].Vector.FilePath)
	}
	if results[0].Score < 0.99 {
		t.Errorf("top result score should be ~1.0, got %f", results[0].Score)
	}

	if !store.HasIndex("test-repo") {
		t.Error("store should have index for test-repo")
	}
	if store.HasIndex("nonexistent") {
		t.Error("store should not have index for nonexistent repo")
	}
}

func TestVectorStore_Clear(t *testing.T) {
	store := NewVectorStore(100)
	store.Index("repo", []*Vector{{ID: "1", Values: []float64{1.0}}})
	store.Clear("repo")
	if store.HasIndex("repo") {
		t.Error("store should not have index after clear")
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}
	if s := cosineSimilarity(a, b); s < 0.99 {
		t.Errorf("identical vectors should have similarity ~1.0, got %f", s)
	}

	c := []float64{0, 1, 0}
	if s := cosineSimilarity(a, c); s > 0.01 {
		t.Errorf("orthogonal vectors should have similarity ~0.0, got %f", s)
	}

	if s := cosineSimilarity(a, nil); s != 0 {
		t.Errorf("nil vector should return 0, got %f", s)
	}
}

func TestFormatContextForPrompt(t *testing.T) {
	results := []*SearchResult{
		{Vector: &Vector{FilePath: "main.go", StartLine: 10, EndLine: 20, Content: "func main() {}"}, Score: 0.95},
	}
	formatted := FormatContextForPrompt(results, 1000)
	if !strings.Contains(formatted, "main.go") {
		t.Error("formatted context should contain file path")
	}
	if !strings.Contains(formatted, "func main()") {
		t.Error("formatted context should contain code")
	}
}
