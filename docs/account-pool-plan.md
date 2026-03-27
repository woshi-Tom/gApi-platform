# Claw AI 项目合并 + 账号池功能开发计划

## 📋 项目概述

**目标**：
1. 合并后端 (`claw-ai`) 和前端 (`claw-ai-frontend`) 为统一项目
2. 基于 One API 设计模式，开发账号池功能

**当前状态**：
- 后端：Flask 3.0.3 + SQLAlchemy + MySQL，OpenAPI v1.4.9
- 前端：React 19 + Ant Design 5 + TypeScript，Vite 构建

---

## 🎯 阶段一：项目合并（1-2天）

### 1.1 目录结构调整

**目标**：将前后端合并为 monorepo 结构

```
claw-ai/
├── backend/                    # 原 source_code/ 改名
│   ├── app/
│   │   ├── __init__.py
│   │   ├── config.py
│   │   ├── blueprints/
│   │   ├── models/
│   │   ├── services/
│   │   └── utils/
│   ├── run.py
│   └── requirements.txt
├── frontend/                   # 原 claw-ai-frontend/claw-ai_-backend_-ui-main/
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── vite.config.ts
├── docs/                       # 项目文档
├── scripts/                    # 部署脚本
├── docker-compose.yml          # 统一容器编排
├── .gitignore
└── README.md
```

### 1.2 Git 仓库合并

**方法 A：保留历史（推荐）**
```bash
cd /home/claw/claw-ai
# 将前端作为子目录加入
git remote add frontend /home/claw/claw-ai-frontend
git fetch frontend
git merge frontend/main --allow-unrelated-histories
# 移动文件到 frontend/ 目录
git mv claw-ai_-backend_-ui-main/ frontend/
git rm -rf claw-ai_-backend_-ui-main/
```

**方法 B：简单复制**
```bash
# 直接复制前端到项目
cp -r /home/claw/claw-ai-frontend/claw-ai_-backend_-ui-main /home/claw/claw-ai/frontend
```

### 1.3 后端目录重命名

```bash
cd /home/claw/claw-ai
git mv source_code/ backend/
# 更新所有引用路径
```

### 1.4 统一开发环境

**创建 `docker-compose.yml`**：
```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: claw_ai
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  backend:
    build: ./backend
    ports:
      - "5005:5005"
    environment:
      - DB_ENABLED=1
      - DB_HOST=mysql
      - DB_NAME=claw_ai
    depends_on:
      - mysql

  frontend:
    build: ./frontend
    ports:
      - "5173:5173"
    environment:
      - VITE_API_BASE_URL=http://localhost:5005/api
    depends_on:
      - backend

volumes:
  mysql_data:
```

---

## 🎯 阶段二：账号池核心功能（3-5天）

基于 One API 的设计模式，实现账号池功能。

### 2.1 数据库设计

#### 2.1.1 新增表

**渠道表 `channels`**（存储上游 API 配置）
```sql
CREATE TABLE channels (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,           -- 渠道名称
    type VARCHAR(50) NOT NULL,            -- 渠道类型: openai, azure, claude, nvidia_nim, etc.
    base_url VARCHAR(500),                -- API 基础 URL
    api_key TEXT NOT NULL,                -- API 密钥 (加密存储)
    status INT DEFAULT 1,                 -- 1=启用, 2=禁用, 3=自动禁用, 4=检测中, 5=已死亡
    weight INT DEFAULT 1,                 -- 权重（加权随机）
    priority INT DEFAULT 0,               -- 优先级
    models TEXT,                          -- 支持的模型列表 (JSON)
    model_mapping TEXT,                   -- 模型映射 (JSON)
    balance DECIMAL(10,4) DEFAULT 0,     -- 余额
    used_quota BIGINT DEFAULT 0,         -- 已使用配额
    group_name VARCHAR(50) DEFAULT 'default',  -- 分组
    config TEXT,                          -- 扩展配置 (JSON)
    
    -- 健康检测字段（新增）
    last_check_at DATETIME,              -- 上次检测时间
    last_success_at DATETIME,            -- 上次成功时间
    failure_count INT DEFAULT 0,         -- 连续失败次数
    response_time_avg INT DEFAULT 0,    -- 平均响应时间(ms)
    is_healthy BOOLEAN DEFAULT TRUE,     -- 健康状态
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_status (status),
    INDEX idx_type (type),
    INDEX idx_is_healthy (is_healthy),
    INDEX idx_group_name (group_name)
);
```

**健康检测策略**：
| 参数 | 推荐值 | 说明 |
|------|--------|------|
| 检测频率 | 每 5 分钟 | APScheduler 定时任务 |
| 检测方式 | 实际请求（轻量查询） | HEAD 可能被拦截 |
| 失败阈值 | 3 次连续失败 | 自动标记 status=5 (dead) |
| 恢复检测 | dead 后每小时重试 | 有机会自动恢复 |
| 超时时间 | 30 秒 | 请求超时自动失败 |

**能力表 `abilities`**（模型-渠道关联）
```sql
CREATE TABLE abilities (
    group_name VARCHAR(50) NOT NULL,      -- 用户分组
    model VARCHAR(100) NOT NULL,          -- 模型名称
    channel_id INT NOT NULL,              -- 渠道 ID
    enabled BOOLEAN DEFAULT TRUE,         -- 是否启用
    priority INT DEFAULT 0,               -- 优先级
    PRIMARY KEY (group_name, model, channel_id),
    FOREIGN KEY (channel_id) REFERENCES channels(id)
);
```

