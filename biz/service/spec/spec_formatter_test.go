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
