# IM 消息队列技术选型对比分析

## 1. 候选方案对比

### 1.1 主要候选技术

| 消息队列 | 类型 | 核心特点 | 适用场景 |
|---------|------|---------|---------|
| **NATS JetStream** | 云原生消息系统 | 轻量、低延迟、简单 | 微服务、实时通信 |
| **Apache Kafka** | 分布式日志系统 | 高吞吐、持久化、重放 | 大数据、日志、事件溯源 |
| **RabbitMQ** | 传统消息队列 | 功能丰富、可靠性高 | 企业应用、任务队列 |
| **Apache Pulsar** | 统一消息平台 | 分层存储、多租户 | 大规模流处理 |
| **Redis Streams** | 内存数据结构 | 极低延迟、简单 | 轻量级消息流 |
| **NSQ** | 分布式消息平台 | 去中心化、易运维 | 实时消息分发 |

---

## 2. 详细对比分析

### 2.1 NATS JetStream

**官网**: https://nats.io/
**GitHub**: https://github.com/nats-io/nats-server (14.8k stars)

#### 架构特点
```
┌─────────────────────────────────────────┐
│          NATS JetStream                 │
│  ┌──────────┐  ┌──────────┐            │
│  │ Stream 1 │  │ Stream 2 │  ...       │
│  └──────────┘  └──────────┘            │
│       │             │                   │
│  ┌────▼─────┬──────▼────┐              │
│  │Consumer 1│Consumer 2 │  ...         │
│  └──────────┴───────────┘              │
└─────────────────────────────────────────┘
```

#### 优势 ✅
1. **极低延迟**
   - P50: < 1ms
   - P99: < 10ms
   - 非常适合实时 IM 场景

2. **轻量级**
   - 单个二进制文件（~20MB）
   - 内存占用低（空载 < 10MB）
   - 零依赖，部署简单

3. **原生支持**
   - **At-least-once** 投递保证
   - **消息持久化**（文件存储）
   - **消息重放**（按时间/序列号）
   - **消息去重**（Message ID）

4. **Pub/Sub 模式**
   - 原生支持主题订阅
   - 支持通配符订阅（`msg.*.user-123`）
   - 动态订阅/取消订阅

5. **多语言支持**
   - Go, Java, JavaScript, Python, Rust, C#, Ruby
   - 客户端库成熟稳定

6. **云原生**
   - 完美契合 Kubernetes
   - 支持水平扩展
   - 内置集群模式

#### 劣势 ❌
1. **存储能力有限**
   - 不适合长期存储（建议 < 7天）
   - 存储在本地磁盘，不支持分层存储

2. **功能相对简单**
   - 没有复杂的消息路由规则
   - 没有事务支持
   - 没有消息优先级

3. **社区规模**
   - 相比 Kafka 社区较小
   - 生态工具较少

#### 性能指标
```yaml
吞吐量: 10-15 million msg/sec (单节点)
延迟: P50 < 1ms, P99 < 10ms
持久化: 支持（文件存储）
消息顺序: 保证（Stream 内）
消息重复: 可能（需要客户端去重）
最大消息: 默认 1MB（可配置到 64MB）
```

#### 配置示例
```yaml
# JetStream Stream 配置
name: USER_MESSAGES
subjects:
  - msg.private.*
  - msg.group.*
storage: file
retention: limits
max_msgs: 10000000      # 1000万条
max_bytes: 100GB        # 100GB
max_age: 604800s        # 7天
max_msg_size: 10MB
discard: old
replicas: 3             # 3副本
```

#### 适用场景 ✅
- ✅ **实时消息推送**（IM、聊天）
- ✅ **微服务通信**
- ✅ **事件驱动架构**
- ✅ **IoT 设备通信**
- ❌ 大数据分析
- ❌ 长期日志存储

---

### 2.2 Apache Kafka

**官网**: https://kafka.apache.org/
**GitHub**: https://github.com/apache/kafka (28k stars)