**令牌表 `tokens`**（用户 API 令牌）
```sql
CREATE TABLE tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,                 -- 所属用户
    key VARCHAR(100) NOT NULL UNIQUE,     -- 令牌 (sk-xxx)
    name VARCHAR(100),                    -- 令牌名称
    status INT DEFAULT 1,                 -- 1=启用, 2=禁用, 3=过期, 4=额度耗尽
    remain_quota BIGINT DEFAULT 0,       -- 剩余配额
    used_quota BIGINT DEFAULT 0,         -- 已使用配额
    unlimited_quota BOOLEAN DEFAULT FALSE,-- 是否无限配额
    models TEXT,                          -- 允许的模型 (JSON)
    subnet VARCHAR(100),                  -- IP 限制
    expired_time BIGINT DEFAULT -1,       -- 过期时间 (-1=永不过期)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

**使用日志表 `usage_logs`**（API 调用记录）
```sql
CREATE TABLE usage_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    token_id INT,
    channel_id INT,
    model VARCHAR(100) NOT NULL,
    prompt_tokens INT DEFAULT 0,
    completion_tokens INT DEFAULT 0,
    quota BIGINT DEFAULT 0,              -- 消耗配额
    request_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    INDEX idx_user_request (user_id, request_at),
    INDEX idx_channel_request (channel_id, request_at),
    INDEX idx_model (model)
);
```

**审计日志表 `audit_logs`**（溯源和合规必需）
```sql
CREATE TABLE audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_type VARCHAR(50) NOT NULL,     -- login/logout/token_create/quota_change/channel_change
    user_id INT,
    operator_id INT,                     -- 操作者（管理员）
    target_type VARCHAR(50),            -- 目标类型：channel/token/user
    target_id INT,                       -- 目标 ID
    action VARCHAR(20) NOT NULL,        -- create/update/delete/enable/disable
    detail TEXT,                         -- 详细信息 (JSON)
    ip_address VARCHAR(45),             -- IP 地址
    user_agent TEXT,                     -- 浏览器 UA
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_event_type (event_type),
    INDEX idx_user_time (user_id, created_at),
    INDEX idx_target (target_type, target_id)
);

-- 审计日志事件类型枚举
-- login: 用户登录
-- logout: 用户登出
-- token_create: 创建 Token
-- token_update: 更新 Token
-- token_delete: 删除 Token
-- channel_create: 创建渠道
-- channel_update: 更新渠道
-- channel_delete: 删除渠道
-- channel_enable: 启用渠道
-- channel_disable: 禁用渠道
-- quota_grant: 配额发放
-- quota_deduct: 配额扣除
```

#### 2.1.2 修改现有表

**用户表 `users`** 添加字段：
```sql
ALTER TABLE users ADD COLUMN quota BIGINT DEFAULT 0;        -- 用户总配额
ALTER TABLE users ADD COLUMN used_quota BIGINT DEFAULT 0;   -- 已使用配额
ALTER TABLE users ADD COLUMN api_group VARCHAR(50) DEFAULT 'default';  -- API 分组
```

---

### 2.2 后端实现

#### 2.2.1 新增模块

```
backend/app/
├── models/
│   ├── db_models.py           # 现有模型
│   ├── channel_model.py       # 渠道模型（新增）
│   ├── ability_model.py       # 能力模型（新增）
│   └── token_model.py         # 令牌模型（新增）
├── services/
│   ├── channel_service.py     # 渠道管理服务（新增）
│   ├── channel_health.py      # 渠道健康检测服务（新增）
│   ├── distributor_service.py # 分发服务（新增）
│   ├── billing_service.py     # 计费服务（新增）
│   └── audit_service.py       # 审计日志服务（新增）
├── blueprints/
│   ├── frontend.py            # 现有前台
│   ├── admin.py               # 现有后台
│   └── api_proxy.py           # API 代理（新增）
└── adapters/                  # 适配器目录（新增）
    ├── base_adapter.py        # 基础适配器
    ├── openai_adapter.py      # OpenAI 适配器
    ├── claude_adapter.py      # Claude 适配器
    └── gemini_adapter.py      # Gemini 适配器
```

#### 2.2.2 核心服务实现

**分发服务 `distributor_service.py`**：
```python
class DistributorService:
    def select_channel(self, group: str, model: str) -> Channel:
        """选择可用渠道（优先级 + 随机）"""
        abilities = Ability.query.filter_by(
            group_name=group,
            model=model,
            enabled=True
        ).order_by(Ability.priority.desc()).all()
        
        if not abilities:
            raise NoChannelError(f"No available channel for {model}")
        
        # 按优先级分组，随机选择
        max_priority = abilities[0].priority
        top_abilities = [a for a in abilities if a.priority == max_priority]
        selected = random.choice(top_abilities)
        
        return Channel.query.get(selected.channel_id)
    
    def select_channel_with_retry(self, group: str, model: str, 
                                   exclude_ids: List[int] = None) -> Channel:
        """带重试的渠道选择"""
        exclude_ids = exclude_ids or []
        abilities = Ability.query.filter(
            Ability.group_name == group,
            Ability.model == model,
            Ability.enabled == True,
            Ability.channel_id.notin_(exclude_ids)
        ).order_by(Ability.priority.desc()).all()
        
        if not abilities:
            raise NoChannelError("All channels exhausted")
        
        return Channel.query.get(random.choice(abilities).channel_id)
```

**计费服务 `billing_service.py`**：
```python
class BillingService:
    # 模型倍率配置
    MODEL_RATES = {
        'gpt-3.5-turbo': 1.0,
        'gpt-4': 20.0,
        'gpt-4-turbo': 10.0,
        'claude-3-opus': 15.0,
        'claude-3-sonnet': 3.0,
    }
    
    # 分组倍率配置
    GROUP_RATES = {
        'default': 1.0,
        'vip': 0.8,
        'enterprise': 0.6,
    }
    
    def calculate_quota(self, group: str, model: str, 
                        prompt_tokens: int, completion_tokens: int) -> int:
        """计算消耗配额"""
        model_rate = self.MODEL_RATES.get(model, 1.0)
        group_rate = self.GROUP_RATES.get(group, 1.0)
        completion_rate = 1.33 if 'gpt-3.5' in model else 2.0
        
        quota = int(group_rate * model_rate * 
                    (prompt_tokens + completion_tokens * completion_rate))
        return quota
    
    def pre_consume(self, user_id: int, token_id: int, estimated_quota: int):
        """预扣配额"""
        user = User.query.get(user_id)
        token = Token.query.get(token_id)
        
        if not token.unlimited_quota and token.remain_quota < estimated_quota:
            raise QuotaInsufficientError("Token quota insufficient")
        if user.quota - user.used_quota < estimated_quota:
            raise QuotaInsufficientError("User quota insufficient")
        
        user.used_quota += estimated_quota
        token.used_quota += estimated_quota
        if not token.unlimited_quota:
            token.remain_quota -= estimated_quota
        db.session.commit()
    
    def post_consume(self, user_id: int, token_id: int, 
                     actual_quota: int, pre_consumed: int):
        """请求完成后调整配额"""
        delta = pre_consumed - actual_quota
        user = User.query.get(user_id)
        token = Token.query.get(token_id)
        
        user.used_quota -= delta
        token.used_quota -= delta
        if not token.unlimited_quota:
            token.remain_quota += delta
        db.session.commit()
