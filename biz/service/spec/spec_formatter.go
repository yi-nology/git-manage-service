package spec

import (
	"regexp"
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
	Type   SectionType
	Name   string
	Lines  []string
	SubPkg string
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
	PreambleOrder   bool
	AlignValues     bool
	PathMacros      bool
	UtilMacros      bool
	CommonCleanup   bool
	ConditionalTrim bool
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
		PreambleOrder:   true,
		AlignValues:     true,
		PathMacros:      true,
		UtilMacros:      true,
		CommonCleanup:   true,
		ConditionalTrim: true,
	}
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
	"GPL": "GPL-1.0-only", "GPLv1": "GPL-1.0-only", "GPL v1": "GPL-1.0-only",
	"GPL2": "GPL-2.0-only", "GPLv2": "GPL-2.0-only", "GPL v2": "GPL-2.0-only",
	"GPLv2+": "GPL-2.0-or-later", "GPL v2+": "GPL-2.0-or-later", "GPL v2 or later": "GPL-2.0-or-later",
	"GPL3": "GPL-3.0-only", "GPLv3": "GPL-3.0-only", "GPL v3": "GPL-3.0-only",
	"GPLv3+": "GPL-3.0-or-later", "GPL v3+": "GPL-3.0-or-later", "GPL v3 or later": "GPL-3.0-or-later",
	"LGPL": "LGPL-2.0-only", "LGPLv2": "LGPL-2.0-only", "LGPL v2": "LGPL-2.0-only",
	"LGPLv2+": "LGPL-2.0-or-later", "LGPL v2+": "LGPL-2.0-or-later",
	"LGPL2.1": "LGPL-2.1-only", "LGPLv2.1": "LGPL-2.1-only", "LGPL v2.1": "LGPL-2.1-only",
	"LGPLv2.1+": "LGPL-2.1-or-later", "LGPL v2.1+": "LGPL-2.1-or-later",
	"LGPL3": "LGPL-3.0-only", "LGPLv3": "LGPL-3.0-only", "LGPL v3": "LGPL-3.0-only",
	"LGPLv3+": "LGPL-3.0-or-later", "LGPL v3+": "LGPL-3.0-or-later",
	"AGPL": "AGPL-3.0-only", "AGPLv3": "AGPL-3.0-only", "AGPL v3": "AGPL-3.0-only", "AGPLv3+": "AGPL-3.0-or-later",
	"MIT License": "MIT", "MIT license": "MIT", "The MIT License": "MIT",
	"BSD": "BSD-3-Clause", "BSD License": "BSD-3-Clause", "BSD license": "BSD-3-Clause",
	"New BSD License": "BSD-3-Clause", "3-Clause BSD": "BSD-3-Clause", "Revised BSD": "BSD-3-Clause",
	"2-Clause BSD": "BSD-2-Clause", "Simplified BSD": "BSD-2-Clause",
	"Apache": "Apache-2.0", "Apache License": "Apache-2.0", "Apache 2": "Apache-2.0",
	"Apache 2.0": "Apache-2.0", "ASL 2.0": "Apache-2.0", "ASL2": "Apache-2.0",
	"Artistic": "Artistic-1.0-Perl", "Artistic License": "Artistic-1.0-Perl", "Artistic 2.0": "Artistic-2.0",
	"Boost": "BSL-1.0", "Boost Software License": "BSL-1.0", "BSL": "BSL-1.0",
	"CDDL": "CDDL-1.0", "CDDLv1": "CDDL-1.0",
	"EPL": "EPL-1.0", "EPL 2.0": "EPL-2.0",
	"MPL": "MPL-2.0", "MPLv2": "MPL-2.0",
	"ISC License": "ISC", "ISC license": "ISC",
	"PSF": "Python-2.0", "Python License": "Python-2.0",
	"ZPL": "ZPL-2.1", "WTFPL": "WTFPL", "Unlicense": "Unlicense",
	"Public Domain": "LicenseRef-Public-Domain", "public domain": "LicenseRef-Public-Domain",
}

var depTags = []string{
	"BuildRequires", "Requires", "Recommends", "Suggests", "Supplements",
	"Enhances", "Conflicts", "Obsoletes", "Provides",
	"Requires(pre)", "Requires(post)", "Requires(preun)", "Requires(postun)",
	"Requires(pretrans)", "Requires(posttrans)",
}

var depTagPrefixes []string

