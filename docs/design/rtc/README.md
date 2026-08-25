# RTC Service (Real-Time Communication 音视频服务)

## 1. 服务概述

**职责**: 音视频通话、视频会议（基于LiveKit）

**核心功能**:
- 一对一音视频（发起、接听、拒绝、挂断）
- 群组音视频（多人通话、邀请）
- 视频会议（创建、加入、预约）
- 会议功能（屏幕共享、全员静音、录制）

## 2. 文档导航

| 功能 | 文档 | 说明 |
|------|------|------|
| 音视频通话 | [call.md](call.md) | 一对一/群组通话 |
| 视频会议 | [meeting.md](meeting.md) | 会议创建与管理 |

## 3. 数据模型

- **CallSession**: 通话会话
- **CallLog**: 通话记录
- **MeetingRoom**: 会议室
- **MeetingParticipant**: 会议参与者
- **MeetingRecording**: 会议录制

## 4. 推送通知

- `notification.livekit.call_invite.{user_id}` - 音视频通话邀请
- `notification.livekit.call_status.{call_id}` - 通话状态变更
- `notification.livekit.call_rejected.{from_user_id}` - 通话拒绝

## 5. 依赖服务

- **User Service**: 用户信息
- **Group Service**: 群信息
- **Message Service**: 通话消息记录
- **Push Service**: 通话推送
- **PostgreSQL**: 通话记录

---

返回: [后端总体设计](../backend-design.md)
