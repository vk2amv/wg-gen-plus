<template>
  <v-card class="elevation-12">
    <v-card-title class="headline">Login</v-card-title>
    <v-card-text>
      <v-alert
        v-if="error"
        type="error"
        dismissible
      >
        {{ error }}
      </v-alert>
      
      <v-form @submit.prevent="login">
        <v-text-field
          v-model="username"
          label="Username"
          required
          prepend-icon="mdi-account"
          :disabled="loading"
        ></v-text-field>
        <v-text-field
          v-model="password"
          label="Password"
          type="password"
          required
          prepend-icon="mdi-lock"
          :disabled="loading"
        ></v-text-field>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" type="submit" :loading="loading">Login</v-btn>
        </v-card-actions>
      </v-form>
    </v-card-text>
  </v-card>
</template>

<script>
import { mapActions, mapState } from 'vuex';

export default {
  name: 'LoginForm',
  data: () => ({
    username: '',
    password: '',
    loading: false
  }),
  computed: {
    ...mapState('auth', ['error', 'status'])
  },
  methods: {
    ...mapActions('auth', ['local_login']),
    async login() {
      if (!this.username || !this.password) {
        this.$store.commit('auth/AUTH_ERROR', 'Username and password are required');
        return;
      }
      
      this.loading = true;
      console.log("Login attempt with:", this.username);
      
      try {
        await this.local_login({
          username: this.username,
          password: this.password
        });
        
        // Redirect after successful login
        this.$router.push('/clients');
      } catch (error) {
        console.error("Login error in component:", error);
      } finally {
        this.loading = false;
      }
    }
  }
}
</script>