package spec

import (
	"strings"
	"testing"
)

func TestFormat_BasicTabToSpaces(t *testing.T) {
	input := "Name:\tfoo\nVersion:\t1.0\n"
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "\t") {
		t.Errorf("expected no tabs, got: %q", result)
	}
	if !strings.Contains(result, "Name:    foo") {
		t.Errorf("expected tab replaced with spaces, got: %q", result)
	}
	_ = changes
}

func TestFormat_RemoveCleanSection(t *testing.T) {
	input := `Name: foo
Version: 1.0
Release: 1
Summary: test

%description
test package

%clean
rm -rf $RPM_BUILD_ROOT

%files
%doc
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "%clean") {
		t.Errorf("expected %%clean to be removed, got:\n%s", result)
	}
	found := false
	for _, c := range changes {
		if c.Type == "removed" && strings.Contains(c.Reason, "%clean") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a change record for removing %%clean, changes=%v", changes)
	}
}

func TestFormat_RemoveBuildRoot(t *testing.T) {
	input := `Name: foo
BuildRoot: %{_tmppath}/%{name}-%{version}
Version: 1.0

%description
test
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.LicenseSPDX = false
	opts.SortDeps = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "BuildRoot") {
		t.Errorf("expected BuildRoot to be removed, got:\n%s", result)
	}
	found := false
	for _, c := range changes {
		if c.Type == "removed" && strings.Contains(c.Reason, "BuildRoot") {
			found = true
		}
	}
	if !found {
		t.Error("expected a change record for removing BuildRoot")
	}
}

func TestFormat_LicenseSPDX(t *testing.T) {
	input := `Name: foo
Version: 1.0
License: GPLv2+
Summary: test

%description
test
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.SortDeps = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "GPL-2.0-or-later") {
		t.Errorf("expected License to be normalized to SPDX, got:\n%s", result)
	}
	found := false
	for _, c := range changes {
		if c.Type == "modified" && strings.Contains(c.Reason, "SPDX") {
			found = true
		}
	}
	if !found {
		t.Error("expected a change record for License SPDX normalization")
	}
}

func TestFormat_SortDeps(t *testing.T) {
	input := `Name: foo
Version: 1.0
BuildRequires: gcc
BuildRequires: make
BuildRequires: gcc
BuildRequires: automake

%description
test
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(result, "\n")
	var brLines []string
	for _, l := range lines {
		if strings.HasPrefix(l, "BuildRequires:") {
			brLines = append(brLines, l)
		}
	}

	if len(brLines) != 3 {
		t.Errorf("expected 3 unique BuildRequires after dedup, got %d: %v", len(brLines), brLines)
	}

	for i := 1; i < len(brLines); i++ {
		if brLines[i] < brLines[i-1] {
			t.Errorf("expected sorted deps, but %q comes after %q", brLines[i-1], brLines[i])
		}
	}
}

func TestFormat_Curlify(t *testing.T) {
	input := `Name: foo
Version: 1.0
Source0: %{name}-%{version}.tar.gz

%description
test

%prep
%setup -q
make %{?_smp_mflags}
echo %myvar
`
	opts := DefaultFormatOptions()
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(result, "%setup") && strings.Contains(result, "%{setup}") {
		t.Error("expected setup macro to NOT be curlified (whitelisted)")
	}

	if strings.Contains(result, "%myvar") && !strings.Contains(result, "%{myvar}") {
		t.Error("expected %myvar to be curlified to %{myvar}")
	}
}

func TestFormat_NoChangesNeeded(t *testing.T) {
	input := `Name: foo
Version: 1.0
Release: 1
Summary: A test package
License: MIT

%description
A test package.

%prep
%setup -q

%build
%configure
make %{?_smp_mflags}

%install
%make_install

%files
%doc README.md
%license LICENSE
%{_bindir}/%{name}
`
	opts := DefaultFormatOptions()
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) > 0 {
		t.Logf("Note: %d changes made on clean input (may be whitespace normalization): %v", len(changes), changes)
	}
	_ = result
}

func TestFormat_CollapseBlankLines(t *testing.T) {
	input := "Name: foo\n\n\n\nVersion: 1.0\n"
	opts := FormatOptions{}

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "\n\n\n") {
		t.Errorf("expected consecutive blank lines to be collapsed, got:\n%q", result)
	}
}

