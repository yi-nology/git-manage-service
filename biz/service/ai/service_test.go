package ai

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/api"
)

func TestParseJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkVal func(*api.AIDiagnosisResponse) bool
	}{
		{
			name:    "valid json in markdown",
			input:   "```json\n{\"rootCause\":\"test\",\"canAutoFix\":true}\n```",
			wantErr: false,
			checkVal: func(r *api.AIDiagnosisResponse) bool {
				return r.RootCause == "test" && r.CanAutoFix
			},
		},
		{
			name:    "valid json without markdown",
			input:   "{\"rootCause\":\"direct\",\"canAutoFix\":false}",
			wantErr: false,
			checkVal: func(r *api.AIDiagnosisResponse) bool {
				return r.RootCause == "direct" && !r.CanAutoFix
			},
		},
		{
			name:    "invalid json",
			input:   "not json at all",
			wantErr: true,
		},
		{
			name:    "partial json with text around",
			input:   "some text {\"rootCause\":\"embedded\"} more text",
			wantErr: false,
			checkVal: func(r *api.AIDiagnosisResponse) bool {
				return r.RootCause == "embedded"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result api.AIDiagnosisResponse
			err := parseJSONResponse(tt.input, &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkVal != nil && !tt.checkVal(&result) {
				t.Errorf("parseJSONResponse() got = %+v, check failed", result)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int64
	}{
		{"int", 42, 42},
		{"int64", int64(100), 100},
		{"uint", uint(50), 50},
		{"nil", nil, 0},
		{"bool", true, 0},
		{"string", "123", 0},
		{"float64", float64(3.14), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toInt64(tt.input); got != tt.want {
				t.Errorf("toInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildSyncFailureContext(t *testing.T) {
	logs := "error: failed to push some refs"
	stderr := "! [rejected]        main -> main (non-fast-forward)"
	currentBranch := "main"
	trackingBranch := "origin/main"
	recentActions := []string{"commit", "pull", "push"}
	maxLen := 500

	result := BuildSyncFailureContext(logs, stderr, currentBranch, trackingBranch, recentActions, maxLen)

	if len(result) > maxLen {
		t.Errorf("BuildSyncFailureContext() result too long: got %d, max %d", len(result), maxLen)
	}

	if !contains(result, "main") || !contains(result, "non-fast-forward") {
		t.Errorf("BuildSyncFailureContext() missing expected content: got %s", result)
	}
}

func TestBuildRepoContext(t *testing.T) {
	repoKey := "test/repo"
	defaultBranch := "main"
	branchCount := int64(5)
	tagCount := int64(10)
	commitCount := int64(1000)

	result := BuildRepoContext(repoKey, defaultBranch, branchCount, tagCount, commitCount)

	if !contains(result, repoKey) || !contains(result, defaultBranch) {
		t.Errorf("BuildRepoContext() missing expected content: got %s", result)
	}
}

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
	if svc.runner == nil {
		t.Error("NewService() runner is nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
