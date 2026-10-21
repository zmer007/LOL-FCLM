<template>
  <main class="about">
    <section class="hero">
      <h1>关于路边小站</h1>
      <p class="tagline">个人向的英雄联盟对线技巧学习站点</p>
    </section>

    <section class="section">
      <div class="section-title">站点简介</div>
      <p>
        路边小站是个人使用的英雄联盟对线技巧网站。
      </p>
      <p>
        站点内容由站长个人整理与维护，用于记录自己的对线理解与出装思路，也方便随时查阅回顾。
      </p>
    </section>

    <section class="section">
      <div class="section-title">版本信息</div>
      <div class="kv"><span>应用版本</span><span>v0.1.0</span></div>
      <div class="kv"><span>技术栈</span><span>Vue 3 + Go + Gin + MySQL</span></div>
      <div class="kv"><span>更新日期</span><span>2026-10-18</span></div>
    </section>

    <div class="disclaimer">
      本站为学习交流用途的非官方站点，不包含正式游戏数据。<br />
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
