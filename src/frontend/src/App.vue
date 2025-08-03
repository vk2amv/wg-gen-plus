<template>
  <v-app id="inspire">
    <Notification v-bind:notification="notification"/>
    
    <div v-if="isAuthenticated">
      <Header/>
      <v-main>
        <v-container>
          <router-view />
        </v-container>
      </v-main>
      <Footer/>
    </div>
    <div v-else>
      <router-view />
    </div>
  </v-app>
</template>

<script>
  import Notification from './components/Notification'
  import Header from "./components/Header";
  import Footer from "./components/Footer";
  import LoginForm from "./components/LoginForm";
  import {mapActions, mapGetters} from "vuex";

  export default {
    name: 'App',

    components: {
      Footer,
      Header,
      Notification,
      LoginForm
    },

    data: () => ({
      notification: {
        show: false,
        color: '',
        text: '',
      },
    }),

    computed:{
      ...mapGetters({
        isAuthenticated: 'auth/isAuthenticated',
        authStatus: 'auth/authStatus',
        authRedirectUrl: 'auth/authRedirectUrl',
        authError: 'auth/error',
        clientError: 'client/error',
        serverError: 'server/error',
        isLocalAuth: 'auth/isLocalAuth',
      })
    },

    created () {
      this.$vuetify.theme.dark = true
    },

    async mounted() {
      // First check the auth type
      const isLocal = await this.check_auth_type();
      console.log("Auth type check result:", isLocal);
      
      // If using local auth, stop here - don't trigger oauth2 flow
      if (isLocal) {
        console.log("Local authentication enabled, skipping OAuth2 flow");
        return;
      }
      
      // Only proceed with OAuth2 if not authenticated and not local auth
      if (!this.isAuthenticated) {
        console.log("Using OAuth2 flow");
        if (this.$route.query.code && this.$route.query.state) {
          this.oauth2_exchange({
            code: this.$route.query.code,
            state: this.$route.query.state
          });
        } else {
          this.oauth2_url();
        }
      }
    },

    watch: {
      authError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      clientError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      serverError(newValue, oldValue) {
        console.log(newValue)
        this.notify('error', newValue);
      },

      isAuthenticated(newValue, oldValue) {
        console.log(`Updating isAuthenticated from ${oldValue} to ${newValue}`);
        if (newValue === true) {
          this.$router.push('/clients')
        }
      },

      authStatus(newValue, oldValue) {
        console.log(`Updating authStatus from ${oldValue} to ${newValue}`);
        if (newValue === 'redirect') {
          window.location.replace(this.authRedirectUrl)
        }
      },
    },

    methods: {
      ...mapActions('auth', {
        oauth2_exchange: 'oauth2_exchange',
        oauth2_url: 'oauth2_url',
        check_auth_type: 'check_auth_type',
      }),

      notify(color, msg) {
        this.notification.show = true;
        this.notification.color = color;
        this.notification.text = msg;
      }
    }
  };
</script>
