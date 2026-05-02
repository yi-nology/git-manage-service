package spec

import (
	"strings"
	"testing"
)

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
