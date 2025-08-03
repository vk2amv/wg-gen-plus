import ApiService from "../../services/api.service";

const state = {
  error: null,
  errorVisible: false,
  users: [],
  currentUser: null
}

const getters = {
  error(state) {
    return state.error;
  },
  
  errorVisible(state) {
    return state.errorVisible;
  },

  users(state) {
    return state.users;
  },

  currentUser(state) {
    return state.currentUser;
  }
}

const actions = {
  error({ commit }, errorMessage) {
    // Reset error state first
    commit('setErrorVisible', false);
    commit('setError', '');
    
    // Set a small timeout to ensure UI updates
    setTimeout(() => {
      let message = 'Unknown error occurred';
      
      // Check for different error formats
      if (errorMessage?.response?.data?.error) {
        message = errorMessage.response.data.error;
      } else if (errorMessage?.response?.data?.message) {
        message = errorMessage.response.data.message;
      } else if (errorMessage?.message) {
        message = errorMessage.message;
      } else if (typeof errorMessage === 'string') {
        message = errorMessage;
      }
      
      console.log("Setting error message:", message);
      
      // Set the error and make it visible
      commit('setError', message);
      commit('setErrorVisible', true);
      
      // Auto-dismiss after 5 seconds
      setTimeout(() => {
        commit('setErrorVisible', false);
      }, 5000);
    }, 10);
  },
  
  fetchUsers({ commit }) {
    return ApiService.get("/users")
      .then(resp => {
        commit('setUsers', resp)
      })
      .catch(err => {
        console.error("Fetch users error:", err.response?.data);
        commit('error', err);
      })
  },

  fetchUser({ commit }, id) {
    return ApiService.get(`/users/${id}`)
      .then(resp => {
        commit('setCurrentUser', resp)
        return resp
      })
      .catch(err => {
        console.error("Fetch user error:", err.response?.data);
        commit('error', err);
        throw err
      })
  },

  createUser({ commit, dispatch }, user) {
    return ApiService.post('/users', user)
      .then(resp => {
        dispatch('fetchUsers')
        return resp
      })
      .catch(err => {
        console.log("Create user error - raw:", err);
        
        // Extract the specific error message from the response
        let errorMessage = "Failed to create user";
        
        if (err.response && err.response.data && err.response.data.error) {
          // This is the format your backend returns: { "error": "message" }
          errorMessage = err.response.data.error;
        } else if (err.message) {
          errorMessage = err.message;
        }
        
        console.log("Error message to display:", errorMessage);
        
        // Set the actual error message, not the full error object
        commit('setError', errorMessage);
        commit('setErrorVisible', true);
        
        throw err; // Re-throw for component handling
      })
  },

  updateUser({ commit, dispatch }, user) {
    return ApiService.patch(`/users/${user.sub}`, user)
      .then(resp => {
        dispatch('fetchUsers')
        return resp
      })
      .catch(err => {
        console.error("Update user error:", err.response?.data);
        
        // Extract the specific error message
        let errorMessage = "Failed to update user";
        
        if (err.response && err.response.data && err.response.data.error) {
          errorMessage = err.response.data.error;
        } else if (err.message) {
          errorMessage = err.message;
        }
        
        commit('setError', errorMessage);
        commit('setErrorVisible', true);
        
        throw err
      })
  },

  deleteUser({ commit, dispatch }, id) {
    return ApiService.delete(`/users/${id}`)
      .then(() => {
        dispatch('fetchUsers')
      })
      .catch(err => {
        console.error("Delete user error:", err.response?.data);
        
        // Extract the specific error message
        let errorMessage = "Failed to delete user";
        
        if (err.response && err.response.data && err.response.data.error) {
          errorMessage = err.response.data.error;
        } else if (err.message) {
          errorMessage = err.message;
        }
        
        commit('setError', errorMessage);
        commit('setErrorVisible', true);
        
        throw err
      })
  },

  clearError({ commit }) {
    commit('setError', '');
    commit('setErrorVisible', false);
  },

  fetchCurrentUser({ commit }) {
    return ApiService.get("/users/me")  // Assuming you have an endpoint for current user
      .then(resp => {
        commit('setCurrentUser', resp);
        return resp;
      })
      .catch(err => {
        console.error("Fetch current user error:", err.response?.data);
        commit('error', err);
        
        // If we can't fetch the current user, set a default with no admin rights
        commit('setCurrentUser', {
          sub: '',
          name: 'Unknown User',
          email: '',
          isAdmin: false
        });
      });
  },
}

const mutations = {
  setError(state, error) {
    state.error = error;
  },
  
  setErrorVisible(state, visible) {
    state.errorVisible = visible;
  },

  setUsers(state, users) {
    state.users = users
  },

  setCurrentUser(state, user) {
    state.currentUser = user
  }
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}