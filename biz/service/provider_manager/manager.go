package provider_manager

import (
	"fmt"
	"sync"

	sdkprov "github.com/yi-nology/git-platform-sdk/provider"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

var (
	instance *ProviderManager
	once     sync.Once
)

type ProviderManager struct {
	mu    sync.RWMutex
	cache map[uint]sdkprov.Provider
}

func GetManager() *ProviderManager {
	once.Do(func() {
		instance = &ProviderManager{
			cache: make(map[uint]sdkprov.Provider),
		}
	})
	return instance
}

func (m *ProviderManager) GetProvider(configID uint) (sdkprov.Provider, error) {
	m.mu.RLock()
	if p, ok := m.cache[configID]; ok {
		m.mu.RUnlock()
		return p, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	dao := db.NewProviderConfigDAO()
	cfg, err := dao.FindByID(configID)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}

	cred, err := resolveCredential(cfg.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	p, err := newProvider(cfg, cred)
	if err != nil {
		return nil, err
	}
	m.cache[configID] = p
	return p, nil
}

func (m *ProviderManager) Invalidate(configID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cache, configID)
}

func (m *ProviderManager) DetectAndCreate(remoteURL string, credentialID uint) (sdkprov.Provider, *sdkprov.DetectResult, error) {
	result, err := sdkprov.DetectPlatform(remoteURL)
	if err != nil {
		return nil, nil, err
	}

	cred, err := resolveCredential(credentialID)
	if err != nil {
		return nil, nil, err
	}

	p, err := sdkprov.NewProvider(sdkprov.Config{
		Platform: result.Platform,
		BaseURL:  result.BaseURL,
		Token:    cred.Secret,
		SkipTLS:  false,
	})
	if err != nil {
		return nil, nil, err
	}
	return p, result, nil
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

func newProvider(cfg *po.ProviderConfig, cred *po.Credential) (sdkprov.Provider, error) {
	return sdkprov.NewProvider(sdkprov.Config{
		Platform: sdkprov.Platform(cfg.Platform),
		BaseURL:  cfg.BaseURL,
		Token:    cred.Secret,
		SkipTLS:  cfg.SkipTLS,
	})
}
