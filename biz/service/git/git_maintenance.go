package git

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

import (
	"github.com/yi-nology/git-manage-service/biz/model/api"
)

type MaintenanceService struct{}

func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{}
}

// 共享 helper：被本包多个维护文件以及 maintenance_snapshot.go 使用
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, infoErr := d.Info()
			if infoErr != nil {
				return infoErr
			}
			size += info.Size()
		}
		return nil
	})
	return size, err
}
func sumSizeBytes(files []api.PrefixFileEntry) int64 {
	var total int64
	for _, f := range files {
		total += f.SizeBytes
	}
	return total
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