```

**健康检测服务 `channel_health.py`**：
```python
from apscheduler.schedulers.background import BackgroundScheduler
import requests
import time

class ChannelHealthService:
    """渠道健康检测服务"""
    
    def __init__(self):
        self.scheduler = BackgroundScheduler()
        self.FAILURE_THRESHOLD = 3
        self.CHECK_INTERVAL_MINUTES = 5
        self.DEAD_RETRY_INTERVAL_HOURS = 1
        self.REQUEST_TIMEOUT = 30
    
    def start(self):
        """启动定时检测任务"""
        self.scheduler.add_job(
            self.check_all_channels,
            'interval',
            minutes=self.CHECK_INTERVAL_MINUTES,
            id='health_check'
        )
        self.scheduler.start()
    
    def check_all_channels(self):
        """检测所有启用的渠道"""
        channels = Channel.query.filter(Channel.status.in_([1, 4])).all()
        for channel in channels:
            try:
                self.check_channel(channel)
            except Exception as e:
                logging.error(f"Channel {channel.id} health check failed: {e}")
    
    def check_channel(self, channel: Channel) -> dict:
        """检测单个渠道"""
        channel.status = 4  # 检测中
        db.session.commit()
        
        start_time = time.time()
        result = {'success': False, 'error': None, 'response_time': 0}
        
        try:
            # 根据渠道类型选择检测方式
            if channel.type == 'openai':
                result = self._check_openai(channel)
            elif channel.type == 'claude':
                result = self._check_claude(channel)
            else:
                result = self._check_generic(channel)
            
            result['response_time'] = int((time.time() - start_time) * 1000)
            
        except Exception as e:
            result['error'] = str(e)
        
        # 更新渠道状态
        self._update_channel_status(channel, result)
        
        return result
    
    def _check_openai(self, channel: Channel) -> dict:
        """检测 OpenAI 渠道"""
        headers = {
            'Authorization': f'Bearer {decrypt_api_key(channel.api_key)}',
            'Content-Type': 'application/json'
        }
        url = f'{channel.base_url}/models'
        response = requests.get(url, headers=headers, timeout=self.REQUEST_TIMEOUT)
        
        if response.status_code == 200:
            return {'success': True}
        else:
            return {'success': False, 'error': f'HTTP {response.status_code}'}
    
    def _check_generic(self, channel: Channel) -> dict:
        """通用检测（使用 models 接口）"""
        headers = {
            'Authorization': f'Bearer {decrypt_api_key(channel.api_key)}',
        }
        url = f'{channel.base_url}/models'
        response = requests.get(url, headers=headers, timeout=self.REQUEST_TIMEOUT)
        
        if response.status_code == 200:
            return {'success': True}
        else:
            return {'success': False, 'error': f'HTTP {response.status_code}'}
    
    def _update_channel_status(self, channel: Channel, result: dict):
        """根据检测结果更新渠道状态"""
        if result['success']:
            channel.status = 1  # 启用
            channel.is_healthy = True
            channel.failure_count = 0
            channel.last_success_at = datetime.now()
            channel.last_check_at = datetime.now()
            
            # 更新平均响应时间（滑动平均）
            if result.get('response_time'):
                old_avg = channel.response_time_avg or 0
                channel.response_time_avg = int((old_avg * 4 + result['response_time']) / 5)
        else:
            channel.failure_count += 1
            channel.last_check_at = datetime.now()
            
            if channel.failure_count >= self.FAILURE_THRESHOLD:
                channel.status = 5  # 已死亡
                channel.is_healthy = False
                logging.warning(f"Channel {channel.id} marked as dead after {channel.failure_count} failures")
        
        db.session.commit()
    
    def reset_dead_channel(self, channel_id: int) -> bool:
        """重置死亡渠道（手动恢复）"""
        channel = Channel.query.get(channel_id)
        if not channel:
            return False
        
        channel.status = 1
        channel.is_healthy = True
        channel.failure_count = 0
        db.session.commit()
        
        # 立即检测
        self.check_channel(channel)
        return True
```

**审计日志服务 `audit_service.py`**：
```python
from functools import wraps
from flask import request
import json

class AuditService:
    """审计日志服务"""
    
    @staticmethod
    def log(event_type: str, user_id: int = None, operator_id: int = None,
            target_type: str = None, target_id: int = None,
            action: str = None, detail: dict = None):
        """记录审计日志"""
        audit = AuditLog(
            event_type=event_type,
            user_id=user_id,
            operator_id=operator_id,
            target_type=target_type,
            target_id=target_id,
            action=action,
            detail=json.dumps(detail) if detail else None,
            ip_address=request.remote_addr,
            user_agent=request.headers.get('User-Agent')
        )
        db.session.add(audit)
        db.session.commit()
    
    @staticmethod
    def query_logs(event_type: str = None, user_id: int = None,
                   target_type: str = None, target_id: int = None,
                   start_time: datetime = None, end_time: datetime = None,
                   page: int = 1, page_size: int = 50):
        """查询审计日志"""
        query = AuditLog.query
        
        if event_type:
            query = query.filter_by(event_type=event_type)
        if user_id:
            query = query.filter_by(user_id=user_id)
        if target_type:
            query = query.filter_by(target_type=target_type)
        if target_id:
            query = query.filter_by(target_id=target_id)
        if start_time:
            query = query.filter(AuditLog.created_at >= start_time)
        if end_time:
            query = query.filter(AuditLog.created_at <= end_time)
        
        query = query.order_by(AuditLog.created_at.desc())
        
        return query.paginate(page=page, per_page=page_size)

