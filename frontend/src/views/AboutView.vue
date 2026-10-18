<template>
  <main class="about">
    <section class="hero">
      <h1>关于路边英雄</h1>
      <p class="tagline">英雄联盟对线技巧学习与分享平台</p>
    </section>

    <section class="section">
      <div class="section-title">平台简介</div>
      <p>
        路边英雄是一个聚焦英雄联盟对线技巧的学习与分享平台。你可以在首页搜索英雄，查看该英雄的推荐出装、符文与对线视频；
        也可以浏览全英雄资料，查看装备与符文图鉴，或在个人中心管理收藏与投稿。
      </p>
      <p>
        平台鼓励召唤师分享自己的对线理解与出装思路，所有投稿内容经审核后公开展示，帮助更多玩家提升对线水平。
      </p>
    </section>

    <section class="section">
      <div class="section-title">核心功能</div>
      <ul class="feature-list">
        <li>
          <span class="feat-icon">🔍</span>
          <div>
            <div class="feat-name">英雄搜索</div>
            <div class="feat-desc">快速查找英雄，查看对线出装、符文与视频攻略</div>
          </div>
        </li>
        <li>
          <span class="feat-icon">⚔️</span>
          <div>
            <div class="feat-name">装备符文图鉴</div>
            <div class="feat-desc">浏览全装备、召唤师技能与符文，数据同步自官方 Data Dragon</div>
          </div>
        </li>
        <li>
          <span class="feat-icon">⭐</span>
          <div>
            <div class="feat-name">收藏管理</div>
            <div class="feat-desc">收藏喜欢的内容，在个人中心随时回顾</div>
          </div>
        </li>
      </ul>
    </section>

    <section class="section">
      <div class="section-title">版本信息</div>
      <div class="kv"><span>应用版本</span><span>v0.1.0</span></div>
      <div class="kv"><span>技术栈</span><span>Vue 3 + Go + Gin + MySQL</span></div>
      <div class="kv"><span>更新日期</span><span>2026-10-18</span></div>
    </section>

    <div class="disclaimer">
      本站为学习交流用途的非官方平台，不包含正式游戏数据。<br />
      英雄联盟及相关名称、形象的权利归其权利人所有。
    </div>
  </main>
</template>

<style scoped>
.about {
  max-width: 760px;
  margin: 0 auto;
  padding: 32px 24px 60px;
}

.hero {
  text-align: center;
  padding: 24px 0 36px;
}
.hero h1 {
  font-size: 28px;
  color: var(--gold-bright);
  letter-spacing: 2px;
}
.hero .tagline {
  margin-top: 10px;
  font-size: 14px;
  color: var(--mute);
}

.section {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  padding: 24px 28px 26px;
  margin-bottom: 20px;
}
.section-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--gold-bright);
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}
.section-title::before {
  content: "";
  width: 3px;
  height: 15px;
  background: var(--gold);
  border-radius: 2px;
}
.section p {
  font-size: 13.5px;
  line-height: 1.9;
  color: var(--mute);
}
.section p + p {
  margin-top: 12px;
}

.feature-list {
  list-style: none;
  display: grid;
  gap: 18px;
}
.feature-list li {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}
.feat-icon {
  font-size: 22px;
  line-height: 1;
  flex-shrink: 0;
}
.feat-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--gold-bright);
}
.feat-desc {
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--mute);
}

.kv {
  display: flex;
  justify-content: space-between;
  padding: 11px 2px;
  font-size: 13px;
}
.kv + .kv {
  border-top: 1px solid var(--line-soft);
}
.kv span:first-child {
  color: var(--faint);
}
.kv span:last-child {
  color: var(--mute);
}

.disclaimer {
  margin-top: 26px;
  text-align: center;
  font-size: 12px;
  color: var(--faint);
  line-height: 1.9;
}
</style>