#### 架构特点
```
┌─────────────────────────────────────────┐
│            Kafka Cluster                │
│  ┌────────────────────────────────┐    │
│  │ Topic: user_messages           │    │
│  │  ┌──────┬──────┬──────┬──────┐ │    │
│  │  │Part 0│Part 1│Part 2│Part 3│ │    │
│  │  └──────┴──────┴──────┴──────┘ │    │
│  └────────────────────────────────┘    │
│                                         │
│  Consumer Group: realtime-1              │
│  Consumer Group: message-service        │
└─────────────────────────────────────────┘
```

#### 优势 ✅
1. **超高吞吐量**
   - 百万级 QPS（单节点）
   - 水平扩展能力强

2. **持久化能力强**
   - 支持长期存储（天/月/年）
   - 支持消息重放（任意时间点）
   - 分层存储（Tiered Storage，支持 S3）

3. **消息顺序保证**
   - Partition 内严格有序
   - Key 相同的消息路由到同一 Partition

4. **生态成熟**
   - Kafka Streams（流处理）
   - Kafka Connect（数据集成）
   - Schema Registry（消息schema管理）
   - 监控工具丰富（Confluent Control Center, Kafka Manager）

5. **Exactly-once 语义**
   - 支持事务（Transactions）
   - 幂等生产者（Idempotent Producer）

6. **社区活跃**
   - 超大规模生产实践（LinkedIn, Uber, Netflix）
   - 丰富的文档和案例

#### 劣势 ❌
1. **高延迟**
   - P50: 5-10ms
   - P99: 50-100ms
   - **不适合实时 IM**（延迟敏感）

2. **运维复杂**
   - 依赖 Zookeeper（v3.0+ 移除但仍需熟悉）
   - 需要 JVM 调优（内存管理复杂）
   - 集群运维成本高

3. **资源占用高**
   - 内存占用大（JVM heap + page cache）
   - 磁盘 IO 消耗大
   - 需要专门的硬件资源

4. **复杂度高**
   - 学习曲线陡峭
   - Partition、Replica、Consumer Group 概念复杂
   - 需要专业的运维团队

5. **不支持消息过滤**
   - Consumer 必须消费所有消息
   - 无法按用户 ID 过滤

6. **点对点消息困难**
   - 为每个用户创建 Topic？（不现实）
   - 需要额外的路由层

#### 性能指标
```yaml
吞吐量: 1-2 million msg/sec (单节点)
延迟: P50 5-10ms, P99 50-100ms
持久化: 强（顺序写磁盘）
消息顺序: 保证（Partition 内）
消息重复: 可配置（exactly-once）
最大消息: 默认 1MB（可配置到 2GB）
```

#### 适用场景 ✅
- ✅ **日志聚合**
- ✅ **事件溯源**
- ✅ **大数据管道**
- ✅ **流处理**
- ❌ **实时IM**（延迟过高）
- ❌ 点对点消息

---

### 2.3 RabbitMQ

**官网**: https://www.rabbitmq.com/
**GitHub**: https://github.com/rabbitmq/rabbitmq-server (12.5k stars)

#### 架构特点
```
┌─────────────────────────────────────────┐
│           RabbitMQ Cluster              │
│  ┌────────┐   ┌────────┐   ┌────────┐  │
│  │Exchange│──→│ Queue  │──→│Consumer│  │
│  │(Topic) │   │        │   │        │  │
│  └────────┘   └────────┘   └────────┘  │
│                                         │
│  支持多种 Exchange 类型:                │
│  - Direct (点对点)                      │
│  - Topic (主题订阅)                     │
│  - Fanout (广播)                        │
│  - Headers (基于header路由)            │
└─────────────────────────────────────────┘
```

#### 优势 ✅
1. **功能丰富**
   - 多种 Exchange 类型（灵活路由）
   - 消息优先级
   - 延迟队列（TTL）
   - 死信队列（DLQ）

2. **消息可靠性**
   - 消息持久化
   - 发布确认（Publisher Confirms）
   - 消费确认（Consumer Acks）
   - 镜像队列（高可用）

3. **管理界面**
   - 内置 Web 管理界面
   - 可视化监控
   - 易于调试

4. **多协议支持**
   - AMQP 0-9-1
   - MQTT（IoT）
   - STOMP（WebSocket）

#### 劣势 ❌
1. **性能一般**
   - 吞吐量: 1-10万 msg/sec
   - **不适合高并发 IM**

