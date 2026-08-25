# 文件上传下载设计

## 1. 概述

文件服务负责文件上传、下载、管理，基于MinIO对象存储。

## 2. 功能列表

- [x] 生成上传凭证
- [x] 完成上传确认
- [x] 生成下载链接
- [x] 获取文件信息
- [x] 删除文件
- [x] 批量获取文件信息

## 3. 业务流程

### 3.1 文件上传

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant FileService
    participant MinIO
    participant DB

    Client->>Gateway: POST /file/upload/token<br/>Header: Authorization: Bearer {token}<br/>Body: {file_type, file_name, file_size}
    Gateway->>FileService: gRPC GenerateUploadToken(userId, fileType, fileName, fileSize)
    FileService->>FileService: 生成FileID(UUID)
    FileService->>MinIO: 生成上传凭证(Presigned URL)
    MinIO-->>FileService: 上传URL
    FileService->>DB: 创建文件记录(初始状态pending)
    FileService-->>Gateway: 返回fileId + uploadUrl + fileUrl
    Gateway-->>Client: 200 OK

    Client->>MinIO: 上传文件(直接PUT到Presigned URL)
    MinIO-->>Client: 上传成功

    Client->>Gateway: POST /file/upload/complete<br/>Header: Authorization: Bearer {token}<br/>Body: {file_id}
    Gateway->>FileService: gRPC CompleteUpload(fileId, userId)
    FileService->>MinIO: 验证文件存在
    FileService->>DB: 更新文件状态为已完成
    FileService-->>Gateway: 返回文件信息
    Gateway-->>Client: 200 OK
```

### 3.2 文件下载

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant FileService
    participant MinIO

    Client->>Gateway: GET /file/download/{fileId}?expires=60<br/>Header: Authorization: Bearer {token}
    Gateway->>FileService: gRPC GenerateDownloadURL(fileId, userId, expiresMinutes)
    FileService->>MinIO: 生成下载链接(Presigned URL)
    MinIO-->>FileService: 下载URL
    FileService-->>Gateway: 返回downloadUrl
    Gateway-->>Client: 200 OK(redirect或直接返回URL)
```

### 3.3 删除文件

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant FileService
    participant MinIO
    participant DB

    Client->>Gateway: DELETE /file/{fileId}<br/>Header: Authorization: Bearer {token}
    Gateway->>FileService: gRPC DeleteFile(fileId, userId)
    FileService->>DB: 检查文件归属
    FileService->>MinIO: 删除对象
    FileService->>DB: 删除文件记录
    FileService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

## 4. API设计

### 4.1 生成上传凭证

```protobuf
message GenerateUploadTokenRequest {
    string user_id = 1;
    int32 file_type = 2; // 1-image/2-video/3-audio/4-file/5-log
    string file_name = 3;
    int64 file_size = 4;
}

message GenerateUploadTokenResponse {
    string file_id = 1;
    string upload_url = 2;
    string file_url = 3;
}
```

### 4.2 生成下载链接

```protobuf
message GenerateDownloadURLRequest {
    string file_id = 1;
    string user_id = 2;
    int32 expires_minutes = 3;
}

message GenerateDownloadURLResponse {
    string download_url = 1;
}
```

## 5. 存储桶设计

| 存储桶 | 用途 | 权限 |
|--------|------|------|
| avatars | 用户头像 | 公开读 |
| group-avatars | 群头像 | 公开读 |
| chat-images | 聊天图片 | 私有 |
| chat-videos | 聊天视频 | 私有 |
| chat-voices | 聊天语音 | 私有 |
| chat-files | 聊天文件 | 私有 |

## 6. 依赖服务

- **MinIO**: 对象存储
- **PostgreSQL**: 文件元信息
