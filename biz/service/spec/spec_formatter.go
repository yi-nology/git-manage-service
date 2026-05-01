package spec

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type SectionType int

const (
	SectionPreamble SectionType = iota
	SectionDescription
	SectionPrep
	SectionBuild
	SectionInstall
	SectionCheck
	SectionClean
	SectionFiles
	SectionChangelog
	SectionPackage
	SectionScriptlet
)

type Section struct {
	Type     SectionType
	Name     string
	Lines    []string
	SubPkg   string
}

type FormatOptions struct {
	Curlify         bool
	RemoveClean     bool
	RemoveBuildRoot bool
	RemoveGroup     bool
	LicenseSPDX     bool
	SortDeps        bool
	TabToSpaces     bool
	IndentSize      int
}

type FormatChange struct {
	Line   int    `json:"line"`
	Type   string `json:"type"`
	Before string `json:"before"`
	After  string `json:"after"`
	Reason string `json:"reason"`
}

func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		Curlify:         true,
		RemoveClean:     true,
		RemoveBuildRoot: true,
		RemoveGroup:     false,
		LicenseSPDX:     true,
		SortDeps:        true,
		TabToSpaces:     true,
		IndentSize:      4,
	}
}

var sectionHeaders = []string{
	"%description",
	"%prep",
	"%build",
	"%install",
	"%check",
	"%clean",
	"%files",
	"%changelog",
	"%package",
	"%post",
	"%postun",
	"%pre",
	"%preun",
	"%posttrans",
	"%pretrans",
	"%triggerin",
	"%triggerun",
	"%triggerpostun",
	"%verifyscript",
	"%filetriggerin",
	"%filetriggerun",
	"%filetriggerpostun",
	"%transfiletriggerin",
	"%transfiletriggerun",
	"%transfiletriggerpostun",
}

var macroWhitelist = map[string]bool{
	"prep": true, "build": true, "install": true, "check": true, "clean": true,
	"files": true, "changelog": true, "description": true, "package": true,
	"post": true, "postun": true, "pre": true, "preun": true, "posttrans": true,
	"pretrans": true, "setup": true, "autosetup": true, "autopatch": true,
	"configure": true, "make_install": true, "cmake": true, "meson": true,
	"find_lang": true, "license": true, "doc": true, "ghost": true,
	"config": true, "attr": true, "defattr": true, "dir": true, "verify": true,
	"nil": true, "define": true, "global": true, "if": true, "else": true,
	"endif": true, "ifarch": true, "ifnarch": true, "ifos": true, "ifnos": true,
	"with": true, "without": true, "bcond_with": true, "bcond_without": true,
	"mklibname": true, "configure2_5x": true, "cmake_kde4": true,
	"make": true, "make_build": true, "makeinstall": true,
	"remove_parsepolicy": true, "perl_vendorarch": true, "perl_vendorlib": true,
	"perl_sitelib": true, "perl_sitearch": true,
	"name": true, "version": true, "release": true, "epoch": true,
	"summary": true, "url": true, "group": true,
	"source0": true, "source1": true, "source2": true, "source3": true,
	"source4": true, "source5": true, "source6": true, "source7": true,
	"source8": true, "source9": true,
	"patch0": true, "patch1": true, "patch2": true, "patch3": true,
	"patch4": true, "patch5": true, "patch6": true, "patch7": true,
	"patch8": true, "patch9": true,
	"dist": true, "rhel": true, "fedora": true, "centos": true, "suse": true,
	"opensuse": true, "sles": true, "mandriva": true, "mdk": true,
	"_prefix": true, "_sysconfdir": true, "_bindir": true,
	"_sbindir": true, "_libdir": true, "_libexecdir": true, "_datadir": true,
	"_mandir": true, "_docdir": true, "_includedir": true, "_infodir": true,
	"_localstatedir": true, "_sharedstatedir": true, "optflags": true,
	"buildroot": true, "_builddir": true, "_rpmdir": true, "_srpmdir": true,
	"_sourcedir": true, "_specdir": true, "_srcrpmdir": true,
	"_tmppath": true, "_tmpdir": true, "_var": true, "_usr": true,
	"_exec_prefix": true, "_initddir": true, "_initrddir": true,
	"_unitdir": true, "_udevrulesdir": true, "_udevdir": true,
	"_sysusersdir": true, "_tmpfilesdir": true, "_environmentdir": true,
}