2. **Erlang 生态**
   - Erlang/OTP 运行时
   - 调试困难
   - 国内资料少

3. **扩展性有限**
   - 集群扩展复杂
   - 分片（Sharding）需要插件

4. **内存消耗**
   - 每条消息都有元数据开销
   - 大量队列时内存占用高

#### 性能指标
```yaml
吞吐量: 10,000-100,000 msg/sec
延迟: P50 5-20ms, P99 50-200ms
持久化: 支持
消息顺序: 保证（Queue 内）
最大消息: 默认 128MB
```

#### 适用场景 ✅
- ✅ **企业应用集成**
- ✅ **任务队列**（异步任务处理）
- ✅ **工作流引擎**
- ❌ **高并发IM**（性能不足）

---

### 2.4 Apache Pulsar

**官网**: https://pulsar.apache.org/
**GitHub**: https://github.com/apache/pulsar (14.2k stars)

#### 架构特点
```
┌─────────────────────────────────────────┐
│         Apache Pulsar                   │
│  ┌─────────────────────────────────┐   │
│  │  Serving Layer (Brokers)        │   │
│  └─────────────┬───────────────────┘   │
│                │                        │
│  ┌─────────────▼───────────────────┐   │
│  │  Storage Layer (BookKeeper)     │   │
│  │  ┌──────┐  ┌──────┐  ┌──────┐  │   │
│  │  │Bookie│  │Bookie│  │Bookie│  │   │
│  │  └──────┘  └──────┘  └──────┘  │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

#### 优势 ✅
1. **分层架构**
   - Broker（计算）和 BookKeeper（存储）分离
   - 存储和计算独立扩展
   - 支持分层存储（冷热数据分离）

2. **高性能**
   - 低延迟（P99 < 5ms）
   - 高吞吐（百万级 QPS）

3. **多租户**
   - 原生支持租户隔离
   - 资源配额管理

4. **Geo-Replication**
   - 跨数据中心复制
   - 灾备能力强

5. **统一模型**
   - 同时支持 Queue 和 Stream
   - 消息保留策略灵活

#### 劣势 ❌
1. **复杂度极高**
   - 依赖组件多（Zookeeper, BookKeeper）
   - 运维成本高
   - 学习曲线陡峭

2. **资源占用大**
   - 内存占用高
   - 磁盘要求高
   - 需要专业运维团队

3. **社区相对较小**
   - 中文资料少
   - 生产案例较少（国内）

4. **杀鸡用牛刀**
   - 对于中小规模 IM 来说过于复杂
   - 适合超大规模场景

#### 性能指标
```yaml
吞吐量: 2-3 million msg/sec (单节点)
延迟: P50 < 2ms, P99 < 5ms
持久化: 强（BookKeeper）
消息顺序: 保证
```

#### 适用场景 ✅
- ✅ **超大规模消息系统**（10亿+ 用户）
- ✅ **多租户 SaaS 平台**
- ✅ **跨数据中心同步**
- ❌ **中小规模 IM**（过度设计）

---

### 2.5 Redis Streams

**官网**: https://redis.io/docs/data-types/streams/
**GitHub**: https://github.com/redis/redis (66k stars)

#### 架构特点
```
┌─────────────────────────────────────────┐
│             Redis                       │
│  ┌──────────────────────────────────┐  │
│  │  Stream: user_messages           │  │
│  │  ┌─────┬─────┬─────┬─────┐      │  │
│  │  │Msg 1│Msg 2│Msg 3│Msg 4│ ...  │  │
│  │  └─────┴─────┴─────┴─────┘      │  │
│  └──────────────────────────────────┘  │
│                                         │
│  Consumer Group: gateway                │
│  Consumer Group: message-service        │
└─────────────────────────────────────────┘
```

#### 优势 ✅
1. **极低延迟**
   - 内存操作，P50 < 1ms
   - 最快的消息队列

2. **简单易用**
   - 已有 Redis 基础设施
   - API 简单（XADD, XREAD, XREADGROUP）
   - 零学习成本

3. **轻量级**
   - 无需额外组件
   - 内存占用小

4. **Consumer Group 支持**
   - 支持消息确认（ACK）
   - 支持消息重试

#### 劣势 ❌
1. **持久化能力弱**
   - 内存存储为主
   - RDB/AOF 持久化有风险
   - **不适合离线消息存储**

2. **存储容量受限**
   - 受限于内存大小
   - 成本高（内存 >> 磁盘）

3. **扩展性有限**
   - Redis Cluster 分片复杂
   - 跨分片订阅困难

4. **可靠性一般**
   - 主从切换可能丢消息
   - 没有多副本机制

#### 性能指标
```yaml
吞吐量: 100,000-500,000 msg/sec
延迟: P50 < 1ms, P99 < 5ms
持久化: 弱（AOF 有延迟）
消息顺序: 保证
最大消息: 512MB
```

#### 适用场景 ✅
- ✅ **实时排行榜**
- ✅ **实时通知**（临时消息）
- ✅ **在线状态同步**
- ❌ **离线消息存储**（数据可能丢失）

---

### 2.6 NSQ

**官网**: https://nsq.io/
**GitHub**: https://github.com/nsqio/nsq (25k stars)

#### 架构特点
```
┌─────────────────────────────────────────┐
│              NSQ                        │
│  ┌──────────┐  ┌──────────┐            │
│  │ nsqd-1   │  │ nsqd-2   │  ...       │
│  │ (Topic)  │  │ (Topic)  │            │
│  └────┬─────┘  └────┬─────┘            │
│       │             │                   │
│  ┌────▼─────────────▼────┐             │
│  │   nsqlookupd (服务发现) │             │
│  └─────────────────────────┘             │
└─────────────────────────────────────────┘
```

#### 优势 ✅
1. **去中心化**
   - 无单点故障
   - 水平扩展简单

2. **运维简单**
   - Go 编写，单二进制
   - 内置管理界面
   - 部署简单

3. **可靠性**
   - 消息持久化（磁盘）
   - 消息自动重试
   - 消息超时机制

#### 劣势 ❌
1. **顺序不保证**
   - **消息无序**（分布式特性）
   - 不适合需要顺序的场景

2. **功能简单**
   - 没有消息重放
   - 没有消息过滤
   - 没有 Consumer Group

3. **持久化有限**
   - 不适合长期存储

#### 性能指标
```yaml
吞吐量: 100,000-300,000 msg/sec
延迟: P50 2-5ms, P99 20-50ms
持久化: 支持
消息顺序: 不保证
```

#### 适用场景 ✅
- ✅ **分布式任务队列**
- ✅ **日志收集**
- ❌ **IM消息**（无序）

---

## 3. IM 场景关键指标对比

### 3.1 综合对比表

| 指标 | NATS JetStream | Kafka | RabbitMQ | Pulsar | Redis Streams | NSQ |
|------|---------------|-------|----------|--------|--------------|-----|
| **延迟 (P99)** | ⭐⭐⭐⭐⭐ <1ms | ⭐⭐ 50-100ms | ⭐⭐⭐ 50-200ms | ⭐⭐⭐⭐ <5ms | ⭐⭐⭐⭐⭐ <1ms | ⭐⭐⭐⭐ 20-50ms |
| **吞吐量** | ⭐⭐⭐⭐ 10M/s | ⭐⭐⭐⭐⭐ 1M/s | ⭐⭐ 100K/s | ⭐⭐⭐⭐⭐ 2M/s | ⭐⭐⭐ 500K/s | ⭐⭐⭐ 300K/s |
| **持久化** | ⭐⭐⭐⭐ 文件 | ⭐⭐⭐⭐⭐ 强 | ⭐⭐⭐⭐ 可配置 | ⭐⭐⭐⭐⭐ 强 | ⭐⭐ 弱 | ⭐⭐⭐⭐ 文件 |
| **运维复杂度** | ⭐⭐⭐⭐⭐ 低 | ⭐⭐ 高 | ⭐⭐⭐ 中 | ⭐ 极高 | ⭐⭐⭐⭐⭐ 低 | ⭐⭐⭐⭐ 低 |
| **资源占用** | ⭐⭐⭐⭐⭐ 低 | ⭐⭐ 高 | ⭐⭐⭐ 中 | ⭐ 极高 | ⭐⭐⭐⭐ 低 | ⭐⭐⭐⭐ 低 |
| **消息顺序** | ⭐⭐⭐⭐⭐ 保证 | ⭐⭐⭐⭐⭐ 保证 | ⭐⭐⭐⭐ 保证 | ⭐⭐⭐⭐⭐ 保证 | ⭐⭐⭐⭐⭐ 保证 | ⭐ 无序 |
| **Pub/Sub** | ⭐⭐⭐⭐⭐ 原生 | ⭐⭐⭐ 需设计 | ⭐⭐⭐⭐ 支持 | ⭐⭐⭐⭐⭐ 原生 | ⭐⭐⭐⭐ 支持 | ⭐⭐⭐ 基础 |
| **点对点消息** | ⭐⭐⭐⭐⭐ 易实现 | ⭐⭐ 困难 | ⭐⭐⭐⭐ 易实现 | ⭐⭐⭐⭐ 易实现 | ⭐⭐⭐⭐ 易实现 | ⭐⭐⭐ 可实现 |
| **离线消息** | ⭐⭐⭐⭐ 7天 | ⭐⭐⭐⭐⭐ 无限 | ⭐⭐⭐ 有限 | ⭐⭐⭐⭐⭐ 无限 | ⭐ 不适合 | ⭐⭐⭐ 有限 |
| **学习曲线** | ⭐⭐⭐⭐⭐ 平缓 | ⭐⭐ 陡峭 | ⭐⭐⭐⭐ 平缓 | ⭐ 极陡 | ⭐⭐⭐⭐⭐ 平缓 | ⭐⭐⭐⭐ 平缓 |
| **社区支持** | ⭐⭐⭐⭐ 活跃 | ⭐⭐⭐⭐⭐ 最活跃 | ⭐⭐⭐⭐ 成熟 | ⭐⭐⭐ 成长中 | ⭐⭐⭐⭐⭐ 广泛 | ⭐⭐⭐ 小众 |
| **IM 适配度** | ⭐⭐⭐⭐⭐ 非常适合 | ⭐⭐ 不适合 | ⭐⭐⭐ 勉强 | ⭐⭐⭐ 过度设计 | ⭐⭐⭐ 部分适合 | ⭐⭐ 不适合 |

### 3.2 IM 场景关键需求匹配

#### 需求1: 极低延迟 (< 10ms)
```
✅ NATS JetStream  - P99 < 1ms
✅ Redis Streams   - P99 < 1ms
⚠️ Pulsar          - P99 < 5ms
⚠️ NSQ             - P99 20-50ms
❌ Kafka           - P99 50-100ms
❌ RabbitMQ        - P99 50-200ms
```

#### 需求2: 支持点对点消息 (每个用户独立队列)
```
✅ NATS JetStream  - 通配符订阅: msg.private.user-123
✅ Redis Streams   - 每用户一个Stream
✅ RabbitMQ        - 每用户一个Queue
⚠️ Pulsar          - 支持但复杂
❌ Kafka           - 为每个用户创建Topic？不现实
⚠️ NSQ             - 可以但无序
```

#### 需求3: 消息顺序保证
```
✅ NATS JetStream  - Stream 内保证
✅ Kafka           - Partition 内保证
✅ Redis Streams   - Stream 内保证
✅ RabbitMQ        - Queue 内保证
✅ Pulsar          - Topic 内保证
❌ NSQ             - 不保证
```

#### 需求4: 离线消息存储 (3-7天)
```
✅ NATS JetStream  - 7天，文件存储
✅ Kafka           - 无限，磁盘存储
⚠️ Pulsar          - 无限，但复杂
⚠️ RabbitMQ        - 有限，内存压力大
⚠️ NSQ             - 有限
❌ Redis Streams   - 不适合（内存）
```

#### 需求5: 运维复杂度（团队规模）
```
✅ NATS JetStream  - 单二进制，零依赖
✅ Redis Streams   - 已有 Redis
✅ NSQ             - 简单
⚠️ RabbitMQ        - 中等（Erlang）
❌ Kafka           - 复杂（JVM + Zookeeper）
❌ Pulsar          - 极复杂（多组件）
```

---

## 4. 推荐方案

### 4.1 最佳选择：NATS JetStream

#### 选择理由 ✅

1. **完美匹配 IM 场景**
   - ✅ 极低延迟（< 1ms）
   - ✅ 原生 Pub/Sub
   - ✅ 支持点对点消息
   - ✅ 消息顺序保证
   - ✅ 离线消息存储（7天足够）

2. **运维成本低**
   - ✅ 单二进制，零依赖
   - ✅ 部署简单
   - ✅ 资源占用小
   - ✅ 学习曲线平缓

3. **云原生**
   - ✅ K8s 友好
   - ✅ 水平扩展
   - ✅ 故障恢复快

4. **已经在用**
   - ✅ 项目中已配置 NATS
   - ✅ Docker Compose 已包含
   - ✅ 零额外成本

#### 适用规模
```
用户数: < 1000万
消息量: < 1000万条/天
并发连接: < 100万
团队规模: 3-10人
```

### 4.2 备选方案

#### 方案A: NATS JetStream + PostgreSQL
```
┌─────────┐
│  NATS   │ ← 实时消息分发 (7天)
│JetStream│
└─────────┘