def audit_log(event_type: str, detail_func=None):
    """审计日志装饰器"""
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            result = func(*args, **kwargs)
            
            user_id = get_current_user_id()
            detail = detail_func(*args, **kwargs) if detail_func else None
            
            AuditService.log(
                event_type=event_type,
                user_id=user_id,
                action='update',
                detail=detail
            )
            
            return result
        return wrapper
    return decorator
```

**API 代理蓝图 `api_proxy.py`**：
```python
from flask import Blueprint, request, Response, stream_with_context
import requests

api_proxy = Blueprint('api_proxy', __name__, url_prefix='/v1')

@api_proxy.route('/chat/completions', methods=['POST'])
@require_token_auth
def chat_completions():
    """OpenAI 兼容的聊天接口"""
    user = get_current_user()
    token = get_current_token()
    data = request.get_json()
    model = data.get('model', '')
    
    # 1. 选择渠道
    distributor = DistributorService()
    channel = distributor.select_channel(user.api_group, model)
    
    # 2. 预扣配额
    billing = BillingService()
    estimated_quota = billing.calculate_quota(user.api_group, model, 1000, 500)
    billing.pre_consume(user.id, token.id, estimated_quota)
    
    try:
        # 3. 获取适配器
        adapter = get_adapter(channel.type)
        
        # 4. 转发请求
        if data.get('stream'):
            return Response(
                stream_with_context(adapter.stream_chat(channel, data)),
                mimetype='text/event-stream'
            )
        else:
            response = adapter.chat(channel, data)
            actual_quota = billing.calculate_quota(
                user.api_group, model,
                response['usage']['prompt_tokens'],
                response['usage']['completion_tokens']
            )
            billing.post_consume(user.id, token.id, actual_quota, estimated_quota)
            return response
    except Exception as e:
        # 失败重试
        return retry_with_other_channel(user, token, data, exclude_ids=[channel.id])

@api_proxy.route('/models', methods=['GET'])
@require_token_auth
def list_models():
    """列出可用模型"""
    user = get_current_user()
    abilities = Ability.query.filter_by(
        group_name=user.api_group,
        enabled=True
    ).distinct(Ability.model).all()
    
    return {
        'data': [{'id': a.model, 'object': 'model'} for a in abilities]
    }
```

#### 2.2.3 后台管理接口

在 `admin.py` 中添加渠道管理接口：

```python
# 渠道管理
@admin.route('/channels', methods=['GET'])
def list_channels():
    """渠道列表"""
    page = request.args.get('page', 1, type=int)
    page_size = request.args.get('pageSize', 10, type=int)
    channels = Channel.query.paginate(page=page, per_page=page_size)
    return channels_to_dict(channels)

@admin.route('/channels', methods=['POST'])
def create_channel():
    """创建渠道"""
    data = request.get_json()
    channel = Channel(**data)
    db.session.add(channel)
    db.session.commit()
    # 自动创建能力记录
    create_abilities_for_channel(channel)
    return channel_to_dict(channel), 201

@admin.route('/channels/<int:channel_id>', methods=['PUT'])
def update_channel(channel_id):
    """更新渠道"""
    channel = Channel.query.get_or_404(channel_id)
    data = request.get_json()
    for key, value in data.items():
        setattr(channel, key, value)
    db.session.commit()
    return channel_to_dict(channel)

@admin.route('/channels/<int:channel_id>', methods=['DELETE'])
def delete_channel(channel_id):
    """删除渠道"""
    channel = Channel.query.get_or_404(channel_id)
    # 删除关联的能力记录
    Ability.query.filter_by(channel_id=channel_id).delete()
    db.session.delete(channel)
    db.session.commit()
    return '', 204

# === 批量操作 API（新增）===

@admin.route('/channels/batch', methods=['POST'])
def batch_create_channels():
    """批量导入渠道（支持 YAML 配置）"""
    data = request.get_json()
    channels_data = data.get('channels', [])
    results = []
    
    for item in channels_data:
        channel = Channel(
            name=item['name'],
            type=item['type'],
            base_url=item.get('base_url'),
            api_key=encrypt_api_key(item['api_key']),
            models=json.dumps(item.get('models', [])),
            weight=item.get('weight', 1),
            priority=item.get('priority', 0),
            group_name=item.get('group', 'default')
        )
        db.session.add(channel)
        results.append(channel)
    
    db.session.commit()
    
    # 自动创建能力记录
    for channel in results:
        create_abilities_for_channel(channel)
    
    return {'created': len(results), 'channels': [c.id for c in results]}, 201

@admin.route('/channels/<int:channel_id>/test', methods=['POST'])
def test_channel(channel_id):
    """手动测试渠道连通性"""
    channel = Channel.query.get_or_404(channel_id)
    result = test_channel_health(channel)
    return {'channel_id': channel_id, 'result': result}

@admin.route('/channels/<int:channel_id>/enable', methods=['POST'])
def enable_channel(channel_id):
    """启用渠道"""
    channel = Channel.query.get_or_404(channel_id)
    channel.status = 1
    channel.is_healthy = True
    channel.failure_count = 0
    db.session.commit()
    return {'status': 'enabled'}

@admin.route('/channels/<int:channel_id>/disable', methods=['POST'])
def disable_channel(channel_id):
    """禁用渠道"""
    channel = Channel.query.get_or_404(channel_id)
    channel.status = 2
    db.session.commit()
    return {'status': 'disabled'}

# === 健康检测 API ===

@admin.route('/health-check/status', methods=['GET'])
def health_check_status():
    """健康检测状态概览"""
    stats = {
        'total': Channel.query.count(),
        'healthy': Channel.query.filter_by(is_healthy=True).count(),
        'unhealthy': Channel.query.filter_by(is_healthy=False).count(),
        'dead': Channel.query.filter_by(status=5).count(),
        'checking': Channel.query.filter_by(status=4).count(),
    }
    return stats

