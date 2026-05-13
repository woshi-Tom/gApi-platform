#!/usr/bin/env python3
"""
gAPI Platform API Test Script
=============================
功能测试 + 压力测试 一体化脚本。

用法:
  # 运行功能测试（健康检查、渠道管理、兑换码等）
  python test_api.py

  # 运行压力测试（需要先创建 API Token）
  python test_api.py --pressure --token sk-xxx
  python test_api.py --pressure --token sk-xxx --base-url http://localhost:8080
"""
import requests
import json
import sys

BASE_URL = "http://localhost:8080/api/v1"


class APITester:
    """
    功能测试类。
    覆盖管理后台的核心 API：健康检查、登录、渠道管理、兑换码等。
    """

    def __init__(self):
        self.session = requests.Session()  # 复用 TCP 连接，提高测试效率
        self.token = None                  # 登录后保存 JWT token
        self.user_id = None
        self.failures = []                 # 记录失败的测试

    def request(self, method, path, **kwargs):
        """
        统一的 HTTP 请求方法。
        - 自动拼接 BASE_URL + path
        - 如果有 token 自动带上 Authorization header
        - 连接失败时直接退出（fail-fast）
        """
        url = f"{BASE_URL}{path}"
        if self.token:
            kwargs.setdefault('headers', {})['Authorization'] = f'Bearer {self.token}'
        kwargs.setdefault('headers', {}).setdefault('Content-Type', 'application/json')

        try:
            response = self.session.request(method, url, **kwargs)
            return response
        except requests.exceptions.ConnectionError:
            print(f"❌ Failed to connect to {url}")
            sys.exit(1)

    def assert_eq(self, name, actual, expected, msg=""):
        """断言 actual == expected，失败则记录并返回 False"""
        if actual != expected:
            detail = f" (expected {expected}, got {actual}) {msg}"
            self.failures.append(name + detail)
            print(f"  ❌ {name}{detail}")
            return False
        return True

    def assert_true(self, name, condition, msg=""):
        """断言 condition 为真"""
        if not condition:
            self.failures.append(name + f" {msg}")
            print(f"  ❌ {name} {msg}")
            return False
        return True

    def test_health(self):
        """测试服务健康检查端点（无需认证）"""
        print("\n=== Test 1: Health Check ===")
        try:
            resp = requests.get("http://localhost:8080/health")
            self.assert_eq("health status", resp.status_code, 200)
            data = resp.json()
            self.assert_true("health has status", "status" in data or "code" in data)
            print(f"  Health: {resp.text[:200]}")
        except Exception as e:
            self.failures.append(f"health check connection failed: {e}")
            print(f"  ❌ Connection failed: {e}")

    def test_register(self):
        """测试系统初始化状态探测"""
        print("\n=== Test 2: Init Status ===")
        resp = self.request('GET', '/init/status')
        self.assert_eq("init status", resp.status_code, 200)
        print(f"  Init status: {resp.text[:200]}")

    def test_login(self):
        """
        测试管理员登录。
        - 验证未认证返回 401
        - 验证错误路径返回 404
        - 用 admin/admin123 登录获取 JWT token
        """
        print("\n=== Test 3: User Login ===")

        # 验证未认证请求被拒绝
        resp = self.request('GET', '/user/info')
        self.assert_eq("unauth returns 401", resp.status_code, 401)

        # 验证错误路径返回 404
        resp2 = self.request('POST', '/login', json={"username": "admin", "password": "admin123"})
        self.assert_eq("wrong path returns 404", resp2.status_code, 404)

        # 管理员登录
        admin_data = {"username": "admin", "password": "admin123"}
        resp = self.request('POST', '/admin/login', json=admin_data)
        self.assert_eq("admin login", resp.status_code, 200)

        if resp.status_code == 200:
            result = resp.json()
            has_token = 'data' in result and 'token' in result['data']
            self.assert_true("login returns token", has_token)
            if has_token:
                self.token = result['data']['token']
                print(f"  Token: {self.token[:20]}...")

    def test_channels(self):
        """测试渠道管理 API — 列出渠道并验证响应格式"""
        print("\n=== Test 4: Channel Management ===")

        if not self.token:
            self.failures.append("channels: no auth token")
            print("  ❌ No auth token")
            return

        resp = self.request('GET', '/admin/channels')
        self.assert_eq("channels list", resp.status_code, 200)

        if resp.status_code == 200:
            result = resp.json()
            has_list = 'data' in result and 'list' in result['data']
            self.assert_true("channels has list", has_list)
            if has_list:
                channels = result['data']['list']
                print(f"  Found {len(channels)} channels")
                for ch in channels[:3]:
                    print(f"  - {ch.get('name')} ({ch.get('type')})")

    def test_register_settings(self):
        """测试注册设置 API"""
        print("\n=== Test 5: Register Settings ===")

        resp = self.request('GET', '/admin/settings/register')
        self.assert_eq("register settings", resp.status_code, 200)

        if resp.status_code == 200:
            result = resp.json()
            has_data = 'data' in result
            self.assert_true("settings has data", has_data)
            if has_data:
                data = result['data']
                self.assert_true("settings has enabled", 'enabled' in data or 'enable_register' in data or len(data) > 0)
                print(f"  Settings: {json.dumps(data, indent=2)}")

    def test_redemption_codes(self):
        """测试兑换码创建 — 批量生成兑换码"""
        print("\n=== Test 6: Create Redemption Code ===")

        if not self.token:
            self.failures.append("redemption: no auth token")
            print("  ❌ No auth token")
            return

        data = {
            "code_type": "quota",
            "prefix": "TEST",
            "quota_amount": 10000,
            "count": 5,
            "note": "Test batch"
        }

        resp = self.request('POST', '/admin/redemption/codes', json=data)
        self.assert_eq("create redemption codes", resp.status_code, 200)

        if resp.status_code == 200:
            result = resp.json()
            has_codes = 'data' in result and result['data']
            self.assert_true("codes created", has_codes)
            if has_codes:
                print(f"  Created: {json.dumps(result['data'], indent=2)}")

    def run_all(self):
        """按顺序执行所有功能测试，返回 0=全部通过 1=有失败"""
        print("=" * 50)
        print("gAPI Platform API Test Suite")
        print("=" * 50)

        self.test_health()
        self.test_register()
        self.test_login()
        self.test_channels()
        self.test_register_settings()
        self.test_redemption_codes()

        print("\n" + "=" * 50)
        if self.failures:
            print(f"FAILED: {len(self.failures)} assertion(s) failed")
            for f in self.failures:
                print(f"  - {f}")
            print("=" * 50)
            return 1
        else:
            print("ALL PASSED")
            print("=" * 50)
            return 0


