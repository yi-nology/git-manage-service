package spec

import (
	"fmt"
	"sort"
	"strings"
)

func (f *SpecFormatter) commonCleanup(line string, opts FormatOptions) string {
	if opts.CommonCleanup {
		line = reBuildRootVar.ReplaceAllString(line, "%{buildroot}")
		line = reOptFlagsVar.ReplaceAllString(line, "%{optflags}")
		line = reDeprecatedGrep.ReplaceAllStringFunc(line, func(m string) string {
			if m == "egrep" {
				return "grep -E"
			}
			return "grep -F"
		})
	}

	if opts.PathMacros {
		for _, pm := range pathToMacro {
			if strings.Contains(line, pm.From) {
				line = strings.ReplaceAll(line, pm.From, pm.To)
			}
		}
	}

	if opts.UtilMacros {
		line = reUtilMacroRe.ReplaceAllStringFunc(line, func(m string) string {
			sub := reUtilMacroRe.FindStringSubmatch(m)
			if len(sub) >= 2 {
				if replacement, ok := utilMacroMap[sub[1]]; ok {
					return replacement
				}
			}
			return m
		})
	}

	return line
}

func (f *SpecFormatter) basicLineCleanup(line string, opts FormatOptions, lineNum int, changes *[]FormatChange) string {
	if opts.TabToSpaces {
		line = strings.ReplaceAll(line, "\t", strings.Repeat(" ", opts.IndentSize))
	}
	line = strings.TrimRight(line, " \t")
	line = f.commonCleanup(line, opts)

	if opts.Curlify {
		newLine := f.curlifyLine(line)
		if newLine != line {
			*changes = append(*changes, FormatChange{
				Line:   lineNum,
				Type:   "modified",
				Before: line,
				After:  newLine,
				Reason: "normalized macro brackets",
			})
			line = newLine
		}
	}
	return line
}

func (f *SpecFormatter) formatPreamble(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	depGroups := make(map[string][]string)
	var depOrder []string
	inDeps := false
	lastDepTag := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			if inDeps {
				flushDeps(&result, depGroups, depOrder, changes, opts)
				depGroups = make(map[string][]string)
				depOrder = nil
				inDeps = false
			}
			result = append(result, line)
			continue
		}

		line = strings.ReplaceAll(line, "\t", strings.Repeat(" ", opts.IndentSize))
		line = strings.TrimRight(line, " \t")

		if opts.RemoveBuildRoot && reBuildRootTag.MatchString(trimmed) {
			*changes = append(*changes, FormatChange{
				Line: i + 1, Type: "removed", Before: trimmed,
				Reason: "removed deprecated BuildRoot field",
			})
			continue
		}
		if opts.RemoveGroup && reGroupTag.MatchString(trimmed) {
			*changes = append(*changes, FormatChange{
				Line: i + 1, Type: "removed", Before: trimmed,
				Reason: "removed deprecated Group field",
			})
			continue
		}

		if opts.LicenseSPDX {
			if m := reLicenseTag.FindStringSubmatch(trimmed); m != nil {
				licenseVal := strings.TrimSpace(m[1])
				if fixed, ok := licenseMap[licenseVal]; ok {
					*changes = append(*changes, FormatChange{
						Line: i + 1, Type: "modified", Before: trimmed,
						After:  "License: " + fixed,
						Reason: fmt.Sprintf("normalized License to SPDX: %s -> %s", licenseVal, fixed),
					})
					line = "License: " + fixed
					trimmed = line
				}
			}
		}

		if reNoSourcePatch.MatchString(trimmed) {
			newLine := reNoSourcePatch.ReplaceAllString(trimmed, "${1}0:")
			if newLine != trimmed {
				*changes = append(*changes, FormatChange{
					Line: i + 1, Type: "modified", Before: trimmed, After: newLine,
					Reason: "normalized Source/Patch to explicit number 0",
				})
				line = newLine
				trimmed = newLine
			}
		}

		line = f.commonCleanup(line, opts)

		if isDepTag(trimmed) {
			m := reConditionTag.FindStringSubmatch(trimmed)
			if m != nil {
				tag := m[1]
				if m[2] != "" {
					tag += m[2]
				}
				value := strings.TrimSpace(m[3])
				value = f.normalizeDepOps(value)

				if inDeps && tag != lastDepTag {
					flushDeps(&result, depGroups, depOrder, changes, opts)
					depGroups = make(map[string][]string)
					depOrder = nil
				}
				inDeps = true
				lastDepTag = tag
				if _, exists := depGroups[tag]; !exists {
					depOrder = append(depOrder, tag)
				}
				depGroups[tag] = append(depGroups[tag], value)
				continue
			}
		}

		if inDeps {
			flushDeps(&result, depGroups, depOrder, changes, opts)
			depGroups = make(map[string][]string)
			depOrder = nil
			inDeps = false
		}

		if opts.Curlify {
			newLine := f.curlifyLine(line)
			if newLine != line {
				*changes = append(*changes, FormatChange{
					Line: i + 1, Type: "modified", Before: line, After: newLine,
					Reason: "normalized macro brackets",
				})
				line = newLine
			}
		}
		result = append(result, line)
	}

	if inDeps {
		flushDeps(&result, depGroups, depOrder, changes, opts)
	}

	result = f.collapseBlankLines(result)

	if opts.PreambleOrder {
		result = f.orderPreamble(result, changes)
	}
	if opts.AlignValues {
		result = f.alignTagValues(result, changes)
	}
	return result
}

