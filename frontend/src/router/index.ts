import { createRouter, createWebHistory } from 'vue-router'
import { TOKEN_KEY } from '../api/http'

// 现有 12 个设计页将在后续里程碑逐个迁入
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: () => import('../views/HomeView.vue') },
    {
      // 对线页：首页搜索带 ?champion= 进入；里程碑 3 完善为 /matchup/:heroId
      path: '/matchup',
      name: 'matchup',
      component: () => import('../views/MatchupView.vue'),
    },
    { path: '/about', name: 'about', component: () => import('../views/AboutView.vue') },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/AuthView.vue'),
      props: { mode: 'login' },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/AuthView.vue'),
      props: { mode: 'register' },
    },
    {
      path: '/me',
      name: 'me',
      component: () => import('../views/MeView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

// 登录守卫：未登录访问需鉴权页面时跳登录，并记录回跳地址
router.beforeEach((to) => {
  if (to.meta.requiresAuth && !localStorage.getItem(TOKEN_KEY)) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