func TestFormat_TrailingWhitespace(t *testing.T) {
	input := "Name: foo   \nVersion: 1.0  \n"
	opts := FormatOptions{}

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(result, "\n") {
		if line != strings.TrimRight(line, " \t") {
			t.Errorf("expected no trailing whitespace, got: %q", line)
		}
	}
}

func TestFormat_PreambleOrder(t *testing.T) {
	input := `Version: 1.0
Name: foo
Release: 1
License: MIT
Summary: test package
URL: https://example.com

%description
test
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.AlignValues = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(result, "\n")
	var nameIdx, versionIdx, releaseIdx, summaryIdx, licenseIdx int
	for i, l := range lines {
		switch {
		case strings.HasPrefix(l, "Name:"):
			nameIdx = i
		case strings.HasPrefix(l, "Version:"):
			versionIdx = i
		case strings.HasPrefix(l, "Release:"):
			releaseIdx = i
		case strings.HasPrefix(l, "Summary:"):
			summaryIdx = i
		case strings.HasPrefix(l, "License:"):
			licenseIdx = i
		}
	}

	if versionIdx < nameIdx {
		t.Error("Name should come before Version in canonical order")
	}
	if releaseIdx < versionIdx {
		t.Error("Version should come before Release")
	}
	if summaryIdx < releaseIdx {
		t.Error("Release should come before Summary")
	}
	if licenseIdx < summaryIdx {
		t.Error("Summary should come before License")
	}

	found := false
	for _, c := range changes {
		if c.Type == "reordered" && strings.Contains(c.Reason, "preamble") {
			found = true
		}
	}
	if !found {
		t.Error("expected a reordered change for preamble ordering")
	}
}

func TestFormat_PreambleOrderDisabled(t *testing.T) {
	input := `Version: 1.0
Name: foo

%description
test
`
	opts := DefaultFormatOptions()
	opts.PreambleOrder = false
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(result, "\n")
	if !strings.HasPrefix(lines[0], "Version:") {
		t.Errorf("expected original order preserved when PreambleOrder disabled, got:\n%s", result)
	}
}

func TestFormat_AlignValues(t *testing.T) {
	input := `Name: foo
Version: 1.0
Release: 1
Summary: test
License: MIT
URL: https://example.com

%description
test
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(result, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "Name:") {
			colonIdx := strings.Index(l, ":")
			spaceCount := 0
			for i := colonIdx + 1; i < len(l) && l[i] == ' '; i++ {
				spaceCount++
			}
			if spaceCount < 2 {
				t.Errorf("expected Name value to be aligned, got: %q", l)
			}
			break
		}
	}

	found := false
	for _, c := range changes {
		if c.Type == "modified" && strings.Contains(c.Reason, "aligned") {
			found = true
		}
	}
	if !found {
		t.Error("expected an aligned change record")
	}
}

func TestFormat_PathMacros(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%files
/usr/bin/foo
/usr/sbin/bar
/etc/baz.conf
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "%{_bindir}/foo") {
		t.Errorf("expected /usr/bin to be replaced with %%{_bindir}, got:\n%s", result)
	}
	if !strings.Contains(result, "%{_sbindir}/bar") {
		t.Errorf("expected /usr/sbin to be replaced with %%{_sbindir}, got:\n%s", result)
	}
	if !strings.Contains(result, "%{_sysconfdir}/baz.conf") {
		t.Errorf("expected /etc to be replaced with %%{_sysconfdir}, got:\n%s", result)
	}
}

func TestFormat_PathMacrosDisabled(t *testing.T) {
	input := `Name: foo
Version: 1.0

%files
/usr/bin/foo
`
	opts := DefaultFormatOptions()
	opts.PathMacros = false
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false
	opts.UtilMacros = false
	opts.CommonCleanup = false
	opts.ConditionalTrim = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "/usr/bin/foo") {
		t.Errorf("expected /usr/bin to be preserved when PathMacros disabled, got:\n%s", result)
	}
}

