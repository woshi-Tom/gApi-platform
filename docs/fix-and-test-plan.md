# 代码修复和测试文档开发计划

## 一、修复清单

### 🔴 高优先级修复

| 序号 | 问题 | 文件 | 修复方案 | 状态 |
|------|------|------|----------|------|
| 1 | `_decrypt_key` 未实现 | health_check_service.py | 调用 encryption.decrypt_api_key() | 待修复 |
| 2 | `TokenBatch` 表未定义 | pool_models.py | 添加 TokenBatch 模型 | 待修复 |
| 3 | `import json` 位置错误 | token_batch_service.py | 移到文件顶部 | 待修复 |
| 4 | `User` 模型跨表引用 | pool_models.py / billing_service.py | 从 db_models 导入 User | 待修复 |

### 🟡 中优先级优化

| 序号 | 问题 | 说明 | 状态 |
|------|------|------|------|
| 5 | encryption 单例 LSP 警告 | 添加 @property 装饰器 | 待优化 |
| 6 | 数据库 session 不一致 | 统一使用 context manager 模式 | 待优化 |

---

## 二、修复步骤

### Step 1: 添加 TokenBatch 模型

**文件**: `/home/claw/claw-ai/source_code/app/models/pool_models.py`

```python
class TokenBatch(Base):
    """令牌批次表"""
    __tablename__ = "token_batches"

    batch_id = Column(String(50), primary_key=True)
    batch_name = Column(String(100))
    total_count = Column(Integer, default=0)
    success_count = Column(Integer, default=0)
    failed_count = Column(Integer, default=0)
    created_by = Column(Integer)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)
```

### Step 2: 修复 health_check_service.py

**文件**: `/home/claw/claw-ai/source_code/app/services/health_check_service.py`

修复 `_decrypt_key` 方法：
```python
def _decrypt_key(self, encrypted_key: str) -> str:
    from app.utils.encryption import decrypt_api_key
    return decrypt_api_key(encrypted_key)
```

### Step 3: 修复 token_batch_service.py

**文件**: `/home/claw/claw-ai/source_code/app/services/token_batch_service.py`

将 `import json` 从文件末尾移到顶部。

### Step 4: 修复 billing_service.py 的 User 引用

**文件**: `/home/claw/claw-ai/source_code/app/services/billing_service.py`

将 `from app.models.pool_models import Token, User` 改为从 db_models 导入 User：
```python
from app.models.db_models import User
from app.models.pool_models import Token
```

---

## 三、测试用例设计

### 3.1 单元测试

#### test_billing_service.py

```python
class TestBillingService:
    def test_calculate_quota_gpt35(self):
        service = BillingService(mock_db)
        quota = service.calculate_quota("default", "gpt-3.5-turbo", 1000, 500)
        assert quota > 0
    
    def test_calculate_quota_gpt4(self):
        gpt35_quota = service.calculate_quota("default", "gpt-3.5-turbo", 1000, 500)
        gpt4_quota = service.calculate_quota("default", "gpt-4", 1000, 500)
        assert gpt4_quota > gpt35_quota
    
    def test_pre_consume_success(self):
        pass
    
    def test_pre_consume_insufficient(self):
        pass
```

#### test_encryption.py

```python
class TestEncryptionService:
    def test_encrypt_decrypt_roundtrip(self):
        original = "sk-test-api-key-12345"
        encrypted = encrypt_api_key(original)
        decrypted = decrypt_api_key(encrypted)
        assert decrypted == original
        assert encrypted != original
```

### 3.2 集成测试

```python
class TestIntegration:
    def test_full_billing_flow(self):
        pass
```

---

## 四、验收标准

| 任务 | 验收标准 |
|------|----------|
| TokenBatch 模型 | 模型定义正确，可创建表 |
| health_check_service | _decrypt_key 返回正确解密值 |
| token_batch_service | import json 在文件顶部 |
| billing_service | User 引用无错误 |
| 单元测试 | 每个服务至少 3 个测试用例 |
| 测试文档 | README.md 包含运行指南 |