func init() {
	for _, t := range depTags {
		depTagPrefixes = append(depTagPrefixes, t+":")
	}
}

var preambleTagOrder = map[string]int{
	"define": 0, "global": 1,
	"bcond_with": 2, "bcond_without": 3,
	"Name": 10, "Epoch": 11, "Version": 12, "Release": 13,
	"Summary": 14, "License": 15, "Group": 16,
	"URL": 20, "Url": 21,
	"Source": 30, "Source0": 30, "Source1": 31, "Source2": 32, "Source3": 33,
	"Source4": 34, "Source5": 35, "Source6": 36, "Source7": 37, "Source8": 38, "Source9": 39,
	"Patch": 40, "Patch0": 40, "Patch1": 41, "Patch2": 42, "Patch3": 43,
	"Patch4": 44, "Patch5": 45, "Patch6": 46, "Patch7": 47, "Patch8": 48, "Patch9": 49,
	"BuildRequires": 60, "BuildConflicts": 61,
	"Requires": 70, "Requires(pre)": 71, "Requires(post)": 72, "Requires(preun)": 73,
	"Requires(postun)": 74, "Requires(pretrans)": 75, "Requires(posttrans)": 76,
	"Recommends": 80, "Suggests": 81, "Enhances": 82, "Supplements": 83,
	"Conflicts": 90, "Obsoletes": 91, "Provides": 92,
	"BuildArch": 100, "ExclusiveArch": 101, "ExcludeArch": 102,
	"Vendor": 110, "Packager": 111,
}

var pathToMacro = []struct{ From, To string }{
	{"/usr/bin", "%{_bindir}"},
	{"/usr/sbin", "%{_sbindir}"},
	{"/usr/libexec", "%{_libexecdir}"},
	{"/usr/include", "%{_includedir}"},
	{"/usr/share", "%{_datadir}"},
	{"/usr/share/man", "%{_mandir}"},
	{"/usr/share/info", "%{_infodir}"},
	{"/usr/share/doc/packages", "%{_docdir}"},
	{"/etc/init.d", "%{_initddir}"},
	{"/etc", "%{_sysconfdir}"},
	{"/var", "%{_localstatedir}"},
	{"/usr", "%{_prefix}"},
}

var utilMacroMap = map[string]string{
	"__make": "make", "__rm": "rm", "__cp": "cp", "__mv": "mv",
	"__install": "install", "__mkdir_p": "mkdir -p", "__mkdir": "mkdir",
	"__ln_s": "ln -s", "__chmod": "chmod", "__chown": "chown",
	"__sed": "sed", "__awk": "awk", "__grep": "grep", "__cat": "cat",
	"__id_u": "id -u", "__cc": "gcc", "__cxx": "g++",
	"__ln_t": "ln", "__tar": "tar", "__gzip": "gzip",
	"__bzip2": "bzip2", "__xz": "xz", "__lzma": "xz --format-lzma",
}

