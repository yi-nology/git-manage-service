package ai

import (
	"encoding/json"
	"regexp"
	"strings"
)

func ExtractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(raw); i++ {
		ch := raw[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return raw[start : i+1]
			}
		}
	}
	return ""
}

func DecodeJSON(raw string, out any) bool {
	jsonStr := ExtractJSON(raw)
	if jsonStr == "" {
		return false
	}
	return json.Unmarshal([]byte(jsonStr), out) == nil
}

func StripFencedCode(raw string, langs ...string) string {
	langPattern := "[a-zA-Z0-9_-]*"
	if len(langs) > 0 {
		langPattern = "(?:" + strings.Join(langs, "|") + ")?"
	}
	re := regexp.MustCompile("(?s)```" + langPattern + "\\n(.*?)```")
	if matches := re.FindStringSubmatch(raw); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(raw)
}
