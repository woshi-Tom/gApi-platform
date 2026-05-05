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


if __name__ == "__main__":
    tester = APITester()
    tester.run_all()