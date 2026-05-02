package codereview

import (
	"testing"
)

func TestCalculateRisk_NoFindings(t *testing.T) {
	risk := calculateRisk(nil)
	if risk != SeverityInfo {
		t.Errorf("expected info for no findings, got %s", risk)
	}
}

func TestCalculateRisk_CriticalFinding(t *testing.T) {
	risk := calculateRisk([]*Finding{{Severity: SeverityCritical}})
	if risk != SeverityCritical {
		t.Errorf("expected critical, got %s", risk)
	}
}

func TestCalculateRisk_HighFinding(t *testing.T) {
	risk := calculateRisk([]*Finding{{Severity: SeverityHigh}})
	if risk != SeverityHigh {
		t.Errorf("expected high, got %s", risk)
	}
}

func TestCalculateRisk_MultipleMediumEscalates(t *testing.T) {
	findings := []*Finding{
		{Severity: SeverityMedium},
		{Severity: SeverityMedium},
		{Severity: SeverityMedium},
	}
	risk := calculateRisk(findings)
	if risk != SeverityHigh {
		t.Errorf("expected high for 3+ medium, got %s", risk)
	}
}

func TestCalculateRisk_TwoMedium(t *testing.T) {
	findings := []*Finding{
		{Severity: SeverityMedium},
		{Severity: SeverityMedium},
	}
	risk := calculateRisk(findings)
	if risk != SeverityMedium {
		t.Errorf("expected medium for 2 medium, got %s", risk)
	}
}

func TestCalculateRisk_MultipleLowEscalates(t *testing.T) {
	findings := []*Finding{
		{Severity: SeverityLow},
		{Severity: SeverityLow},
		{Severity: SeverityLow},
		{Severity: SeverityLow},
		{Severity: SeverityLow},
	}
	risk := calculateRisk(findings)
	if risk != SeverityMedium {
		t.Errorf("expected medium for 5+ low, got %s", risk)
	}
}

func TestCalculateRisk_FourLow(t *testing.T) {
	findings := []*Finding{
		{Severity: SeverityLow},
		{Severity: SeverityLow},
		{Severity: SeverityLow},
		{Severity: SeverityLow},
	}
	risk := calculateRisk(findings)
	if risk != SeverityLow {
		t.Errorf("expected low for 4 low, got %s", risk)
	}
}

func TestCalculateRisk_SingleLow(t *testing.T) {
	risk := calculateRisk([]*Finding{{Severity: SeverityLow}})
	if risk != SeverityLow {
		t.Errorf("expected low, got %s", risk)
	}
}

func TestCalculateRisk_MixedSeverities(t *testing.T) {
	findings := []*Finding{
		{Severity: SeverityLow},
		{Severity: SeverityMedium},
		{Severity: SeverityCritical},
	}
	risk := calculateRisk(findings)
	if risk != SeverityCritical {
		t.Errorf("critical should dominate, got %s", risk)
	}
}

func TestDeduplicate_Empty(t *testing.T) {
	result := deduplicate(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestDeduplicate_AllSame(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "abc"},
		{Fingerprint: "abc"},
		{Fingerprint: "abc"},
	}
	result := deduplicate(findings)
	if len(result) != 1 {
		t.Errorf("expected 1 unique, got %d", len(result))
	}
}

func TestDeduplicate_AllDifferent(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "a"},
		{Fingerprint: "b"},
		{Fingerprint: "c"},
	}
	result := deduplicate(findings)
	if len(result) != 3 {
		t.Errorf("expected 3 unique, got %d", len(result))
	}
}

func TestCountBySeverity(t *testing.T) {
	findings := []*Finding{
		{Severity: SeverityCritical},
		{Severity: SeverityCritical},
		{Severity: SeverityHigh},
		{Severity: SeverityLow},
	}
	if n := countBySeverity(findings, SeverityCritical); n != 2 {
		t.Errorf("expected 2 critical, got %d", n)
	}
	if n := countBySeverity(findings, SeverityHigh); n != 1 {
		t.Errorf("expected 1 high, got %d", n)
	}
	if n := countBySeverity(findings, SeverityMedium); n != 0 {
		t.Errorf("expected 0 medium, got %d", n)
	}
}

func TestAggregate_EmptyFindings(t *testing.T) {
	result := Aggregate(nil, 0, 0, 0, true)
	if result.Blocked {
		t.Error("empty findings should not be blocked")
	}
	if result.RiskLevel != SeverityInfo {
		t.Errorf("expected info risk, got %s", result.RiskLevel)
	}
}

func TestAggregate_StatsPreserved(t *testing.T) {
	result := Aggregate(nil, 100, 50, 10, false)
	if result.TotalAdd != 100 {
		t.Errorf("expected TotalAdd=100, got %d", result.TotalAdd)
	}
	if result.TotalDel != 50 {
		t.Errorf("expected TotalDel=50, got %d", result.TotalDel)
	}
	if result.FileCount != 10 {
		t.Errorf("expected FileCount=10, got %d", result.FileCount)
	}
}

func TestAggregate_HighBlockReason(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "a", Severity: SeverityHigh, Title: "Issue"},
	}
	result := Aggregate(findings, 10, 5, 3, true)
	if !result.Blocked {
		t.Error("expected blocked on high")
	}
	if result.BlockReason == "" {
		t.Error("expected block reason")
	}
}

func TestAggregate_CriticalBlocked(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "a", Severity: SeverityCritical, Title: "Critical"},
	}
	result := Aggregate(findings, 10, 5, 3, true)
	if !result.Blocked {
		t.Error("expected blocked on critical")
	}
}

func TestAggregate_MediumNotBlocked(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "a", Severity: SeverityMedium, Title: "Medium"},
	}
	result := Aggregate(findings, 10, 5, 3, true)
	if result.Blocked {
		t.Error("medium should not block")
	}
}