func flushDeps(result *[]string, groups map[string][]string, order []string, changes *[]FormatChange, opts FormatOptions) {
	for _, tag := range order {
		deps := groups[tag]
		if opts.SortDeps {
			origLen := len(deps)
			deps = sortUniq(deps)
			if len(deps) < origLen {
				*changes = append(*changes, FormatChange{
					Line: len(*result) + 1, Type: "reordered",
					Reason: fmt.Sprintf("sorted and deduplicated %s (%d -> %d)", tag, origLen, len(deps)),
				})
			}
		}
		for _, dep := range deps {
			*result = append(*result, tag+": "+dep)
		}
	}
}

func sortUniq(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}
	sort.Strings(result)
	return result
}

func isDepTag(line string) bool {
	for _, prefix := range depTagPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func (f *SpecFormatter) normalizeDepOps(dep string) string {
	dep = strings.ReplaceAll(dep, "=<", "<=")
	dep = strings.ReplaceAll(dep, "=>", ">=")
	return dep
}

type preambleEntry struct {
	line     string
	tag      string
	tagOrder int
	isDefine bool
	isBcond  bool
	isCond   bool
	isComment bool
	original int
}

func (f *SpecFormatter) orderPreamble(lines []string, changes *[]FormatChange) []string {
	var entries []preambleEntry
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		entry := preambleEntry{line: line, tagOrder: 999, original: i}

		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			entry.isComment = true
			entries = append(entries, entry)
			continue
		}
		if strings.HasPrefix(trimmed, "%define ") || strings.HasPrefix(trimmed, "%global ") {
			entry.isDefine = true
			entry.tagOrder = 0
			entries = append(entries, entry)
			continue
		}
		if strings.HasPrefix(trimmed, "%bcond_") {
			entry.isBcond = true
			entry.tagOrder = 2
			entries = append(entries, entry)
			continue
		}
		if strings.HasPrefix(trimmed, "%if") || strings.HasPrefix(trimmed, "%else") || strings.HasPrefix(trimmed, "%endif") {
			entry.isCond = true
			entries = append(entries, entry)
			continue
		}

		matches := reTagColon.FindStringSubmatch(trimmed)
		if matches != nil {
			entry.tag = matches[1]
			if order, ok := preambleTagOrder[matches[1]]; ok {
				entry.tagOrder = order
			} else {
				for prefix, order := range preambleTagOrder {
					if strings.HasPrefix(matches[1], prefix) {
						entry.tagOrder = order
						break
					}
				}
			}
		}
		if isDepTag(trimmed) {
			matches := reConditionTag.FindStringSubmatch(trimmed)
			if matches != nil {
				tag := matches[1]
				if order, ok := preambleTagOrder[tag]; ok {
					entry.tagOrder = order
				}
			}
		}
		entries = append(entries, entry)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.isDefine != b.isDefine {
			return a.isDefine
		}
		if a.isBcond != b.isBcond {
			return a.isBcond
		}
		if a.tagOrder != b.tagOrder {
			return a.tagOrder < b.tagOrder
		}
		return a.original < b.original
	})

	reordered := false
	for i, e := range entries {
		if e.original != i {
			reordered = true
			break
		}
	}
	if reordered {
		*changes = append(*changes, FormatChange{
			Type:   "reordered",
			Reason: "reordered preamble tags to canonical order",
		})
	}

	var result []string
	for _, e := range entries {
		result = append(result, e.line)
	}
	return result
}

func (f *SpecFormatter) alignTagValues(lines []string, changes *[]FormatChange) []string {
	var tagLines []int
	maxTagLen := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "%") {
			continue
		}
		matches := reTagColon.FindStringSubmatch(trimmed)
		if matches != nil {
			tagPart := matches[1] + ":"
			if len(tagPart) > maxTagLen {
				maxTagLen = len(tagPart)
			}
			tagLines = append(tagLines, i)
		}
	}
	if maxTagLen == 0 || len(tagLines) < 2 {
		return lines
	}

	result := make([]string, len(lines))
	copy(result, lines)
	aligned := false

	for _, i := range tagLines {
		matches := reTagColon.FindStringSubmatch(result[i])
		if matches == nil {
			continue
		}
		tagPart := matches[1] + ":"
		value := matches[3]
		if len(tagPart) < maxTagLen {
			padding := strings.Repeat(" ", maxTagLen-len(tagPart))
			newLine := tagPart + padding + " " + value
			if newLine != result[i] {
				if !aligned {
					*changes = append(*changes, FormatChange{
						Type:   "modified",
						Reason: fmt.Sprintf("aligned tag values to column %d", maxTagLen+1),
					})
					aligned = true
				}
				result[i] = newLine
			}
		}
	}
	return result
}
