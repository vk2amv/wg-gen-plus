<template>
  <v-container class="fill-height" fluid>
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <LoginForm v-if="isLocalAuth" />
        <v-card v-else class="elevation-12">
          <v-card-title class="headline">Authentication</v-card-title>
          <v-card-text>
            <p>Please wait while we redirect you to the authentication provider...</p>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import LoginForm from "../components/LoginForm";
import { mapGetters } from "vuex";

export default {
  name: "Home",
  components: {
    LoginForm
  },
  computed: {
    ...mapGetters({
      isLocalAuth: "auth/isLocalAuth"
    })
  }
};
</script>