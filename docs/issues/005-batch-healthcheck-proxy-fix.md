# Issue #005: 批量健康检查全异常 — Adapter 代理支持

## 问题

管理后台「批量检测」按钮触发后，所有配置了代理的 NVIDIA 渠道均显示「异常」，
但手动逐个测试（操作 → 测试 → 获取模型列表）成功。

## 根因

手动测试通过 `newProxyClient()` 读取渠道的代理配置，但所有 adapter（共 8 个文件）
使用硬编码的 `&http.Client{Timeout: 120s}`，完全忽略代理设置。

## 修复方案（方案 A 修正版）

在 adapter `Channel` 结构体中增加代理字段，所有 adapter 的 HTTP 请求统一通过
`DoChannelRequest()` 函数，根据 channel 代理配置决定是否使用代理 client。

## 本次变更文件

| 文件 | 变更 |
|------|------|
| `backend/internal/pkg/adapter/types.go` | Channel 增加 ProxyEnabled/ProxyType/ProxyURL 字段；新增 NewHTTPClient() 和 DoChannelRequest() 函数 |
| `backend/internal/pkg/adapter/nvidia.go` | 4 处 a.client.Do → DoChannelRequest |
| `backend/internal/pkg/adapter/openai.go` | 4 处 a.client.Do → DoChannelRequest |
| `backend/internal/pkg/adapter/azure.go` | 3 处 a.client.Do → DoChannelRequest |
| `backend/internal/pkg/adapter/claude.go` | 1 处 a.client.Do → DoChannelRequest |
| `backend/internal/pkg/adapter/deepseek.go` | 3 处 a.client.Do → DoChannelRequest |
| `backend/internal/pkg/adapter/gemini.go` | 2 处 a.client.Do → DoChannelRequest |
| `backend/internal/pkg/adapter/others.go` | 15 处 a.client.Do → DoChannelRequest（含 zhipu/baidu/yi/ollama/localai/groq） |
| `backend/internal/service/health_check.go` | testChannel 增加代理参数；checkChannel 和 CheckChannelManually 传递代理配置 |

## 待验证

1. `go build ./...` 编译通过
2. `go vet ./...` 无警告
3. 配置了代理的 NVIDIA 渠道，批量检测应返回 is_healthy: true
4. 未配置代理的渠道，行为与修复前一致

## 设计要点

- adapter 以单例模式注册（factory.go），不能修改共享 client
- DoChannelRequest() 在有代理配置时创建临时 client，无代理时复用默认 client
- socks5 代理需要 `golang.org/x/net/proxy` 依赖（如有需要后续添加）
