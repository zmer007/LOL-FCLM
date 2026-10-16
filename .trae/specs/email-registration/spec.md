# 邮箱注册与登录 - Product Requirements Document

## Overview
- **Summary**: 为 LOL 辅助平台新增基于邮箱的用户注册与登录能力。注册时邮箱为必填项，通过发送 6 位验证码到用户邮箱完成邮箱归属验证；验证通过后账号激活。登录时支持用户名或邮箱 + 密码。
- **Purpose**: 当前注册仅依赖用户名+密码，无法进行密码找回、重要通知触达，且缺少邮箱归属校验。引入邮箱可完善账号体系，支撑后续找回密码、通知等能力。
- **Target Users**: 平台新注册用户、已注册用户（登录）。

## Goals
- 注册时收集并校验邮箱，确保邮箱归属真实（6 位验证码）。
- 未验证邮箱的账号不可登录。
- 登录支持用户名或邮箱 + 密码。
- 邮件发送通过可配置的 SMTP 完成，验证码存储在 Valkey 中并设 TTL。

## Non-Goals
- 不实现"忘记密码 / 重置密码"流程（本次只做注册与登录，预留邮箱字段供后续）。
- 不实现邮箱修改功能。
- 不接入第三方邮件 API（如 SendGrid），仅用 SMTP。
- 不实现图形验证码 / 频率限制的复杂风控（仅做基础的发送频率限制）。

## Background & Context
- 当前 `User` 模型仅有 `username`、`password_hash`、`avatar`、`role`、`status`，无邮箱字段。
- 当前注册接口 `POST /api/auth/register` 入参为 `{username, password}`，直接创建活跃用户。
- 后端已部署 Valkey（Redis 协议）容器，但 Go 代码中尚无 Valkey 客户端依赖。
- 后端无 SMTP 邮件发送能力。
- 前端 `AuthView.vue` 注册表单仅有用户名、密码两个字段。

## Functional Requirements

- **FR-1**: 注册接口接受 `username`、`email`、`password`，邮箱为必填，格式需校验。
- **FR-2**: 注册前需先调用"发送验证码"接口，后端生成 6 位数字验证码，通过 SMTP 发送到该邮箱，并将验证码存入 Valkey（key 含邮箱，TTL 10 分钟）。
- **FR-3**: 同一邮箱 60 秒内只能发送一次验证码（频率限制）。
- **FR-4**: 注册时校验验证码是否匹配且未过期；验证通过后创建用户，邮箱标记为已验证，账号状态为 active。
- **FR-5**: 邮箱在系统中唯一，重复邮箱注册返回冲突错误。
- **FR-6**: 未验证邮箱的账号无法登录（本次注册流程已强制验证，故此条为防御性约束）。
- **FR-7**: 登录接口接受 `account`（用户名或邮箱）+ `password`，后端按用户名或邮箱查找用户。
- **FR-8**: 提供"重新发送验证码"能力（复用发送接口，受 60 秒频率限制）。
- **FR-9**: 验证码使用后立即失效（一次性）。

## Non-Functional Requirements

- **NFR-1**: 验证码 10 分钟内有效，超时需重新获取。
- **NFR-2**: 验证码为 6 位数字，不可预测（使用加密随机源）。
- **NFR-3**: SMTP 配置（host、port、username、password、from）通过环境变量注入，默认不开启邮件功能（未配置时发送接口返回友好提示）。
- **NFR-4**: 邮件发送失败时接口返回错误，不泄露 SMTP 凭据等内部信息。
- **NFR-5**: 数据库迁移通过 GORM AutoMigrate 自动添加 `email`、`email_verified` 字段，不破坏现有数据。

## Constraints
- **Technical**: Go 1.26 + Gin + GORM(MySQL) + Valkey；前端 Vue3 + Vite。
- **Business**: 邮箱必填、6 位验证码、支持邮箱登录（用户已确认）。
- **Dependencies**: 需要可用的 SMTP 服务器（由用户通过环境变量配置）；Valkey 服务已部署。

