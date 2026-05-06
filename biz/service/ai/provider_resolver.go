package ai

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

type ProviderResolver struct{}

func NewProviderResolver() *ProviderResolver {
	return &ProviderResolver{}
}

func (r *ProviderResolver) Resolve(sel ProviderSelection) (llm.Provider, error) {
	if sel.ID != 0 {
		return llm.ResolveProviderByID(sel.ID)
	}
	if sel.Name != "" {
		if p, err := llm.GetProvider(sel.Name); err == nil {
			return p, nil
		}
		if p, err := llm.GetProviderByDBName(sel.Name); err == nil {
			return p, nil
		}
		return nil, fmt.Errorf("LLM provider %q not found", sel.Name)
	}
	return llm.GetDefaultProvider()
}

func HasAvailableProvider() bool {
	return llm.HasDefaultProvider()
}
