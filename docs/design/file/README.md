# File Service (文件服务)

## 1. 服务概述

**职责**: 文件上传、下载、管理（基于MinIO）

**核心功能**:
- 文件上传（头像、聊天图片/视频/语音/文件）
- 文件下载（带签名URL）
- 文件管理（元信息、去重、清理）
- 图片处理（压缩、缩略图）
- 视频处理（压缩、封面）
- 存储桶管理

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 文件上传 | [upload.md](upload.md) | 上传、下载、断点续传 |

## 3. 数据模型

- **File**: 文件元信息
- **FileUpload**: 上传记录

## 4. 推送通知

- `notification.file.upload_completed.{user_id}` - 文件上传完成通知
- `notification.file.processing.{user_id}` - 文件处理进度通知
- `notification.file.expiring.{user_id}` - 文件过期提醒

## 5. 依赖服务

- **MinIO**: 对象存储
- **Redis**: 上传进度缓存
- **PostgreSQL**: 文件元信息

---

返回: [后端总体设计](../backend-design.md)