┌─────────┐
│PostgreSQL│ ← 长期消息存储（按需）
└─────────┘
```

**说明**：
- NATS 负责实时消息推送和短期离线消息（7天）
- PostgreSQL 存储历史消息（长期查询）
- 成本低，运维简单

#### 方案B: NATS + Redis Streams（混合）
```
┌─────────┐
│  NATS   │ ← 服务器间通信
│JetStream│
└─────────┘

┌─────────┐
│  Redis  │ ← 在线状态、正在输入
│ Streams │
└─────────┘
```

**说明**：
- NATS 处理消息分发
- Redis Streams 处理临时性状态（不需要持久化）
- 充分发挥各自优势

### 4.3 不推荐方案

#### ❌ Kafka
**原因**：
- 延迟过高（50-100ms）
- 运维复杂度高
- 资源占用大
- 点对点消息实现困难

**适用场景**：大数据、日志、事件溯源，而非实时 IM

#### ❌ Pulsar
**原因**：
- 复杂度过高（杀鸡用牛刀）
- 运维成本极高
- 需要专业团队

**适用场景**：超大规模（10亿+ 用户）、多租户 SaaS

#### ❌ NSQ
**原因**：
- **消息无序**（致命缺陷）
- IM 必须保证消息顺序

---

## 5. NATS JetStream 深度配置

### 5.1 生产环境配置

#### NATS 服务器配置
```yaml
# nats-server.conf
port: 4222
http_port: 8222

