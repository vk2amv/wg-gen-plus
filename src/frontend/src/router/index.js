import Vue from 'vue'
import VueRouter from 'vue-router'
import store from "../store";

Vue.use(VueRouter);

const routes = [
  {
    path: '/clients',
    name: 'clients',
    component: function () {
      return import(/* webpackChunkName: "Clients" */ '../views/Clients.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/server',
    name: 'server',
    component: function () {
      return import(/* webpackChunkName: "Server" */ '../views/Server.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/status',
    name: 'status',
    component: function () {
      return import(/* webpackChunkName: "Status" */ '../views/Status.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/users',
    name: 'users',
    component: function () {
      return import(/* webpackChunkName: "Users" */ '../views/Users.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/',
    name: 'home',
    component: function () {
      return import(/* webpackChunkName: "Home" */ '../views/Home.vue')
    }
  },
  {
    path: '*',
    redirect: '/'
  }
];

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
});

router.beforeEach((to, from, next) => {
  if(to.matched.some(record => record.meta.requiresAuth)) {
    if (store.getters["auth/isAuthenticated"]) {
      next()
      return
    }
    next('/')
  } else if (to.path === '/' && store.getters["auth/isAuthenticated"]) {
    next('/clients')
  } else {
    next()
  }
})

export default router