var (
	reMacroNoBrace   = regexp.MustCompile(`%([A-Za-z_][A-Za-z0-9_]*)`)
	reConditionTag   = regexp.MustCompile(`^(BuildRequires|Requires|Recommends|Suggests|Supplements|Enhances|Conflicts|Obsoletes|Provides)(\([^)]*\))?\s*:\s*(.*)`)
	reLicenseTag     = regexp.MustCompile(`^(?i)License\s*:\s*(.*)`)
	reGroupTag       = regexp.MustCompile(`^(?i)Group\s*:`)
	reBuildRootTag   = regexp.MustCompile(`^(?i)BuildRoot\s*:`)
	reSectionHeader  = regexp.MustCompile(`^%([a-zA-Z][a-zA-Z0-9_]*)\s*(.*)`)
	reTagColon       = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_]*(\([^)]*\))?)\s*:\s*(.*)`)
	reSetupLine      = regexp.MustCompile(`^%setup\s+(.*)`)
	rePatchLine      = regexp.MustCompile(`^%patch(\d*)\s+(.*)`)
	reMakeLine       = regexp.MustCompile(`^(make)\b(.*)`)
	reRmBuildRoot    = regexp.MustCompile(`(?i)rm\s+-rf\s+.(?:RPM_BUILD_ROOT|\{RPM_BUILD_ROOT\}|\{buildroot\}|buildroot)`)
	reDefAttrLine    = regexp.MustCompile(`^%defattr\(\s*-?\s*,\s*root\s*,\s*root\s*[-,\s]*\)`)
	reUtilMacroRe    = regexp.MustCompile(`%\{(__\w+)\}`)
	reBuildRootVar   = regexp.MustCompile(`\$(?:RPM_BUILD_ROOT|\{RPM_BUILD_ROOT\})`)
	reOptFlagsVar    = regexp.MustCompile(`\$(?:RPM_OPT_FLAGS|\{RPM_OPT_FLAGS\})`)
	reDeprecatedGrep = regexp.MustCompile(`\b(egrep|fgrep)\b`)
	reDepOperator    = regexp.MustCompile(`(>=|<=|>|<|=)\s*([=<>)])`)
	reMakeinstall    = regexp.MustCompile(`%makeinstall\b`)
	reMakeDestdir    = regexp.MustCompile(`make\s+.*DESTDIR=%\{?buildroot\}?.*install`)
	reSetupQn        = regexp.MustCompile(`-qn\s+`)
	rePatchP0        = regexp.MustCompile(`\s+-p0\b`)
	reNoSourcePatch  = regexp.MustCompile(`^(Source|Patch)\s*:`)
)

type SpecFormatter struct{}

func NewSpecFormatter() *SpecFormatter {
	return &SpecFormatter{}
}

func (f *SpecFormatter) Format(content string, opts FormatOptions) (string, []FormatChange, error) {
	if strings.Contains(content, "nospeccleaner") {
		return content, nil, nil
	}

	var changes []FormatChange
	lines := strings.Split(content, "\n")

	multilineRanges := f.findMultilineDefines(lines)
	sections := f.parseSections(lines, multilineRanges)

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
			formatted = f.formatPreamble(sec.Lines, opts, &changes)
		case SectionDescription:
			formatted = f.formatDescription(sec.Lines, opts, &changes)
		case SectionPrep:
			formatted = f.formatPrep(sec.Lines, opts, &changes)
		case SectionBuild, SectionCheck:
			formatted = f.formatBuild(sec.Lines, opts, &changes)
		case SectionInstall:
			formatted = f.formatInstall(sec.Lines, opts, &changes)
		case SectionFiles:
			formatted = f.formatFiles(sec.Lines, opts, &changes)
		case SectionPackage:
			formatted = f.formatPreamble(sec.Lines, opts, &changes)
		case SectionChangelog:
			formatted = sec.Lines
		default:
			formatted = f.formatScriptlet(sec.Lines, opts, &changes)
		}

		if opts.ConditionalTrim {
			formatted = f.trimConditionals(formatted, &changes)
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

type lineRange struct{ start, end int }

func (f *SpecFormatter) findMultilineDefines(lines []string) []lineRange {
	var ranges []lineRange
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "%define ") || strings.HasPrefix(trimmed, "%global ") {
			if strings.HasSuffix(trimmed, "\\") {
				start := i
				for i < len(lines)-1 && strings.HasSuffix(strings.TrimSpace(lines[i]), "\\") {
					i++
				}
				ranges = append(ranges, lineRange{start, i})
			}
		}
	}
	return ranges
}

func isInMultiline(lineIdx int, ranges []lineRange) bool {
	for _, r := range ranges {
		if lineIdx >= r.start && lineIdx <= r.end {
			return true
		}
	}
	return false
}

func (f *SpecFormatter) lineNumberOf(allLines []string, firstLine string, sectionIdx int) int {
	for i, l := range allLines {
		if l == firstLine {
			return i + 1
		}
	}
	return sectionIdx + 1
}

func (f *SpecFormatter) parseSections(lines []string, multilineRanges []lineRange) []Section {
	var sections []Section
	if len(lines) == 0 {
		return sections
	}

	current := Section{Type: SectionPreamble, Lines: []string{}}

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)

		if isInMultiline(idx, multilineRanges) {
			current.Lines = append(current.Lines, line)
			continue
		}

		if strings.HasPrefix(trimmed, "%") && !strings.HasPrefix(trimmed, "%%") {
			matches := reSectionHeader.FindStringSubmatch(trimmed)
			if matches != nil {
				secName := strings.ToLower(matches[1])
				if isSectionKeyword(secName) {
					sections = append(sections, current)
					sType := sectionTypeFromName(secName)
					current = Section{
						Type:  sType,
						Name:  trimmed,
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
		"triggerin", "triggerun", "triggerpostun", "verifyscript",
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