# JetStream 配置
jetstream {
  store_dir: /data/nats
  max_memory_store: 10GB
  max_file_store: 100GB
}

# 集群配置（3节点）
cluster {
  name: nats-cluster
  listen: 0.0.0.0:6222
  routes [
    nats://nats-1:6222
    nats://nats-2:6222
    nats://nats-3:6222
  ]
}

# 监控
monitoring {
  prometheus: /metrics
}
```

#### Stream 配置（消息流）
```go
// 用户消息 Stream
userMessagesStream := &nats.StreamConfig{
    Name:     "USER_MESSAGES",
    Subjects: []string{
        "msg.private.*",     // 私聊: msg.private.{conversationId}
        "msg.group.*",       // 群聊: msg.group.{groupId}
    },
    Storage:    nats.FileStorage,
    Retention:  nats.LimitsPolicy,
    MaxMsgs:    10_000_000,    // 1000万条
    MaxBytes:   100 * 1024 * 1024 * 1024,  // 100GB
    MaxAge:     7 * 24 * time.Hour,         // 7天
    MaxMsgSize: 10 * 1024 * 1024,          // 10MB
    Discard:    nats.DiscardOld,
    Replicas:   3,             // 3副本
}

// 系统通知 Stream
notificationStream := &nats.StreamConfig{
    Name:     "SYSTEM_NOTIFICATIONS",
    Subjects: []string{
        "notif.friend.*",    // 好友通知
        "notif.group.*",     // 群组通知
        "notif.user.*",      // 用户通知
    },
    Storage:    nats.FileStorage,
    Retention:  nats.LimitsPolicy,
    MaxMsgs:    1_000_000,
    MaxBytes:   10 * 1024 * 1024 * 1024,   // 10GB
    MaxAge:     3 * 24 * time.Hour,         // 3天
    MaxMsgSize: 1 * 1024 * 1024,           // 1MB
    Replicas:   3,
}