var licenseMap = map[string]string{
	"GPL":                      "GPL-1.0-only",
	"GPLv1":                    "GPL-1.0-only",
	"GPL v1":                   "GPL-1.0-only",
	"GPL2":                     "GPL-2.0-only",
	"GPLv2":                    "GPL-2.0-only",
	"GPL v2":                   "GPL-2.0-only",
	"GPLv2+":                   "GPL-2.0-or-later",
	"GPL v2+":                  "GPL-2.0-or-later",
	"GPL v2 or later":          "GPL-2.0-or-later",
	"GPL3":                     "GPL-3.0-only",
	"GPLv3":                    "GPL-3.0-only",
	"GPL v3":                   "GPL-3.0-only",
	"GPLv3+":                   "GPL-3.0-or-later",
	"GPL v3+":                  "GPL-3.0-or-later",
	"GPL v3 or later":          "GPL-3.0-or-later",
	"LGPL":                     "LGPL-2.0-only",
	"LGPLv2":                   "LGPL-2.0-only",
	"LGPL v2":                  "LGPL-2.0-only",
	"LGPLv2+":                  "LGPL-2.0-or-later",
	"LGPL v2+":                 "LGPL-2.0-or-later",
	"LGPL2.1":                  "LGPL-2.1-only",
	"LGPLv2.1":                 "LGPL-2.1-only",
	"LGPL v2.1":                "LGPL-2.1-only",
	"LGPLv2.1+":                "LGPL-2.1-or-later",
	"LGPL v2.1+":               "LGPL-2.1-or-later",
	"LGPL3":                    "LGPL-3.0-only",
	"LGPLv3":                   "LGPL-3.0-only",
	"LGPL v3":                  "LGPL-3.0-only",
	"LGPLv3+":                  "LGPL-3.0-or-later",
	"LGPL v3+":                 "LGPL-3.0-or-later",
	"AGPL":                     "AGPL-3.0-only",
	"AGPLv3":                   "AGPL-3.0-only",
	"AGPL v3":                  "AGPL-3.0-only",
	"AGPLv3+":                  "AGPL-3.0-or-later",
	"MIT License":              "MIT",
	"MIT license":              "MIT",
	"The MIT License":          "MIT",
	"BSD":                      "BSD-3-Clause",
	"BSD License":              "BSD-3-Clause",
	"BSD license":              "BSD-3-Clause",
	"New BSD License":          "BSD-3-Clause",
	"New BSD license":          "BSD-3-Clause",
	"3-Clause BSD":             "BSD-3-Clause",
	"Revised BSD":              "BSD-3-Clause",
	"2-Clause BSD":             "BSD-2-Clause",
	"Simplified BSD":           "BSD-2-Clause",
	"Apache":                   "Apache-2.0",
	"Apache License":           "Apache-2.0",
	"Apache 2":                 "Apache-2.0",
	"Apache 2.0":               "Apache-2.0",
	"ASL 2.0":                  "Apache-2.0",
	"ASL2":                     "Apache-2.0",
	"Apache Software License":  "Apache-2.0",
	"Artistic":                 "Artistic-1.0-Perl",
	"Artistic License":         "Artistic-1.0-Perl",
	"Artistic 2.0":             "Artistic-2.0",
	"Boost":                    "BSL-1.0",
	"Boost Software License":   "BSL-1.0",
	"BSL":                      "BSL-1.0",
	"CDDL":                     "CDDL-1.0",
	"CDDLv1":                   "CDDL-1.0",
	"CDDL 1.0":                 "CDDL-1.0",
	"EPL":                      "EPL-1.0",
	"EPL 1.0":                  "EPL-1.0",
	"EPL 2.0":                  "EPL-2.0",
	"MPL":                      "MPL-2.0",
	"MPLv2":                    "MPL-2.0",
	"MPL 2.0":                  "MPL-2.0",
	"MPLv1.1":                  "MPL-1.1",
	"MPL 1.1":                  "MPL-1.1",
	"ISC License":              "ISC",
	"ISC license":              "ISC",
	"PSF":                      "Python-2.0",
	"Python":                   "Python-2.0",
	"Python License":           "Python-2.0",
	"PSF License":              "Python-2.0",
	"ZPL":                      "ZPL-2.1",
	"Zope":                     "ZPL-2.1",
	"WTFPL":                    "WTFPL",
	"Unlicense":                "Unlicense",
	"Public Domain":            "LicenseRef-Public-Domain",
	"public domain":            "LicenseRef-Public-Domain",
	"LGPLv2+ with exceptions":  "LGPL-2.1-or-later WITH GCC-exception-2.0",
}

var depTags = []string{
	"BuildRequires",
	"Requires",
	"Recommends",
	"Suggests",
	"Supplements",
	"Enhances",
	"Conflicts",
	"Obsoletes",
	"Provides",
	"Requires(pre)",
	"Requires(post)",
	"Requires(preun)",
	"Requires(postun)",
	"Requires(pretrans)",
	"Requires(posttrans)",
}

