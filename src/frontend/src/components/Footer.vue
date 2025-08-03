<template>
    <v-container>
        <v-footer app>
            <v-row justify="start" no-gutters>
                <v-col cols="12" lg="6" md="12" sm="12">
                    <div :align="$vuetify.breakpoint.smAndDown ? 'center' : 'left'">
                        <small>Copyright &copy; {{ new Date().getFullYear() }}, Wg Gen Plus. </small>
                        <small>This work is licensed under the<a class="pr-1 pl-1" href="https://www.apache.org/licenses/LICENSE-2.0.html" target="_blank">Apache 2.0 License.</a></small>
                    </div>
                </v-col>
            </v-row>
            <v-row justify="end" no-gutters>
                <v-col cols="12" lg="6" md="12" sm="12">
                    <div :align="$vuetify.breakpoint.smAndDown ? 'center' : 'right'">
                        <small>Created with</small>
                        <v-icon class="pr-1 pl-1">mdi-heart</v-icon><span>by</span><a class="pr-2 pl-1" href="mailto:lindsay@127-0-0-1.au">Lindsay</a>
                        <a :href="'https://gitea.lgharvey.net/Lindsay/wg-gen-plus/commit/' + version"><kbd>Version: {{ version.substring(0,7) }}</kbd></a>
                    </div>
                </v-col>
            </v-row>
        </v-footer>
    </v-container>
</template>

<script>
  import {mapActions, mapGetters} from "vuex";

  export default {
    name: 'Footer',

    data: () => ({

    }),

    computed:{
      ...mapGetters({
        version: 'server/version',
      }),
    },

    mounted() {
      this.versionServer()
    },

    methods: {
      ...mapActions('server', {
        versionServer: 'version',
      }),
    }
  }
</script>
