# OpenList Workers 文件系统功能

基于用户驱动配置的云存储文件系统管理 API，支持多种云存储服务的统一文件操作。

## 概述

OpenList Workers 现在支持基于用户驱动配置的文件系统操作，用户可以：
- 管理个人的驱动配置
- 通过统一的 API 操作不同的云存储服务
- 进行文件和目录的增删改查操作
- 支持文件上传下载功能

## 架构特点

### 用户隔离
- 每个用户有独立的驱动配置
- 配置与用户ID绑定，确保数据安全
- 支持多租户架构

### 驱动支持
- 本地存储 (Local)
- Amazon S3
- 阿里云盘 (Aliyundrive)
- OneDrive
- Google Drive
- 更多驱动持续添加...

### 接口设计
- RESTful API 设计
- 统一的错误处理
- 支持分页查询
- 类型安全的接口检查

## API 文档

### 基础信息

**Base URL**: `http://localhost:8787` (开发环境)

**必需参数**:
- `user_id`: 用户ID
- `config_id`: 驱动配置ID

### 文件系统操作

#### 1. 列出文件和目录
```http
GET /api/fs/list?user_id={user_id}&config_id={config_id}&path={path}&page={page}&per_page={per_page}
```

**参数**:
- `path`: 目录路径 (默认: "/")
- `page`: 页码 (默认: 1)
- `per_page`: 每页条数 (默认: 20)

**响应**:
```json
{
  "code": 200,
  "data": {
    "files": [
      {
        "name": "test.txt",
        "size": 1024,
        "is_dir": false,
        "modified": "2024-01-01T00:00:00Z"
      }
    ],
    "path": "/",
    "user_id": 1,
    "config_id": 1
  }
}
```

#### 2. 获取文件信息
```http
GET /api/fs/get?user_id={user_id}&config_id={config_id}&path={path}
```

#### 3. 列出目录（仅目录）
```http
GET /api/fs/dirs?user_id={user_id}&config_id={config_id}&path={path}
```

#### 4. 创建目录
```http
POST /api/fs/mkdir
Content-Type: application/x-www-form-urlencoded

user_id={user_id}&config_id={config_id}&path={parent_path}&dir_name={dir_name}
```

#### 5. 重命名文件/目录
```http
POST /api/fs/rename
Content-Type: application/x-www-form-urlencoded

user_id={user_id}&config_id={config_id}&path={file_path}&new_name={new_name}
```

#### 6. 移动文件/目录
```http
POST /api/fs/move
Content-Type: application/x-www-form-urlencoded

user_id={user_id}&config_id={config_id}&path={src_path}&dst_path={dst_path}
```

#### 7. 复制文件/目录
```http
POST /api/fs/copy
Content-Type: application/x-www-form-urlencoded

user_id={user_id}&config_id={config_id}&path={src_path}&dst_path={dst_path}
```

#### 8. 删除文件/目录
```http
POST /api/fs/remove
Content-Type: application/x-www-form-urlencoded

user_id={user_id}&config_id={config_id}&path={file_path}
```

#### 9. 上传文件
```http
PUT /api/fs/upload?user_id={user_id}&config_id={config_id}&path={dir_path}&filename={filename}
Content-Type: application/octet-stream

[文件内容]
```

#### 10. 下载文件
```http
GET /d/?user_id={user_id}&config_id={config_id}&path={file_path}
```
返回重定向到实际文件下载链接。

### 驱动配置管理

#### 获取用户驱动列表
```http
GET /api/drivers?user_id={user_id}&enabled=true
```

#### 创建驱动配置
```http
POST /api/user/driver/create
Content-Type: application/json

{
  "name": "MyS3",
  "display_name": "我的 S3 存储",
  "description": "私人 S3 存储配置",
  "config": "{\"bucket\":\"my-bucket\",\"region\":\"us-east-1\"...}",
  "enabled": true,
  "order": 1
}
```

#### 更新驱动配置
```http
POST /api/user/driver/update
Content-Type: application/json

{
  "id": 1,
  "name": "MyS3",
  "config": "{\"bucket\":\"new-bucket\"...}",
  "enabled": true
}
```

#### 删除驱动配置
```http
POST /api/user/driver/delete?id={config_id}
```

## 错误处理

### 常见错误码

- `400`: 请求参数错误
- `404`: 文件/驱动不存在
- `501`: 驱动不支持该操作
- `500`: 服务器内部错误

### 错误响应格式
```json
{
  "code": 501,
  "message": "Driver does not support copy operation"
}
```

## 驱动兼容性

| 驱动类型 | 列表 | 创建目录 | 重命名 | 移动 | 复制 | 删除 | 上传 | 下载 |
|---------|------|----------|--------|------|------|------|------|------|
| Local   | ✅   | ✅       | ✅     | ✅   | ✅   | ✅   | ✅   | ✅   |
| S3      | ✅   | ✅       | ❌     | ✅   | ✅   | ✅   | ✅   | ✅   |
| 阿里云盘 | ✅   | ✅       | ✅     | ✅   | ✅   | ✅   | ✅   | ✅   |
| OneDrive| ✅   | ❌       | ✅     | ❌   | ✅   | ✅   | ✅   | ✅   |
| Google Drive| ✅| ✅       | ✅     | ✅   | ❌   | ✅   | ✅   | ✅   |

