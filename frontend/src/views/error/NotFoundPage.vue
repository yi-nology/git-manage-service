<template>
  <div class="not-found-page">
    <div class="not-found-card">
      <h1 class="not-found-code">404</h1>
      <EmptyState title="页面未找到" description="您访问的页面不存在或已被移除">
        <template #action>
          <el-button type="primary" @click="router.push('/')">
            <el-icon><House /></el-icon>
            返回首页
          </el-button>
        </template>
      </EmptyState>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { House } from '@element-plus/icons-vue'
import EmptyState from '@/components/common/EmptyState.vue'

const router = useRouter()
</script>

<style scoped>
.not-found-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - var(--header-height));
  padding: var(--spacing-xl);
}

.not-found-card {
  text-align: center;
  background: var(--bg-color-page);
  border-radius: var(--border-radius-lg);
  padding: var(--spacing-xl);
  box-shadow: var(--box-shadow-lg);
  border: 1px solid var(--border-color);
  max-width: 500px;
  width: 100%;
}

.not-found-code {
  font-size: 48px;
  font-weight: 700;
  color: var(--text-color-primary);
  margin: 0 0 var(--spacing-md) 0;
}

@media (max-width: 768px) {
  .not-found-card {
    padding: var(--spacing-lg);
  }

  .not-found-code {
    font-size: 36px;
  }
}
</style>
