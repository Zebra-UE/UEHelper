<template>
  <el-container class="main-container">
    <!-- 左侧树形结构 -->
    <el-aside width="480px" class="aside">
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="treeProps"
        :load="loadNode"
        lazy
        @node-click="handleNodeClick"
        v-loading="loading"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node">
            <span>{{ node.label }}</span>
            <span v-if="loading && currentLoadingNode?.id === data.id" class="loading-indicator">
              <el-icon class="is-loading"><Loading /></el-icon>
            </span>
          </span>
        </template>
      </el-tree>
    </el-aside>
    
    <!-- 右侧内容区 -->
    <el-container class="right-container">
      <el-header height="50%" class="right-top">
        <el-card class="full-height">
          <template #header>
            <div class="card-header">
              <span>节点详情</span>
            </div>
          </template>
          <div class="section-content">
            <pre v-if="selectedNode">{{ JSON.stringify(selectedNode, null, 2) }}</pre>
            <el-empty v-else description="请选择一个节点" />
          </div>
        </el-card>
      </el-header>
      
      <el-main height="50%" class="right-bottom">
        <el-card class="full-height">
          <template #header>
            <div class="card-header">
              <span>操作记录</span>
              <el-button type="primary" size="small" @click="clearLogs">清空</el-button>
            </div>
          </template>
          <div class="section-content logs">
            <div v-for="(log, index) in logs" :key="index" class="log-item">
              {{ log }}
            </div>
          </div>
        </el-card>
      </el-main>
    </el-container>
  </el-container>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { fetchRootNodes, fetchChildNodes, type TreeNode } from './api/tree'
import type Node from 'element-plus/es/components/tree/src/model/node'
import type { LoadFunction } from 'element-plus/es/components/tree/src/tree.type'

// 树的引用
const treeRef = ref()

// 树形结构数据
const treeData = ref<TreeNode[]>([])
const loading = ref(false)
const currentLoadingNode = ref<TreeNode | null>(null)
const selectedNode = ref<TreeNode | null>(null)
const logs = ref<string[]>([])

// 树的配置
const treeProps = {
  children: 'children',
  label: 'label',
  isLeaf: 'isLeaf'
}

// 加载节点数据
const loadNode: LoadFunction<TreeNode, Node> = async (node, resolve) => {
  if (node.level === 0) {
    // 加载根节点
    loading.value = true
    try {
      const rootNodes = await fetchRootNodes()
      resolve(rootNodes)
      addLog('加载根节点成功')
    } catch (error) {
      addLog('加载根节点失败: ' + (error as Error).message)
      resolve([])
    } finally {
      loading.value = false
    }
  } else {
    // 加载子节点
    currentLoadingNode.value = node.data
    loading.value = true
    try {
      const children = await fetchChildNodes(node.data.id)
      resolve(children)
      addLog(`加载节点 "${node.data.label}" 的子节点成功`)
    } catch (error) {
      addLog(`加载节点 "${node.data.label}" 的子节点失败: ` + (error as Error).message)
      resolve([])
    } finally {
      loading.value = false
      currentLoadingNode.value = null
    }
  }
}

// 树节点点击事件
const handleNodeClick = (data: TreeNode) => {
  selectedNode.value = data
  addLog(`选中节点: ${data.label}`)
}

// 添加日志
const addLog = (message: string) => {
  const time = new Date().toLocaleTimeString()
  logs.value.unshift(`[${time}] ${message}`)
}

// 清空日志
const clearLogs = () => {
  logs.value = []
}

// 页面加载时自动展开第一层
onMounted(() => {
  // 树加载完成后自动触发根节点加载
  if (treeRef.value) {
    treeRef.value.store.setData([])
  }
})
</script>

<style scoped>
.main-container {
  height: 100vh;
  width: 100%;
}

.aside {
  background-color: #f5f7fa;
  border-right: solid 1px #e6e6e6;
  padding: 20px 0;
}

.right-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.right-top {
  height: 50% !important;
  padding: 20px;
  overflow: hidden;
}

.right-bottom {
  height: 50% !important;
  padding: 20px;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.full-height {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.section-content {
  flex: 1;
  overflow-y: auto;
  padding: 10px 0;
}

:deep(.el-tree) {
  background-color: transparent;
}

:deep(.el-card) {
  height: 100%;
}

:deep(.el-card__body) {
  height: calc(100% - 60px);
  padding: 10px;
  overflow: hidden;
}

.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  padding-right: 8px;
}

.loading-indicator {
  color: var(--el-color-primary);
  margin-left: 8px;
}

.logs {
  font-family: monospace;
  font-size: 12px;
}

.log-item {
  padding: 4px 0;
  border-bottom: 1px dashed #eee;
}

.log-item:last-child {
  border-bottom: none;
}
</style>