# 邮箱注册与登录 - Independent Review

## Checkpoints

- [x] CP-R1: 注册需邮箱+验证码，错误/过期验证码返回 400
  - **Type**: `rule`
  - **Covers**: AC-1
  - **Evidence**: 注入错误码注册返回 HTTP 400 "验证码错误或已过期"

- [x] CP-R2: 发送验证码接口生成 6 位码存 Valkey(TTL 600s) 并 SMTP 发送
  - **Type**: `rule`
  - **Covers**: AC-2
  - **Evidence**: `generateCode()` 使用 crypto/rand；`valkey.Set` TTL=codeTTL(600s)；`mail.SendVerificationCode` 发送

- [x] CP-R3: 同一邮箱 60s 内重复发送被拒绝
  - **Type**: `rule`
  - **Covers**: AC-3
  - **Evidence**: `SetNX(rateKey, "1", codeResendGap)` 原子操作，失败返回 ErrCodeRateLimited → HTTP 429

- [x] CP-R4: 验证码一次性，使用后即删除
  - **Type**: `rule`
  - **Covers**: AC-4
  - **Evidence**: 注册成功后 `valkey GET code:verify:{email}` 返回 nil

- [x] CP-R5: 邮箱唯一，重复注册返回 409
  - **Type**: `rule`
  - **Covers**: AC-5
  - **Evidence**: 重复邮箱（有效码）注册返回 HTTP 409

- [x] CP-R6: 登录支持用户名或邮箱
  - **Type**: `rule`
  - **Covers**: AC-6
  - **Evidence**: 用邮箱登录返回 HTTP 200 + token

- [x] CP-R7: 数据库自动迁移 email/email_verified 列，现有用户不受影响
  - **Type**: `rule`
  - **Covers**: AC-7
  - **Evidence**: `DESCRIBE users` 含 email(varchar 128 UNI)、email_verified(tinyint)；admin 用户名登录 HTTP 200

- [x] CP-R8: SMTP 未配置时发码返回 503 降级，不崩溃
  - **Type**: `rule`
  - **Covers**: AC-8
  - **Evidence**: 未配置 SMTP 时 send-code 返回 HTTP 503 "邮件服务未配置"

- [x] CP-U1: 前端注册表单可用性
  - **Type**: `rubric`
  - **Covers**: AC-9
  - **Scale**: 1-5
  - **Anchors**: 1 = 无邮箱/验证码；3 = 有字段无倒计时；5 = 邮箱+验证码+60s倒计时+错误提示完整
  - **Pass Threshold**: >= 4
  - **Score**: 5
  - **Evidence**: AuthView.vue 含邮箱输入框、验证码输入框、发送按钮、60s 倒计时、错误提示

- [x] CP-U2: 前端登录表单灵活性
  - **Type**: `rubric`
  - **Covers**: AC-10
  - **Scale**: 1-5
  - **Anchors**: 1 = 仅用户名；3 = 支持邮箱提示不清；5 = 提示"用户名/邮箱"且两者均可登录
  - **Pass Threshold**: >= 4
  - **Score**: 5
  - **Evidence**: 登录模式 label 为"用户名 / 邮箱"，后端 accept account 字段

## Review History

### Review R1
- **Result**: `pass`
- **Evidence**: 独立审查员核查全部 10 项 AC 通过；提出 3 个 actionable 改进（A1 发码顺序、A2 唯一键冲突映射、A3 频率限制原子性）及 B1 邮箱验证防御性检查、B2 邮箱格式严格性。
- **Remediation**: 
  - A1: 改为先存 Valkey 再发邮件，邮件失败回删验证码
  - A2: `ErrDuplicatedKey` 时查询用户名/邮箱判断冲突来源
  - A3: 频率限制改用 `SetNX` 原子操作
  - B1: Login 增加 `EmailVerified` 防御性检查
  - B2: `validateEmail` 校验 `addr.Address == email`，拒绝带显示名地址
- **Post-remediation verification**: 完整流程（发码→注册→登录）再次通过，验证码一次性、邮箱唯一、邮箱登录均正常。
