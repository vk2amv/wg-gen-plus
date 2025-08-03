import Vue from "vue";
import TokenService from "./token.service";

const ApiService = {

  setHeader() {
    const token = TokenService.getToken();
    if (token) {
      Vue.axios.defaults.headers.common['x-wg-gen-plus-auth'] = token;
    }
  },

  get(resource) {
    return Vue.axios.get(resource)
      .then(response => response.data)
      .catch(error => {
        console.error("API GET error:", error);
        // Don't wrap the error, just rethrow it
        throw error;
      });
  },

  post(resource, params) {
    this.setHeader();
    
    // Ensure content type is set correctly for JSON data
    const headers = { 'Content-Type': 'application/json' };
    
    return Vue.axios.post(resource, params, { headers })
      .then(response => response.data)
      .catch(error => {
        console.error("API POST error:", error);
        console.error("Request was:", { resource, params });
        console.error("Response was:", error.response?.data);
        throw error;
      });
  },

  put(resource, params) {
    return Vue.axios.put(resource, params)
      .then(response => response.data)
      .catch(error => {
        console.error("API PUT error:", error);
        throw error;
      });
  },

  patch(resource, params) {
    return Vue.axios.patch(resource, params)
      .then(response => response.data)
      .catch(error => {
        console.error("API PATCH error:", error);
        throw error;
      });
  },

  delete(resource) {
    return Vue.axios.delete(resource)
      .then(response => response.data)
      .catch(error => {
        console.error("API DELETE error:", error);
        throw error;
      });
  },

  getWithConfig(resource, config) {
    return Vue.axios.get(resource, config)
      .then(response => response.data)
      .catch(error => {
        console.error("API GET with config error:", error);
        throw error;
      });
  },
};

export default ApiService;
