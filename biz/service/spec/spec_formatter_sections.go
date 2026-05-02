package spec

import (
	"strings"
)

func (f *SpecFormatter) formatDescription(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		line = f.basicLineCleanup(line, opts, i+1, changes)
		if i > 0 && strings.HasPrefix(strings.TrimSpace(line), "Author(s):") {
			break
		}
		result = append(result, line)
	}
	return f.collapseBlankLines(result)
}

func (f *SpecFormatter) formatPrep(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		line = f.basicLineCleanup(line, opts, i+1, changes)
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "%setup ") {
			if reSetupQn.MatchString(trimmed) {
				newLine := reSetupQn.ReplaceAllString(line, "-q -n ")
				*changes = append(*changes, FormatChange{
					Line: i + 1, Type: "modified", Before: line, After: newLine,
					Reason: "split %setup -qn to -q -n",
				})
				line = newLine
			}
		}

		if rePatchLine.MatchString(trimmed) {
			newLine := rePatchP0.ReplaceAllString(line, "")
			if newLine != line {
				*changes = append(*changes, FormatChange{
					Line: i + 1, Type: "modified", Before: line, After: strings.TrimSpace(newLine),
					Reason: "removed redundant -p0 from %patch",
				})
				line = strings.TrimSpace(newLine)
			}
			matches := rePatchLine.FindStringSubmatch(strings.TrimSpace(line))
			if matches != nil && matches[1] == "" {
				newLine := strings.Replace(line, "%patch ", "%patch0 ", 1)
				*changes = append(*changes, FormatChange{
					Line: i + 1, Type: "modified", Before: line, After: newLine,
					Reason: "normalized bare %patch to %patch0",
				})
				line = newLine
			} else if strings.TrimSpace(line) == "%patch" {
				newLine := strings.Replace(line, "%patch", "%patch0", 1)
				*changes = append(*changes, FormatChange{
					Line: i + 1, Type: "modified", Before: line, After: newLine,
					Reason: "normalized bare %patch to %patch0",
				})
				line = newLine
			}
		}

		result = append(result, line)
	}
	return f.collapseBlankLines(result)
}

func (f *SpecFormatter) formatBuild(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		line = f.basicLineCleanup(line, opts, i+1, changes)
		trimmed := strings.TrimSpace(line)

		if reMakeLine.MatchString(trimmed) && !strings.Contains(trimmed, "$(MAKE)") && !strings.Contains(trimmed, "`make") {
			newLine := reMakeLine.ReplaceAllString(line, "%make_build$2")
			*changes = append(*changes, FormatChange{
				Line: i + 1, Type: "modified", Before: line, After: newLine,
				Reason: "replaced make with %make_build",
			})
			line = newLine
		}

		_ = trimmed
		result = append(result, line)
	}
	return f.collapseBlankLines(result)
}

func (f *SpecFormatter) formatInstall(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		line = f.basicLineCleanup(line, opts, i+1, changes)
		trimmed := strings.TrimSpace(line)

		if reRmBuildRoot.MatchString(trimmed) {
			*changes = append(*changes, FormatChange{
				Line: i + 1, Type: "removed", Before: trimmed,
				Reason: "removed redundant rm -rf %{buildroot}",
			})
			continue
		}

		if reMakeinstall.MatchString(trimmed) {
			newLine := reMakeinstall.ReplaceAllString(line, "%make_install")
			*changes = append(*changes, FormatChange{
				Line: i + 1, Type: "modified", Before: trimmed, After: "%make_install",
				Reason: "replaced %makeinstall with %make_install",
			})
			line = newLine
		}

		if reMakeDestdir.MatchString(trimmed) {
			newLine := "%make_install"
			*changes = append(*changes, FormatChange{
				Line: i + 1, Type: "modified", Before: trimmed, After: newLine,
				Reason: "replaced make DESTDIR install with %make_install",
			})
			line = newLine
		}

		result = append(result, line)
	}
	return f.collapseBlankLines(result)
}

func (f *SpecFormatter) formatFiles(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		line = f.basicLineCleanup(line, opts, i+1, changes)
		trimmed := strings.TrimSpace(line)

		if reDefAttrLine.MatchString(trimmed) {
			*changes = append(*changes, FormatChange{
				Line: i + 1, Type: "removed", Before: trimmed,
				Reason: "removed redundant %defattr(-,root,root)",
			})
			continue
		}

		if strings.HasPrefix(trimmed, "%doc ") {
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) == 2 {
				files := strings.Fields(parts[1])
				var docFiles, licenseFiles []string
				for _, fn := range files {
					upper := strings.ToUpper(fn)
					if strings.Contains(upper, "LICENSE") || strings.Contains(upper, "LICENCE") || strings.Contains(upper, "COPYING") {
						licenseFiles = append(licenseFiles, fn)
					} else {
						docFiles = append(docFiles, fn)
					}
				}
				if len(licenseFiles) > 0 {
					*changes = append(*changes, FormatChange{
						Line: i + 1, Type: "modified", Before: trimmed,
						Reason: "extracted license files from %doc to %license",
					})
					if len(docFiles) > 0 {
						result = append(result, "%doc "+strings.Join(docFiles, " "))
					}
					result = append(result, "%license "+strings.Join(licenseFiles, " "))
					continue
				}
			}
		}

		result = append(result, line)
	}
	return f.collapseBlankLines(result)
}

func (f *SpecFormatter) formatScriptlet(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		line = f.basicLineCleanup(line, opts, i+1, changes)
		result = append(result, line)
	}
	return f.collapseBlankLines(result)
}

func (f *SpecFormatter) curlifyLine(line string) string {
	return reMacroNoBrace.ReplaceAllStringFunc(line, func(match string) string {
		sub := reMacroNoBrace.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		name := sub[1]
		if macroWhitelist[name] || name[0] == '_' {
			return match
		}
		if len(name) == 1 && name[0] >= '0' && name[0] <= '9' {
			return match
		}
		if isSectionKeyword(name) {
			return match
		}
		return "%{" + name + "}"
	})
}

func (f *SpecFormatter) trimConditionals(lines []string, changes *[]FormatChange) []string {
	var result []string
	prevWasCond := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isCond := strings.HasPrefix(trimmed, "%if") || strings.HasPrefix(trimmed, "%else") || strings.HasPrefix(trimmed, "%endif")
		isBlank := trimmed == ""

		if isBlank && prevWasCond {
			continue
		}

		result = append(result, line)
		prevWasCond = isCond
	}

	return result
}

func (f *SpecFormatter) collapseBlankLines(lines []string) []string {
	var result []string
	prevBlank := false
	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && prevBlank {
			continue
		}
		result = append(result, line)
		prevBlank = isBlank
	}
	return result
}

func (f *SpecFormatter) normalizeTrailingNewlines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	return lines
}
