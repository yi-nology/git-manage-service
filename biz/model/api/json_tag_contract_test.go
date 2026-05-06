package api

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestHandwrittenAPIJSONTagsDoNotAddSnakeCase(t *testing.T) {
	allowed := loadSnakeCaseJSONTagBaseline(t)
	actual := collectSnakeCaseJSONTags(t)

	var violations []string
	for key, count := range actual {
		maxAllowed, ok := allowed[key]
		if !ok {
			violations = append(violations, fmt.Sprintf("%s: new snake_case JSON tag", key))
			continue
		}
		if count > maxAllowed {
			violations = append(violations, fmt.Sprintf("%s: count %d exceeds baseline %d", key, count, maxAllowed))
		}
	}

	sort.Strings(violations)
	if len(violations) > 0 {
		t.Fatalf("new hand-written API DTO fields must use camelCase JSON tags:\n%s", strings.Join(violations, "\n"))
	}
}

func loadSnakeCaseJSONTagBaseline(t *testing.T) map[string]int {
	t.Helper()

	data, err := os.ReadFile("json_tag_baseline.txt")
	if err != nil {
		t.Fatalf("failed to read JSON tag baseline: %v", err)
	}

	allowed := make(map[string]int)
	for lineNo, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 3 {
			t.Fatalf("invalid baseline line %d: %q", lineNo+1, line)
		}

		count, err := strconv.Atoi(fields[2])
		if err != nil {
			t.Fatalf("invalid baseline count on line %d: %v", lineNo+1, err)
		}

		allowed[baselineKey(fields[0], fields[1])] = count
	}

	return allowed
}

func collectSnakeCaseJSONTags(t *testing.T) map[string]int {
	t.Helper()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("failed to list Go files: %v", err)
	}

	counts := make(map[string]int)
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") || file == "api.pb.go" {
			continue
		}

		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, file, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", file, err)
		}

		ast.Inspect(parsed, func(node ast.Node) bool {
			field, ok := node.(*ast.Field)
			if !ok || field.Tag == nil {
				return true
			}

			jsonName := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json")
			jsonName = strings.Split(jsonName, ",")[0]
			if jsonName == "" || jsonName == "-" || !strings.Contains(jsonName, "_") {
				return true
			}

			counts[baselineKey(file, jsonName)]++
			return true
		})
	}

	return counts
}

func baselineKey(file, tag string) string {
	return file + " " + tag
}
