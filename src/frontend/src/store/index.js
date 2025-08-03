import Vue from 'vue'
import Vuex from 'vuex'
import auth from './modules/auth'
import client from './modules/client'
import server from './modules/server'
import status from './modules/status'
import users from './modules/users'

Vue.use(Vuex)

export default new Vuex.Store({
  modules: {
    auth,
    client,
    server,
    status,
    users
  }
})
