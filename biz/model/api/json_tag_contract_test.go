package api

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// TestHandwrittenAPIJSONTagsAreSnakeCase enforces the API naming convention:
// every JSON tag on a hand-written DTO field in this package must use
// snake_case. camelCase JSON tags are rejected so the HTTP API stays
// consistent with the codegen output (--snake_tag) and the frontend.
//
// Decision: API JSON fields are snake_case (see docs/dev-notes/API_NAMING_CONVENTION.md).
// This supersedes the earlier "camelCase long-term / snake_case freeze" policy.
func TestHandwrittenAPIJSONTagsAreSnakeCase(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list go files: %v", err)
	}

	var violations []string
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") || file == "api.pb.go" {
			continue
		}

		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}

		ast.Inspect(parsed, func(node ast.Node) bool {
			field, ok := node.(*ast.Field)
			if !ok || field.Tag == nil {
				return true
			}
			jsonName := strings.Split(reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json"), ",")[0]
			if jsonName == "" || jsonName == "-" {
				return true
			}
			for _, r := range jsonName {
				if r >= 'A' && r <= 'Z' {
					violations = append(violations, fmt.Sprintf("%s: json tag %q is not snake_case", file, jsonName))
					break
				}
			}
			return true
		})
	}

	sort.Strings(violations)
	if len(violations) > 0 {
		t.Fatalf("hand-written API DTO fields must use snake_case JSON tags:\n%s", strings.Join(violations, "\n"))
	}
}
