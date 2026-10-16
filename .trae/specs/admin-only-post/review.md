# 权限调整与留言功能 - Review

## Review History

### Cycle 1 — Independent Review

**Reviewer**: independent (general_purpose_task subagent)
**Date**: 2026-10-17
**Result**: `pass`

## Checkpoint Results

| AC | Type | Result | Evidence |
|----|------|--------|----------|
| AC-1 普通用户投稿 403 | rule | ✅ pass | 接口测试 HTTP 403；路由在 adminPost 组（RequiredAuth+RequireAdmin） |
| AC-2 管理员可投稿 | rule | ✅ pass | 接口测试 HTTP 400（校验失败，非 403，证明通过权限校验） |
| AC-3 对线页投稿按钮仅管理员可见 | rule | ✅ pass | 浏览器验证普通用户无按钮；MatchupView v-if="isAdmin" |
| AC-4 留言提交接口 | rule | ✅ pass | 接口测试 HTTP 200，留言入库（id:1） |
| AC-5 留言频率限制 | rule | ✅ pass | 接口测试 10s 内重复返回 429 |
| AC-6 管理员看全部留言 | rule | ✅ pass | 接口测试 HTTP 200 返回列表；浏览器验证后台留言标签 |
| AC-7 留言无公开接口 | rule | ✅ pass | 未登录 401；普通用户 /admin/messages 403 |
| AC-8 我的留言可见 | rule | ✅ pass | 接口测试 HTTP 200 返回该用户留言 |
| AC-9 前端留言功能 | rubric | ✅ pass (5/5) | 我的页面有输入框+字数计数+提交反馈+历史列表 |
| AC-10 管理后台留言查看 | rubric | ✅ pass (5/5) | 后台留言列表展示用户名/内容/时间 |

## Security Checks
- 留言端点全部受 RequiredAuth 保护，无公开路由 ✅
- Mine 接口用 JWT claims 的 UserID 过滤，无法越权读取他人留言 ✅
- 投稿写路由全部 RequireAdmin ✅

## Non-blocking Observations (advisory, not actioned)
1. 前后端字符计数口径（length vs rune）对 emoji 类略有偏差，不影响功能。
2. `/admin` 路由仅 requiresAuth，由 AdminView 内部校验角色，后端强制，无安全风险。
3. 留言列表无分页，数据量大时可优化。

## Remediation
- 采纳：message_repository 排序由 `id DESC` 改为 `created_at DESC`，贴合 spec "按时间倒序" 语义。已重建验证通过。

## Final Verdict
**pass** — 所有 10 条 AC 通过，安全控制正确，构建与端到端测试通过。