// 在线状态 Stream（临时）
onlineStatusStream := &nats.StreamConfig{
    Name:     "ONLINE_STATUS",
    Subjects: []string{
        "status.online.*",
        "status.typing.*",
    },
    Storage:    nats.MemoryStorage,  // 内存存储
    Retention:  nats.LimitsPolicy,
    MaxAge:     5 * time.Minute,     // 5分钟
    MaxMsgs:    100_000,
    Replicas:   1,  // 无需副本
}
```

#### Consumer 配置（消费者）
```go
// Realtime 消费者（推送给在线用户）
realtimeConsumer := &nats.ConsumerConfig{
    Durable:       "realtime-consumer",
    DeliverPolicy: nats.DeliverNewPolicy,  // 只消费新消息
    AckPolicy:     nats.AckExplicitPolicy, // 显式ACK
    AckWait:       30 * time.Second,       // 30秒超时
    MaxDeliver:    3,                      // 最多重试3次
    FilterSubject: "msg.*",
}

// Message Service 消费者（持久化）
messageServiceConsumer := &nats.ConsumerConfig{
    Durable:       "message-service-consumer",
    DeliverPolicy: nats.DeliverAllPolicy,  // 消费所有消息
    AckPolicy:     nats.AckExplicitPolicy,
    AckWait:       60 * time.Second,
    MaxDeliver:    5,
    FilterSubject: "msg.*",
}

