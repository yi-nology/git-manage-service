package adapter

import (
	"fmt"
	"sync"
)

var (
	globalInitOnce sync.Once
)

// InitGlobalGitSyncService 全局初始化 git-sync-service
func InitGlobalGitSyncService() {
	globalInitOnce.Do(func() {
		gsCfg := DefaultConfig()
		svc := GetGitSyncService()

		if err := svc.Initialize(gsCfg); err != nil {
			fmt.Printf("[INFO] git-sync-service not initialized: %v\n", err)
			return
		}

		svc.Start()
		fmt.Println("[INFO] git-sync-service initialized and started successfully")
	})
}

// ShutdownGlobalGitSyncService 关闭全局服务
func ShutdownGlobalGitSyncService() {
	svc := GetGitSyncService()
	if svc.IsReady() {
		svc.Stop()
		fmt.Println("[INFO] git-sync-service stopped")
	}
}
