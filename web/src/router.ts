import { createRouter, createWebHistory } from 'vue-router'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'foryou', component: () => import('@/views/ForYou.vue') },
    { path: '/modes', name: 'modes', component: () => import('@/views/Modes.vue') },
    { path: '/detail/:kind/:source/:id', name: 'detail', component: () => import('@/views/PlaylistDetail.vue') },
    { path: '/search', name: 'search', component: () => import('@/views/Search.vue') },
    { path: '/mine', name: 'mine', component: () => import('@/views/Mine.vue') },
    { path: '/mine/likes', name: 'likes', component: () => import('@/views/MySongs.vue'), meta: { list: 'likes' } },
    { path: '/mine/history', name: 'history', component: () => import('@/views/MySongs.vue'), meta: { list: 'history' } },
    { path: '/mine/playlist/:id', name: 'created', component: () => import('@/views/CreatedPlaylist.vue') },
    { path: '/mine/downloads', name: 'downloads', component: () => import('@/views/Downloads.vue') },
    { path: '/settings', name: 'settings', component: () => import('@/views/Settings.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})