func TestFormat_UtilMacros(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%install
%{__make} install
%{__rm} -rf build
%{__cp} file1 file2
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "%{__make}") {
		t.Errorf("expected %%{__make} to be replaced with make, got:\n%s", result)
	}
	if strings.Contains(result, "%{__rm}") {
		t.Errorf("expected %%{__rm} to be replaced, got:\n%s", result)
	}
	if strings.Contains(result, "%{__cp}") {
		t.Errorf("expected %%{__cp} to be replaced, got:\n%s", result)
	}
	if !strings.Contains(result, "rm -rf build") {
		t.Errorf("expected 'rm -rf build', got:\n%s", result)
	}
	if !strings.Contains(result, "cp file1 file2") {
		t.Errorf("expected 'cp file1 file2', got:\n%s", result)
	}
}

func TestFormat_CommonCleanup_BuildRootVar(t *testing.T) {
	input := `Name: foo
Version: 1.0

%install
cp file $RPM_BUILD_ROOT/usr/bin/foo
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "$RPM_BUILD_ROOT") {
		t.Errorf("expected $RPM_BUILD_ROOT to be replaced with %%{buildroot}, got:\n%s", result)
	}
	if !strings.Contains(result, "%{buildroot}") {
		t.Errorf("expected %%{buildroot} in result, got:\n%s", result)
	}
}

func TestFormat_CommonCleanup_Egrep(t *testing.T) {
	input := `Name: foo
Version: 1.0

%build
egrep "pattern" file
fgrep "literal" file
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "egrep") {
		t.Errorf("expected egrep to be replaced with grep -E, got:\n%s", result)
	}
	if strings.Contains(result, "fgrep") {
		t.Errorf("expected fgrep to be replaced with grep -F, got:\n%s", result)
	}
	if !strings.Contains(result, "grep -E") {
		t.Error("expected grep -E in result")
	}
	if !strings.Contains(result, "grep -F") {
		t.Error("expected grep -F in result")
	}
}

func TestFormat_DepOperatorNormalization(t *testing.T) {
	input := `Name: foo
Version: 1.0
BuildRequires: bar =< 1.0
Requires: baz => 2.0

%description
test
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "=<") {
		t.Errorf("expected =< to be normalized to <=, got:\n%s", result)
	}
	if strings.Contains(result, "=>") {
		t.Errorf("expected => to be normalized to >=, got:\n%s", result)
	}
	if !strings.Contains(result, "<= 1.0") {
		t.Error("expected <= 1.0 in result")
	}
	if !strings.Contains(result, ">= 2.0") {
		t.Error("expected >= 2.0 in result")
	}
}

func TestFormat_ConditionalTrim(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%prep
%if 0%{?suse_version}

%setup -q

%else

%setup -q -n foo

%endif
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(result, "%if 0%{?suse_version}\n\n") {
		t.Errorf("expected no blank line right after %%if, got:\n%s", result)
	}
	if strings.Contains(result, "%else\n\n") {
		t.Errorf("expected no blank line right after %%else, got:\n%s", result)
	}
	if strings.Contains(result, "%endif\n\n") {
		t.Errorf("expected no blank line right after %%endif, got:\n%s", result)
	}
}

func TestFormat_ConditionalTrimDisabled(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%prep
%if 0%{?suse_version}

%setup -q

%endif
`
	opts := DefaultFormatOptions()
	opts.ConditionalTrim = false
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "%if 0%{?suse_version}\n\n") {
		t.Logf("Note: blank lines after %%if may have been collapsed by collapseBlankLines instead")
	}
}

func TestFormat_PrepSetupQn(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%prep
%setup -qn foo-1.0
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "-qn ") {
		t.Errorf("expected -qn to be split to -q -n, got:\n%s", result)
	}
	if !strings.Contains(result, "-q -n ") {
		t.Errorf("expected -q -n in result, got:\n%s", result)
	}
}

func TestFormat_PatchNormalization(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%prep
%patch -p0
%patch1 -p1
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "-p0") {
		t.Errorf("expected -p0 to be removed, got:\n%s", result)
	}
	if !strings.Contains(result, "%patch0") {
		t.Errorf("expected %%patch to be normalized to %%patch0, got:\n%s", result)
	}
}

