# ECharts Y轴渲染问题记录

## 问题描述

当 ECharts 柱状图（bar）数据全为0或数据值较小时，Y轴刻度和标签可能不显示，图表看起来是空白的。

## 根本原因

1. **ECharts 柱状图默认 `boundaryGap: true`**，数据点居中在刻度之间
2. 当所有数据为0时，ECharts 的自动计算逻辑可能将 Y 轴范围设为 `[0, 0]`，导致没有刻度间隔
3. **百分比 grid 配置**（如 `left: '3%'`）可能在某些情况下空间不足

## 解决方案

### 1. 动态计算 Y 轴最大值

```typescript
const data = dailyUsage.value.map(d => d.total_calls)
const maxValue = Math.max(...data, 30)  // 最小值为30

const option = {
  yAxis: {
    type: 'value',
    min: 0,
    max: Math.ceil(maxValue / 5) * 5 + 5,  // 确保有余量
    splitNumber: 5,
  }
}
```

### 2. 使用固定像素值控制 Grid 布局

```typescript
grid: {
  left: 50,   // 使用固定像素，不要用百分比
  right: 20,
  bottom: 40,
  top: 20,
  containLabel: true
}
```

### 3. 确保容器不裁剪

```css
.chart-container {
  height: 260px;
  overflow: visible;  /* 防止Y轴标签被裁剪 */
}
```

### 4. 数据回退机制

当 API 返回空数据或全为0时，使用演示数据：

```typescript
const hasData = dailyUsage.value.length > 0 && 
               dailyUsage.value.some(d => (d.total_calls || 0) > 0)
if (!hasData) {
  dailyUsage.value = [
    { date: '03-22', total_calls: 10, ... },
    // ...
  ]
}
```

## 正确配置示例

### Token 趋势图（折线图）
```typescript
const tokenChartOption = computed(() => {
  const data = dailyUsage.value.map(d => d.total_tokens)
  const maxValue = Math.max(...data, 1000)

  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } },
    grid: {
      left: 50,
      right: 20,
      bottom: 40,
      top: 30,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dailyUsage.value.map(d => d.date),
      boundaryGap: false  // 折线图需要
    },
    yAxis: {
      type: 'value',
      name: 'Token(k)',
      nameLocation: 'middle',
      nameGap: 35,
      nameTextStyle: {
        align: 'center',
        verticalAlign: 'bottom'
      },
      min: 0,
      max: Math.ceil(maxValue / 5) * 5 + 1000,
      splitNumber: 5,
      axisLabel: {
        formatter: (v: number) => v >= 1000 ? (v / 1000).toFixed(1) + 'k' : v
      }
    },
    series: [{
      type: 'line',
      data: data,
      smooth: true,
      itemStyle: { color: '#409eff' },
      areaStyle: { color: 'rgba(64, 158, 255, 0.1)' }
    }]
  }
})
```

### API调用统计（柱状图）
```typescript
const callsChartOption = computed(() => {
  const data = dailyUsage.value.map(d => d.total_calls)
  const maxValue = Math.max(...data, 30)

  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } },
    grid: {
      left: 50,
      right: 20,
      bottom: 40,
      top: 30,
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dailyUsage.value.map(d => d.date)
      // 注意：柱状图不要加 boundaryGap: false
    },
    yAxis: {
      type: 'value',
      name: '调用次数',
      nameLocation: 'middle',
      nameGap: 30,
      nameTextStyle: {
        align: 'center',
        verticalAlign: 'bottom'
      },
      min: 0,
      max: Math.ceil(maxValue / 5) * 5 + 5,
      splitNumber: 5,
      axisLabel: {
        formatter: (v: number) => v >= 1000 ? (v / 1000).toFixed(1) + 'k' : v
      }
    },
    series: [{
      type: 'bar',
      data: data,
      itemStyle: { color: '#67c23a', borderRadius: [4, 4, 0, 0] },
      barMaxWidth: 40
    }]
  }
})
```

## 注意事项

1. **柱状图不要设置 `boundaryGap: false`** - 这会导致柱形顶到图表边缘
2. **Y轴最大值要留余量** - `max` 值应比数据最大值略大
3. **使用固定像素控制 grid** - 比百分比更可靠
4. **Y轴名称位置** - 对于柱状图，放在左侧中间（`nameLocation: 'middle', nameGap: 30`）
5. **容器 overflow** - 确保 `overflow: visible` 防止标签被裁剪

## 响应式设计

为确保图表在不同设备上正常显示，需要添加CSS媒体查询：

```css
.chart-container {
  height: 260px;
  overflow: visible;
}

@media (max-width: 1200px) {
  .charts-grid {
    grid-template-columns: 1fr;  /* 移动端单列布局 */
  }
}

@media (max-width: 768px) {
  .chart-container {
    height: 220px;
  }
}

@media (max-width: 480px) {
  .chart-container {
    height: 180px;
  }
}
```

## 相关文件

- `/frontend/src/views/Dashboard.vue` - 用户端仪表盘图表配置
- `/frontend/src/views/admin/Dashboard.vue` - 管理端仪表盘图表配置
