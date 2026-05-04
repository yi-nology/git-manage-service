package llm

type ModelOption struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type ProviderPreset struct {
	ID                string        `json:"id"`
	DisplayName       string        `json:"display_name"`
	Category          string        `json:"category"`
	Type              string        `json:"type"`
	BaseURL           string        `json:"base_url"`
	AnthropicURL      string        `json:"anthropic_url,omitempty"`
	SupportsAnthropic bool          `json:"supports_anthropic"`
	Models            []ModelOption `json:"models"`
	DefaultModel      string        `json:"default_model"`
	MaxTokens         int           `json:"max_tokens"`
	Icon              string        `json:"icon"`
	Region            string        `json:"region"`
	RequiresKey       bool          `json:"requires_key"`
	IsCodingPlan      bool          `json:"is_coding_plan"`
	CodingPlanPrice   string        `json:"coding_plan_price,omitempty"`
	Warning           string        `json:"warning,omitempty"`
	SubscribeURL      string        `json:"subscribe_url,omitempty"`
}

var Presets = []ProviderPreset{
	// ====== 📦 Coding Plan ======
	{
		ID: "aliyun-coding-plan", DisplayName: "阿里云百炼 Coding Plan",
		Category: "coding_plan", Type: "openai_compatible",
		BaseURL:           "https://coding.dashscope.aliyuncs.com/v1",
		AnthropicURL:      "https://coding.dashscope.aliyuncs.com/apps/anthropic",
		SupportsAnthropic: true,
		Models: []ModelOption{
			{ID: "qwen3.5-plus", DisplayName: "Qwen3.5-Plus (推荐，支持图片理解)"},
			{ID: "qwen3-max-2026-01-23", DisplayName: "Qwen3-Max"},
			{ID: "qwen3-coder-next", DisplayName: "Qwen3-Coder-Next"},
			{ID: "qwen3-coder-plus", DisplayName: "Qwen3-Coder-Plus"},
			{ID: "kimi-k2.5", DisplayName: "Kimi-K2.5 (支持图片理解)"},
			{ID: "glm-5", DisplayName: "GLM-5"},
			{ID: "MiniMax-M2.5", DisplayName: "MiniMax-M2.5"},
		},
		DefaultModel: "qwen3.5-plus", MaxTokens: 8192,
		Icon: "aliyun", Region: "cn", RequiresKey: true,
		IsCodingPlan: true, CodingPlanPrice: "¥200/月",
		Warning:      "Coding Plan 专属 Key（sk-sp-开头）与通用 API Key 不互通，请勿混用。",
		SubscribeURL: "https://www.aliyun.com/benefit/scene/codingplan",
	},
	{
		ID: "zhipu-coding-plan", DisplayName: "智谱 GLM Coding Plan",
		Category: "coding_plan", Type: "openai_compatible",
		BaseURL:           "https://open.bigmodel.cn/api/coding/paas/v4",
		AnthropicURL:      "https://open.bigmodel.cn/api/anthropic",
		SupportsAnthropic: true,
		Models: []ModelOption{
			{ID: "glm-5", DisplayName: "GLM-5 (旗舰)"},
			{ID: "glm-4.7", DisplayName: "GLM-4.7"},
			{ID: "glm-4-flash", DisplayName: "GLM-4-Flash (快速)"},
		},
		DefaultModel: "glm-5", MaxTokens: 8192,
		Icon: "zhipu", Region: "cn", RequiresKey: true,
		IsCodingPlan: true, CodingPlanPrice: "¥49~469/月",
		Warning:      "请使用 Coding Plan 专属端点，勿用通用 API 端点。",
		SubscribeURL: "https://www.bigmodel.cn/glm-coding",
	},
	{
		ID: "volcengine-coding-plan", DisplayName: "火山引擎方舟 Coding Plan",
		Category: "coding_plan", Type: "openai_compatible",
		BaseURL:           "https://ark.cn-beijing.volces.com/api/coding/v3",
		SupportsAnthropic: true,
		AnthropicURL:      "https://ark.cn-beijing.volces.com/api/coding/v3/anthropic",
		Models: []ModelOption{
			{ID: "ark-code-latest", DisplayName: "Auto (自动调度)"},
			{ID: "Doubao-Seed-2.0-pro", DisplayName: "Doubao-Seed-2.0-Pro"},
			{ID: "Doubao-Seed-Code", DisplayName: "Doubao-Seed-Code"},
			{ID: "glm-4.7", DisplayName: "GLM-4.7"},
			{ID: "deepseek-v3.2", DisplayName: "DeepSeek-V3.2"},
			{ID: "kimi-k2.5", DisplayName: "Kimi-K2.5"},
			{ID: "MiniMax-M2.5", DisplayName: "MiniMax-M2.5"},
		},
		DefaultModel: "ark-code-latest", MaxTokens: 8192,
		Icon: "volcengine", Region: "cn", RequiresKey: true,
		IsCodingPlan: true, CodingPlanPrice: "¥40~200/月",
		Warning:      "请使用 Coding Plan 专用 URL，勿用 ark.cn-beijing.volces.com/api/v3（会产生额外费用）。",
		SubscribeURL: "https://www.volcengine.com/activity/codingplan",
	},
	{
		ID: "tencent-coding-plan", DisplayName: "腾讯云 Coding Plan",
		Category: "coding_plan", Type: "openai_compatible",
		BaseURL: "https://coding.tencentcloudapi.com/v1",
		Models: []ModelOption{
			{ID: "hunyuan-2.0-instruct", DisplayName: "Hunyuan-2.0-Instruct"},
			{ID: "glm-5", DisplayName: "GLM-5"},
			{ID: "kimi-k2.5", DisplayName: "Kimi-K2.5"},
			{ID: "MiniMax-M2.5", DisplayName: "MiniMax-M2.5"},
		},
		DefaultModel: "hunyuan-2.0-instruct", MaxTokens: 8192,
		Icon: "tencent", Region: "cn", RequiresKey: true,
		IsCodingPlan: true, CodingPlanPrice: "¥40~200/月",
		SubscribeURL: "https://cloud.tencent.com/act/pro/codingplan",
	},
	{
		ID: "minimax-token-plan", DisplayName: "MiniMax Token Plan",
		Category: "coding_plan", Type: "openai_compatible",
		BaseURL:           "https://api.minimaxi.com/v1",
		SupportsAnthropic: true,
		Models: []ModelOption{
			{ID: "MiniMax-M2.5", DisplayName: "MiniMax-M2.5 (标准)"},
			{ID: "MiniMax-M2.7-highspeed", DisplayName: "MiniMax-M2.7 (极速，~100TPS)"},
		},
		DefaultModel: "MiniMax-M2.5", MaxTokens: 8192,
		Icon: "minimax", Region: "cn", RequiresKey: true,
		IsCodingPlan: true, CodingPlanPrice: "¥290~8990/年",
		Warning:      "Token Plan 专属 Key 与按量计费 Key 不互通。",
		SubscribeURL: "https://platform.minimaxi.com/subscribe/token-plan",
	},
	{
		ID: "kimi-code-plan", DisplayName: "Kimi Code Plan",
		Category: "coding_plan", Type: "openai_compatible",
		BaseURL: "https://api.moonshot.cn/v1",
		Models: []ModelOption{
			{ID: "kimi-k2.5", DisplayName: "Kimi-K2.5"},
		},
		DefaultModel: "kimi-k2.5", MaxTokens: 8192,
		Icon: "kimi", Region: "cn", RequiresKey: true,
		IsCodingPlan: true, CodingPlanPrice: "¥49~699/月",
		SubscribeURL: "https://www.kimi.com/code",
	},

	// ====== 🇨🇳 国内直连 API ======
	{
		ID: "qwen", DisplayName: "通义千问 Qwen",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Models: []ModelOption{
			{ID: "qwen-max", DisplayName: "Qwen-Max (最强)"},
			{ID: "qwen-plus", DisplayName: "Qwen-Plus (均衡)"},
			{ID: "qwen-turbo", DisplayName: "Qwen-Turbo (快速)"},
			{ID: "qwen-coder", DisplayName: "Qwen-Coder (代码专用)"},
		},
		DefaultModel: "qwen-plus", MaxTokens: 8192,
		Icon: "qwen", Region: "cn", RequiresKey: true,
	},
	{
		ID: "deepseek", DisplayName: "DeepSeek",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://api.deepseek.com/v1",
		Models: []ModelOption{
			{ID: "deepseek-coder", DisplayName: "DeepSeek-Coder (代码专用)"},
			{ID: "deepseek-chat", DisplayName: "DeepSeek-Chat"},
		},
		DefaultModel: "deepseek-coder", MaxTokens: 8192,
		Icon: "deepseek", Region: "cn", RequiresKey: true,
	},
	{
		ID: "zhipu", DisplayName: "智谱 AI",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
		Models: []ModelOption{
			{ID: "glm-4", DisplayName: "GLM-4"},
			{ID: "glm-4-flash", DisplayName: "GLM-4-Flash"},
			{ID: "codegeex-4", DisplayName: "CodeGeeX-4 (代码专用)"},
		},
		DefaultModel: "glm-4", MaxTokens: 8192,
		Icon: "zhipu", Region: "cn", RequiresKey: true,
	},
	{
		ID: "baidu", DisplayName: "百度文心",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://qianfan.baidubce.com/v2",
		Models: []ModelOption{
			{ID: "ernie-4.0-8k", DisplayName: "ERNIE-4.0-8K"},
			{ID: "ernie-3.5-8k", DisplayName: "ERNIE-3.5-8K"},
		},
		DefaultModel: "ernie-4.0-8k", MaxTokens: 8192,
		Icon: "baidu", Region: "cn", RequiresKey: true,
	},
	{
		ID: "iflytek", DisplayName: "讯飞星火",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://spark-api.xf-yun.com/v1",
		Models: []ModelOption{
			{ID: "generalv3.5", DisplayName: "星火 v3.5"},
			{ID: "4.0Ultra", DisplayName: "星火 4.0 Ultra"},
		},
		DefaultModel: "generalv3.5", MaxTokens: 8192,
		Icon: "iflytek", Region: "cn", RequiresKey: true,
	},
	{
		ID: "yi", DisplayName: "零一万物 Yi",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://api.lingyiwanwu.com/v1",
		Models: []ModelOption{
			{ID: "yi-large", DisplayName: "Yi-Large"},
			{ID: "yi-medium", DisplayName: "Yi-Medium"},
		},
		DefaultModel: "yi-large", MaxTokens: 8192,
		Icon: "yi", Region: "cn", RequiresKey: true,
	},

	// ====== 🌏 国际直连 API ======
	{
		ID: "openai", DisplayName: "OpenAI",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://api.openai.com/v1",
		Models: []ModelOption{
			{ID: "gpt-4o", DisplayName: "GPT-4o (推荐)"},
			{ID: "gpt-4-turbo", DisplayName: "GPT-4-Turbo"},
			{ID: "gpt-3.5-turbo", DisplayName: "GPT-3.5-Turbo (经济)"},
			{ID: "o1-mini", DisplayName: "o1-mini (推理)"},
		},
		DefaultModel: "gpt-4o", MaxTokens: 8192,
		Icon: "openai", Region: "global", RequiresKey: true,
	},
	{
		ID: "anthropic", DisplayName: "Anthropic Claude",
		Category: "direct", Type: "anthropic",
		BaseURL: "https://api.anthropic.com",
		Models: []ModelOption{
			{ID: "claude-3-5-sonnet-20241022", DisplayName: "Claude 3.5 Sonnet (推荐)"},
			{ID: "claude-3-opus-20240229", DisplayName: "Claude 3 Opus (最强)"},
			{ID: "claude-3-haiku-20240307", DisplayName: "Claude 3 Haiku (快速)"},
		},
		DefaultModel: "claude-3-5-sonnet-20241022", MaxTokens: 8192,
		Icon: "anthropic", Region: "global", RequiresKey: true,
	},
	{
		ID: "google-gemini", DisplayName: "Google Gemini",
		Category: "direct", Type: "gemini",
		BaseURL: "https://generativelanguage.googleapis.com",
		Models: []ModelOption{
			{ID: "gemini-1.5-pro", DisplayName: "Gemini 1.5 Pro"},
			{ID: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash (快速)"},
		},
		DefaultModel: "gemini-1.5-pro", MaxTokens: 8192,
		Icon: "google", Region: "global", RequiresKey: true,
	},
	{
		ID: "mistral", DisplayName: "Mistral",
		Category: "direct", Type: "openai_compatible",
		BaseURL: "https://api.mistral.ai/v1",
		Models: []ModelOption{
			{ID: "codestral-latest", DisplayName: "Codestral (代码专用)"},
			{ID: "mistral-large-latest", DisplayName: "Mistral Large"},
		},
		DefaultModel: "codestral-latest", MaxTokens: 8192,
		Icon: "mistral", Region: "global", RequiresKey: true,
	},

	// ====== 🖥 本地部署 ======
	{
		ID: "ollama", DisplayName: "Ollama",
		Category: "local", Type: "ollama",
		BaseURL: "http://localhost:11434",
		Models: []ModelOption{
			{ID: "codellama", DisplayName: "CodeLlama"},
			{ID: "deepseek-coder", DisplayName: "DeepSeek-Coder"},
			{ID: "qwen2.5-coder", DisplayName: "Qwen2.5-Coder"},
			{ID: "starcoder2", DisplayName: "StarCoder2"},
			{ID: "codegemma", DisplayName: "CodeGemma"},
		},
		DefaultModel: "qwen2.5-coder", MaxTokens: 4096,
		Icon: "ollama", Region: "local", RequiresKey: false,
	},
	{
		ID: "vllm", DisplayName: "vLLM",
		Category: "local", Type: "openai_compatible",
		BaseURL:      "",
		Models:       []ModelOption{},
		DefaultModel: "", MaxTokens: 4096,
		Icon: "vllm", Region: "local", RequiresKey: false,
	},
	{
		ID: "lm-studio", DisplayName: "LM Studio",
		Category: "local", Type: "openai_compatible",
		BaseURL:      "http://localhost:1234/v1",
		Models:       []ModelOption{},
		DefaultModel: "", MaxTokens: 4096,
		Icon: "lmstudio", Region: "local", RequiresKey: false,
	},
}

func GetPresets() []ProviderPreset {
	return Presets
}

func GetPresetsByCategory(category string) []ProviderPreset {
	var result []ProviderPreset
	for _, p := range Presets {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

func GetPresetByID(id string) *ProviderPreset {
	for i := range Presets {
		if Presets[i].ID == id {
			return &Presets[i]
		}
	}
	return nil
}