@admin.route('/health-check/trigger', methods=['POST'])
def trigger_health_check():
    """手动触发全量健康检测"""
    # 更新所有启用渠道为检测中
    Channel.query.filter_by(status=1).update({'status': 4})
    db.session.commit()
    
    # 异步触发检测任务
    from tasks import check_all_channels
    check_all_channels.delay()
    
    return {'status': 'triggered', 'message': '健康检测任务已触发'}

# 令牌管理
@admin.route('/tokens', methods=['GET'])
def list_tokens():
    """令牌列表"""
    user_id = request.args.get('userId')
    query = Token.query
    if user_id:
        query = query.filter_by(user_id=user_id)
    return tokens_to_dict(query.all())

@admin.route('/tokens', methods=['POST'])
def create_token():
    """创建令牌"""
    data = request.get_json()
    token = Token(
        user_id=data['userId'],
        key=generate_token_key(),  # 生成 sk-xxx 格式
        name=data.get('name'),
        remain_quota=data.get('quota', 0),
        unlimited_quota=data.get('unlimited', False)
    )
    db.session.add(token)
    db.session.commit()
    
    # 记录审计日志
    AuditService.log(
        event_type='token_create',
        operator_id=get_current_admin_id(),
        target_type='token',
        target_id=token.id,
        action='create',
        detail={'user_id': data['userId'], 'name': data.get('name')}
    )
    
    return token_to_dict(token), 201

# === 审计日志 API ===

@admin.route('/audit-logs', methods=['GET'])
def list_audit_logs():
    """查询审计日志"""
    event_type = request.args.get('eventType')
    user_id = request.args.get('userId', type=int)
    target_type = request.args.get('targetType')
    target_id = request.args.get('targetId', type=int)
    start_time = request.args.get('startTime')
    end_time = request.args.get('endTime')
    page = request.args.get('page', 1, type=int)
    page_size = request.args.get('pageSize', 50, type=int)
    
    logs = AuditService.query_logs(
        event_type=event_type,
        user_id=user_id,
        target_type=target_type,
        target_id=target_id,
        start_time=datetime.fromisoformat(start_time) if start_time else None,
        end_time=datetime.fromisoformat(end_time) if end_time else None,
        page=page,
        page_size=page_size
    )
    
    return {
        'data': [log_to_dict(log) for log in logs.items],
        'total': logs.total,
        'page': page,
        'pageSize': page_size
    }

@admin.route('/audit-logs/export', methods=['GET'])
def export_audit_logs():
    """导出审计日志（CSV）"""
    logs = AuditService.query_logs(page=1, page_size=10000)
    
    output = io.StringIO()
    writer = csv.writer(output)
    writer.writerow(['时间', '事件类型', '用户', '操作者', '目标类型', '目标ID', '操作', '详情', 'IP'])
    
    for log in logs.items:
        writer.writerow([
            log.created_at,
            log.event_type,
            log.user_id,
            log.operator_id,
            log.target_type,
            log.target_id,
            log.action,
            log.detail,
            log.ip_address
        ])
    
    return Response(
        output.getvalue(),
        mimetype='text/csv',
        headers={'Content-Disposition': 'attachment; filename=audit_logs.csv'}
    )

# === 使用统计 API ===

@admin.route('/usage-stats', methods=['GET'])
def usage_stats():
    """使用统计概览"""
    period = request.args.get('period', 'day')  # day/week/month
    
    # 计算时间范围
    if period == 'day':
        start_time = datetime.now() - timedelta(days=1)
    elif period == 'week':
        start_time = datetime.now() - timedelta(weeks=1)
    else:
        start_time = datetime.now() - timedelta(days=30)
    
    # 基础统计
    stats = {
        'total_requests': UsageLog.query.filter(UsageLog.request_at >= start_time).count(),
        'total_quota': db.session.query(func.sum(UsageLog.quota)).filter(UsageLog.request_at >= start_time).scalar() or 0,
        'total_users': db.session.query(func.count(func.distinct(UsageLog.user_id))).filter(UsageLog.request_at >= start_time).scalar(),
        'total_tokens': db.session.query(func.count(func.distinct(UsageLog.token_id))).filter(UsageLog.request_at >= start_time).scalar(),
    }
    
    # 模型使用排行
    model_stats = db.session.query(
        UsageLog.model,
        func.count(UsageLog.id).label('count'),
        func.sum(UsageLog.quota).label('quota')
    ).filter(UsageLog.request_at >= start_time).group_by(UsageLog.model).order_by(desc('count')).limit(10).all()
    
    stats['top_models'] = [{'model': m[0], 'requests': m[1], 'quota': m[2]} for m in model_stats]
    
    # 渠道使用统计
    channel_stats = db.session.query(
        Channel.name,
        func.count(UsageLog.id).label('count'),
        func.sum(UsageLog.quota).label('quota')
    ).join(UsageLog, Channel.id == UsageLog.channel_id).filter(
        UsageLog.request_at >= start_time
    ).group_by(Channel.name).all()
    
    stats['channel_usage'] = [{'channel': c[0], 'requests': c[1], 'quota': c[2]} for c in channel_stats]
    
    # 每日趋势
    daily_stats = db.session.query(
        func.date(UsageLog.request_at).label('date'),
        func.count(UsageLog.id).label('count'),
        func.sum(UsageLog.quota).label('quota')
    ).filter(UsageLog.request_at >= start_time).group_by(func.date(UsageLog.request_at)).all()
    
    stats['daily_trend'] = [{'date': str(d[0]), 'requests': d[1], 'quota': d[2]} for d in daily_stats]
    
    return stats
```

---

### 2.3 前端实现

#### 2.3.1 新增页面

```
frontend/src/
├── pages/
│   ├── channels_page.tsx      # 渠道管理（新增）
│   ├── tokens_page.tsx        # 令牌管理（新增）
│   ├── usage_logs_page.tsx    # 使用日志（新增）
│   └── ...existing pages
├── api/
│   └── channel_api.ts         # 渠道 API（新增）
└── types/
    └── channel.ts             # 渠道类型（新增）
