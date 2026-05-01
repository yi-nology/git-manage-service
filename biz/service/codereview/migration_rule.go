package codereview

import (
	"path/filepath"
	"strings"
)

type MigrationRule struct{}

func (r *MigrationRule) ID() string { return "migration-check" }

var migrationDirs = []string{
	"migrations", "migrate", "db/migrate", "db/migrations",
	"alembic/versions", "prisma/migrations", "drizzle",
}

func (r *MigrationRule) Check(ctx *RuleContext) ([]*Finding, error) {
	var findings []*Finding

	for _, f := range ctx.Files {
		if f.IsDeleted {
			continue
		}
		if !isMigrationFile(f.NewPath) {
			continue
		}

		hasIndex := false
		hasFK := false

		for _, hunk := range f.Hunks {
			for _, line := range hunk.Lines {
				if line.Type != "add" {
					continue
				}
				lower := strings.ToLower(line.Content)

				if strings.Contains(lower, "drop_") || strings.Contains(lower, "drop table") || strings.Contains(lower, "drop column") {
					findings = append(findings, &Finding{
						RuleID:      r.ID(),
						Source:      "rule",
						Severity:    SeverityHigh,
						FilePath:    f.NewPath,
						NewLine:     line.NewLine,
						Title:       "Destructive migration detected",
						Message:     "Migration contains DROP operations which may cause data loss. Ensure this is intentional.",
						Suggestion:  "Use reversible migrations. Consider soft-delete or deprecation periods.",
						Fingerprint: computeFingerprint(r.ID(), f.NewPath, line.NewLine, line.Content),
					})
				}

				if strings.Contains(lower, "create index") || strings.Contains(lower, "addindex") || strings.Contains(lower, "add_index") {
					hasIndex = true
				}
				if strings.Contains(lower, "foreign") || strings.Contains(lower, "references") {
					hasFK = true
				}
			}
		}

		if hasFK && !hasIndex {
			findings = append(findings, &Finding{
				RuleID:      r.ID(),
				Source:      "rule",
				Severity:    SeverityMedium,
				FilePath:    f.NewPath,
				Title:       "Foreign key without index",
				Message:     "Migration adds a foreign key but no corresponding index. This may cause slow queries.",
				Suggestion:  "Add an index on the foreign key column for better query performance.",
				Fingerprint: computeFingerprint(r.ID(), f.NewPath, 0, "fk-no-index"),
			})
		}
	}

	return findings, nil
}

func isMigrationFile(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != ".sql" && ext != ".go" && ext != ".py" && ext != ".ts" && ext != ".js" && ext != ".rb" {
		return false
	}
	dir := filepath.Dir(path)
	for _, md := range migrationDirs {
		if strings.Contains(dir, md) {
			return true
		}
	}
	if strings.Contains(base, "migration") || strings.HasPrefix(base, "V") || strings.HasPrefix(base, "U") {
		return true
	}
	return false
}
