<template>
  <div class="repo-detail-page" v-loading="loading">
    <div class="page-header-wrap">
      <PageHeader :title="repo?.name || '仓库详情'" show-back back-route="/local-repos">
        <template #title-suffix>
          <StatusBadge v-if="currentVersion" variant="success" :text="currentVersion" :show-dot="false" />
        </template>
        <template #actions>
          <ActionPill variant="green" :icon="Share" @click="$router.push(`/local-repos/${repoKey}/branches`)">
            分支管理
          </ActionPill>
          <ActionPill variant="amber" :icon="Refresh" @click="$router.push(`/local-repos/${repoKey}/sync`)">
            同步任务
          </ActionPill>
        </template>
      </PageHeader>
    </div>

    <div class="layout-container">
      <div class="left-nav">
        <div class="sidebar-card">
          <div
            v-for="item in sidebarItems"
            :key="item.key"
            class="sidebar-item"
            :class="{ active: activeTab === item.key && !(item as any).route }"
            @click="handleNavSelect(item.key)"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </div>
        </div>
      </div>

      <div class="content-area">
        <div v-show="activeTab === 'info'">
          <div v-if="repo" class="info-card">
            <div class="info-top-row">
              <div class="info-left-col">
                <div class="info-section-header">
                  <SectionTitle title="基本信息" />
                  <ActionPill variant="outline" :icon="Edit" @click="openEditDialog">
                    编辑仓库
                  </ActionPill>
                </div>
                <div class="info-row">
                  <div class="info-field"><span class="info-label">名称</span><span class="info-value info-value--bold">{{ repo.name }}</span></div>
                  <div class="info-field"><span class="info-label">当前版本</span><StatusBadge v-if="currentVersion" variant="success" :text="currentVersion" :show-dot="false" /><span v-else class="info-value">-</span></div>
                </div>
                <div class="info-field"><span class="info-label">本地路径</span><span class="info-value mono">{{ repo.path }}</span></div>
                <div class="info-row">
                  <div class="info-field"><span class="info-label">Repo Key</span><span class="info-value info-value--accent">{{ repo.key }}<button class="copy-btn-sm" @click="copyKey">复制</button></span></div>
                  <div class="info-field"><span class="info-label">远程 URL</span><span class="info-value">{{ repo.remote_url || '-' }}</span></div>
                </div>
                <div class="info-row">
                  <div class="info-field"><span class="info-label">创建时间</span><span class="info-value">{{ formatDate(repo.created_at) }}</span></div>
                  <div class="info-field"><span class="info-label">更新时间</span><span class="info-value">{{ formatDate(repo.updated_at) }}</span></div>
                </div>
              </div>

              <div class="info-v-divider"></div>

              <div class="info-right-col">
                <BindingPanel
                  :bindings="bindings"
                  @add="openBindingDialog"
                  @delete="handleDeleteBinding"
                  @set-primary="handleSetPrimaryBinding"
                  @register-webhook="handleRegisterWebhook"
                  @delete-webhook="handleDeleteWebhook"
                />
              </div>
            </div>

            <template v-if="scanData">
              <div class="info-divider"></div>
              <div class="scan-section">
                <div class="info-section-header" style="margin-bottom:12px">
                  <SectionTitle title="远程配置" />
                  <span class="info-subtitle">来自 .git/config</span>
                </div>
                <div class="scan-remote-list">
                  <div v-for="r in scanData.remotes" :key="r.name" class="scan-remote-row">
                    <span class="remote-name">{{ r.name }}</span>
                    <span class="remote-url">{{ r.fetch_url }}</span>
                    <StatusBadge v-if="r.is_mirror" variant="warning" text="Mirror" :show-dot="false" />
                  </div>
                </div>
                <div v-if="scanData.branches?.length" class="tracking-tags">
                  <StatusBadge v-for="b in scanData.branches" :key="b.name" variant="info" :text="`${b.name} -> ${b.upstream_ref}`" :show-dot="false" />
                </div>
              </div>
            </template>
          </div>
        </div>

        <BindingDialog
          v-model:visible="showBindingDialog"
          :repo-key="repoKey"
          :providers="availableProviders"
          @created="loadBindings"
        />

        <div v-show="activeTab === 'spec'" class="spec-full-area">
          <SpecEditor ref="specEditorRef" :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'stats'">
          <el-card>
            <el-form inline class="filter-form">
              <el-form-item label="分支">
                <el-select v-model="statsFilter.branch" placeholder="全部" clearable @change="loadStats" style="width: 220px">
                  <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
                </el-select>
              </el-form-item>
              <el-form-item label="提交人">
                <el-select v-model="statsFilter.author" placeholder="全部" clearable filterable @change="loadStats" style="width: 220px">
                  <el-option v-for="a in statsAuthors" :key="a.email" :label="`${a.name}(${a.email})`" :value="a.name" />
                </el-select>
              </el-form-item>
              <el-form-item label="开始日期">
                <el-date-picker v-model="statsFilter.since" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
              </el-form-item>
              <el-form-item label="结束日期">
                <el-date-picker v-model="statsFilter.until" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="loadStats">
                  <el-icon><Search /></el-icon> 查询
                </el-button>
                <el-button @click="handleExportCsv('stats')">
                  <el-icon><Download /></el-icon> 导出 CSV
                </el-button>
              </el-form-item>
            </el-form>

            <div v-if="statsData">
              <el-row :gutter="16" class="mb-4">
                <el-col :span="12">
                  <el-statistic title="总有效行数" :value="statsData.total_lines" />
                </el-col>
                <el-col :span="12">
                  <el-statistic title="活跃贡献者" :value="statsData.authors?.length || 0" />
                </el-col>
              </el-row>

              <GitStatsCharts :stats-data="statsData" />

              <el-card shadow="never" class="mt-4">
                <template #header><span style="font-weight:600;font-size:14px">提交历史（最近100条）</span></template>
                <el-table :data="commitHistory" border size="small" max-height="400">
                  <el-table-column prop="hash" label="Hash" width="100">
                    <template #default="{ row }">
                      <el-text class="mono-text" size="small">{{ row.hash?.substring(0, 8) }}</el-text>
                    </template>
                  </el-table-column>
                  <el-table-column prop="author" label="作者" width="120" />
                  <el-table-column prop="date" label="时间" width="160">
                    <template #default="{ row }">{{ formatRelativeTime(row.date) }}</template>
                  </el-table-column>
                  <el-table-column prop="message" label="信息" />
                </el-table>
              </el-card>
            </div>
            <el-empty v-else description="点击查询按钮加载数据" />
          </el-card>
        </div>

        <div v-show="activeTab === 'lines'">
          <el-card>
            <el-form inline class="filter-form">
              <el-form-item label="分支">
                <el-select v-model="lineStatsFilter.branch" placeholder="当前工作区" clearable @change="loadLineStats" style="width: 220px">
                  <el-option v-for="b in statsBranches" :key="b" :label="b" :value="b" />
                </el-select>
              </el-form-item>
              <el-form-item label="提交人">
                <el-select v-model="lineStatsFilter.author" placeholder="全部" clearable filterable style="width: 220px">
                  <el-option v-for="a in statsAuthors" :key="a.email" :label="`${a.name}(${a.email})`" :value="a.name" />
                </el-select>
              </el-form-item>
              <el-form-item label="开始日期">
                <el-date-picker v-model="lineStatsFilter.since" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
              </el-form-item>
              <el-form-item label="结束日期">
                <el-date-picker v-model="lineStatsFilter.until" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="loadLineStats">
                  <el-icon><Search /></el-icon> 查询
                </el-button>
                <el-button @click="openExcludeConfig">
                  <el-icon><Setting /></el-icon> 排除配置
                </el-button>
                <el-button @click="handleExportCsv('lines')">
                  <el-icon><Download /></el-icon> 导出 CSV
                </el-button>
              </el-form-item>
            </el-form>

            <el-alert type="info" :closable="false" show-icon class="mb-4">
              选择分支/提交人/时间范围后将使用 git blame 分析代码归属，统计速度会较慢
            </el-alert>

            <div v-loading="lineStatsLoading">
              <div v-if="lineStatsData">
                <el-row :gutter="16" class="mb-4">
                  <el-col :span="6"><el-statistic title="代码行数" :value="lineStatsData.code_lines" /></el-col>
                  <el-col :span="6"><el-statistic title="注释行数" :value="lineStatsData.comment_lines" /></el-col>
                  <el-col :span="6"><el-statistic title="空白行数" :value="lineStatsData.blank_lines" /></el-col>
                  <el-col :span="6"><el-statistic title="文件总数" :value="lineStatsData.total_files" /></el-col>
                </el-row>

                <el-alert v-if="lineStatsData.status === 'processing'" title="正在统计中..." type="info" :closable="false" show-icon>
                  {{ lineStatsData.progress }}
                </el-alert>

                <LineStatsCharts :line-stats-data="lineStatsData" />
              </div>
              <el-empty v-else description="点击查询按钮加载数据" />
            </div>
          </el-card>
        </div>

        <div v-show="activeTab === 'versions'">
          <el-card>
            <template #header>
              <div class="card-header-row">
                <span>版本标签管理</span>
                <div class="header-actions">
                  <el-button size="small" @click="handleFetchTags" :loading="fetchTagsLoading">
                    <el-icon><Download /></el-icon> 拉取远端 Tags
                  </el-button>
                  <el-button size="small" type="primary" @click="openCreateTagDialog">
                    <el-icon><Plus /></el-icon> 创建 Tag
                  </el-button>
                  <el-button size="small" @click="loadVersions">
                    <el-icon><Refresh /></el-icon>
                  </el-button>
                </div>
              </div>
            </template>
            <div v-if="versionList.length === 0 && !versionsLoading">
              <el-empty description="暂无版本标签">
                <el-button type="primary" @click="openCreateTagDialog">创建第一个 Tag</el-button>
              </el-empty>
            </div>
            <el-table v-else :data="versionList" v-loading="versionsLoading" stripe border size="small">
              <el-table-column prop="name" label="标签名称" width="160">
                <template #default="{ row }">
                  <el-tag type="success" size="small">{{ row.name }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="hash" label="Commit" width="120">
                <template #default="{ row }">
                  <el-text class="mono-text" size="small">{{ row.hash?.substring(0, 8) }}</el-text>
                </template>
              </el-table-column>
              <el-table-column prop="tagger" label="作者" width="120" />
              <el-table-column prop="date" label="日期" width="160">
                <template #default="{ row }">{{ formatDate(row.date) }}</template>
              </el-table-column>
              <el-table-column prop="message" label="说明" min-width="200" show-overflow-tooltip />
              <el-table-column label="操作" width="200" fixed="right">
                <template #default="{ row }">
                  <el-button-group size="small">
                    <el-button @click="handlePushTag(row.name)">
                      <el-icon><Top /></el-icon> 推送
                    </el-button>
                    <el-button @click="handleCopyHash(row.hash)">
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                    <el-button type="danger" @click="handleDeleteTag(row.name)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </el-button-group>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>

        <div v-show="activeTab === 'files'" style="height: 100%; min-height: 600px;">
          <FileExplorer :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'commits'">
          <CommitSearch :repo-key="repoKey" :branches="allRefs" :authors="statsAuthors" />
        </div>

        <div v-show="activeTab === 'stash'">
          <StashManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'submodules'">
          <SubmoduleManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'patches'">
          <PatchManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'slim'">
          <SlimManager :repo-key="repoKey" />
        </div>

        <div v-show="activeTab === 'author'">
          <AuthorFix :repo-key="repoKey" :remotes="remoteNames" />
        </div>
      </div>
    </div>

    <el-dialog v-model="showEditDialog" title="编辑仓库" width="750px" destroy-on-close>
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="editForm.name" placeholder="仓库名称" />
        </el-form-item>
        <el-form-item label="本地路径" required>
          <el-input v-model="editForm.path" placeholder="本地仓库路径" />
        </el-form-item>
        <el-form-item label="远程 URL">
          <div class="url-input-group">
            <el-radio-group v-model="editUrlMode" size="small" class="url-mode-switch">
              <el-radio-button value="ssh">SSH</el-radio-button>
              <el-radio-button value="https">HTTPS</el-radio-button>
            </el-radio-group>
            <el-input
              v-model="editForm.remote_url"
              :placeholder="editUrlMode === 'ssh' ? 'git@github.com:user/repo.git' : 'https://github.com/user/repo.git'"
              @blur="validateEditUrl"
              :class="{ 'is-error-input': editUrlError }"
            />
          </div>
          <div v-if="editUrlError" class="field-error">{{ editUrlError }}</div>
        </el-form-item>
        <el-form-item label="默认凭证">
          <CredentialSelector
            v-model="editDefaultCredentialId"
            :url="editForm.remote_url"
            placeholder="选择默认凭证（可选）"
          />
        </el-form-item>

        <el-divider content-position="left">远程仓库配置</el-divider>

        <el-form-item label="">
          <div class="remotes-section">
            <div class="remotes-header">
              <span>配置多个远程仓库及其凭证</span>
              <el-button size="small" type="primary" @click="addEditRemote">+ 新增远程</el-button>
            </div>
            <div v-for="(remote, index) in editRemotes" :key="index" class="edit-remote-item">
              <el-card shadow="hover">
                <div class="edit-remote-row">
                  <el-input v-model="remote.name" size="small" placeholder="名称 (如 origin)" style="width: 120px;" />
                  <el-radio-group v-model="remoteUrlModes[index]" size="small" class="url-mode-switch-sm">
                    <el-radio-button value="ssh">SSH</el-radio-button>
                    <el-radio-button value="https">HTTPS</el-radio-button>
                  </el-radio-group>
                  <el-input v-model="remote.fetch_url" size="small" :placeholder="remoteUrlModes[index] === 'ssh' ? 'git@host:user/repo.git' : 'https://host/repo.git'" style="flex: 1;" @blur="validateRemoteUrl(index)" :class="{ 'is-error-input': remoteUrlErrors[index] }" />
                  <el-button size="small" :icon="Connection" circle @click="testEditRemote(index)" title="测试连接" :loading="remote._testing" />
                  <el-button size="small" :icon="Delete" circle type="danger" @click="removeEditRemote(index)" title="删除" />
                </div>
                <div v-if="remoteUrlErrors[index]" class="field-error" style="margin-left: 128px;">{{ remoteUrlErrors[index] }}</div>
                <div class="edit-remote-cred">
                  <span class="cred-label">凭证:</span>
                  <CredentialSelector
                    :model-value="editRemoteCredentials[remote.name]"
                    :url="remote.fetch_url"
                    placeholder="选择凭证（可选）"
                    @update:model-value="(v) => updateEditRemoteCred(remote.name, v)"
                  />
                </div>
              </el-card>
            </div>
            <el-empty v-if="editRemotes.length === 0" description="无远程仓库配置" :image-size="60" />
          </div>
        </el-form-item>

        <el-form-item v-if="editTrackingBranches.length > 0" label="分支追踪">
          <div class="tracking-branches">
            <el-tag v-for="b in editTrackingBranches" :key="b.name" size="small" style="margin: 2px 4px;">
              {{ b.name }} -> {{ b.upstream_ref }}
            </el-tag>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveEdit" :loading="editSaving">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showExcludeDialog" title="排除配置" width="550px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="排除目录">
          <el-input v-model="excludeDirsText" type="textarea" :rows="4" placeholder="每行一个目录路径" />
        </el-form-item>
        <el-form-item label="排除规则">
          <el-input v-model="excludePatternsText" type="textarea" :rows="4" placeholder="每行一个 glob 规则" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExcludeDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveExclude">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showCreateTagDialog" title="创建版本标签" width="550px" destroy-on-close>
      <el-form :model="createTagForm" label-width="100px">
        <el-form-item label="版本类型">
          <el-radio-group v-model="createTagForm.versionType" @change="handleVersionTypeChange">
            <el-radio-button value="patch">Patch (修复)</el-radio-button>
            <el-radio-button value="minor">Minor (功能)</el-radio-button>
            <el-radio-button value="major">Major (大版本)</el-radio-button>
            <el-radio-button value="custom">自定义</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="当前版本" v-if="nextVersionInfo">
          <el-tag type="info">{{ nextVersionInfo.current || '无' }}</el-tag>
        </el-form-item>
        <el-form-item label="标签名称" required>
          <el-input v-model="createTagForm.name" :disabled="createTagForm.versionType !== 'custom'" placeholder="v1.0.0" />
        </el-form-item>
        <el-form-item label="目标引用">
          <el-input v-model="createTagForm.ref" placeholder="HEAD (默认当前分支最新提交)" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="createTagForm.message" type="textarea" :rows="2" placeholder="版本说明" />
        </el-form-item>
        <el-form-item label="推送到远端">
          <el-select v-model="createTagForm.push_remote" placeholder="不推送" clearable>
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateTagDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateTag" :loading="createTagLoading">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPushTagDialog" :title="`推送标签: ${pushTagName}`" width="480px" destroy-on-close>
      <el-form label-width="90px">
        <el-form-item label="目标远端">
          <el-select v-model="pushTagRemote" placeholder="选择远端">
            <el-option v-for="r in remoteNames" :key="r" :label="r" :value="r" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPushTagDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitPushTag" :loading="pushTagLoading">确认推送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Edit, Search, Download, Setting, Plus, Top, Delete, CopyDocument, Connection, DocumentCopy, InfoFilled, Document, DataAnalysis, Files, Timer, Folder, Box, Link, Share, Operation, User, Grid } from '@element-plus/icons-vue'
import { getRepoDetail, scanRepo, updateRepo, fetchRepo } from '@/api/modules/repo'
import { testConnection } from '@/api/modules/system'
import { testCredential } from '@/api/modules/credential'
import { getStatsAnalyze, getStatsAuthors, getStatsBranches, getStatsCommits, getLineStats, getLineStatsConfig, saveLineStatsConfig, exportStatsCsv } from '@/api/modules/stats'
import { getVersionList, getCurrentVersion, getNextVersion } from '@/api/modules/version'
import type { VersionTag, NextVersionInfo } from '@/api/modules/version'
import { createTag, deleteTag, pushTag } from '@/api/modules/branch'
import { showGitError } from '@/utils/git'
import type { RepoDTO, ScanResult, GitRemote, TrackingBranch } from '@/types/repo'
import type { StatsResponse, LineStatsResponse } from '@/types/stats'
import { formatDate, formatRelativeTime } from '@/utils/format'
import GitStatsCharts from '@/components/stats/GitStatsCharts.vue'
import LineStatsCharts from '@/components/stats/LineStatsCharts.vue'
import FileExplorer from '@/components/repo/FileExplorer.vue'
import CommitSearch from '@/components/repo/CommitSearch.vue'
import StashManager from '@/components/repo/StashManager.vue'
import SubmoduleManager from '@/components/repo/SubmoduleManager.vue'
import PatchManager from '@/components/patch/PatchManager.vue'
import CredentialSelector from '@/components/credential/CredentialSelector.vue'
import SpecEditor from '@/components/spec/SpecEditor.vue'
import SlimManager from '@/components/repo/SlimManager.vue'
import AuthorFix from '@/components/repo/AuthorFix.vue'

import { validateGitRemoteUrl, detectGitProtocol, convertGitUrl } from '@/utils/git'
import { getProvider } from '@/api/modules/provider'
import type { ProviderConfigDTO } from '@/api/modules/provider'
import { listBindings, deleteBinding, setPrimaryBinding, registerBindingWebhook, deleteBindingWebhook } from '@/api/modules/binding'
import type { RepoProviderBindingDTO } from '@/types/binding'
import BindingPanel from '@/components/binding/BindingPanel.vue'
import BindingDialog from '@/components/binding/BindingDialog.vue'
import { useProviderStore } from '@/stores/useProviderStore'

import PageHeader from '@/components/common/PageHeader.vue'
import ActionPill from '@/components/common/ActionPill.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import SectionTitle from '@/components/common/SectionTitle.vue'

const providerStore = useProviderStore()


const route = useRoute()
const router = useRouter()
const repoKey = route.params.repoKey as string

const loading = ref(false)
const repo = ref<RepoDTO | null>(null)
const scanData = ref<ScanResult | null>(null)
const activeTab = ref('info')
const currentVersion = ref('')
const providerInfo = ref<ProviderConfigDTO | null>(null)
const showBindingDialog = ref(false)
const availableProviders = ref<ProviderConfigDTO[]>([])
const bindings = ref<RepoProviderBindingDTO[]>([])
const specEditorRef = ref<{ refresh: () => void; clearEditor: () => void } | null>(null)

const sidebarItems = [
  { key: 'info', label: '基本信息', icon: InfoFilled },
  { key: 'spec', label: 'Spec 编辑器', icon: Document },
  { key: 'stats', label: 'Git 有效提交度量', icon: DataAnalysis },
  { key: 'lines', label: '真实工程代码度量', icon: Files },
  { key: 'versions', label: '版本历史', icon: Timer },
  { key: 'files', label: '文件', icon: Folder },
  { key: 'commits', label: 'Commit 搜索', icon: Search },
  { key: 'stash', label: 'Stash 管理', icon: Box },
  { key: 'submodules', label: 'Submodule', icon: Link },
  { key: 'patches', label: 'Patch 管理', icon: DocumentCopy },
  { key: 'slim', label: '仓库瘦身', icon: Operation },
  { key: 'author', label: '作者修复', icon: User },
]

const statsFilter = ref({ branch: '', author: '', since: '', until: '' })
const lineStatsFilter = ref({ branch: '', author: '', since: '', until: '' })
const statsBranches = ref<string[]>([])
const statsAuthors = ref<{ name: string; email: string }[]>([])
const statsData = ref<StatsResponse | null>(null)
const lineStatsData = ref<LineStatsResponse | null>(null)
const lineStatsLoading = ref(false)
const commitHistory = ref<{ hash: string; author: string; date: string; message: string }[]>([])

const versionList = ref<VersionTag[]>([])
const versionsLoading = ref(false)
const fetchTagsLoading = ref(false)
const remoteNames = ref<string[]>([])

const allRefs = computed(() => {
  const tags = (versionList.value || []).map(v => v.name)
  return [...(statsBranches.value || []), ...tags]
})

const showCreateTagDialog = ref(false)
const createTagLoading = ref(false)
const nextVersionInfo = ref<NextVersionInfo | null>(null)
const createTagForm = ref({
  versionType: 'patch' as 'patch' | 'minor' | 'major' | 'custom',
  name: '',
  ref: 'HEAD',
  message: '',
  push_remote: '',
})

const showPushTagDialog = ref(false)
const pushTagName = ref('')
const pushTagRemote = ref('origin')
const pushTagLoading = ref(false)

const showEditDialog = ref(false)
const editSaving = ref(false)
const editForm = ref({ name: '', path: '', remote_url: '' })

interface EditRemoteRow extends GitRemote {
  _testing?: boolean
}
const editRemotes = ref<EditRemoteRow[]>([])
const editTrackingBranches = ref<TrackingBranch[]>([])
const editDefaultCredentialId = ref<number | undefined>()
const editRemoteCredentials = ref<Record<string, number | undefined>>({})
const editUrlError = ref('')
const remoteUrlErrors = ref<Record<number, string>>({})
const editUrlMode = ref<'ssh' | 'https'>('ssh')
const remoteUrlModes = ref<Record<number, 'ssh' | 'https'>>({})

const showExcludeDialog = ref(false)
const excludeDirsText = ref('')
const excludePatternsText = ref('')

 onMounted(async () => {
  if (route.query.tab && typeof route.query.tab === 'string') {
    activeTab.value = route.query.tab === 'workspace' ? 'files' : route.query.tab
  }
  loading.value = true
  try {
    repo.value = await getRepoDetail(repoKey)
    if (repo.value?.path) {
      try {
        scanData.value = await scanRepo(repo.value.path)
        remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
      } catch { /* ignore */ }
    }
    try {
      statsBranches.value = (await getStatsBranches(repoKey)) || []
      if (statsBranches.value.length > 0) {
        statsFilter.value.branch = statsBranches.value[0]!
        lineStatsFilter.value.branch = statsBranches.value[0]!
      }
    } catch { statsBranches.value = [] }
    try { statsAuthors.value = (await getStatsAuthors(repoKey)) || [] } catch { statsAuthors.value = [] }
    try { currentVersion.value = (await getCurrentVersion(repoKey)) || '' } catch { /* ignore */ }
    try { versionList.value = (await getVersionList(repoKey)) || [] } catch { versionList.value = [] }
    loadProviderInfo()
    loadBindings()
  } finally {
    loading.value = false
  }
})

watch(activeTab, (val) => {
  if (val === 'versions' && (versionList.value || []).length === 0) {
    loadVersions()
  }
  if (val === 'spec') {
    nextTick(() => {
      specEditorRef.value?.clearEditor()
      specEditorRef.value?.refresh()
    })
  }
})

watch(editUrlMode, (newMode, oldMode) => {
  if (oldMode && newMode !== oldMode && editForm.value.remote_url) {
    editForm.value.remote_url = convertGitUrl(editForm.value.remote_url, newMode)
  }
})

watch(remoteUrlModes, (newModes, oldModes) => {
  if (!oldModes) return
  for (const [idxStr, newMode] of Object.entries(newModes)) {
    const idx = parseInt(idxStr)
    const oldMode = oldModes[idx]
    if (oldMode && newMode !== oldMode && editRemotes.value[idx]?.fetch_url) {
      editRemotes.value[idx]!.fetch_url = convertGitUrl(editRemotes.value[idx]!.fetch_url, newMode)
    }
  }
}, { deep: true })

async function loadStats() {
  try {
    statsData.value = await getStatsAnalyze(repoKey, {
      branch: statsFilter.value.branch || undefined,
      author: statsFilter.value.author || undefined,
      since: statsFilter.value.since || undefined,
      until: statsFilter.value.until || undefined,
    })
    const res = await getStatsCommits(repoKey, {
      branch: statsFilter.value.branch || undefined,
      author: statsFilter.value.author || undefined,
      since: statsFilter.value.since || undefined,
      until: statsFilter.value.until || undefined,
    })
    commitHistory.value = (Array.isArray(res) ? res : []).slice(0, 100)
  } catch { /* ignore */ }
}

async function loadLineStats() {
  try {
    lineStatsLoading.value = true
    const result = await getLineStats(repoKey, {
      branch: lineStatsFilter.value.branch || undefined,
      author: lineStatsFilter.value.author || undefined,
      since: lineStatsFilter.value.since || undefined,
      until: lineStatsFilter.value.until || undefined,
    })
    lineStatsData.value = result
    if (result && result.status === 'processing') {
      lineStatsLoading.value = false
      pollLineStats()
    } else {
      lineStatsLoading.value = false
    }
  } catch {
    lineStatsLoading.value = false
  }
}

function pollLineStats() {
  const timer = setTimeout(async () => {
    try {
      const result = await getLineStats(repoKey, {
        branch: lineStatsFilter.value.branch || undefined,
        author: lineStatsFilter.value.author || undefined,
        since: lineStatsFilter.value.since || undefined,
        until: lineStatsFilter.value.until || undefined,
      })
      lineStatsData.value = result
      if (result && result.status === 'processing') {
        pollLineStats()
      }
    } catch { /* ignore */ }
  }, 2000)
  void timer
}

async function loadVersions() {
  versionsLoading.value = true
  try {
    versionList.value = (await getVersionList(repoKey)) || []
  } catch { /* ignore */ }
  finally {
    versionsLoading.value = false
  }
}

async function handleFetchTags() {
  fetchTagsLoading.value = true
  try {
    await fetchRepo(repoKey)
    ElMessage.success('远端 Tags 拉取成功')
    await loadVersions()
  } catch {
    ElMessage.error('拉取远端 Tags 失败')
  } finally {
    fetchTagsLoading.value = false
  }
}

async function openCreateTagDialog() {
  createTagForm.value = { versionType: 'patch', name: '', ref: 'HEAD', message: '', push_remote: '' }
  nextVersionInfo.value = null
  showCreateTagDialog.value = true
  try {
    nextVersionInfo.value = await getNextVersion(repoKey)
    handleVersionTypeChange(createTagForm.value.versionType)
  } catch { /* ignore */ }
}

function handleVersionTypeChange(type: string | number | boolean) {
  if (!nextVersionInfo.value) return
  switch (type) {
    case 'patch':
      createTagForm.value.name = nextVersionInfo.value.next_patch
      break
    case 'minor':
      createTagForm.value.name = nextVersionInfo.value.next_minor
      break
    case 'major':
      createTagForm.value.name = nextVersionInfo.value.next_major
      break
    case 'custom':
      createTagForm.value.name = ''
      break
  }
}

async function handleCreateTag() {
  if (!createTagForm.value.name) {
    ElMessage.warning('标签名称不能为空')
    return
  }
  createTagLoading.value = true
  try {
    await createTag({
      repo_key: repoKey,
      name: createTagForm.value.name,
      ref: createTagForm.value.ref || 'HEAD',
      message: createTagForm.value.message,
      push_remote: createTagForm.value.push_remote || undefined,
    })
    ElMessage.success(`标签 ${createTagForm.value.name} 创建成功`)
    showCreateTagDialog.value = false
    await loadVersions()
    try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch { /* ignore */ }
  } catch (e: unknown) {
    const err = e as { message?: string }
    ElMessage.error('创建标签失败: ' + (err.message || '未知错误'))
  } finally {
    createTagLoading.value = false
  }
}

function handlePushTag(tagName: string) {
  pushTagName.value = tagName
  pushTagRemote.value = remoteNames.value[0] || 'origin'
  showPushTagDialog.value = true
}

async function handleSubmitPushTag() {
  pushTagLoading.value = true
  try {
    await pushTag({ repo_key: repoKey, tag_name: pushTagName.value, remote_name: pushTagRemote.value })
    ElMessage.success(`标签 ${pushTagName.value} 已推送到 ${pushTagRemote.value}`)
    showPushTagDialog.value = false
  } catch (e: unknown) {
    showGitError(e, '推送标签')
  } finally {
    pushTagLoading.value = false
  }
}

async function handleDeleteTag(tagName: string) {
  try {
    await ElMessageBox.confirm(`确定要删除标签 "${tagName}" 吗？`, '删除标签', {
      confirmButtonText: '仅删除本地',
      cancelButtonText: '取消',
      type: 'warning',
      distinguishCancelAndClose: true,
    })
    await deleteTag({ repo_key: repoKey, name: tagName })
    ElMessage.success(`标签 ${tagName} 已删除`)
    await loadVersions()
    try { currentVersion.value = await getCurrentVersion(repoKey) || '' } catch { /* ignore */ }
  } catch (action) {
    if (action === 'cancel' || action === 'close') return
    const err = action as { message?: string }
    ElMessage.error('删除标签失败: ' + (err.message || '未知错误'))
  }
}

function handleCopyHash(hash: string) {
  if (hash) {
    navigator.clipboard.writeText(hash)
    ElMessage.success('已复制 Commit Hash')
  }
}

function copyKey() {
  if (repo.value?.key) {
    navigator.clipboard.writeText(repo.value.key)
    ElMessage.success('已复制 Repo Key')
  }
}

async function loadProviderInfo() {
  if (!repo.value?.provider_config_id) { providerInfo.value = null; return }
  const cached = providerStore.getProviderById(repo.value.provider_config_id)
  if (cached) { providerInfo.value = cached; return }
  try {
    providerInfo.value = await getProvider(repo.value.provider_config_id)
  } catch { providerInfo.value = null }
}

async function loadBindings() {
  try {
    const result = await listBindings({ repo_key: repoKey })
    bindings.value = result || []
  } catch { bindings.value = [] }
}

async function openBindingDialog() {
  try {
    await providerStore.fetchProviders()
    availableProviders.value = providerStore.providers
  } catch { availableProviders.value = [] }
  showBindingDialog.value = true
}

async function handleDeleteBinding(id: number) {
  try {
    await ElMessageBox.confirm('确认取消此关联？', '取消关联', { type: 'warning' })
    await deleteBinding(id, true)
    ElMessage.success('关联已取消')
    loadBindings()
  } catch {}
}

async function handleSetPrimaryBinding(id: number) {
  try {
    await setPrimaryBinding(id)
    ElMessage.success('已设为主关联')
    loadBindings()
  } catch (e: any) {
    ElMessage.error('操作失败: ' + (e?.message || ''))
  }
}

async function handleRegisterWebhook(id: number) {
  try {
    await registerBindingWebhook(id)
    ElMessage.success('Webhook 已注册')
    loadBindings()
  } catch (e: any) {
    ElMessage.error('注册失败: ' + (e?.message || ''))
  }
}

async function handleDeleteWebhook(id: number) {
  try {
    await deleteBindingWebhook(id)
    ElMessage.success('Webhook 已删除')
    loadBindings()
  } catch (e: any) {
    ElMessage.error('删除失败: ' + (e?.message || ''))
  }
}

function openEditDialog() {
  if (!repo.value) return
  editForm.value = {
    name: repo.value.name,
    path: repo.value.path,
    remote_url: repo.value.remote_url || '',
  }
  editRemotes.value = []
  editTrackingBranches.value = []
  editDefaultCredentialId.value = repo.value.default_credential_id
  editRemoteCredentials.value = { ...(repo.value.remote_credentials || {}) }
  editUrlError.value = ''
  remoteUrlErrors.value = {}
  remoteUrlModes.value = {}
  const mainProto = detectGitProtocol(repo.value.remote_url || '')
  editUrlMode.value = mainProto === 'http' ? 'https' : 'ssh'
  showEditDialog.value = true

  if (repo.value.path) {
    scanRepo(repo.value.path).then(result => {
      editRemotes.value = (result.remotes || []).map(r => ({
        ...r,
        _testing: false,
      }))
      editTrackingBranches.value = result.branches || []
      editRemotes.value.forEach((r, i) => {
        const p = detectGitProtocol(r.fetch_url || '')
        remoteUrlModes.value[i] = p === 'http' ? 'https' : 'ssh'
      })
      if (!editForm.value.remote_url && editRemotes.value.length > 0) {
        editForm.value.remote_url = editRemotes.value[0]!.fetch_url
        const p = detectGitProtocol(editForm.value.remote_url)
        editUrlMode.value = p === 'http' ? 'https' : 'ssh'
      }
    }).catch(() => { /* ignore scan failure */ })
  }
}

async function handleSaveEdit() {
  if (!editForm.value.name || !editForm.value.path) {
    ElMessage.warning('名称和路径不能为空')
    return
  }
  if (editForm.value.remote_url) {
    const err = validateGitRemoteUrl(editForm.value.remote_url)
    if (err) {
      editUrlError.value = err
      return
    }
  }
  for (let i = 0; i < editRemotes.value.length; i++) {
    const r = editRemotes.value[i]!
    if (r.fetch_url) {
      const err = validateGitRemoteUrl(r.fetch_url)
      if (err) {
        remoteUrlErrors.value[i] = err
        ElMessage.warning(`远程 "${r.name || 'unnamed'}" 的 URL 格式不正确`)
        return
      }
    }
  }
  editSaving.value = true
  try {
    const remotes: GitRemote[] = editRemotes.value
      .filter(r => r.name && r.fetch_url)
      .map(r => ({
        name: r.name,
        fetch_url: r.fetch_url,
        push_url: r.push_url || r.fetch_url,
        is_mirror: r.is_mirror,
      }))

    const rc: Record<string, number> = {}
    for (const [k, v] of Object.entries(editRemoteCredentials.value)) {
      if (v) rc[k] = v
    }

    await updateRepo({
      key: repoKey,
      name: editForm.value.name,
      path: editForm.value.path,
      remote_url: editForm.value.remote_url || undefined,
      remotes,
      default_credential_id: editDefaultCredentialId.value,
      remote_credentials: Object.keys(rc).length > 0 ? rc : undefined,
    })
    ElMessage.success('保存成功')
    showEditDialog.value = false
    repo.value = await getRepoDetail(repoKey)
    if (repo.value?.path) {
      try {
        scanData.value = await scanRepo(repo.value.path)
        remoteNames.value = (scanData.value?.remotes || []).map((r: { name: string }) => r.name)
      } catch { /* ignore */ }
    }
  } finally {
    editSaving.value = false
  }
}

function addEditRemote() {
  editRemotes.value.push({
    name: '',
    fetch_url: '',
    push_url: '',
    is_mirror: false,
    _testing: false,
  })
}

function updateEditRemoteCred(name: string, val: number | undefined) {
  if (val) {
    editRemoteCredentials.value[name] = val
  } else {
    delete editRemoteCredentials.value[name]
  }
}

function validateEditUrl() {
  const url = editForm.value.remote_url
  if (!url) {
    editUrlError.value = ''
    return
  }
  const proto = detectGitProtocol(url)
  if (proto === 'ssh') editUrlMode.value = 'ssh'
  else if (proto === 'http') editUrlMode.value = 'https'
  editUrlError.value = validateGitRemoteUrl(url)
}

function validateRemoteUrl(index: number) {
  const remote = editRemotes.value[index]
  if (!remote) return
  if (!remote.fetch_url) {
    delete remoteUrlErrors.value[index]
    return
  }
  const proto = detectGitProtocol(remote.fetch_url)
  if (proto === 'ssh') remoteUrlModes.value[index] = 'ssh'
  else if (proto === 'http') remoteUrlModes.value[index] = 'https'
  const err = validateGitRemoteUrl(remote.fetch_url)
  if (err) {
    remoteUrlErrors.value[index] = err
  } else {
    delete remoteUrlErrors.value[index]
  }
}

function removeEditRemote(index: number) {
  editRemotes.value.splice(index, 1)
  delete remoteUrlErrors.value[index]
  delete remoteUrlModes.value[index]
}

async function testEditRemote(index: number) {
  const row = editRemotes.value[index]
  if (!row || !row.fetch_url) {
    ElMessage.warning('请输入 Fetch URL')
    return
  }
  row._testing = true
  try {
    const credentialId = editRemoteCredentials.value[row.name]
    if (credentialId) {
      const result = await testCredential(credentialId, row.fetch_url)
      if (result.success) {
        ElMessage.success(`${row.name || 'Remote'} 连接成功`)
      } else {
        ElMessage.error('连接失败: ' + (result.message || '未知错误'))
      }
    } else if (editDefaultCredentialId.value) {
      const result = await testCredential(editDefaultCredentialId.value, row.fetch_url)
      if (result.success) {
        ElMessage.success(`${row.name || 'Remote'} 连接成功`)
      } else {
        ElMessage.error('连接失败: ' + (result.message || '未知错误'))
      }
    } else {
      const result = await testConnection(row.fetch_url)
      if (result.status === 'success') {
        ElMessage.success(`${row.name || 'Remote'} 连接成功`)
      } else {
        ElMessage.error('连接失败: ' + (result.error || '未知错误'))
      }
    }
  } catch (e: any) {
    ElMessage.error('连接测试请求失败: ' + (e?.message || ''))
  } finally {
    row._testing = false
  }
}

async function handleExportCsv(type: string) {
  try {
    const params: Record<string, string> = { type }
    if (type === 'stats') {
      if (statsFilter.value.branch) params.branch = statsFilter.value.branch
      if (statsFilter.value.author) params.author = statsFilter.value.author
      if (statsFilter.value.since) params.since = statsFilter.value.since
      if (statsFilter.value.until) params.until = statsFilter.value.until
    } else {
      if (lineStatsFilter.value.branch) params.branch = lineStatsFilter.value.branch
    }
    const response = await exportStatsCsv(repoKey, params) as unknown as Blob
    const url = window.URL.createObjectURL(response)
    const a = document.createElement('a')
    a.href = url
    a.download = `${repo.value?.name || repoKey}-${type}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  } catch { ElMessage.error('导出失败') }
}

async function openExcludeConfig() {
  try {
    const config = await getLineStatsConfig(repoKey)
    excludeDirsText.value = (config.exclude_dirs || []).join('\n')
    excludePatternsText.value = (config.exclude_patterns || []).join('\n')
  } catch { /* ignore */ }
  showExcludeDialog.value = true
}

async function handleSaveExclude() {
  try {
    await saveLineStatsConfig(repoKey, {
      exclude_dirs: excludeDirsText.value.split('\n').map(s => s.trim()).filter(Boolean),
      exclude_patterns: excludePatternsText.value.split('\n').map(s => s.trim()).filter(Boolean),
    })
    ElMessage.success('排除配置已保存')
    showExcludeDialog.value = false
  } catch { /* handled */ }
}

function handleNavSelect(key: string) {
  const item = sidebarItems.find(i => i.key === key)
  if (item && (item as any).route) {
    router.push(`/local-repos/${repoKey}/${key}`)
  } else {
    activeTab.value = key
  }
}
</script>

<style scoped>
.page-header-wrap {
  padding: 16px 32px;
  border-bottom: 1px solid var(--border-color);
}

.info-card {
  border-radius: 12px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.info-top-row {
  display: flex;
  gap: 0;
}

.info-left-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-right: 16px;
}

.info-v-divider {
  width: 1px;
  background: var(--border-color);
  align-self: stretch;
}

.info-right-col {
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-left: 16px;
}

.info-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-subtitle {
  font-size: 12px;
  color: var(--text-color-secondary, #94A3B8);
}

.info-row {
  display: flex;
  gap: 20px;
}

.info-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.info-label {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.info-value {
  font-size: 14px;
  color: var(--text-color-primary);
}
.info-value--bold { font-weight: 500; }
.info-value--accent { color: var(--accent-primary); font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; display: flex; align-items: center; gap: 8px; }
.info-value.mono { font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace; font-size: 13px; }

.copy-btn-sm {
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
  background: transparent;
  font-size: 11px;
  color: var(--text-color-secondary);
  cursor: pointer;
  transition: all 0.2s;
}
.copy-btn-sm:hover { border-color: var(--accent-primary); color: var(--accent-primary); }

.info-divider {
  height: 1px;
  background: var(--border-color);
  margin: 16px 0;
}

.scan-section { display: flex; flex-direction: column; }

.scan-remote-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.scan-remote-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  background: #F8F9FC;
  font-size: 13px;
}

.remote-name {
  font-weight: 600;
  color: var(--accent-primary);
  min-width: 60px;
}

.remote-url {
  font-family: 'SF Mono', 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-color-secondary);
  flex: 1;
}

.tracking-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 10px;
}

.platform-card-mini {
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.platform-card-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.platform-icon-sm {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
}

.platform-card-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.platform-card-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-primary);
}

.platform-card-repo {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.platform-connected-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
  padding: 3px 8px;
  border-radius: 9999px;
  background: #ECFDF5;
  color: #059669;
  font-size: 10px;
}

.dot-green {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #10B981;
}

.platform-card-meta {
  display: flex;
  gap: 16px;
}

.platform-card-meta .meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.platform-card-meta .meta-label {
  font-size: 11px;
  color: var(--text-color-secondary, #94A3B8);
}

.platform-card-meta .meta-value {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color-primary);
}

.no-platform-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px;
  border: 1px dashed var(--border-color);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-color-secondary, #94A3B8);
}

.link-platform-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 6px;
  background: var(--accent-primary);
  color: #fff;
  border: none;
  font-size: 12px;
  cursor: pointer;
  transition: opacity 0.2s;
}
.link-platform-btn:hover { opacity: 0.9; }

.layout-container {
  display: flex;
  gap: 20px;
  padding: 20px;
}

.left-nav {
  width: 220px;
  flex-shrink: 0;
}

.sidebar-card {
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: var(--bg-color-page);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  padding: 8px;
  height: calc(100vh - 180px);
  overflow-y: auto;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: var(--border-radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  color: var(--text-color-primary);
  font-size: var(--font-size-sm);
}

.sidebar-item:hover {
  background: var(--border-color-extra-light);
}

.sidebar-item.active {
  background: var(--accent-bg);
  color: var(--accent-primary);
}

.sidebar-item.active .el-icon {
  color: var(--accent-primary);
}

.sidebar-item .el-icon {
  color: var(--text-color-secondary);
  font-size: 16px;
}

.content-area {
  flex: 1;
  min-height: calc(100vh - 180px);
}

.spec-full-area {
  height: calc(100vh - 180px);
}

.spec-full-area :deep(.spec-editor-container) {
  height: 100%;
}

.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: var(--spacing-md);
}

.mono-text {
  font-family: monospace;
}

.mt-3 {
  margin-top: 12px;
}

.mt-4 {
  margin-top: var(--spacing-md);
}

.mb-4 {
  margin-bottom: var(--spacing-md);
}

.ml-1 {
  margin-left: var(--spacing-xs);
}

.mt-1 {
  margin-top: var(--spacing-xs);
}

.version-timeline {
  padding: var(--spacing-md) 0;
}

.version-card {
  max-width: 500px;
}

.version-header {
  margin-bottom: var(--spacing-sm);
}

.version-info {
  font-size: var(--font-size-md);
  line-height: 1.8;
  color: var(--text-color-regular);
}

.edit-remote-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
}

.edit-remote-cred {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.edit-remote-cred .cred-label {
  font-size: var(--font-size-sm);
  color: var(--text-color-regular);
  flex-shrink: 0;
}

.remotes-section {
  width: 100%;
}

.remotes-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: var(--font-size-md);
  color: var(--text-color-regular);
}

.tracking-branches {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}

.field-error {
  color: var(--danger-color);
  font-size: var(--font-size-xs);
  margin-top: var(--spacing-xs);
}

.is-error-input :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--danger-color) inset;
}

.url-input-group {
  display: flex;
  gap: var(--spacing-sm);
  width: 100%;
}

.url-mode-switch {
  flex-shrink: 0;
}

.url-mode-switch-sm {
  flex-shrink: 0;
}

.url-input-group .el-input {
  flex: 1;
}

@media (max-width: 1024px) {
  .left-nav {
    width: 200px;
  }
}

@media (max-width: 768px) {
  .layout-container {
    flex-direction: column;
    padding: var(--spacing-md);
  }

  .left-nav {
    width: 100%;
  }

  .sidebar-card {
    height: auto;
    max-height: 300px;
    flex-direction: row;
    flex-wrap: wrap;
  }

  .content-area {
    min-height: auto;
  }
}

</style>
