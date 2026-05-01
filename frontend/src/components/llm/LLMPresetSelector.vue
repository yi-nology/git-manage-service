<template>
  <div class="preset-selector">
    <div v-for="group in groups" :key="group.key" class="preset-group">
      <div class="group-header">
        <span class="group-icon">{{ group.icon }}</span>
        <span class="group-title">{{ group.title }}</span>
        <span class="group-desc">{{ group.desc }}</span>
      </div>
      <div class="preset-grid">
        <div
          v-for="preset in group.items"
          :key="preset.id"
          class="preset-card"
          :class="{ 'preset-card--cp': preset.is_coding_plan }"
          @click="$emit('select', preset)"
        >
          <div class="preset-card__icon" :class="'icon--' + preset.icon">
            {{ iconChar(preset) }}
          </div>
          <div class="preset-card__info">
            <div class="preset-card__name">{{ preset.display_name }}</div>
            <div v-if="preset.is_coding_plan && preset.coding_plan_price" class="preset-card__price">
              {{ preset.coding_plan_price }}
            </div>
            <div v-else-if="preset.requires_key" class="preset-card__key-hint">需密钥</div>
            <div v-else class="preset-card__key-hint preset-card__key-hint--free">无需密钥</div>
          </div>
          <div v-if="preset.is_coding_plan" class="preset-card__badge">Coding Plan</div>
          <div v-if="preset.models.length > 0" class="preset-card__models">
            {{ preset.models.length }} 个模型
          </div>
        </div>
      </div>
    </div>

    <div class="custom-entry" @click="$emit('custom')">
      <el-icon :size="16"><Setting /></el-icon>
      <span>自定义配置</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Setting } from '@element-plus/icons-vue'
import type { LLMPresetDTO } from '@/api/modules/llm-settings'

const props = defineProps<{ presets: LLMPresetDTO[] }>()
defineEmits<{ select: [preset: LLMPresetDTO]; custom: [] }>()

interface PresetGroup {
  key: string
  icon: string
  title: string
  desc: string
  items: LLMPresetDTO[]
}

const groups = computed<PresetGroup[]>(() => [
  {
    key: 'coding_plan',
    icon: '📦',
    title: 'Coding Plan',
    desc: '固定月费，一个 Key 访问多个顶级模型',
    items: props.presets.filter(p => p.category === 'coding_plan'),
  },
  {
    key: 'cn',
    icon: '🇨🇳',
    title: '国内模型',
    desc: '按量计费',
    items: props.presets.filter(p => p.category === 'direct' && p.region === 'cn'),
  },
  {
    key: 'global',
    icon: '🌏',
    title: '国际模型',
    desc: '按量计费',
    items: props.presets.filter(p => p.category === 'direct' && p.region === 'global'),
  },
  {
    key: 'local',
    icon: '🖥',
    title: '本地部署',
    desc: '本地运行，无需网络',
    items: props.presets.filter(p => p.category === 'local'),
  },
].filter(g => g.items.length > 0))

function iconChar(preset: LLMPresetDTO): string {
  const map: Record<string, string> = {
    aliyun: '阿', zhipu: '智', volcengine: '火', tencent: '腾',
    minimax: 'M', kimi: 'K', qwen: 'Q', deepseek: 'D', baidu: '百',
    iflytek: '讯', yi: 'Y', openai: 'G', anthropic: 'C', google: 'G',
    mistral: 'M', ollama: 'O', vllm: 'V', lmstudio: 'L',
  }
  return map[preset.icon] || preset.display_name.charAt(0)
}
</script>

<style scoped>
.preset-selector { display: flex; flex-direction: column; gap: 24px; }
.group-header { display: flex; align-items: center; gap: 8px; margin-bottom: 12px; }
.group-icon { font-size: 18px; }
.group-title { font-size: 15px; font-weight: 600; color: var(--text-color-primary); }
.group-desc { font-size: 12px; color: var(--text-color-secondary); }
.preset-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 12px; }
.preset-card {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 16px 12px; border-radius: 10px; border: 1px solid var(--border-color);
  background: var(--bg-color-page); cursor: pointer; transition: all 0.2s;
  position: relative;
}
.preset-card:hover { border-color: var(--accent-primary); box-shadow: 0 2px 8px rgba(99,102,241,0.12); transform: translateY(-1px); }
.preset-card--cp { border-color: #10B981; background: linear-gradient(135deg, #F0FDF4 0%, var(--bg-color-page) 100%); }
.preset-card--cp:hover { border-color: #059669; box-shadow: 0 2px 8px rgba(16,185,129,0.15); }
.preset-card__icon {
  width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center;
  justify-content: center; font-size: 16px; font-weight: 700; color: #FFF; flex-shrink: 0;
}
.icon--aliyun { background: #FF6A00; }
.icon--zhipu { background: #4F46E5; }
.icon--volcengine { background: #1664FF; }
.icon--tencent { background: #006EFF; }
.icon--minimax { background: #7C3AED; }
.icon--kimi { background: #1A1A2E; }
.icon--qwen { background: #FF6A00; }
.icon--deepseek { background: #4A90D9; }
.icon--baidu { background: #2932E1; }
.icon--iflytek { background: #0066CC; }
.icon--yi { background: #6366F1; }
.icon--openai { background: #10A37F; }
.icon--anthropic { background: #D4A574; color: #1E293B; }
.icon--google { background: #4285F4; }
.icon--mistral { background: #F97316; }
.icon--ollama { background: #6366F1; }
.icon--vllm { background: #3B82F6; }
.icon--lmstudio { background: #8B5CF6; }
.preset-card__info { text-align: center; display: flex; flex-direction: column; gap: 4px; }
.preset-card__name { font-size: 13px; font-weight: 500; color: var(--text-color-primary); line-height: 1.3; }
.preset-card__price { font-size: 11px; font-weight: 600; color: #059669; background: #ECFDF5; padding: 2px 8px; border-radius: 4px; }
.preset-card__key-hint { font-size: 11px; color: var(--text-color-secondary); }
.preset-card__key-hint--free { color: #059669; }
.preset-card__badge {
  position: absolute; top: 6px; right: 6px; font-size: 9px; font-weight: 600;
  color: #059669; background: #ECFDF5; padding: 2px 6px; border-radius: 4px;
}
.preset-card__models { font-size: 11px; color: var(--text-color-secondary); }
.custom-entry {
  display: flex; align-items: center; gap: 8px; padding: 12px 16px;
  border-radius: 8px; border: 1px dashed var(--border-color); cursor: pointer;
  color: var(--text-color-secondary); font-size: 13px; transition: all 0.2s;
}
.custom-entry:hover { border-color: var(--accent-primary); color: var(--accent-primary); background: var(--accent-bg); }
</style>
