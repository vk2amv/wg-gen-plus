import ApiService from "../../services/api.service";
import TokenService from "../../services/token.service";

// state
const state = {
  error: null,
  user: JSON.parse(localStorage.getItem('user')) || null,
  authStatus: '',
  authRedirectUrl: '',
  status: '',
  token: localStorage.getItem('token') || '',
  redirectUrl: null,
  isLocalAuth: false, // Add this to track auth type
};

// getters
const getters = {
  error(state) {
    return state.error;
  },
  user(state) {
    return state.user;
  },
  isAuthenticated(state) {
    return !!state.token && state.user !== null;
  },
  authRedirectUrl(state) {
    return state.authRedirectUrl
  },
  authStatus(state) {
    return state.authStatus
  },
  isLocalAuth: state => state.isLocalAuth, // Add this getter
};

// actions
const actions = {
  user({ commit }){
    ApiService.get("/auth/user")
      .then( resp => {
        commit('user', resp)
      })
      .catch(err => {
        commit('error', err);
        commit('logout')
      });
  },

  // Initialize auth state from stored token
  initAuth({ commit, dispatch, state }) {
    return new Promise((resolve) => {
      const token = TokenService.getToken();
      if (token) {
        console.log("Found existing token, restoring session");
        ApiService.setHeader();
        
        // If we don't already have a user object in state, fetch it
        if (!state.user) {
          dispatch('user');
        } else {
          console.log("User data already in state:", state.user);
        }
      } else {
        console.log("No token found, user needs to authenticate");
      }
      resolve();
    });
  },

  oauth2_url({ commit, dispatch }){
    if (TokenService.getToken()) {
      ApiService.setHeader();
      dispatch('user');
      return
    }
    ApiService.get("/auth/oauth2_url")
      .then(resp => {
        if (resp.codeUrl === '_magic_string_fake_auth_no_redirect_'){
          console.log("server report oauth2 is disabled, fake exchange")
          commit('authStatus', 'disabled')
          TokenService.saveClientId(resp.clientId)
          dispatch('oauth2_exchange', {code: "", state: resp.state})
        } else {
          commit('authStatus', 'redirect')
          commit('authRedirectUrl', resp)
        }
      })
      .catch(err => {
        commit('authStatus', 'error')
        commit('error', err);
        commit('logout')
      })
  },

  oauth2_exchange({ commit, dispatch }, data){
    data.clientId = TokenService.getClientId()
    ApiService.post("/auth/oauth2_exchange", data)
      .then(resp => {
        commit('authStatus', 'success')
        commit('token', resp)
        dispatch('user');
      })
      .catch(err => {
        commit('authStatus', 'error')
        commit('error', err);
        commit('logout')
      })
  },

  logout({ commit }){
    ApiService.get("/auth/logout")
      .then(resp => {
        commit('logout')
      })
      .catch(err => {
        commit('authStatus', '')
        commit('error', err);
        commit('logout')
      })
  },

  // Check auth type
  async check_auth_type({ commit }) {
    try {
      const response = await ApiService.get('/auth/type');
      console.log("Auth type response:", response);
      const isLocal = response.isLocal;
      commit('SET_AUTH_TYPE', isLocal);
      return isLocal;
    } catch (err) {
      console.error('Error checking auth type:', err);
      return false;
    }
  },

  // Local login
  async local_login({ commit }, { username, password }) {
    commit('AUTH_REQUEST');
    try {
      console.log("Attempting login with:", username, "password length:", password?.length);
      
      // Send as JSON instead of FormData
      const response = await ApiService.post('/auth/login', {
        username: username,
        password: password
      });
      
      console.log("Login response:", response);
      
      const token = response.token;
      const user = response.user;
      
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      ApiService.setHeader();
      
      commit('AUTH_SUCCESS', { token, user });
    } catch (err) {
      console.error("Login error:", err);
      console.error("Response data:", err.response?.data);
      commit('AUTH_ERROR', err.response?.data?.error || 'Authentication failed');
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      throw err;
    }
  },
};

// mutations
const mutations = {
  error(state, error) {
    state.error = error;
  },
  authStatus(state, authStatus) {
    state.authStatus = authStatus;
  },
  authRedirectUrl(state, resp) {
    state.authRedirectUrl = resp.codeUrl;
    TokenService.saveClientId(resp.clientId);
  },
  token(state, token) {
    TokenService.saveToken(token);
    ApiService.setHeader();
    TokenService.destroyClientId();
  },
  user(state, user) {
    state.user = user;
  },
  logout(state) {
    state.user = null;
    state.token = '';
    state.status = '';
    TokenService.destroyToken();
    TokenService.destroyClientId();
    localStorage.removeItem('user');
  },
  AUTH_REQUEST(state) {
    state.status = 'loading';
    state.error = null;
  },
  AUTH_SUCCESS(state, { token, user }) {
    state.status = 'success';
    state.token = token;
    state.user = user;
    state.error = null;
  },
  AUTH_ERROR(state, error) {
    state.status = 'error';
    state.error = error;
  },
  AUTH_LOGOUT(state) {
    state.status = '';
    state.token = '';
    state.user = null;
  },
  SET_REDIRECT(state, url) {
    state.status = 'redirect';
    state.redirectUrl = url;
  },
  SET_AUTH_TYPE(state, isLocal) {
    state.isLocalAuth = isLocal;
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
