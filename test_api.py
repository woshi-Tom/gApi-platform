#!/usr/bin/env python3
"""
gAPI Platform API Test Script
Tests key API endpoints
"""
import requests
import json
import sys

BASE_URL = "http://localhost:8080/api/v1"

class APITester:
    def __init__(self):
        self.session = requests.Session()
        self.token = None
        self.user_id = None
    
    def request(self, method, path, **kwargs):
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
    
    def test_health(self):
        """Test health endpoint"""
        print("\n=== Test 1: Health Check ===")
        try:
            resp = requests.get("http://localhost:8080/health")
            print(f"✅ Health: {resp.json()}")
            return True
        except Exception as e:
            print(f"❌ Health check failed: {e}")
            return False
    
    def test_register(self):
        """Test user registration"""
        print("\n=== Test 2: User Registration ===")
        # Use a unique email with timestamp
        import time
        email = f"test{int(time.time())}@example.com"
        
        data = {
            "username": f"testuser{int(time.time())}",
            "email": email,
            "password": "Test123456"
        }
        
        # First, verify email code (we'll skip this for now and test login directly)
        # Let's test the init endpoint to see what's available
        resp = self.request('GET', '/init/status')
        print(f"Init status: {resp.status_code} - {resp.text[:200]}")
        
        # Try to register without email verification - should fail or require verification
        # Let's check what the error looks like
        return True
    
    def test_login(self):
        """Test user login"""
        print("\n=== Test 3: User Login ===")
        
        # Try login with a test user (need to create one first or use existing)
        # Let's first check if we can get user info without auth
        resp = self.request('GET', '/user/info')
        print(f"Get info without auth: {resp.status_code}")
        
        if resp.status_code == 401:
            print("✅ Auth required (expected)")
        
        # Try admin login - check both paths
        admin_data = {
            "username": "admin",
            "password": "admin123"
        }
        
        # Try /admin/login
        resp = self.request('POST', '/admin/login', json=admin_data)
        print(f"Admin login (/admin/login): {resp.status_code}")
        
        # Try /login (admin group)
        resp2 = self.request('POST', '/login', json=admin_data)
        print(f"Admin login (/login): {resp2.status_code} - {resp2.text[:200] if resp2.text else 'empty'}")
        
        if resp.status_code == 200:
            try:
                result = resp.json()
                if 'data' in result and 'token' in result['data']:
                    self.token = result['data']['token']
                    print(f"✅ Admin login successful, token: {self.token[:20]}...")
                else:
                    print(f"Response: {result}")
            except:
                print(f"Response: {resp.text}")
        else:
            print(f"❌ Admin login failed: {resp.text}")
        
        return self.token is not None
    
    def test_channels(self):
        """Test channel management (admin)"""
        print("\n=== Test 4: Channel Management ===")
        
        if not self.token:
            print("❌ No auth token, skipping channel test")
            return False
        
        # List channels
        resp = self.request('GET', '/admin/channels')
        print(f"List channels: {resp.status_code}")
        
        if resp.status_code == 200:
            try:
                result = resp.json()
                channels = result.get('data', {}).get('list', [])
                print(f"✅ Found {len(channels)} channels")
                
                for ch in channels[:3]:
                    print(f"  - {ch.get('name')} ({ch.get('type')}) @ {ch.get('base_url')}")
                
                return True
            except:
                print(f"Response: {resp.text}")
        
        return False
    
    def test_register_settings(self):
        """Test registration settings"""
        print("\n=== Test 5: Register Settings ===")
        
        # This might require admin auth
        resp = self.request('GET', '/admin/settings/register')
        print(f"Register settings: {resp.status_code}")
        
        if resp.status_code == 200:
            try:
                result = resp.json()
                print(f"✅ Register settings: {json.dumps(result.get('data', {}), indent=2)}")
                return True
            except:
                print(f"Response: {resp.text}")
        
        return False
    
    def test_redemption_codes(self):
        """Test redemption code creation (admin)"""
        print("\n=== Test 6: Create Redemption Code ===")
        
        if not self.token:
            print("❌ No auth token, skipping")
            return False
        
        data = {
            "code_type": "quota",
            "prefix": "TEST",
            "quota_amount": 10000,
            "count": 5,
            "note": "Test batch"
        }
        
        resp = self.request('POST', '/admin/redemption/codes', json=data)
        print(f"Create codes: {resp.status_code}")
        
        if resp.status_code == 200:
            try:
                result = resp.json()
                print(f"✅ Created: {json.dumps(result.get('data', {}), indent=2)}")
                return True
            except:
                print(f"Response: {resp.text}")
        else:
            print(f"Response: {resp.text[:200]}")
        
        return False
    
    def run_all(self):
        """Run all tests"""
        print("=" * 50)
        print("gAPI Platform API Test Suite")
        print("=" * 50)
        
        # Test 1: Health
        self.test_health()
        
        # Test 2: Register
        self.test_register()
        
        # Test 3: Login (get token)
        self.test_login()
        
        # Test 4: Channels
        self.test_channels()
        
        # Test 5: Register settings
        self.test_register_settings()
        
        # Test 6: Create redemption codes
        self.test_redemption_codes()
        
        print("\n" + "=" * 50)
        print("Test Complete")
        print("=" * 50)


