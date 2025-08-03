import ApiService from "../../services/api.service";

const state = {
  error: null,
  server: null,
  config: '',
  version: '_ci_build_not_run_properly_',
}

const getters = {
  error(state) {
    return state.error;
  },

  server(state) {
    return state.server;
  },

  version(state) {
    return state.version;
  },

  config(state) {
    return state.config;
  },
}

const actions = {

    error({ commit }, error) {
    commit('errorVisible', false);
    commit('error', '');
  
    setTimeout(() => {
      let msg = 'Unknown error occurred';
  
      // Axios-style errors
      if (error?.response?.data?.message) {
        msg = error.response.data.message;
      }
      // Generic .message error
      else if (error?.message) {
        msg = error.message;
      }
      // String fallback
      else if (typeof error === 'string') {
        msg = error;
      }
  
      commit('error', msg);
      commit('errorVisible', true);
    }, 10);
  },

  read({ commit, dispatch }){
    ApiService.get("/server")
      .then(resp => {
        commit('server', resp)
        dispatch('config')
      })
      .catch(err => {
        commit('error', err)
      })
  },

  update({ commit }, server){
    ApiService.patch(`/server`, server)
      .then(resp => {
        commit('server', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  config({ commit }){
    ApiService.getWithConfig("/server/config", {responseType: 'arraybuffer'})
      .then(resp => {
        commit('config', resp)
      })
      .catch(err => {
        commit('error', err)
      })
  },

  version({ commit }){
    ApiService.get("/server/version")
      .then(resp => {
        commit('version', resp.version)
      })
      .catch(err => {
        commit('error', err)
      })
  },

}

const mutations = {
  error(state, error) {
    state.error = error;
  },

  server(state, server){
    state.server = server
  },

  config(state, config){
    state.config = config
  },

  version(state, version){
    state.version = version
  },
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