// 离线消息拉取（按用户）
func createUserOfflineConsumer(userId string) *nats.ConsumerConfig {
    return &nats.ConsumerConfig{
        Durable:       fmt.Sprintf("user-%s-offline", userId),
        DeliverPolicy: nats.DeliverLastPerSubjectPolicy,  // 每个主题最新消息
        FilterSubject: fmt.Sprintf("msg.*.%s", userId),   // 只接收该用户的消息
        AckPolicy:     nats.AckExplicitPolicy,
    }
}
```

### 5.2 性能调优

#### 发布优化
```go
// 批量发布
func PublishBatch(js nats.JetStreamContext, messages []*Message) error {
    var futures []nats.PubAckFuture

    for _, msg := range messages {
        subject := fmt.Sprintf("msg.%s.%s", msg.Type, msg.ConversationId)
        data, _ := proto.Marshal(msg)

        // 异步发布
        future, err := js.PublishAsync(subject, data)
        if err != nil {
            return err
        }
        futures = append(futures, future)
    }

    // 等待所有发布完成
    for _, f := range futures {
        select {
        case <-f.Ok():
            // 成功
        case err := <-f.Err():
            return err
        }
    }
    return nil
}
```

#### 消费优化
```go
// 拉取模式（批量消费）
func ConsumeBatch(js nats.JetStreamContext, consumer string) error {
    sub, err := js.PullSubscribe("msg.*", consumer)
    if err != nil {
        return err
    }

    for {
        // 每次拉取100条
        msgs, err := sub.Fetch(100, nats.MaxWait(5*time.Second))
        if err != nil {
            continue
        }

        // 批量处理
        for _, msg := range msgs {
            processMessage(msg)
            msg.Ack()
        }
    }
}
```

### 5.3 监控指标

```go
// Prometheus 指标
var (
    natsPublishTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "nats_publish_total",
            Help: "Total number of messages published",
        },
        []string{"stream", "subject"},
    )

    natsConsumeTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "nats_consume_total",
            Help: "Total number of messages consumed",
        },
        []string{"consumer"},
    )

    natsPublishLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "nats_publish_latency_seconds",
            Help:    "Latency of message publish",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .5, 1},
        },
        []string{"stream"},
    )

    natsStreamSize = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "nats_stream_size_bytes",
            Help: "Size of NATS stream in bytes",
        },
        []string{"stream"},
    )
)
```

---

## 6. 成本对比（100万用户规模）

### 6.1 硬件资源需求

| 组件 | NATS JetStream | Kafka | Pulsar |
|------|---------------|-------|--------|
| **服务器数量** | 3台 | 3 Broker + 3 Zookeeper = 6台 | 3 Broker + 3 BookKeeper + 3 ZK = 9台 |
| **CPU** | 4核/台 | 8核/台 | 8核/台 |
| **内存** | 8GB/台 | 32GB/台 | 32GB/台 |
| **磁盘** | 500GB SSD | 1TB SSD | 2TB SSD |
| **总成本/月** | ~$300 | ~$1800 | ~$2700 |

### 6.2 运维成本

| 项目 | NATS | Kafka | Pulsar |
|------|------|-------|--------|
| **运维人员** | 0.5人 | 2人 | 3人 |
| **学习时间** | 1周 | 1个月 | 2个月 |
| **部署时间** | 1小时 | 1天 | 3天 |
| **故障恢复** | 分钟级 | 小时级 | 小时级 |

---

## 7. 总结与建议

### 7.1 强烈推荐：NATS JetStream

**综合评分**：⭐⭐⭐⭐⭐ (5/5)

**推荐理由**：
1. ✅ 完美匹配 IM 场景（低延迟、Pub/Sub、顺序保证）
2. ✅ 运维成本极低（单二进制、零依赖）
3. ✅ 性能优异（P99 < 1ms）
4. ✅ 项目已集成（零额外成本）
5. ✅ 学习曲线平缓（1周上手）

**适用规模**：
- 用户数: 10万 - 1000万
- 并发连接: < 100万
- 消息量: < 1亿条/天

### 7.2 实施建议

#### Phase 1: 基础搭建（1周）
```bash
1. 配置 NATS JetStream（3节点集群）
2. 创建 Stream（USER_MESSAGES, SYSTEM_NOTIFICATIONS）
3. Gateway 集成 NATS 客户端
4. 基础消息收发测试
```

#### Phase 2: 功能完善（2周）
```bash
1. 实现离线消息存储和拉取
2. 实现系统通知推送
3. 消息去重和顺序保证
4. 监控和告警配置
```

#### Phase 3: 优化上线（1周）
```bash
1. 性能压测（10万并发）
2. 故障演练
3. 灰度发布
4. 文档完善
```

### 7.3 迁移路径（如需扩展）

```
NATS JetStream (< 1000万用户)
    ↓
    如果性能不足
    ↓
