<template>
  <div class="author-settings-page">
    <PageHeader title="Git 作者管理" subtitle="管理提交身份和别名" show-back back-route="/settings" />

    <div class="content-area">
      <div v-if="activeIdentity" class="active-card">
        <div class="active-header">
          <span class="active-title">当前激活身份</span>
          <span class="active-badge"><span class="dot"></span>已激活</span>
        </div>
        <div class="active-body">
          <div class="avatar">{{ activeIdentity.canonicalName.charAt(0).toUpperCase() }}</div>
          <div class="active-info">
            <div class="active-name">{{ activeIdentity.canonicalName }}</div>
            <div class="active-email">{{ activeIdentity.canonicalEmail }}</div>
          </div>
        </div>
        <div v-if="activeIdentity.aliases.length > 0" class="active-aliases">
          <span class="alias-label">别名:</span>
          <el-tag v-for="a in activeIdentity.aliases" :key="a.email" size="small" type="info" class="alias-tag">
            {{ a.name }} &lt;{{ a.email }}&gt;
          </el-tag>
        </div>
      </div>

      <div class="list-card">
        <div class="list-header">
          <span class="list-title">所有身份</span>
          <span class="list-count">{{ identities.length }} 个身份</span>
          <el-button type="primary" size="small" :icon="Plus" @click="openCreateDialog">新建身份</el-button>
        </div>

        <div v-loading="loading" class="list-body">
          <div v-for="item in identities" :key="item.id" class="identity-row">
            <div class="row-left">
              <div class="row-avatar" :style="{ background: item.isDefault ? '#6366F1' : '#F59E0B' }">
                {{ item.canonicalName.charAt(0).toUpperCase() }}
              </div>
              <div class="row-info">
                <div class="row-name">{{ item.canonicalName }}</div>
                <div class="row-email">{{ item.canonicalEmail }} · {{ item.aliases.length }} 个别名</div>
              </div>
            </div>
            <div class="row-actions">
              <el-tag v-if="item.isDefault" type="success" size="small" effect="plain">
                <span class="dot-sm"></span>激活
              </el-tag>
              <el-button v-else size="small" @click="handleActivate(item.id)">激活</el-button>
              <el-button size="small" @click="openEditDialog(item)">编辑</el-button>
              <el-button v-if="!item.isDefault" size="small" type="danger" plain @click="handleDelete(item.id)">删除</el-button>
            </div>
          </div>
          <el-empty v-if="identities.length === 0 && !loading" description="暂无身份，请创建" />
        </div>
      </div>

      <div class="hint-card">
        <el-icon :size="16"><InfoFilled /></el-icon>
        <span>激活身份后，新提交将使用该身份的名称和邮箱。别名用于识别历史提交中属于你的记录，支持精确匹配和 Email 匹配两种模式。</span>
      </div>
    </div>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑身份' : '新建身份'" width="520px" destroy-on-close>
      <el-form :model="dialogForm" label-width="100px">
        <el-form-item label="主名" required>
          <el-input v-model="dialogForm.canonicalName" placeholder="如 murphyyi" />
        </el-form-item>
        <el-form-item label="主邮箱" required>
          <el-input v-model="dialogForm.canonicalEmail" placeholder="如 zy84338719@hotmail.com" />
        </el-form-item>
        <el-form-item label="别名">
          <div class="alias-list">
            <div v-for="(a, i) in dialogForm.aliases" :key="i" class="alias-row">
              <el-input v-model="a.name" size="small" placeholder="名称" style="width: 140px" />
              <el-input v-model="a.email" size="small" placeholder="邮箱" style="flex: 1" />
              <el-button size="small" type="danger" :icon="Delete" circle @click="dialogForm.aliases.splice(i, 1)" />
            </div>
            <el-button size="small" @click="dialogForm.aliases.push({ name: '', email: '' })">+ 添加别名</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveDialog" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus, Delete, InfoFilled } from '@element-plus/icons-vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useAuthorIdentity } from '@/composables/useAuthor'
import type { AuthorIdentityDTO, AliasEntry } from '@/api/modules/author'

