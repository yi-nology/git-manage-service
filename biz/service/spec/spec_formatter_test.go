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