var depTagPrefixes []string

func init() {
	for _, t := range depTags {
		depTagPrefixes = append(depTagPrefixes, t+":")
	}
}

var (
	reMacroNoBrace = regexp.MustCompile(`%([A-Za-z_][A-Za-z0-9_]*)`)
	reConditionTag = regexp.MustCompile(`^(BuildRequires|Requires|Recommends|Suggests|Supplements|Enhances|Conflicts|Obsoletes|Provides)(\([^)]*\))?\s*:\s*(.*)`)
	reLicenseTag   = regexp.MustCompile(`^(License)\s*:\s*(.*)`)
	reGroupTag     = regexp.MustCompile(`^(?i)Group\s*:`)
	reBuildRootTag = regexp.MustCompile(`^(?i)BuildRoot\s*:`)
	reSectionHeader = regexp.MustCompile(`^%([a-zA-Z][a-zA-Z0-9_]*)\s*(.*)`)
)

type SpecFormatter struct{}

func NewSpecFormatter() *SpecFormatter {
	return &SpecFormatter{}
}

func (f *SpecFormatter) Format(content string, opts FormatOptions) (string, []FormatChange, error) {
	var changes []FormatChange

	lines := strings.Split(content, "\n")

	sections := f.parseSections(lines)

	var result []string
	firstSection := true

	for i, sec := range sections {
		if sec.Type == SectionClean && opts.RemoveClean {
			changes = append(changes, FormatChange{
				Line:   f.lineNumberOf(lines, sec.Lines[0], i),
				Type:   "removed",
				Before: strings.Join(sec.Lines, "\n"),
				Reason: "removed deprecated %clean section",
			})
			continue
		}

		var formatted []string
		switch sec.Type {
		case SectionPreamble:
			formatted = f.formatPreamble(sec.Lines, opts, &changes, lines)
		case SectionDescription:
			formatted = f.formatDescription(sec.Lines, opts, &changes)
		case SectionChangelog:
			formatted = sec.Lines
		default:
			formatted = f.formatScriptlet(sec.Lines, opts, &changes)
		}

		if !firstSection {
			result = append(result, "")
		}
		firstSection = false

		result = append(result, formatted...)
	}

	result = f.normalizeTrailingNewlines(result)

	return strings.Join(result, "\n"), changes, nil
}

func (f *SpecFormatter) lineNumberOf(allLines []string, firstLine string, sectionIdx int) int {
	for i, l := range allLines {
		if l == firstLine {
			return i + 1
		}
	}
	return sectionIdx + 1
}

func (f *SpecFormatter) parseSections(lines []string) []Section {
	var sections []Section

	if len(lines) == 0 {
		return sections
	}

	current := Section{Type: SectionPreamble, Lines: []string{}}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "%") && !strings.HasPrefix(trimmed, "%%") {
			matches := reSectionHeader.FindStringSubmatch(trimmed)
			if matches != nil {
				secName := strings.ToLower(matches[1])

				if isSectionKeyword(secName) {
					if len(current.Lines) > 0 {
						sections = append(sections, current)
					} else {
						sections = append(sections, current)
					}

					sType := sectionTypeFromName(secName)
					current = Section{
						Type: sType,
						Name: trimmed,
						Lines: []string{line},
					}

					if secName == "package" && len(matches) > 2 {
						current.SubPkg = strings.TrimSpace(matches[2])
					}
					continue
				}
			}
		}

		current.Lines = append(current.Lines, line)
	}

	if len(current.Lines) > 0 {
		sections = append(sections, current)
	}

	return sections
}

func isSectionKeyword(name string) bool {
	switch name {
	case "description", "prep", "build", "install", "check", "clean",
		"files", "changelog", "package",
		"post", "postun", "pre", "preun", "posttrans", "pretrans",
		"triggerin", "triggerun", "triggerpostun",
		"verifyscript",
		"filetriggerin", "filetriggerun", "filetriggerpostun",
		"transfiletriggerin", "transfiletriggerun", "transfiletriggerpostun":
		return true
	}
	return false
}

func sectionTypeFromName(name string) SectionType {
	switch name {
	case "description":
		return SectionDescription
	case "prep":
		return SectionPrep
	case "build":
		return SectionBuild
	case "install":
		return SectionInstall
	case "check":
		return SectionCheck
	case "clean":
		return SectionClean
	case "files":
		return SectionFiles
	case "changelog":
		return SectionChangelog
	case "package":
		return SectionPackage
	default:
		return SectionScriptlet
	}
}

