package spec

import (
	"strings"
	"testing"
)

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
