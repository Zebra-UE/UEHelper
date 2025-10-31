import axios from 'axios'

// 定义节点类型
export interface TreeNode {
  id: string | number
  label: string
  children?: TreeNode[]
  isLeaf?: boolean
  hasChildren?: boolean
}

// 创建 axios 实例
const api = axios.create({
  baseURL: 'http://your-api-base-url', // 替换为实际的API地址
  timeout: 5000
})

// 获取根节点数据
export async function fetchRootNodes(): Promise<TreeNode[]> {
  try {
    const response = await api.get('/tree/root')
    return response.data
  } catch (error) {
    console.error('获取根节点失败:', error)
    return []
  }
}

// 获取子节点数据
export async function fetchChildNodes(nodeId: string | number): Promise<TreeNode[]> {
  try {
    const response = await api.get(`/tree/children/${nodeId}`)
    return response.data
  } catch (error) {
    console.error('获取子节点失败:', error)
    return []
  }
}