const { identities, loading, loadIdentities, handleCreate, handleUpdate, handleDelete: doDelete, handleActivate } = useAuthorIdentity()

const activeIdentity = computed(() => identities.value.find(i => i.isDefault) || null)

const showDialog = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const dialogForm = ref<{ canonicalName: string; canonicalEmail: string; aliases: AliasEntry[] }>({
  canonicalName: '',
  canonicalEmail: '',
  aliases: [],
})

function openCreateDialog() {
  editingId.value = null
  dialogForm.value = { canonicalName: '', canonicalEmail: '', aliases: [] }
  showDialog.value = true
}

function openEditDialog(item: AuthorIdentityDTO) {
  editingId.value = item.id
  dialogForm.value = {
    canonicalName: item.canonicalName,
    canonicalEmail: item.canonicalEmail,
    aliases: item.aliases.map(a => ({ ...a })),
  }
  showDialog.value = true
}

async function handleSaveDialog() {
  if (!dialogForm.value.canonicalName || !dialogForm.value.canonicalEmail) return
  saving.value = true
  try {
    if (editingId.value) {
      await handleUpdate(editingId.value, dialogForm.value)
    } else {
      await handleCreate(dialogForm.value)
    }
    showDialog.value = false
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  await doDelete(id)
}

onMounted(() => {
  loadIdentities()
})
</script>

<style scoped>
.author-settings-page { display: flex; flex-direction: column; gap: 20px; }
.content-area { display: flex; flex-direction: column; gap: 20px; padding: 0 0 32px; }

.active-card { display: flex; flex-direction: column; gap: 16px; padding: 24px; background: var(--surface-card); border: 1px solid var(--border-color); border-radius: 12px; }
.active-header { display: flex; justify-content: space-between; align-items: center; }
.active-title { font-size: 16px; font-weight: 600; color: var(--text-color-primary); }
.active-badge { display: flex; align-items: center; gap: 4px; padding: 4px 10px; border-radius: 6px; background: #ECFDF5; color: #059669; font-size: 11px; font-weight: 500; }
.active-badge .dot { width: 6px; height: 6px; border-radius: 50%; background: #10B981; }
.active-body { display: flex; gap: 16px; align-items: center; }
.avatar { width: 56px; height: 56px; border-radius: 28px; background: #6366F1; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 22px; font-weight: 600; flex-shrink: 0; }
.active-info { display: flex; flex-direction: column; gap: 4px; }
.active-name { font-size: 16px; font-weight: 600; color: var(--text-color-primary); }
.active-email { font-size: 13px; color: var(--text-color-secondary); }
.active-aliases { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
.alias-label { font-size: 12px; color: var(--text-color-secondary); font-weight: 500; }
.alias-tag { font-size: 11px; }

.list-card { background: var(--surface-card); border: 1px solid var(--border-color); border-radius: 12px; overflow: hidden; }
.list-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.list-title { font-size: 14px; font-weight: 600; color: var(--text-color-primary); }
.list-count { font-size: 12px; color: var(--text-color-secondary); flex: 1; margin-left: 8px; }
.list-body { display: flex; flex-direction: column; }

.identity-row { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.identity-row:last-child { border-bottom: none; }
.row-left { display: flex; gap: 12px; align-items: center; }
.row-avatar { width: 36px; height: 36px; border-radius: 18px; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 600; flex-shrink: 0; }
.row-info { display: flex; flex-direction: column; gap: 2px; }
.row-name { font-size: 13px; font-weight: 600; color: var(--text-color-primary); }
.row-email { font-size: 12px; color: var(--text-color-secondary); }
.row-actions { display: flex; gap: 8px; align-items: center; }
.dot-sm { display: inline-block; width: 6px; height: 6px; border-radius: 50%; background: #10B981; margin-right: 4px; }

.hint-card { display: flex; gap: 10px; align-items: center; padding: 12px 16px; border-radius: 8px; background: #EFF6FF; border: 1px solid #BFDBFE; font-size: 12px; color: #1E40AF; line-height: 1.5; }

.alias-list { display: flex; flex-direction: column; gap: 8px; width: 100%; }
.alias-row { display: flex; gap: 8px; align-items: center; }
</style>
