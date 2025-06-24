# OpenList Workers API 自动化测试指南

本文档介绍如何使用 OpenList Workers 项目的完整 API 自动化测试套件。

## 测试概览

本项目包含以下测试脚本：

### 1. 综合功能测试 (`test_api_comprehensive.sh`)
全面测试所有 API 功能，包括：
- 系统健康检查和初始化
- 用户认证（注册、登录、登出）
- 用户管理（管理员功能）
- 驱动配置管理
- 文件系统操作
- 离线下载功能
- 错误处理
- 数据一致性验证

### 2. 性能测试 (`test_performance.sh`)
测试 API 的性能指标：
- 响应时间统计
- 并发性能测试
- 压力测试
- 吞吐量分析
- 性能百分位数计算

### 3. 集成测试 (`test_integration.sh`)
模拟真实用户使用场景：
- 新用户完整使用流程
- 管理员操作流程
- 端到端业务场景验证

### 4. 认证测试 (`test_auth_api.sh`)
专门测试认证相关功能：
- JWT Token 生成和验证
- 用户权限验证
- 认证状态管理

### 5. 测试套件运行器 (`run_all_tests.sh`)
统一运行所有测试的主脚本：
- 自动运行所有测试类型
- 生成 HTML 测试报告
- 提供测试结果汇总

## 使用方法

### 前提条件

1. **确保服务器运行**
   ```bash
   # 启动 OpenList Workers 服务器（假设在 localhost:8787）
   wrangler dev --port 8787
   ```

2. **安装依赖**
   - `curl` - 用于发送 HTTP 请求
   - `bash` - 脚本运行环境
   - `python3` - 用于时间戳计算（macOS 性能测试）

### 快速开始

#### 运行所有测试
```bash
./run_all_tests.sh
```

#### 运行特定类型测试
```bash
# 只运行基础功能测试
./run_all_tests.sh basic

# 只运行性能测试
./run_all_tests.sh performance

# 只运行集成测试
./run_all_tests.sh integration

# 只运行认证测试
./run_all_tests.sh auth
```

#### 使用自定义 URL
```bash
# 测试其他环境
./run_all_tests.sh -u https://api.example.com

# 测试本地其他端口
./run_all_tests.sh -u http://localhost:3000
```

### 单独运行测试脚本

#### 1. 综合功能测试
```bash
# 运行所有功能测试
./test_api_comprehensive.sh

# 使用自定义 URL
./test_api_comprehensive.sh -u http://localhost:3000

# 运行特定模块测试
./test_api_comprehensive.sh auth          # 只测试认证
./test_api_comprehensive.sh user          # 只测试用户管理
./test_api_comprehensive.sh driver        # 只测试驱动配置
./test_api_comprehensive.sh filesystem    # 只测试文件系统
./test_api_comprehensive.sh offline       # 只测试离线下载
./test_api_comprehensive.sh error         # 只测试错误处理
```

#### 2. 性能测试
```bash
# 默认性能测试（10个并发用户，每用户5个请求）
./test_performance.sh

# 自定义并发数和请求数
./test_performance.sh -c 20 -r 10

# 运行特定性能测试
./test_performance.sh basic               # 基础性能测试
./test_performance.sh auth                # 认证性能测试
./test_performance.sh api                 # API性能测试
./test_performance.sh stress              # 压力测试
./test_performance.sh error               # 错误处理性能

# 设置超时时间
./test_performance.sh -t 60               # 60秒超时
```

#### 3. 集成测试
```bash
# 运行集成测试
./test_integration.sh

# 使用自定义 URL
./test_integration.sh -u https://api.example.com
```

#### 4. 认证测试
```bash
# 运行认证测试
./test_auth_api.sh http://localhost:8787
```

### 测试报告

运行 `run_all_tests.sh` 后，测试报告会保存在 `test_reports/` 目录中：

- `test_summary_[timestamp].html` - HTML 格式的汇总报告
- `[test_name]_[timestamp].log` - 各个测试的详细日志

#### 查看报告
```bash
# 打开最新的 HTML 报告
open test_reports/test_summary_*.html

# 查看特定测试的日志
cat test_reports/comprehensive_api_tests_*.log
```

## 测试配置

### 环境变量
可以通过环境变量配置测试参数：

```bash
export BASE_URL="http://localhost:8787"
export TEST_TIMEOUT="30"
export CONCURRENT_USERS="10"
export REQUESTS_PER_USER="5"
```

### 自定义配置
编辑测试脚本顶部的配置变量：

```bash
# test_api_comprehensive.sh
BASE_URL="http://localhost:8787"

# test_performance.sh
CONCURRENT_USERS=10
REQUESTS_PER_USER=5
TIMEOUT=30
```

## 测试场景说明

### 综合功能测试场景

1. **系统健康检查**
   - 测试 `/health` 端点
   - 验证系统初始化 `/init`