*注：❌ 表示驱动本身不支持该操作，API 会返回 501 错误*

## 使用示例

### JavaScript/Node.js
```javascript
const axios = require('axios');

class OpenListClient {
  constructor(baseURL, userId) {
    this.baseURL = baseURL;
    this.userId = userId;
  }

  async listFiles(configId, path = '/') {
    const response = await axios.get(`${this.baseURL}/api/fs/list`, {
      params: { user_id: this.userId, config_id: configId, path }
    });
    return response.data;
  }

  async uploadFile(configId, path, filename, fileData) {
    const response = await axios.put(`${this.baseURL}/api/fs/upload`, fileData, {
      params: { user_id: this.userId, config_id: configId, path, filename },
      headers: { 'Content-Type': 'application/octet-stream' }
    });
    return response.data;
  }

  async createFolder(configId, parentPath, dirName) {
    const response = await axios.post(`${this.baseURL}/api/fs/mkdir`,
      `user_id=${this.userId}&config_id=${configId}&path=${encodeURIComponent(parentPath)}&dir_name=${encodeURIComponent(dirName)}`,
      { headers: { 'Content-Type': 'application/x-www-form-urlencoded' }}
    );
    return response.data;
  }
}

// 使用示例
const client = new OpenListClient('http://localhost:8787', 1);

(async () => {
  // 列出文件
  const files = await client.listFiles(1, '/');
  console.log('文件列表:', files);

  // 创建目录
  await client.createFolder(1, '/', 'new_folder');

  // 上传文件
  const fileBuffer = Buffer.from('Hello World');
  await client.uploadFile(1, '/new_folder', 'hello.txt', fileBuffer);
})();
```

### Python
```python
import requests
import json

class OpenListClient:
    def __init__(self, base_url, user_id):
        self.base_url = base_url
        self.user_id = user_id

    def list_files(self, config_id, path='/'):
        response = requests.get(f'{self.base_url}/api/fs/list', params={
            'user_id': self.user_id,
            'config_id': config_id,
            'path': path
        })
        return response.json()

    def upload_file(self, config_id, path, filename, file_data):
        response = requests.put(f'{self.base_url}/api/fs/upload',
                               data=file_data,
                               params={
                                   'user_id': self.user_id,
                                   'config_id': config_id,
                                   'path': path,
                                   'filename': filename
                               },
                               headers={'Content-Type': 'application/octet-stream'})
        return response.json()

# 使用示例
client = OpenListClient('http://localhost:8787', 1)

# 列出文件
files = client.list_files(1, '/')
print('文件列表:', json.dumps(files, indent=2))

# 上传文件
with open('test.txt', 'rb') as f:
    result = client.upload_file(1, '/', 'test.txt', f.read())
    print('上传结果:', result)
```

## 开发和测试

### 运行测试
```bash
# 给测试脚本执行权限
chmod +x test_filesystem_api.sh

# 运行完整的文件系统测试
./test_filesystem_api.sh
```

### 本地开发
```bash
# 启动开发服务器
wrangler dev

# 或使用 go run（如果配置了开发环境）
go run main.go d1_database_dev.go
```

## 部署

### Cloudflare Workers 部署
```bash
# 部署到 Cloudflare Workers
wrangler deploy

# 配置 D1 数据库
wrangler d1 create openlist-db
wrangler d1 execute openlist-db --file=schema.sql
```

### 环境变量配置
```toml
# wrangler.toml
[[d1_databases]]
binding = "DB"
database_name = "openlist-db"
database_id = "your-database-id"
```

## 安全说明

- 所有操作都需要用户ID和配置ID验证
- 用户只能访问自己的驱动配置
- 文件操作都在指定的驱动配置范围内
- 建议在生产环境中添加适当的认证和授权机制

## 常见问题

### Q: 如何添加新的驱动类型？
A: 在 `initDefaultData()` 函数中添加新的驱动配置，确保驱动名称与 OpenList 支持的驱动名称匹配。

### Q: 某些操作返回 501 错误？
A: 这表示所选的驱动不支持该操作。请查看驱动兼容性表格选择支持的操作。

### Q: 如何配置驱动的具体参数？
A: 在驱动配置的 `config` 字段中提供 JSON 格式的配置参数，具体参数格式请参考对应驱动的文档。

### Q: 文件上传大小限制？
A: 具体限制取决于所使用的驱动和 Cloudflare Workers 的限制。建议大文件使用分片上传。

## 更新日志

- **v1.0.0**: 初始版本，支持基本文件系统操作
- **v1.0.1**: 添加驱动兼容性检查
- **v1.0.2**: 优化错误处理和类型安全

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目！