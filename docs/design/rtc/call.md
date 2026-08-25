# 音视频通话设计

## 1. 概述

RTC Service 基于 LiveKit 实现一对一音视频通话功能。

## 2. 功能列表

- [x] 发起通话
- [x] 接听/拒绝通话
- [x] 结束通话
- [x] 通话记录查询

## 3. 业务流程

### 3.1 发起通话

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant RtcService
    participant LiveKit
    participant NATS

    Client->>Gateway: POST /call/initiate<br/>Header: Authorization: Bearer {token}<br/>Body: {callee_id, call_type}
    Gateway->>Gateway: 从JWT解析callerId
    Gateway->>RtcService: gRPC InitiateCall(callerId, calleeId, callType)
    RtcService->>RtcService: 生成CallID(UUID)
    RtcService->>LiveKit: 创建Room
    LiveKit-->>RtcService: Room创建成功
    RtcService->>RtcService: 生成Token
    RtcService->>NATS: 发布通话邀请通知
    RtcService-->>Gateway: 返回CallID + Token
    Gateway-->>Client: 200 OK
```

### 3.2 接听通话

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant RtcService
    participant LiveKit
    participant NATS

    Client->>Gateway: POST /call/join<br/>Header: Authorization: Bearer {token}<br/>Body: {call_id}
    Gateway->>Gateway: 从JWT解析userId
    Gateway->>RtcService: gRPC JoinCall(callId, userId)
    RtcService->>LiveKit: 生成JoinToken
    LiveKit-->>RtcService: Token
    RtcService->>NATS: 发布通话开始通知
    RtcService-->>Gateway: 返回Token + Room信息
    Gateway-->>Client: 200 OK
```

### 3.3 结束通话

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant RtcService
    participant LiveKit
    participant NATS

    Client->>Gateway: POST /call/end<br/>Header: Authorization: Bearer {token}<br/>Body: {call_id}
    Gateway->>RtcService: gRPC EndCall(callId)
    RtcService->>LiveKit: 关闭Room
    RtcService->>NATS: 发布通话结束通知
    RtcService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

### 3.4 拒绝通话

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant RtcService
    participant NATS

    Client->>Gateway: POST /call/reject<br/>Header: Authorization: Bearer {token}<br/>Body: {call_id}
    Gateway->>RtcService: gRPC RejectCall(callId)
    RtcService->>NATS: 发布通话拒绝通知
    RtcService-->>Gateway: 成功
    Gateway-->>Client: 200 OK
```

## 4. API设计

### 4.1 发起通话

```protobuf
message InitiateCallRequest {
    string caller_id = 1;
    string callee_id = 2;
    CallType call_type = 3; // 0-audio/1-video
}

message InitiateCallResponse {
    string call_id = 1;
    string token = 2;
    int64 created_at = 3;
}
```

### 4.2 接听通话

```protobuf
message JoinCallRequest {
    string call_id = 1;
    string user_id = 2;
}

message JoinCallResponse {
    string token = 1;
    string room_name = 2;
}
```

## 5. 通话类型

| 类型 | 说明 |
|------|------|
| audio | 语音通话 |
| video | 视频通话 |

## 6. 通知主题

- `notification.livekit.call_invite.{user_id}` - 通话邀请
- `notification.livekit.call_status.{call_id}` - 通话状态变更
- `notification.livekit.call_rejected.{user_id}` - 通话拒绝