2. **用户认证流程**
   - 用户注册
   - 用户登录
   - 获取当前用户信息
   - 用户登出

3. **用户管理（管理员）**
   - 获取用户列表
   - 创建新用户
   - 更新用户信息
   - 获取单个用户详情

4. **驱动配置管理**
   - 查看可用驱动
   - 创建驱动配置
   - 更新驱动配置
   - 启用/禁用驱动
   - 删除驱动配置

5. **文件系统操作**
   - 列出文件和目录
   - 创建目录
   - 重命名文件/目录
   - 移动文件/目录
   - 复制文件/目录
   - 删除文件/目录
   - 文件上传

6. **离线下载功能**
   - 查看支持的下载工具
   - 配置下载工具（Aria2、qBittorrent等）
   - 创建下载任务
   - 查看任务状态
   - 更新任务状态

7. **错误处理验证**
   - 未认证访问处理
   - 无效Token处理
   - 权限不足处理
   - 资源不存在处理
   - 方法不允许处理

### 性能测试场景

1. **基础性能测试**
   - 健康检查响应时间
   - 系统初始化性能

2. **认证性能测试**
   - 登录请求性能
   - Token验证性能

3. **API性能测试**
   - 用户信息查询性能
   - 驱动列表查询性能
   - 文件列表查询性能

4. **压力测试**
   - 高并发请求处理
   - 系统稳定性验证

5. **错误处理性能**
   - 错误响应时间
   - 异常情况处理效率

### 集成测试场景

1. **新用户完整流程**
   - 系统初始化 → 用户注册 → 配置驱动 → 文件操作

2. **管理员操作流程**
   - 管理员登录 → 用户管理 → 系统配置

## 故障排除

### 常见问题

1. **服务器连接失败**
   ```
   Error: Cannot connect to server at http://localhost:8787
   ```
   **解决方案**：确保 OpenList Workers 服务器正在运行
   ```bash
   wrangler dev --port 8787
   ```

2. **权限错误**
   ```
   bash: ./test_*.sh: Permission denied
   ```
   **解决方案**：添加执行权限
   ```bash
   chmod +x *.sh
   ```

3. **认证失败**
   ```
   Error: Authentication required
   ```
   **解决方案**：检查默认管理员账户是否可用（admin/admin123）

4. **测试超时**
   ```
   Error: Request timeout
   ```
   **解决方案**：增加超时时间或检查服务器性能
   ```bash
   ./test_performance.sh -t 60  # 增加到60秒
   ```

### 调试模式

启用详细日志输出：
```bash
# 添加 -x 参数查看详细执行过程
bash -x ./test_api_comprehensive.sh

# 查看 curl 请求详情
export CURL_VERBOSE=1
./test_api_comprehensive.sh
```

### 测试数据清理

测试可能会创建一些测试数据，如需清理：
```bash
# 清理测试报告
rm -rf test_reports/

# 清理临时文件
rm -rf /tmp/openlist_test/
```

## 持续集成

### GitHub Actions 示例

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Setup Node.js
        uses: actions/setup-node@v2
        with:
          node-version: '18'

      - name: Install Wrangler
        run: npm install -g wrangler

      - name: Start server
        run: wrangler dev --port 8787 &

      - name: Wait for server
        run: sleep 10

      - name: Run tests
        run: ./run_all_tests.sh

      - name: Upload test reports
        uses: actions/upload-artifact@v2
        with:
          name: test-reports
          path: test_reports/
```

## 扩展测试

### 添加新的测试用例

1. **在现有脚本中添加**：
   ```bash
   # 在 test_api_comprehensive.sh 中添加新的测试函数
   test_new_feature() {
       log_info "Testing new feature..."
       # 测试逻辑
   }
   ```

2. **创建新的测试脚本**：
   ```bash
   cp test_template.sh test_new_feature.sh
   # 修改测试内容
   ```

3. **集成到主运行器**：
   ```bash
   # 在 run_all_tests.sh 中添加新的测试类型
   ```

### 自定义断言函数

```bash
# 添加到测试脚本中
assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="$3"

    if [[ "$expected" == "$actual" ]]; then
        log_success "$message"
    else
        log_error "$message: expected '$expected', got '$actual'"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"

    if echo "$haystack" | grep -q "$needle"; then
        log_success "$message"
    else
        log_error "$message: '$haystack' does not contain '$needle'"
        return 1
    fi
}
```

## 最佳实践

1. **定期运行测试**：建议在每次代码变更后运行完整测试套件
2. **环境隔离**：为不同环境（开发、测试、生产）使用不同的测试配置
3. **测试数据管理**：使用专门的测试数据，避免影响生产数据
4. **性能基线**：记录性能测试基线，监控性能退化
5. **测试覆盖率**：定期检查测试覆盖率，补充缺失的测试用例

---

更多信息请参考：
- [OpenList Workers 项目文档](README.md)
- [API 文档](README_D1_Complete.md)
- [部署指南](wrangler.toml)