## Assumptions
- 用户有可用的 SMTP 服务器（如 QQ 邮箱、163 邮箱、企业邮箱等）并愿意提供凭据。
- 容器内可访问 SMTP 服务器的网络。
- 现有用户（admin）无邮箱，仍可通过用户名登录（邮箱字段允许为空，仅对新注册用户必填）。

## Acceptance Criteria

### AC-1: 注册需邮箱+验证码
- **Type**: `rule`
- **Given**: SMTP 已正确配置
- **When**: 调用 `POST /api/auth/register` 且未提供正确的验证码
- **Then**: 返回 400，提示验证码错误或已过期，不创建用户
- **Pass Condition**: 接口拒绝无/错误验证码的注册请求
- **Evidence**: 后端接口测试 + 日志

### AC-2: 发送验证码接口
- **Type**: `rule`
- **Given**: 合法邮箱地址
- **When**: 调用 `POST /api/auth/send-code`，body 含 `email`
- **Then**: 后端生成 6 位码存入 Valkey（TTL 600s），通过 SMTP 发送邮件，返回 200
- **Pass Condition**: 邮件成功发送且 Valkey 中存在对应验证码
- **Evidence**: 邮件收件箱 + Valkey 检查

### AC-3: 验证码频率限制
- **Type**: `rule`
- **Given**: 同一邮箱刚刚发送过验证码
- **When**: 60 秒内再次调用发送接口
- **Then**: 返回 429 或 400，提示"请稍后再试"
- **Pass Condition**: 60 秒内重复发送被拒绝
- **Evidence**: 连续两次调用接口的响应

### AC-4: 验证码一次性 & 过期
- **Type**: `rule`
- **Given**: 验证码已用于一次成功注册，或已超过 10 分钟
- **When**: 再次使用同一验证码注册
- **Then**: 返回验证码错误
- **Pass Condition**: 验证码使用后即删除，过期后不可用
- **Evidence**: Valkey key 在成功注册后被删除

### AC-5: 邮箱唯一性
- **Type**: `rule`
- **Given**: 某邮箱已被注册
- **When**: 用同一邮箱再次注册
- **Then**: 返回 409，提示邮箱已被注册
- **Pass Condition**: 重复邮箱注册被拒绝
- **Evidence**: 数据库唯一索引 + 接口响应

### AC-6: 登录支持邮箱
- **Type**: `rule`
- **Given**: 用户已注册（用户名 user1，邮箱 a@b.com）
- **When**: 用 `account=a@b.com` + 正确密码登录
- **Then**: 返回 200 与 token
- **Pass Condition**: 邮箱登录成功
- **Evidence**: 接口返回 token

### AC-7: 数据库自动迁移
- **Type**: `rule`
- **Given**: 现有数据库已有 users 表
- **When**: 后端启动执行 AutoMigrate
- **Then**: users 表新增 `email`、`email_verified` 列，现有数据不受影响
- **Pass Condition**: 列存在且 admin 用户仍可登录
- **Evidence**: `DESCRIBE users` + admin 登录成功

### AC-8: SMTP 未配置时的降级
- **Type**: `rule`
- **Given**: 未配置 SMTP 环境变量
- **When**: 调用发送验证码接口
- **Then**: 返回 503 或明确错误，提示邮件服务未配置
- **Pass Condition**: 不崩溃，返回友好提示
- **Evidence**: 接口响应

### AC-9: 前端注册表单
- **Type**: `rubric`
- **Dimension**: 注册流程可用性
- **Scale**: 1-5
- **Anchors**: 1 = 表单仍只有用户名密码；3 = 有邮箱+验证码输入框，但交互生硬；5 = 邮箱、密码、验证码输入完整，含"发送验证码"按钮与倒计时，错误提示清晰
- **Pass Threshold**: >= 4
- **Evidence**: 前端页面交互测试

### AC-10: 前端登录表单
- **Type**: `rubric`
- **Dimension**: 登录表单灵活性
- **Scale**: 1-5
- **Anchors**: 1 = 仍只支持用户名；3 = 支持邮箱但提示不明确；5 = 输入框提示"用户名/邮箱"，两者均可登录，体验一致
- **Pass Threshold**: >= 4
- **Evidence**: 前端页面交互测试

## Open Questions
- [ ] SMTP 服务器由用户自行配置（已在假设中说明，无需回答）
