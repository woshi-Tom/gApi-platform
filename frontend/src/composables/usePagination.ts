/**
 * 统一分页逻辑 composable
 * 集中管理分页状态和请求逻辑
 */
import { ref, reactive } from 'vue'

export interface PaginationParams {
  page: number
  pageSize: number
  total: number
}

export interface UsePaginationOptions {
  /** 初始每页条数 */
  defaultPageSize?: number
  /** 是否立即加载 */
  immediate?: boolean
}

/**
 * 服务端分页 composable（用于 API 返回 pagination 的场景）
 * @param fetchFn 数据请求函数，接收 page/pageSize，返回 { list, total }
 */
export function useServerPagination<T>(
  fetchFn: (params: { page: number; pageSize: number }) => Promise<{ list: T[]; total: number }>,
  options: UsePaginationOptions = {}
) {
  const { defaultPageSize = 20 } = options

  const list = ref<T[]>([]) as any
  const loading = ref(false)
  const pagination = reactive<PaginationParams>({
    page: 1,
    pageSize: defaultPageSize,
    total: 0,
  })

  async function load() {
    loading.value = true
    try {
      const result = await fetchFn({
        page: pagination.page,
        pageSize: pagination.pageSize,
      })
      list.value = result.list
      pagination.total = result.total
    } finally {
      loading.value = false
    }
  }

  function handlePageChange() {
    load()
  }

  function handleSizeChange() {
    pagination.page = 1
    load()
  }

  function reset() {
    pagination.page = 1
    load()
  }

  return {
    list,
    loading,
    pagination,
    load,
    handlePageChange,
    handleSizeChange,
    reset,
  }
}

/**
 * 前端分页 composable（用于数据已全部加载，在前端做分页的场景）
 * @param filterFn 可选的筛选函数
 */
export function useClientPagination<T>(
  allData: () => T[],
  options: UsePaginationOptions = {}
) {
  const { defaultPageSize = 10 } = options

  const currentPage = ref(1)
  const pageSize = ref(defaultPageSize)

  const total = () => allData().length

  const paginatedData = (): T[] => {
    const data = allData()
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    return data.slice(start, end)
  }

  function resetPage() {
    currentPage.value = 1
  }

  return {
    currentPage,
    pageSize,
    total,
    paginatedData,
    resetPage,
  }
}