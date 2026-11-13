package ai

import (
	"strings"

	"github.com/oagentkit/guardx/logredact"
)

// RedactSecrets 在把文本送进 LLM 前脱敏（prompt 输入卫生）。
//
// 规则表委托 guardx/logredact（tier0 零第三方依赖）：内置规则覆盖赋值型
// 凭据（api_key/secret/password/token/bearer）、裸凭据（AWS AKIA/OpenAI sk-）、
// URL 凭据、JSON 值与 IP 掩码，且有截断凭据/幂等性的契约测试钉住——本包
// 原先手写的 7 条正则是其子集，收敛到上游单一口径。
func RedactSecrets(text string) string {
	return logredact.Redact(text)
}

func ClampText(text string, maxChars int) string {
	if maxChars <= 0 || len(text) <= maxChars {
		return text
	}
	head := maxChars / 2
	tail := maxChars - head
	var b strings.Builder
	b.Grow(maxChars + 64)
	b.WriteString(text[:head])
	b.WriteString("\n\n... [truncated by AI input budget] ...\n\n")
	b.WriteString(text[len(text)-tail:])
	return b.String()
}
