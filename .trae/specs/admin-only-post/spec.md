# 权限调整与留言功能 - Product Requirements Document

## Overview
- **Summary**: 将内容投稿（视频/图文的创建、编辑、删除）权限收紧为仅管理员可用；普通用户登录后只能浏览；新增"给管理员留言"功能，普通用户可提交留言，留言仅管理员可见，不公开展示。
- **Purpose**: 避免普通用户随意投稿带来的审核负担，集中内容产出由管理员负责；同时为普通用户保留与管理员沟通的渠道。
- **Target Users**: 管理员（投稿、查看留言）、普通用户（浏览、留言）。

## Goals
- 普通用户无法创建/编辑/删除内容（接口层强制 403）。
- 前端对普通用户隐藏所有投稿入口。
- 普通用户可在"我的"页面给管理员留言。
- 管理员可在管理后台查看所有留言。
- 留言不公开展示（无公开接口、不出现在任何公开页面）。

## Non-Goals
- 不实现留言的管理员回复功能（本期留言为单向反馈，管理员可在后台查看）。
- 不修改内容审批流程。
- 不调整英雄/装备/符文等只读数据的权限。

## Background & Context
- 当前 `POST/PUT/DELETE /api/contents` 与 `GET /api/my/contents` 在 `RequiredAuth` 组下，任何登录用户均可投稿。
- 前端对线页、详情页、我的页面均有投稿/编辑入口，未区分角色。
- 系统已有 `RequireAdmin()` 中间件和 `role=admin` 机制。
- 尚无留言/反馈相关数据模型与接口。

## Functional Requirements

- **FR-1**: `POST /api/contents`、`PUT /api/contents/:id`、`DELETE /api/contents/:id` 仅管理员可调用，普通用户返回 403。
- **FR-2**: 前端对线页的"上传视频/发布图文"按钮仅管理员可见。
- **FR-3**: 前端详情页的"编辑"入口仅管理员（或作者本人，即管理员）可见。
- **FR-4**: 前端"我的"页面中，普通用户不展示"我的投稿"编辑入口，改为展示留言功能。
- **FR-5**: 新增 `Message` 模型（id, user_id, username, content, created_at, is_read），通过 AutoMigrate 建表。
- **FR-6**: `POST /api/messages`（需登录）：普通用户提交留言，校验内容非空且长度合理。
- **FR-7**: `GET /api/messages/mine`（需登录）：当前用户查看自己的留言记录。
- **FR-8**: `GET /api/admin/messages`（需管理员）：查看所有用户留言，按时间倒序。
- **FR-9**: 留言无任何公开接口，未登录或非管理员无法读取他人留言。

## Non-Functional Requirements

- **NFR-1**: 留言内容长度限制 1-500 字符。
- **NFR-2**: 留言接口做基础频率限制（同一用户 10 秒内只能提交一次），使用 Valkey。
- **NFR-3**: 权限控制以后端为准，前端隐藏仅为体验优化。

## Constraints
- **Technical**: Go + Gin + GORM(MySQL) + Valkey；Vue3 + Vite。
- **Business**: 管理员独占投稿权；留言仅管理员可见。

## Assumptions
- 现有管理员账号（admin）可正常投稿。
- 普通用户通过邮箱注册后即为 `role=user`。

## Acceptance Criteria

### AC-1: 投稿接口仅管理员
- **Type**: `rule`
- **Given**: 普通用户已登录（role=user）
- **When**: 调用 `POST /api/contents`
- **Then**: 返回 HTTP 403 "需要管理员权限"
- **Pass Condition**: 普通用户投稿被拒
- **Evidence**: 接口测试

### AC-2: 管理员可正常投稿
- **Type**: `rule`
- **Given**: 管理员已登录
- **When**: 调用 `POST /api/contents`
- **Then**: 返回 200，内容创建成功
- **Pass Condition**: 管理员投稿成功
- **Evidence**: 接口测试

### AC-3: 前端投稿入口仅管理员可见
- **Type**: `rule`
- **Given**: 普通用户登录
- **When**: 访问对线页 /matchup
- **Then**: "上传视频"和"发布图文"按钮不渲染
- **Pass Condition**: 普通用户看不到投稿按钮
- **Evidence**: 浏览器检查 DOM

### AC-4: 留言提交接口
- **Type**: `rule`
- **Given**: 普通用户已登录
- **When**: 调用 `POST /api/messages` 提交合法内容
- **Then**: 返回 200，留言入库
- **Pass Condition**: 留言创建成功
- **Evidence**: 数据库查询 + 接口响应

### AC-5: 留言频率限制
- **Type**: `rule`
- **Given**: 普通用户刚刚提交过留言
- **When**: 10 秒内再次提交
- **Then**: 返回 429 或 400，提示"请勿频繁提交"
- **Pass Condition**: 10 秒内重复提交被拒
- **Evidence**: 连续两次调用

### AC-6: 留言仅管理员可见全部
- **Type**: `rule`
- **Given**: 管理员已登录
- **When**: 调用 `GET /api/admin/messages`
- **Then**: 返回所有用户留言列表
- **Pass Condition**: 管理员能看到所有留言
- **Evidence**: 接口返回列表

### AC-7: 留言无公开接口
- **Type**: `rule`
- **Given**: 未登录用户
- **When**: 尝试访问任意留言接口
- **Then**: 返回 401；普通用户无法访问 `/api/admin/messages`（403）
- **Pass Condition**: 留言接口受权限保护
- **Evidence**: 接口测试

### AC-8: 我的留言可见
- **Type**: `rule`
- **Given**: 普通用户已登录且提交过留言
- **When**: 调用 `GET /api/messages/mine`
- **Then**: 返回该用户自己的留言列表
- **Pass Condition**: 用户能看到自己的留言
- **Evidence**: 接口返回该用户留言

### AC-9: 前端留言功能可用性
- **Type**: `rubric`
- **Dimension**: 留言交互体验
- **Scale**: 1-5
- **Anchors**: 1 = 无留言入口；3 = 有输入框但无历史/反馈；5 = 我的页面有留言输入框、提交反馈、历史留言列表
- **Pass Threshold**: >= 4
- **Evidence**: 浏览器交互

### AC-10: 管理后台留言查看
- **Type**: `rubric`
- **Dimension**: 管理员查看留言体验
- **Scale**: 1-5
- **Anchors**: 1 = 管理后台无留言入口；3 = 有列表但信息不全；5 = 管理后台有留言列表，显示用户、内容、时间
- **Pass Threshold**: >= 4
- **Evidence**: 浏览器交互