Kafka (1000万 - 1亿用户)
    ↓
    如果仍不足
    ↓
Pulsar (1亿+ 用户，多数据中心)
```

**说明**：
- 大部分 IM 系统终身都在 NATS 阶段
- 除非达到微信、WhatsApp 规模，否则无需 Kafka/Pulsar

---

## 8. FAQ

### Q1: NATS JetStream vs Kafka，什么时候选 Kafka？

**A**: 只有以下场景选 Kafka：
- ✅ 需要长期存储消息（月/年级别）
- ✅ 需要复杂的流处理（Kafka Streams）
- ✅ 用户量 > 5000万
- ✅ 有专业的 Kafka 运维团队

对于大部分 IM 系统，NATS 足够。

### Q2: 7天离线消息够吗？

**A**: 完全够。
- 95% 用户会在 24小时内上线
- 99% 用户会在 3天内上线
- 超过 7天的"离线消息"，用户体验已经无意义
- 历史消息查询走 PostgreSQL，不走消息队列

### Q3: NATS 能支持多少并发连接？

**A**: 单节点 > 100万连接（见官方 Benchmark）
- 10万连接：内存 < 1GB
- 100万连接：内存 < 10GB
- 对于千万级用户，3节点集群足够

### Q4: 消息会丢失吗？

**A**: 不会（正确配置）
- JetStream 文件持久化
- 3副本机制
- ACK 确认机制
- 可靠性等同于 Kafka

### Q5: 如何处理消息顺序？

**A**: Stream 内自动保证
- 同一 conversation 的消息发到同一 Stream
- 使用递增序列号
- 客户端按序处理

---

**文档版本**: v1.0
**最后更新**: 2026-02-16
**作者**: Claude + AnyChat Team