# ============================================================
# Phase 4 (T-12): Pressure Test
# Usage: python test_api.py --pressure
# ============================================================
import concurrent.futures
import statistics
import argparse
import time
import threading


class PressureTest:
    """Concurrent API pressure testing with latency metrics."""

    def __init__(self, base_url="http://localhost:8080", token=None):
        self.base_url = base_url
        self.token = token
        self.results_lock = threading.Lock()
        self.latencies = []
        self.errors = 0
        self.success = 0

    def _do_chat(self, model="gpt-3.5-turbo", stream=False):
        """Single chat/completions request. Returns (latency_ms, status_code)."""
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
        """Single /v1/completions request."""
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
        """Single /v1/models request."""
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
        """Worker thread: execute fn n times and record results."""
        for _ in range(n):
            lat, code = fn()
            with self.results_lock:
                self.latencies.append(lat)
                if 200 <= code < 300:
                    self.success += 1
                else:
                    self.errors += 1

    def run_scenario(self, name, fn, concurrency, total_requests):
        """Run a single pressure scenario."""
        print(f"\n--- {name} ---")
        print(f"  Concurrency={concurrency}, Total={total_requests}")

        self.latencies = []
        self.errors = 0
        self.success = 0

        per_worker = total_requests // concurrency
        remainder = total_requests % concurrency

        t0 = time.time()

        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as executor:
            futures = []
            for i in range(concurrency):
                n = per_worker + (1 if i < remainder else 0)
                futures.append(executor.submit(self._worker, fn, n))
            concurrent.futures.wait(futures)

        elapsed = time.time() - t0
        tps = total_requests / elapsed if elapsed > 0 else 0

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
        """Run all Phase 4 pressure scenarios."""
        self.token = token

        print("=" * 60)
        print("Phase 4 (T-12): API Pressure Test Suite")
        print("=" * 60)

        results = []

        # Scenario 1: Non-streaming Chat API
        results.append(
            self.run_scenario(
                "Non-stream Chat API (50 concurrency)",
                lambda: self._do_chat(stream=False),
                concurrency=50,
                total_requests=500,
            )
        )

        # Scenario 2: /v1/models list
        results.append(
            self.run_scenario(
                "/v1/models list (100 concurrency)",
                self._do_models,
                concurrency=100,
                total_requests=500,
            )
        )

        # Scenario 3: /v1/completions
        results.append(
            self.run_scenario(
                "/v1/completions (30 concurrency)",
                self._do_completions,
                concurrency=30,
                total_requests=300,
            )
        )

        # Scenario 4: Low-concurrency chat (for environments with limited channels)
        results.append(
            self.run_scenario(
                "Chat API low-concurrency (10 concurrent)",
                lambda: self._do_chat(stream=False),
                concurrency=10,
                total_requests=50,
            )
        )

        # Summary
        print("\n" + "=" * 60)
        print("Pressure Test Summary")
        print("=" * 60)
        for r in results:
            status = "PASS" if r["error_rate_pct"] < 1 and r["p95_ms"] < 3000 else "FAIL"
            print(f"  [{status}] {r['scenario']}: "
                  f"TPS={r['tps']:.1f}, P95={r['p95_ms']:.0f}ms, "
                  f"Err={r['error_rate_pct']:.1f}%")

        return results


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="gAPI Platform API Tester")
    parser.add_argument("--pressure", action="store_true", help="Run pressure tests")
    parser.add_argument("--token", type=str, help="API token for authenticated tests")
    parser.add_argument("--base-url", type=str, default="http://localhost:8080",
                        help="Base URL (default: http://localhost:8080)")
    args = parser.parse_args()

    if args.pressure:
        pt = PressureTest(base_url=args.base_url)
        pt.run_all(token=args.token)
    else:
        tester = APITester()
        tester.run_all()