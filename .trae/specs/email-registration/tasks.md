# 邮箱注册与登录 - Implementation Plan

## Task 1: User 模型增加邮箱字段与仓库方法
- **Status**: `pending`
- **Priority**: high
- **Depends On**: None
- **Description**:
  - 在 `model.User` 增加 `Email`（size 128，唯一索引，允许空）与 `EmailVerified`（bool，默认 false）字段。
  - 在 `UserRepository` 增加 `FindByEmail(email string)` 方法。
  - 确认 `database.Open` 的 AutoMigrate 会自动迁移新字段。
- **Acceptance Criteria Addressed**: AC-5, AC-7
- **Test Requirements**:
  - `rule` TR-1.1: 重启后端后 `DESCRIBE users` 出现 `email`、`email_verified` 列；admin 用户仍可用户名登录。证据：mysql 查询 + 登录接口返回 200。
  - `rule` TR-1.2: `FindByEmail` 能按邮箱查到用户，查不到返回 `repository.ErrNotFound`。证据：调用注册后用邮箱查询命中。

## Task 2: 配置层增加 SMTP 与验证码参数
- **Status**: `pending`
- **Priority**: high
- **Depends On**: None
- **Description**:
  - `config.Config` 增加 `SMTPHost`、`SMTPPort`、`SMTPUser`、`SMTPPassword`、`SMTPFrom`、`CodeTTLSeconds`、`CodeResendSeconds`。
  - 更新 `.env.example` 增加对应变量并给出示例注释。
  - `docker-compose.yml` 的 api 服务 environment 透传 SMTP_* 变量（从 .env 读取）。
- **Acceptance Criteria Addressed**: AC-8, NFR-1, NFR-3
- **Test Requirements**:
  - `rule` TR-2.1: 未配置 SMTP_HOST 时，配置中 `SMTPHost` 为空字符串。证据：打印配置或代码断言。
  - `rule` TR-2.2: `.env.example` 包含 SMTP_HOST、SMTP_PORT、SMTP_USER、SMTP_PASSWORD、SMTP_FROM 字段。证据：文件内容。

## Task 3: Valkey 客户端初始化
- **Status**: `pending`
- **Priority**: high
- **Depends On**: None
- **Description**:
  - 在 go.mod 增加 `github.com/redis/go-redis/v9` 依赖（Valkey 兼容 Redis 协议）。
  - 在 `main.go` 中用 `cfg.ValkeyAddr` 初始化 `*redis.Client`，启动时 Ping 确认连通。
  - 提供包级变量或通过依赖注入传递给 AuthService。
- **Acceptance Criteria Addressed**: AC-2, AC-4
- **Test Requirements**:
  - `rule` TR-3.1: 后端启动日志显示 Valkey 连接成功（无 panic）。证据：容器日志。
  - `rule` TR-3.2: 可通过 redis client Set/Get 一个测试 key。证据：日志或调试输出。

## Task 4: SMTP 邮件服务
- **Status**: `pending`
- **Priority**: high
- **Depends On**: Task 2
- **Description**:
  - 新建 `internal/service/mail_service.go`，封装 `SendVerificationCode(to, code string) error`。
  - 使用 Go 标准库 `net/smtp` 发送 HTML/纯文本邮件，支持 STARTTLS。
  - 未配置 SMTP 时返回明确错误（不 panic）。
  - 邮件正文包含 6 位验证码与平台名称。
- **Acceptance Criteria Addressed**: AC-2, AC-8, NFR-4
- **Test Requirements**:
  - `rule` TR-4.1: SMTP 未配置时调用 `SendVerificationCode` 返回非 nil 错误且不 panic。证据：单元调用或接口响应。
  - `rubric` TR-4.2: 邮件可读性；scale 1-5；anchors 1=乱码/无内容；3=纯文本含验证码；5=格式清晰含平台名+验证码+有效期；threshold >= 4；evidence：收到的邮件内容。

