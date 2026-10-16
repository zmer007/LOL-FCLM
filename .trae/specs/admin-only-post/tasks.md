# 权限调整与留言功能 - Implementation Plan

## Task 1: 后端投稿权限收紧为管理员
- **Status**: `completed`
- **Priority**: high
- **Depends On**: None
- **Completion Evidence**: 将 POST/PUT/DELETE /api/contents 移至无路径前缀的 `adminPost` 组（RequiredAuth + RequireAdmin），保持路径不变。接口测试：普通用户 POST /api/contents → 403；管理员 POST /api/contents → 400（校验失败，非 403，证明通过权限校验）。

## Task 2: Message 模型与仓库
- **Status**: `completed`
- **Priority**: high
- **Depends On**: None
- **Completion Evidence**: 新建 model/message.go、repository/message_repository.go，database.go AutoMigrate 加入 &model.Message{}。DESCRIBE messages 确认字段：id, user_id(MUL), username(32), content(500), is_read(default 0), created_at。

## Task 3: Message 服务与 Handler
- **Status**: `completed`
- **Priority**: high
- **Depends On**: Task 2
- **Completion Evidence**: service/message_service.go（Send 校验 1-500 字 + Valkey SetNX 10s 限流）、handler/message_handler.go（Send/Mine/AdminList，错误映射 400/429）。测试：合法留言 200 入库；空内容 400；10s 内重复 429。

## Task 4: 路由注册
- **Status**: `completed`
- **Priority**: high
- **Depends On**: Task 3
- **Completion Evidence**: main.go 装配 messageRepo/Service/Handler；路由 POST /api/messages、GET /api/messages/mine（auth 组），GET /api/admin/messages（admin 组）。测试：未登录 401；普通用户 /admin/messages 403。

## Task 5: 前端隐藏投稿入口 + 留言功能
- **Status**: `completed`
- **Priority**: high
- **Depends On**: Task 1, Task 4
- **Completion Evidence**: api/message.ts；MatchupView 投稿按钮 v-if="isAdmin"；MeView 投稿区块 v-if="isAdmin"，新增留言输入框+字数计数+提交+历史列表；router 投稿路由 requiresAdmin 守卫。浏览器验证：普通用户对线页无投稿按钮；/me 无投稿区块，有留言模块，提交成功入历史。

## Task 6: 管理后台留言列表
- **Status**: `completed`
- **Priority**: high
- **Depends On**: Task 4
- **Completion Evidence**: AdminView 新增"用户留言"标签，列表展示用户名/内容/时间。浏览器验证：管理员后台可见标签，点击显示留言列表。

## Task 7: 构建部署与端到端验证
- **Status**: `completed`
- **Priority**: high
- **Depends On**: Task 1-6
- **Completion Evidence**: docker compose build 成功；重启 api/frontend；接口测试 AC-1~AC-8 全通过；浏览器验证 AC-3、AC-9(rubric 5/5)、AC-10(rubric 5/5) 通过。
