package provider_manager

import (
	"fmt"
	"sync"
	"time"

	sdkprov "github.com/yi-nology/git-platform-sdk/provider"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

var (
	instance *ProviderManager
	once     sync.Once
)

// ProviderManager wraps the SDK's caching Manager with DB-aware config
// resolution. It maps configID → sdkprov.Config so that Invalidate() can
// remove the correct cache entry.
type ProviderManager struct {
	mgr     *sdkprov.Manager
	configs map[uint]sdkprov.Config // configID → last known config
	mu      sync.RWMutex
}

func GetManager() *ProviderManager {
	once.Do(func() {
		instance = &ProviderManager{
			mgr:     sdkprov.NewManager(30*time.Minute, sdkprov.WithMaxSize(200)),
			configs: make(map[uint]sdkprov.Config),
		}
	})
	return instance
}

func (m *ProviderManager) GetProvider(configID uint) (sdkprov.Provider, error) {
	dao := db.NewProviderConfigDAO()
	cfg, err := dao.FindByID(configID)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}

	cred, err := resolveCredential(cfg.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	sdkCfg := sdkprov.Config{
		Platform: sdkprov.Platform(cfg.Platform),
		BaseURL:  cfg.BaseURL,
		Token:    cred.Secret,
		SkipTLS:  cfg.SkipTLS,
	}

	p, err := m.mgr.Get(sdkCfg)
	if err != nil {
		return nil, err
	}

	// Store config mapping for invalidation.
	m.mu.Lock()
	m.configs[configID] = sdkCfg
	m.mu.Unlock()

	return p, nil
}

func (m *ProviderManager) Invalidate(configID uint) {
	m.mu.RLock()
	cfg, ok := m.configs[configID]
	m.mu.RUnlock()
	if ok {
		m.mgr.Remove(cfg)
		m.mu.Lock()
		delete(m.configs, configID)
		m.mu.Unlock()
	}
}

func resolveCredential(credentialID uint) (*po.Credential, error) {
	if credentialID == 0 {
		return nil, fmt.Errorf("credential ID is 0")
	}
	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(credentialID)
	if err != nil {
		return nil, fmt.Errorf("credential %d not found: %w", credentialID, err)
	}

	switch cred.Type {
	case "http_token", "http_basic", "platform_token":
	default:
		return nil, fmt.Errorf("credential type '%s' cannot be used for platform API (only http_token, http_basic, platform_token are allowed)", cred.Type)
	}

	return cred, nil
}