func (f *SpecFormatter) formatPreamble(lines []string, opts FormatOptions, changes *[]FormatChange, allLines []string) []string {
	var result []string

	depGroups := make(map[string][]string)
	var depOrder []string
	inDeps := false
	lastDepTag := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if opts.TabToSpaces {
			line = strings.ReplaceAll(line, "\t", strings.Repeat(" ", opts.IndentSize))
		}
		line = strings.TrimRight(line, " \t")

		if opts.RemoveBuildRoot && reBuildRootTag.MatchString(trimmed) {
			*changes = append(*changes, FormatChange{
				Line:   i + 1,
				Type:   "removed",
				Before: trimmed,
				Reason: "removed deprecated BuildRoot field",
			})
			continue
		}

		if opts.RemoveGroup && reGroupTag.MatchString(trimmed) {
			*changes = append(*changes, FormatChange{
				Line:   i + 1,
				Type:   "removed",
				Before: trimmed,
				Reason: "removed deprecated Group field",
			})
			continue
		}

		if opts.LicenseSPDX {
			if m := reLicenseTag.FindStringSubmatch(trimmed); m != nil {
				licenseVal := strings.TrimSpace(m[2])
				if fixed, ok := licenseMap[licenseVal]; ok {
					before := trimmed
					trimmed = "License: " + fixed
					line = "License: " + fixed
					*changes = append(*changes, FormatChange{
						Line:   i + 1,
						Type:   "modified",
						Before: before,
						After:  trimmed,
						Reason: fmt.Sprintf("normalized License to SPDX: %s -> %s", licenseVal, fixed),
					})
				}
			}
		}

		if isDepTag(trimmed) {
			m := reConditionTag.FindStringSubmatch(trimmed)
			if m != nil {
				tag := m[1]
				if m[2] != "" {
					tag += m[2]
				}
				value := strings.TrimSpace(m[3])

				if inDeps && tag != lastDepTag {
					inDeps = false
					flushDeps(&result, depGroups, depOrder, changes, opts)
					depGroups = make(map[string][]string)
					depOrder = nil
				}

				if !inDeps {
					inDeps = true
				}
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
					Line:   i + 1,
					Type:   "modified",
					Before: line,
					After:  newLine,
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

	return f.collapseBlankLines(result)
}

func flushDeps(result *[]string, groups map[string][]string, order []string, changes *[]FormatChange, opts FormatOptions) {
	for _, tag := range order {
		deps := groups[tag]
		if opts.SortDeps {
			origLen := len(deps)
			deps = sortUniq(deps)
			if len(deps) < origLen {
				*changes = append(*changes, FormatChange{
					Line:   len(*result) + 1,
					Type:   "reordered",
					Reason: fmt.Sprintf("sorted and deduplicated %s (%d -> %d)", tag, origLen, len(deps)),
				})
			}
		}
		for _, dep := range deps {
			line := tag + ": " + dep
			*result = append(*result, line)
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

func (f *SpecFormatter) formatDescription(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		if opts.TabToSpaces {
			line = strings.ReplaceAll(line, "\t", strings.Repeat(" ", opts.IndentSize))
		}
		line = strings.TrimRight(line, " \t")
		if i > 0 && opts.Curlify {
			newLine := f.curlifyLine(line)
			if newLine != line {
				*changes = append(*changes, FormatChange{
					Line:   i + 1,
					Type:   "modified",
					Before: line,
					After:  newLine,
					Reason: "normalized macro brackets",
				})
				line = newLine
			}
		}
		result = append(result, line)
	}
	return f.collapseBlankLines(result)
}

func (f *SpecFormatter) formatScriptlet(lines []string, opts FormatOptions, changes *[]FormatChange) []string {
	var result []string
	for i, line := range lines {
		if opts.TabToSpaces {
			line = strings.ReplaceAll(line, "\t", strings.Repeat(" ", opts.IndentSize))
		}
		line = strings.TrimRight(line, " \t")
		if opts.Curlify {
			newLine := f.curlifyLine(line)
			if newLine != line {
				*changes = append(*changes, FormatChange{
					Line:   i + 1,
					Type:   "modified",
					Before: line,
					After:  newLine,
					Reason: "normalized macro brackets",
				})
				line = newLine
			}
		}
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

		if macroWhitelist[name] {
			return match
		}

		if name[0] == '_' {
			return match
		}

		if len(name) == 1 && (name[0] >= '0' && name[0] <= '9') {
			return match
		}

		if isSectionKeyword(name) {
			return match
		}

		if name == "define" || name == "global" || name == "with" || name == "without" {
			return match
		}

		return "%{" + name + "}"
	})
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
