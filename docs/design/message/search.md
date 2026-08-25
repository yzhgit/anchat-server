# HTTP 消息搜索设计

## 1. 路由设计

- `GET /api/v1/messages/search?keyword=...&conversation_id=...&content_type=...&limit=...&offset=...`
- gRPC: `MessageService.SearchMessages`

## 2. 参数约束

- `keyword` 必填；
- `conversation_id` 必填（当前实现仅支持会话内搜索）；
- `content_type` 选填（`1..7`，如 `1=text`, `2=image`, `5=file`）；
- `limit` 默认 20，最大 100；
- `offset` 默认 0。

## 3. 时序

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant ConversationService
    participant MessageService
    participant DB

    Client->>Gateway: GET /api/v1/messages/search?keyword=...&conversation_id=...
    Gateway->>ConversationService: GetConversation(user_id, conversation_id)
    ConversationService-->>Gateway: 鉴权通过
    Gateway->>MessageService: SearchMessages(keyword, conversation_id, ...)
    MessageService->>DB: 分页搜索
    DB-->>MessageService: messages,total
    MessageService-->>Gateway: SearchMessagesResponse
    Gateway-->>Client: 200 OK
```

## 4. 说明

- 该接口用于“历史消息检索”，不替代实时同步链路。
- 返回结果按存储层排序规则输出，客户端可按 `sequence` 做二次排序展示。