```

#### 2.3.2 渠道管理页面 `channels_page.tsx`

```tsx
import { Table, Button, Modal, Form, Input, Select, Switch, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useState, useEffect } from 'react'
import { getChannels, createChannel, updateChannel, deleteChannel } from '../api/channel_api'

const CHANNEL_TYPES = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'Azure OpenAI', value: 'azure' },
  { label: 'NVIDIA NIM', value: 'nvidia_nim' },
  { label: 'Anthropic Claude', value: 'claude' },
  { label: 'Google Gemini', value: 'gemini' },
  { label: 'DeepSeek', value: 'deepseek' },
  { label: '智谱 ChatGLM', value: 'zhipu' },
  { label: '通义千问', value: 'tongyi' },
]

export function ChannelsPage() {
  const [channels, setChannels] = useState([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingChannel, setEditingChannel] = useState(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchChannels()
  }, [])

  const fetchChannels = async () => {
    setLoading(true)
    const data = await getChannels()
    setChannels(data)
    setLoading(false)
  }

  const handleCreate = async (values) => {
    if (editingChannel) {
      await updateChannel(editingChannel.id, values)
      message.success('渠道更新成功')
    } else {
      await createChannel(values)
      message.success('渠道创建成功')
    }
    setModalOpen(false)
    form.resetFields()
    fetchChannels()
  }

  const handleDelete = async (id) => {
    await deleteChannel(id)
    message.success('渠道删除成功')
    fetchChannels()
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'type', render: (t) => CHANNEL_TYPES.find(c => c.value === t)?.label || t },
    { title: '状态', dataIndex: 'status', render: (s) => s === 1 ? '启用' : '禁用' },
    { title: '权重', dataIndex: 'weight' },
    { title: '优先级', dataIndex: 'priority' },
    { title: '余额', dataIndex: 'balance', render: (b) => `$${b?.toFixed(2) || '0.00'}` },
    { title: '分组', dataIndex: 'group_name' },
    {
      title: '操作',
      render: (_, record) => (
        <>
          <Button size="small" onClick={() => { setEditingChannel(record); form.setFieldsValue(record); setModalOpen(true) }}>编辑</Button>
          <Button size="small" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </>
      )
    }
  ]

  return (
    <div>
      <Button icon={<PlusOutlined />} type="primary" onClick={() => { setEditingChannel(null); form.resetFields(); setModalOpen(true) }}>
        新建渠道
      </Button>
      <Table columns={columns} dataSource={channels} loading={loading} rowKey="id" />
      
      <Modal title={editingChannel ? '编辑渠道' : '新建渠道'} open={modalOpen} onCancel={() => setModalOpen(false)} onOk={() => form.submit()}>
        <Form form={form} onFinish={handleCreate}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={CHANNEL_TYPES} />
          </Form.Item>
          <Form.Item name="base_url" label="Base URL">
            <Input placeholder="https://api.openai.com" />
          </Form.Item>
          <Form.Item name="api_key" label="API Key" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="models" label="支持模型">
            <Input.TextArea placeholder="gpt-3.5-turbo, gpt-4, ..." />
          </Form.Item>
          <Form.Item name="weight" label="权重" initialValue={1}>
            <Input type="number" />
          </Form.Item>
          <Form.Item name="priority" label="优先级" initialValue={0}>
            <Input type="number" />
          </Form.Item>
          <Form.Item name="group_name" label="分组" initialValue="default">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
```

#### 2.3.3 路由更新

在 `app_router.tsx` 中添加新路由：

```tsx
import { ChannelsPage } from '../pages/channels_page'
import { TokensPage } from '../pages/tokens_page'
import { UsageLogsPage } from '../pages/usage_logs_page'

const routes: RouteObject[] = [
  // ...existing routes
  { path: '/channels', element: protected_element(<ChannelsPage />) },
  { path: '/tokens', element: protected_element(<TokensPage />) },
  { path: '/usage-logs', element: protected_element(<UsageLogsPage />) },
]
```

#### 2.3.4 侧边栏更新

在 `admin_layout.tsx` 中添加菜单项：

```tsx
const menu_items: MenuProps['items'] = [
  // ...existing items
  { key: '/channels', icon: <ApiOutlined />, label: '渠道管理' },
  { key: '/tokens', icon: <KeyOutlined />, label: '令牌管理' },
  { key: '/usage-logs', icon: <FileTextOutlined />, label: '使用日志' },
]
```

---

## 🎯 阶段三：适配器实现（2-3天）

### 3.1 基础适配器接口

```python
# backend/app/adapters/base_adapter.py
from abc import ABC, abstractmethod

class BaseAdapter(ABC):
    """基础适配器接口"""
    
    @abstractmethod
    def chat(self, channel, data: dict) -> dict:
        """非流式聊天"""
        pass
    
    @abstractmethod
    def stream_chat(self, channel, data: dict):
        """流式聊天"""
        pass
    
    @abstractmethod
    def embedding(self, channel, data: dict) -> dict:
        """文本嵌入"""
        pass
    
    @abstractmethod
    def list_models(self, channel) -> list:
        """列出可用模型"""
        pass
```

### 3.2 OpenAI 适配器

```python
# backend/app/adapters/openai_adapter.py
import requests

class OpenAIAdapter(BaseAdapter):
    def chat(self, channel, data: dict) -> dict:
        headers = {
            'Authorization': f'Bearer {channel.api_key}',
            'Content-Type': 'application/json'
        }
        
        url = f'{channel.base_url}/chat/completions'
        response = requests.post(url, json=data, headers=headers)
        response.raise_for_status()
        return response.json()
    
    def stream_chat(self, channel, data: dict):
        headers = {
            'Authorization': f'Bearer {channel.api_key}',
            'Content-Type': 'application/json'
        }
        data['stream'] = True
        
        url = f'{channel.base_url}/chat/completions'
        response = requests.post(url, json=data, headers=headers, stream=True)
        
        for line in response.iter_lines():
            if line:
                yield line.decode('utf-8') + '\n\n'
```

### 3.3 NVIDIA NIM 适配器

```python
# backend/app/adapters/nvidia_nim_adapter.py
class NVIDIANIMAdapter(BaseAdapter):
    """
    NVIDIA NIM 适配器 - OpenAI 兼容格式
    特点：
    - API 格式与 OpenAI 完全兼容
    - 模型名称带命名空间：publisher/model-name
    - API Key 格式：nvapi-xxx
    - 部分模型免费（Kimi-k2, MiniMax-M2.5 等）
    """
    
    BASE_URL = "https://integrate.api.nvidia.com/v1"
    
    def chat(self, channel, data: dict) -> dict:
        headers = {
            'Authorization': f'Bearer {channel.api_key}',
            'Content-Type': 'application/json'
        }
        
        url = f'{channel.base_url or self.BASE_URL}/chat/completions'
        response = requests.post(url, json=data, headers=headers, timeout=120)
        response.raise_for_status()
        return response.json()
    
    def stream_chat(self, channel, data: dict):
        headers = {
            'Authorization': f'Bearer {channel.api_key}',
            'Content-Type': 'application/json'
        }
        data['stream'] = True
        
        url = f'{channel.base_url or self.BASE_URL}/chat/completions'
        response = requests.post(url, json=data, headers=headers, stream=True, timeout=120)
        
        for line in response.iter_lines():
            if line:
                yield line.decode('utf-8') + '\n\n'
```

**NVIDIA NIM 免费模型（推荐优先集成）**：
- `moonshotai/kimi-k2-instruct` - 国产模型
- `minimaxai/minimax-m2.5` - 高质量国产模型
- `meta/llama-3.1-8b-instruct` - 经典开源模型
- `deepseek-ai/deepseek-r1-distill-llama-8b` - DeepSeek 蒸馏版

**注意事项**：
- ⚠️ 国内访问需要 VPN/代理
- ⚠️ 模型名称格式：`publisher/model-name`
- ⚠️ 部分模型需要设置 `max_tokens` 参数

---

### 3.4 Claude 适配器

```python
# backend/app/adapters/claude_adapter.py
class ClaudeAdapter(BaseAdapter):
    def chat(self, channel, data: dict) -> dict:
        headers = {
            'x-api-key': channel.api_key,
            'anthropic-version': '2023-06-01',
            'Content-Type': 'application/json'
        }
        
        # 转换 OpenAI 格式到 Claude 格式
        claude_data = self._convert_request(data)
        
        url = f'{channel.base_url}/v1/messages'
        response = requests.post(url, json=claude_data, headers=headers)
        response.raise_for_status()
        
        # 转换 Claude 响应到 OpenAI 格式
        return self._convert_response(response.json())
    
    def _convert_request(self, openai_data):
        """OpenAI -> Claude 请求转换"""
        return {
            'model': openai_data['model'],
            'max_tokens': openai_data.get('max_tokens', 4096),
            'messages': [
                {'role': m['role'], 'content': m['content']}
                for m in openai_data['messages']
            ]
        }
    
    def _convert_response(self, claude_data):
        """Claude -> OpenAI 响应转换"""
        return {
            'id': claude_data['id'],
            'object': 'chat.completion',
            'choices': [{
                'index': 0,
                'message': {
                    'role': 'assistant',
                    'content': claude_data['content'][0]['text']
                },
                'finish_reason': claude_data['stop_reason']
            }],
            'usage': {
                'prompt_tokens': claude_data['usage']['input_tokens'],
                'completion_tokens': claude_data['usage']['output_tokens']
            }
        }
```

---

## 🎯 阶段四：测试与部署（1-2天）

### 4.1 单元测试

```python
# backend/tests/test_distributor.py
def test_select_channel_random():
    """测试随机选择渠道"""
    # 创建多个相同优先级的渠道
    channel1 = create_test_channel(priority=10)
    channel2 = create_test_channel(priority=10)
    
    distributor = DistributorService()
    selected = distributor.select_channel('default', 'gpt-3.5-turbo')
    
    assert selected.id in [channel1.id, channel2.id]

def test_select_channel_priority():
    """测试优先级选择"""
    low = create_test_channel(priority=1)
    high = create_test_channel(priority=10)
    
    distributor = DistributorService()
    selected = distributor.select_channel('default', 'gpt-3.5-turbo')
    
    assert selected.id == high.id
```

### 4.2 集成测试

```python
# backend/tests/test_api_proxy.py
def test_chat_completions():
    """测试聊天接口"""
    # 创建测试用户和令牌
    user = create_test_user(quota=10000)
    token = create_test_token(user_id=user.id, remain_quota=10000)
    
    # 创建测试渠道
    channel = create_test_channel(
        type='openai',
        api_key='test-key',
        base_url='http://localhost:8080'  # Mock 服务器
    )
    
    response = client.post('/v1/chat/completions', 
        json={'model': 'gpt-3.5-turbo', 'messages': [{'role': 'user', 'content': 'Hi'}]},
        headers={'Authorization': f'Bearer {token.key}'}
    )
    
    assert response.status_code == 200
    assert 'choices' in response.json
```

### 4.3 部署配置

```yaml
# podman-compose.yml (更新)
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: claw_ai
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./backend/init.sql:/docker-entrypoint-initdb.d/init.sql

  backend:
    build: ./backend
    ports:
      - "5005:5005"
    environment:
      - DB_ENABLED=1
      - DB_HOST=mysql
      - DB_NAME=claw_ai
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - RETRY_TIMES=3
    depends_on:
      - mysql
    volumes:
      - ./backend/uploads:/app/uploads

  frontend:
    build: ./frontend
    ports:
      - "5173:5173"
    environment:
      - VITE_API_BASE_URL=/api
    depends_on:
      - backend

  nginx:
    image: nginx:alpine
    ports:
      - "85:85"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - backend
      - frontend

volumes:
  mysql_data:
```

---

## 📊 时间估算

| 阶段 | 任务 | 预计时间 |
|------|------|----------|
| 一 | 项目合并 | 1-2 天 |
| 二 | 账号池核心功能 | 3-5 天 |
| 三 | 适配器实现 | 2-3 天 |
| 四 | 测试与部署 | 1-2 天 |
| **总计** | | **7-12 天** |

---

## 📈 进度管理（新增）

### 进度追踪机制

**每日站会要点**：
1. 昨日完成内容
2. 今日计划任务
3. 阻塞问题清单

**里程碑检查点**：

| 里程碑 | 目标时间 | 验收标准 |
|--------|----------|----------|
| M1: 项目合并完成 | Day 2 | 前端可正常访问后端 API |
| M2: 渠道管理上线 | Day 5 | 支持 CRUD + 健康检测 |
| M3: Token 管理上线 | Day 6 | 支持创建/分发/配额 |
| M4: API 代理可用 | Day 8 | OpenAI 兼容接口可调用 |
| M5: 完整功能上线 | Day 12 | 所有功能联调完成 |

**进度看板字段**：

| 状态 | 说明 |
|------|------|
| 待办 (Todo) | 计划中，未开始 |
| 进行中 (In Progress) | 正在开发 |
| 代码评审 (Code Review) | 等待 Review |
| 已完成 (Done) | 验收通过 |
| 阻塞 (Blocked) | 等待外部依赖 |

### 质量门禁

**每个里程碑必须通过**：

```bash
# 1. 单元测试覆盖率 ≥ 80%
pytest --cov=backend/app --cov-report=term

# 2. 代码风格检查通过
flake8 backend/app --max-line-length=120 --ignore=E501,W503

# 3. 类型检查通过
mypy backend/app --ignore-missing-imports

# 4. 安全扫描通过
bandit -r backend/app
```

---

## ✅ 验收标准

### 功能验收
- [ ] 前后端项目合并完成，可统一启动
- [ ] 后台可管理渠道（CRUD）
- [ ] 后台可管理令牌（CRUD）
- [ ] 用户可通过 OpenAI 兼容接口调用 API
- [ ] 自动负载均衡和失败重试
- [ ] 配额计算准确

### 性能验收
- [ ] API 响应时间 < 500ms（不含上游延迟）
- [ ] 支持并发 100+ 请求
- [ ] 流式响应无明显延迟

### 安全验收
- [ ] Token 验证完整
- [ ] 配额控制有效
- [ ] IP 限制可选
- [ ] API Key 不暴露给前端
- [ ] API Key 加密存储
- [ ] 审计日志完整记录

### 运维验收
- [ ] 健康检测定时任务运行
- [ ] 批量导入渠道可用
- [ ] 审计日志可查询/导出
- [ ] 使用统计面板可用

---

## 🔐 安全设计（重要补充）

### API Key 加密存储

```python
# backend/app/utils/crypto.py
from cryptography.fernet import Fernet
from flask import current_app

class APIKeyCrypto:
    """API Key 加密工具"""
    
    @staticmethod
    def get_cipher():
        """获取加密器"""
        key = current_app.config.get('API_KEY_ENCRYPTION_KEY')
        if not key:
            # 生成并保存新密钥（仅首次）
            key = Fernet.generate_key()
            # 生产环境应从环境变量或密钥管理服务读取
        return Fernet(key if isinstance(key, bytes) else key.encode())
    
    @classmethod
    def encrypt(cls, plaintext: str) -> str:
        """加密 API Key"""
        cipher = cls.get_cipher()
        return cipher.encrypt(plaintext.encode()).decode()
    
    @classmethod
    def decrypt(cls, ciphertext: str) -> str:
        """解密 API Key"""
        cipher = cls.get_cipher()
        return cipher.decrypt(ciphertext.encode()).decode()

# 使用示例
def create_channel(data):
    channel = Channel(
        api_key=APIKeyCrypto.encrypt(data['api_key']),
        # ...
    )

def use_channel(channel):
    api_key = APIKeyCrypto.decrypt(channel.api_key)
    # 发起请求
```

### Token 生成策略

```python
import secrets
import string

def generate_token_key() -> str:
    """生成 sk-xxx 格式的 Token"""
    alphabet = string.ascii_letters + string.digits
    random_part = ''.join(secrets.choice(alphabet) for _ in range(48))
    return f"sk-{random_part}"

def validate_token(token: str) -> bool:
    """验证 Token 格式"""
    if not token.startswith('sk-'):
        return False
    if len(token) != 51:  # sk- + 48 chars
        return False
    return True
```

---

## 📝 任务管理表

### 任务分解（第一阶段）

| ID | 任务 | 负责 | 状态 | 预计 | 实际 | 备注 |
|----|------|------|------|------|------|------|
| T1 | 项目目录合并 | - | 待办 | 1d | - | - |
| T2 | 数据库迁移脚本 | - | 待办 | 0.5d | - | - |
| T3 | 渠道模型与服务 | - | 待办 | 1d | - | - |
| T4 | 健康检测服务 | - | 待办 | 1d | - | - |
| T5 | Token 模型与服务 | - | 待办 | 0.5d | - | - |
| T6 | 审计日志服务 | - | 待办 | 1d | - | - |
| T7 | API 代理接口 | - | 待办 | 1.5d | - | - |
| T8 | 前端渠道管理 | - | 待办 | 1d | - | - |
| T9 | 前端 Token 管理 | - | 待办 | 0.5d | - | - |
| T10 | 集成测试 | - | 待办 | 1d | - | - |

### 任务优先级规则

| 优先级 | 标识 | 说明 |
|--------|------|------|
| P0 | 🔴 阻塞 | 必须立即解决，否则后续无法进行 |
| P1 | 🟠 高 | 当前迭代必须完成 |
| P2 | 🟡 中 | 计划中，可延后 |
| P3 | 🟢 低 | 优化项，可跳过 |

---

## 🔗 参考资料

| 资源 | 链接 |
|------|------|
| One API 仓库 | https://github.com/songquanpeng/one-api |
| OpenAI API 文档 | https://platform.openai.com/docs/api-reference |
| Claude API 文档 | https://docs.anthropic.com/claude/reference |
| Flask 文档 | https://flask.palletsprojects.com/ |
| React Router | https://reactrouter.com/ |
