package ai

import (
	"fmt"
	"strings"
)

type ContextBuilder struct {
	sections []string
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{sections: make([]string, 0, 8)}
}

func (b *ContextBuilder) AddSection(title, content string) *ContextBuilder {
	if content == "" {
		return b
	}
	b.sections = append(b.sections, fmt.Sprintf("## %s\n\n%s", title, content))
	return b
}

func (b *ContextBuilder) AddCodeSection(title, language, code string) *ContextBuilder {
	if code == "" {
		return b
	}
	b.sections = append(b.sections, fmt.Sprintf("## %s\n\n```%s\n%s\n```", title, language, code))
	return b
}

func (b *ContextBuilder) AddListSection(title string, items []string) *ContextBuilder {
	if len(items) == 0 {
		return b
	}
	var bld strings.Builder
	for _, item := range items {
		bld.WriteString("- ")
		bld.WriteString(item)
		bld.WriteString("\n")
	}
	b.sections = append(b.sections, fmt.Sprintf("## %s\n\n%s", title, bld.String()))
	return b
}

func (b *ContextBuilder) Build() string {
	return strings.Join(b.sections, "\n\n")
}

func BuildRepoContext(repoKey, defaultBranch string, branchCount, tagCount, commitCount int64) string {
	b := NewContextBuilder()
	b.AddSection("Repository Information", fmt.Sprintf(
		"Repo: %s\nDefault Branch: %s\nBranches: %d\nTags: %d\nTotal Commits: %d",
		repoKey, defaultBranch, branchCount, tagCount, commitCount,
	))
	return b.Build()
}