## Task 5: AuthService 实现验证码与邮箱注册/登录
- **Status**: `pending`
- **Priority**: high
- **Depends On**: Task 1, Task 3, Task 4
- **Description**:
  - `SendVerificationCode(email string)`: 校验邮箱格式 → 检查 60s 频率限制（Valkey key `code:rate:{email}`）→ 生成 6 位加密随机码 → 存 Valkey `code:verify:{email}` TTL 600s → 调邮件服务发送。
  - `Register(username, email, password, code string)`: 校验用户名/密码/邮箱 → 校验验证码（取 Valkey、比对、删除）→ 检查用户名/邮箱唯一 → 创建用户（Email 填充，EmailVerified=true）→ 发 token。
  - `Login(account, password string)`: account 先按用户名查，查不到再按邮箱查 → 校验密码 → 发 token。
  - 增加 `ErrEmailTaken`、`ErrInvalidCode`、`ErrCodeRateLimited`、`ErrEmailRequired` 等错误。
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6
- **Test Requirements**:
  - `rule` TR-5.1: 错误验证码注册返回 400。证据：接口测试。
  - `rule` TR-5.2: 60s 内重复发送返回频率限制错误。证据：连续调用 send-code 第二次被拒。
  - `rule` TR-5.3: 注册成功后验证码从 Valkey 删除。证据：注册后查 Valkey key 不存在。
  - `rule` TR-5.4: 重复邮箱注册返回 409。证据：接口测试。
  - `rule` TR-5.5: 用邮箱登录返回 token。证据：接口测试。
  - `rule` TR-5.6: 邮箱格式错误返回 400。证据：接口测试。

## Task 6: Handler 与路由
- **Status**: `pending`
- **Priority**: high
- **Depends On**: Task 5
- **Description**:
  - `AuthHandler` 增加 `SendCode(c)` 处理 `POST /api/auth/send-code`（入参 `{email}`）。
  - 修改 `Register` 入参为 `{username, email, password, code}`。
  - 修改 `Login` 入参字段从 `username` 改为 `account`（兼容旧字段也接受 username）。
  - `main.go` 注册新路由；`userVO` 增加 `email` 字段返回。
- **Acceptance Criteria Addressed**: AC-1, AC-2, AC-3, AC-6, AC-8
- **Test Requirements**:
  - `rule` TR-6.1: `POST /api/auth/send-code` 合法邮箱返回 200。证据：curl/接口测试。
  - `rule` TR-6.2: `POST /api/auth/register` 入参含 email+code 成功创建用户。证据：接口返回 token + 数据库有记录。
  - `rule` TR-6.3: `POST /api/auth/login` 接受 `account` 字段（用户名或邮箱）。证据：接口测试。

## Task 7: 前端 API 与注册/登录 UI
- **Status**: `pending`
- **Priority**: high
- **Depends On**: Task 6
- **Description**:
  - `api/auth.ts` 增加 `sendCode(email)`，修改 `register` 入参加 `email`、`code`，`login` 入参改为 `account`。
  - `stores/user.ts` 对应调整方法签名。
  - `AuthView.vue` 注册模式新增邮箱输入框、验证码输入框、"发送验证码"按钮（60s 倒计时）；登录模式将用户名框提示改为"用户名/邮箱"。
  - 错误提示展示后端返回信息。
- **Acceptance Criteria Addressed**: AC-9, AC-10
- **Test Requirements**:
  - `rubric` TR-7.1: 注册表单可用性；scale 1-5；anchors 1=无邮箱/验证码；3=有字段无倒计时；5=邮箱+验证码+60s倒计时+错误提示完整；threshold >= 4；evidence：浏览器交互。
  - `rubric` TR-7.2: 登录表单灵活性；scale 1-5；anchors 1=仅用户名；3=支持邮箱提示不清；5=提示"用户名/邮箱"且两者均可登录；threshold >= 4；evidence：浏览器交互。

## Task 8: 构建部署与端到端验证
- **Status**: `pending`
- **Priority**: high
- **Depends On**: Task 1, Task 2, Task 3, Task 4, Task 5, Task 6, Task 7
- **Description**:
  - `docker compose build api frontend` 成功。
  - 重启服务，验证数据库迁移、发送验证码、注册、登录全流程。
  - 若用户提供 SMTP 凭据，端到端验证邮件收发；否则验证降级路径。
- **Acceptance Criteria Addressed**: AC-1 ~ AC-10
- **Test Requirements**:
  - `rule` TR-8.1: `docker compose build` 无错误。证据：构建成功退出码 0。
  - `rule` TR-8.2: 后端启动无 panic，AutoMigrate 成功。证据：容器日志。
  - `rule` TR-8.3: 完整注册流程可走通（发码→填码→注册→登录）。证据：接口调用序列。