# ============================================================
# Phase 4 (T-12): Pressure Test（压力测试）
# 测试目的：验证平台在高并发下的吞吐量（TPS）和延迟分布（P50/P95/P99）
# 用法：python test_api.py --pressure --token <api_key>
# 注意：需要先在用户中心创建 API Token，否则请求会被 401 拒绝
# ============================================================
import concurrent.futures  # 线程池，实现并发请求
import statistics          # 计算延迟统计值
import argparse            # 命令行参数解析
import time                # 计时
import threading           # 线程锁，保护共享计数器


class PressureTest:
    """
    压力测试类。

    核心指标:
      - TPS（Transactions Per Second）: 吞吐量
      - P50 / P95 / P99: 百分位延迟，反映大多数用户的体验
      - Error Rate: 错误率，反映系统稳定性

    线程模型:
      - 使用 ThreadPoolExecutor 控制并发数
      - 每个 worker 线程执行 N 次请求
      - 用 threading.Lock 保护共享的 latencies/success/errors 计数器
    """

    def __init__(self, base_url="http://localhost:8080", token=None):
        self.base_url = base_url
        self.token = token                          # API Token（Bearer Auth）
        self.results_lock = threading.Lock()        # 保护共享数据
        self.latencies = []                         # 所有请求的延迟（毫秒）
        self.errors = 0                             # 失败请求数
        self.success = 0                            # 成功请求数

    def _do_chat(self, model="gpt-3.5-turbo", stream=False):
        """
        发送一次 chat/completions 请求。
        Returns (latency_ms, status_code)。
        - status_code=0 表示网络错误（如连接超时、DNS 解析失败）
        """
        url = f"{self.base_url}/api/v1/chat/completions"
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        body = {
            "model": model,
            "messages": [{"role": "user", "content": "Say hello in one word."}],
            "max_tokens": 10,
            "stream": stream,
        }
        start = time.time()
        try:
            resp = requests.post(url, json=body, headers=headers, timeout=30)
            latency = (time.time() - start) * 1000
            return latency, resp.status_code
        except Exception as e:
            latency = (time.time() - start) * 1000
            return latency, 0  # 0 = connection error

    def _do_completions(self, model="gpt-3.5-turbo"):
        """
        发送一次 /v1/completions 请求。
        这是 OpenAI 兼容的文本补全端点（与 chat/completions 不同）。
        """
        url = f"{self.base_url}/api/v1/completions"
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        body = {"model": model, "prompt": "Hello", "max_tokens": 10}
        start = time.time()
        try:
            resp = requests.post(url, json=body, headers=headers, timeout=30)
            latency = (time.time() - start) * 1000
            return latency, resp.status_code
        except Exception:
            latency = (time.time() - start) * 1000
            return latency, 0

    def _do_models(self):
        """
        发送一次 /v1/models 请求（获取可用模型列表）。
        这是最轻量的端点，不需要上游渠道，适合测试基础吞吐能力。
        """
        url = f"{self.base_url}/api/v1/models"
        headers = {}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        start = time.time()
        try:
            resp = requests.get(url, headers=headers, timeout=30)
            latency = (time.time() - start) * 1000
            return latency, resp.status_code
        except Exception:
            latency = (time.time() - start) * 1000
            return latency, 0

    def _worker(self, fn, n):
        """
        单个 worker 线程的执行逻辑。
        连续执行 n 次 fn() 调用，并将结果记录到共享计数器。

        注意: 访问共享数据（self.latencies / self.success / self.errors）时必须加锁，
        否则多个线程同时写入会导致数据竞态（data race）。
        """
        for _ in range(n):
            lat, code = fn()
            with self.results_lock:
                self.latencies.append(lat)
                if 200 <= code < 300:
                    self.success += 1
                else:
                    self.errors += 1

    def run_scenario(self, name, fn, concurrency, total_requests):
        """
        运行一个压力测试场景。

        Parameters
        ----------
        name : str
            场景名称（用于输出）
        fn : callable
            要压测的函数（_do_chat / _do_models / _do_completions）
        concurrency : int
            并发数（线程池大小）
        total_requests : int
            总请求数

        输出指标:
          Duration  — 总耗时（秒）
          TPS       — 每秒处理请求数
          P50/P95/P99 — 百分位延迟（毫秒）
          Error Rate — 错误百分比
        """
        print(f"\n--- {name} ---")
        print(f"  Concurrency={concurrency}, Total={total_requests}")

        # 重置计数器
        self.latencies = []
        self.errors = 0
        self.success = 0

        # 计算每个 worker 分配的任务数（尽量平均）
        per_worker = total_requests // concurrency
        remainder = total_requests % concurrency

        t0 = time.time()

        # 用线程池并发执行
        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as executor:
            futures = []
            for i in range(concurrency):
                n = per_worker + (1 if i < remainder else 0)
                futures.append(executor.submit(self._worker, fn, n))
            concurrent.futures.wait(futures)  # 等待所有线程完成

        elapsed = time.time() - t0
        tps = total_requests / elapsed if elapsed > 0 else 0

        # 计算延迟百分位
        if self.latencies:
            sorted_lat = sorted(self.latencies)
            p50 = sorted_lat[int(len(sorted_lat) * 0.50)]
            p95 = sorted_lat[int(len(sorted_lat) * 0.95)]
            p99 = sorted_lat[int(len(sorted_lat) * 0.99)]
            avg_lat = statistics.mean(self.latencies)
            min_lat = min(self.latencies)
            max_lat = max(self.latencies)
        else:
            p50 = p95 = p99 = avg_lat = min_lat = max_lat = 0

        error_rate = (self.errors / total_requests * 100) if total_requests > 0 else 0

        print(f"  Duration:    {elapsed:.1f}s")
        print(f"  TPS:         {tps:.1f}")
        print(f"  Success:     {self.success} / {total_requests}")
        print(f"  Errors:      {self.errors}")
        print(f"  Error Rate:  {error_rate:.1f}%")
        print(f"  P50:         {p50:.0f}ms")
        print(f"  P95:         {p95:.0f}ms")
        print(f"  P99:         {p99:.0f}ms")
        print(f"  Avg:         {avg_lat:.0f}ms")
        print(f"  Min:         {min_lat:.0f}ms")
        print(f"  Max:         {max_lat:.0f}ms")

        return {
            "scenario": name,
            "duration_s": elapsed,
            "tps": tps,
            "success": self.success,
            "errors": self.errors,
            "error_rate_pct": error_rate,
            "p50_ms": p50,
            "p95_ms": p95,
            "p99_ms": p99,
            "avg_ms": avg_lat,
            "min_ms": min_lat,
            "max_ms": max_lat,
        }

    def run_all(self, token=None):
        """
        运行所有预设的压力测试场景。

        场景列表:
          1. chat/completions — 50 并发，500 请求（高并发 Chat）
          2. /v1/models       — 100 并发，500 请求（纯平台吞吐）
          3. /v1/completions  — 30 并发，300 请求（文本补全）
          4. chat (低并发)    — 10 并发，50 请求（保底测试）

        PASS 条件: 错误率 < 1% 且 P95 < 3000ms
        """
        self.token = token

        print("=" * 60)
        print("Phase 4 (T-12): API Pressure Test Suite")
        print("=" * 60)

        results = []

        # 场景 1: 非流式 Chat API（高并发）
        results.append(
            self.run_scenario(
                "Non-stream Chat API (50 concurrency)",
                lambda: self._do_chat(stream=False),
                concurrency=50,
                total_requests=500,
            )
        )

        # 场景 2: 模型列表 API（最适合测纯平台吞吐）
        results.append(
            self.run_scenario(
                "/v1/models list (100 concurrency)",
                self._do_models,
                concurrency=100,
                total_requests=500,
            )
        )

        # 场景 3: 文本补全 API
        results.append(
            self.run_scenario(
                "/v1/completions (30 concurrency)",
                self._do_completions,
                concurrency=30,
                total_requests=300,
            )
        )

        # 场景 4: 低并发 Chat（给没有上游渠道的环境保底测试）
        results.append(
            self.run_scenario(
                "Chat API low-concurrency (10 concurrent)",
                lambda: self._do_chat(stream=False),
                concurrency=10,
                total_requests=50,
            )
        )

        # 汇总输出
        print("\n" + "=" * 60)
        print("Pressure Test Summary")
        print("=" * 60)
        all_pass = True
        for r in results:
            passed = r["error_rate_pct"] < 1 and r["p95_ms"] < 3000
            status = "PASS" if passed else "FAIL"
            if not passed:
                all_pass = False
            print(f"  [{status}] {r['scenario']}: "
                  f"TPS={r['tps']:.1f}, P95={r['p95_ms']:.0f}ms, "
                  f"Err={r['error_rate_pct']:.1f}%")

        print(f"\n{'ALL PASSED' if all_pass else 'SOME SCENARIOS FAILED'}")
        return 0 if all_pass else 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="gAPI Platform API Tester")
    parser.add_argument("--pressure", action="store_true", help="运行压力测试（代替功能测试）")
    parser.add_argument("--token", type=str, help="API Token（压力测试需要传入）")
    parser.add_argument("--base-url", type=str, default="http://localhost:8080",
                        help="API 基础地址（默认: http://localhost:8080）")
    args = parser.parse_args()

    if args.pressure:
        # 压力测试模式
        pt = PressureTest(base_url=args.base_url)
        rc = pt.run_all(token=args.token)
    else:
        # 功能测试模式（默认）
        tester = APITester()
        rc = tester.run_all()

    sys.exit(rc)