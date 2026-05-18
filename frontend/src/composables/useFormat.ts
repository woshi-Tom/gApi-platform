/**
 * 统一格式化工具 composable
 * 集中管理日期、配额、状态等格式化逻辑
 */

/**
 * 格式化配额数字（如 Token 数量）
 * - >= 1M → "1.2M"
 * - >= 1K → "3.4K"
 * - 其他 → 本地化数字
 */
export function formatQuota(n: number | undefined | null): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toLocaleString()
}

/**
 * 格式化为本地日期时间字符串
 */
export function formatDate(dateStr: string | undefined | null): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/**
 * 格式化为相对时间（如 "3分钟前"）
 */
export function formatRelativeTime(timeStr: string | Date | undefined | null): string {
  if (!timeStr) return '-'
  const date = typeof timeStr === 'string' ? new Date(timeStr) : timeStr
  if (isNaN(date.getTime())) return '-'

  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return '刚刚'
  if (diffMins < 60) return `${diffMins}分钟前`
  if (diffHours < 24) return `${diffHours}小时前`
  if (diffDays < 7) return `${diffDays}天前`

  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

/**
 * 订单状态 → 显示文本
 */
export function getOrderStatusName(status: string): string {
  const map: Record<string, string> = {
    completed: '已完成',
    paid: '已支付',
    pending: '待支付',
    cancelled: '已取消',
    refunded: '已退款',
    expired: '已过期',
  }
  return map[status] || status
}

/**
 * 订单状态 → Element Plus Tag 类型
 */
export function getOrderStatusType(status: string): string {
  const map: Record<string, string> = {
    completed: 'success',
    paid: 'success',
    pending: 'warning',
    cancelled: 'info',
    refunded: 'danger',
    expired: 'danger',
  }
  return map[status] || 'info'
}

/**
 * 订单类型 → 显示文本
 */
export function getOrderTypeName(type: string): string {
  const map: Record<string, string> = {
    vip: 'VIP',
    recharge: '充值',
    package: '套餐',
  }
  return map[type] || type
}

/**
 * 用户等级 → 显示文本
 */
export function getLevelName(level: string | undefined | null): string {
  if (!level) return '免费'
  const map: Record<string, string> = {
    free: '免费',
    vip: 'VIP',
    vip_bronze: 'VIP青铜',
    vip_silver: 'VIP白银',
    vip_gold: 'VIP黄金',
    enterprise: '企业版',
  }
  return map[level] || level
}

/**
 * 用户等级 → Element Plus Tag 类型
 */
export function getLevelType(level: string | undefined | null): string {
  if (!level || level === 'free') return 'info'
  const map: Record<string, string> = {
    vip: 'warning',
    vip_bronze: 'warning',
    vip_silver: '',
    vip_gold: 'warning',
    enterprise: 'danger',
  }
  return map[level] || 'info'
}

/**
 * 计算剩余天数
 */
export function getDaysRemaining(dateStr: string | undefined | null): number {
  if (!dateStr) return 0
  const expiry = new Date(dateStr)
  const now = new Date()
  const diff = expiry.getTime() - now.getTime()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}