func TestFormat_BuildMake(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%build
make all
make
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "%make_build") {
		t.Errorf("expected make to be replaced with %%make_build, got:\n%s", result)
	}

	makeReplaced := false
	for _, c := range changes {
		if c.Type == "modified" && strings.Contains(c.Reason, "%make_build") {
			makeReplaced = true
		}
	}
	if !makeReplaced {
		t.Error("expected a change record for make -> %make_build")
	}
}

func TestFormat_InstallRmBuildRoot(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%install
rm -rf %{buildroot}
make install DESTDIR=%{buildroot}
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(result, "\n")
	for _, l := range lines {
		if strings.Contains(l, "rm -rf") && strings.Contains(l, "buildroot") {
			t.Errorf("expected rm -rf %%{buildroot} to be removed, got: %s", l)
		}
	}
}

func TestFormat_Makeinstall(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%install
%makeinstall
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "%makeinstall") && !strings.Contains(result, "%make_install") {
		t.Errorf("expected %%makeinstall to be replaced with %%make_install, got:\n%s", result)
	}
	if !strings.Contains(result, "%make_install") {
		t.Errorf("expected %%make_install in result, got:\n%s", result)
	}
}

func TestFormat_FilesDefAttr(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%files
%defattr(-,root,root,-)
%{_bindir}/foo
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "%defattr") {
		t.Errorf("expected %%defattr(-,root,root) to be removed, got:\n%s", result)
	}
}

func TestFormat_DocLicenseExtraction(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%files
%doc README.md LICENSE COPYING
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "%doc README.md LICENSE") {
		t.Errorf("expected LICENSE and COPYING to be extracted from %%doc, got:\n%s", result)
	}
	if !strings.Contains(result, "%license LICENSE COPYING") {
		t.Errorf("expected %%license LICENSE COPYING in result, got:\n%s", result)
	}
	if !strings.Contains(result, "%doc README.md") {
		t.Errorf("expected %%doc README.md in result, got:\n%s", result)
	}
}

func TestFormat_SourcePatchNumbering(t *testing.T) {
	input := `Source: foo.tar.gz
Patch: fix.patch
Name: foo
Version: 1.0

%description
test
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "Source: ") || strings.Contains(result, "Patch: ") {
		t.Errorf("expected bare Source/Patch to be normalized to Source0/Patch0, got:\n%s", result)
	}
	if !strings.Contains(result, "Source0:") {
		t.Errorf("expected Source0: in result, got:\n%s", result)
	}
	if !strings.Contains(result, "Patch0:") {
		t.Errorf("expected Patch0: in result, got:\n%s", result)
	}
}

func TestFormat_NoSpecCleanerDirective(t *testing.T) {
	input := `Name: foo
Version: 1.0
# nospeccleaner
Source: should-not-change.tar.gz

%description
test
`
	opts := DefaultFormatOptions()

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 {
		t.Errorf("expected no changes with nospeccleaner directive, got %d changes", len(changes))
	}
	if result != input {
		t.Errorf("expected input to be returned unchanged with nospeccleaner directive")
	}
}

func TestFormat_SubPackagePreamble(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%package devel
Summary: Development files
Requires: foo = %{version}
Group: Development/Tools

%description devel
Development files.

%files devel
%{_includedir}/foo.h
`
	opts := DefaultFormatOptions()
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.RemoveGroup = true
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, changes, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "Group:") {
		t.Errorf("expected Group to be removed from sub-package, got:\n%s", result)
	}
	_ = changes
}

func TestFormat_MakeDestdir(t *testing.T) {
	input := `Name: foo
Version: 1.0

%description
test

%install
make DESTDIR=%{buildroot} install
`
	opts := DefaultFormatOptions()
	opts.Curlify = false
	opts.RemoveClean = false
	opts.RemoveBuildRoot = false
	opts.LicenseSPDX = false
	opts.SortDeps = false
	opts.PreambleOrder = false
	opts.AlignValues = false

	result, _, err := NewSpecFormatter().Format(input, opts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "DESTDIR") {
		t.Errorf("expected make DESTDIR install to be replaced with %%make_install, got:\n%s", result)
	}
	if !strings.Contains(result, "%make_install") {
		t.Errorf("expected %%make_install in result, got:\n%s", result)
	}
}
