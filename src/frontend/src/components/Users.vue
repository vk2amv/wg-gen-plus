<template>
  <v-container>
    <!-- Error alert -->
    <v-alert
      v-if="errorVisible"
      type="error"
      prominent
      dismissible
      @input="clearError"
    >
      <div v-if="error">{{ error }}</div>
      <div v-else>An unknown error occurred</div>
    </v-alert>
    
    <v-row>
      <v-col cols="12">
        <v-card dark>
          <v-card-title>
            User Management
            <v-spacer></v-spacer>
            <v-btn
              v-if="isAdmin"
              color="primary"
              @click="openCreateDialog"
            >
              Add New User
              <v-icon right dark>mdi-account-plus</v-icon>
            </v-btn>
          </v-card-title>
          <v-data-table
            :headers="isAdmin ? adminHeaders : userHeaders"
            :items="users"
            :search="search"
            class="elevation-1"
            dark
          >
            <template v-slot:top>
              <v-text-field
                v-model="search"
                label="Search"
                class="mx-4"
                prepend-icon="mdi-magnify"
              ></v-text-field>
            </template>
            <template v-slot:item.isAdmin="{ item }">
              <v-icon v-if="item.isAdmin" color="green">mdi-check</v-icon>
              <v-icon v-else color="red">mdi-close</v-icon>
            </template>
            <template v-slot:item.actions="{ item }">
              <v-icon
                v-if="isAdmin || currentUser.sub === item.sub"
                small
                class="mr-2"
                @click="editUser(item)"
              >
                mdi-pencil
              </v-icon>
              <v-icon
                v-if="isAdmin && currentUser.sub !== item.sub"
                small
                @click="deleteUser(item)"
              >
                mdi-delete
              </v-icon>
            </template>
          </v-data-table>
        </v-card>
      </v-col>
    </v-row>

    <!-- Edit/Create User Dialog -->
    <v-dialog v-model="dialog" max-width="500px">
      <v-card dark>
        <v-card-title>
          <span class="headline">{{ formTitle }}</span>
        </v-card-title>

        <v-card-text>
          <v-container>
            <v-alert
              v-if="errorVisible"
              type="error"
              dense
              text
            >
              {{ error }}
            </v-alert>
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="editedItem.name"
                  label="Name"
                  :rules="[v => !!v || 'Name is required']"
                  required
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="editedItem.email"
                  label="Email"
                  type="email"
                ></v-text-field>
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="editedItem.password"
                  label="Password"
                  type="password"
                  :hint="passwordHint"
                  persistent-hint
                ></v-text-field>
              </v-col>
              <v-col v-if="isAdmin" cols="12">
                <v-switch
                  v-model="editedItem.isAdmin"
                  label="Administrator"
                  color="primary"
                ></v-switch>
              </v-col>
            </v-row>
          </v-container>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="error"
            text
            @click="closeDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="success"
            text
            @click="saveUser"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="500px">
      <v-card dark>
        <v-card-title class="headline">
          Delete User
        </v-card-title>
        <v-card-text>
          Are you sure you want to delete this user? This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="primary"
            text
            @click="closeDeleteDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="error"
            text
            @click="confirmDelete"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script>
import { mapActions, mapGetters } from "vuex";

export default {
  name: 'Users',

  data: () => ({
    search: '',
    dialog: false,
    deleteDialog: false,
    adminHeaders: [
      { text: 'Name', value: 'name' },
      { text: 'Email', value: 'email' },
      { text: 'Admin', value: 'isAdmin' },
      { text: 'Actions', value: 'actions', sortable: false }
    ],
    userHeaders: [
      { text: 'Name', value: 'name' },
      { text: 'Email', value: 'email' },
      { text: 'Actions', value: 'actions', sortable: false }
    ],
    editedIndex: -1,
    editedItem: {
      sub: '',
      name: '',
      email: '',
      password: '',
      isAdmin: false
    },
    defaultItem: {
      sub: '',
      name: '',
      email: '',
      password: '',
      isAdmin: false
    }
  }),

  computed: {
    ...mapGetters({
      users: 'users/users',
      currentUser: 'users/currentUser',
      error: 'users/error',
      errorVisible: 'users/errorVisible'
    }),
    formTitle() {
      return this.editedIndex === -1 ? 'New User' : 'Edit User'
    },
    isAdmin() {
      return this.currentUser && this.currentUser.isAdmin;
    },
    passwordHint() {
      if (this.editedIndex === -1) {
        return 'Enter a new password';
      } else {
        return 'Leave blank to keep current password';
      }
    }
  },

  mounted() {
    this.fetchUsers();
    // Get current user if not already loaded
    if (!this.currentUser) {
      this.getCurrentUser();
    }
  },

  methods: {
    ...mapActions('users', {
      fetchUsers: 'fetchUsers',
      createUser: 'createUser',
      updateUser: 'updateUser',
      removeUser: 'deleteUser',
      setError: 'error',
      doClearError: 'clearError'
    }),

    getCurrentUser() {
      // This would need to be implemented in your store
      // Fetch the currently logged in user's details
      this.$store.dispatch('users/fetchCurrentUser');
    },

    clearError() {
      console.log("Clearing error");
      this.doClearError();
    },

    openCreateDialog() {
      if (!this.isAdmin) return;
      
      this.editedIndex = -1;
      this.editedItem = Object.assign({}, this.defaultItem);
      this.dialog = true;
    },

    editUser(item) {
      // Only allow editing if admin or self
      if (!this.isAdmin && this.currentUser.sub !== item.sub) return;
      
      this.editedIndex = this.users.indexOf(item);
      // Clone the item but exclude the password
      this.editedItem = {
        ...item,
        password: '' // Always start with empty password field when editing
      };
      this.dialog = true;
    },

    deleteUser(item) {
      // Only admins can delete users, and they can't delete themselves
      if (!this.isAdmin || this.currentUser.sub === item.sub) return;
      
      this.editedIndex = this.users.indexOf(item);
      this.editedItem = Object.assign({}, item);
      this.deleteDialog = true;
    },

    closeDialog() {
      this.dialog = false;
      this.$nextTick(() => {
        this.editedItem = Object.assign({}, this.defaultItem);
        this.editedIndex = -1;
      });
    },

    closeDeleteDialog() {
      this.deleteDialog = false;
      this.$nextTick(() => {
        this.editedItem = Object.assign({}, this.defaultItem);
        this.editedIndex = -1;
      });
    },

    async saveUser() {
      // Clear previous errors
      this.clearError();
      
      // Validate required fields
      if (!this.editedItem.name) {
        this.setError('Name is required');
        return;
      }

      try {
        // If not admin, ensure isAdmin stays false
        if (!this.isAdmin) {
          this.editedItem.isAdmin = false;
        }
        
        // When editing, if password is empty, don't send it
        const userToSave = { ...this.editedItem };
        if (this.editedIndex > -1 && !userToSave.password) {
          delete userToSave.password;
        }
        
        if (this.editedIndex > -1) {
          // Update existing user
          await this.updateUser(userToSave);
        } else {
          // Create new user
          await this.createUser(userToSave);
        }
        this.closeDialog();
      } catch (error) {
        console.error('Error saving user:', error);
        
        // Log the full error structure for debugging
        console.log("Error object:", {
          message: error.message,
          response: error.response,
          responseData: error.response?.data,
          errorMessage: error.response?.data?.error
        });
        
        // Handle 400 errors more explicitly
        if (error.response && error.response.status === 400) {
          // Extract the error message from the response
          const errorMsg = error.response.data.error || 'The server rejected your request';
          this.setError(errorMsg);
        } else {
          // For other errors, use a generic message
          this.setError('Failed to save user: ' + (error.message || 'Unknown error'));
        }
      }
    },

    async confirmDelete() {
      try {
        if (!this.isAdmin) return;
        
        await this.removeUser(this.editedItem.sub);
        this.closeDeleteDialog();
      } catch (error) {
        console.error('Error deleting user:', error);
        
        // Handle error display
        if (error.response && error.response.data && error.response.data.error) {
          this.setError(error.response.data.error);
        } else {
          this.setError('Failed to delete user');
        }
      }
    }
  }
}
</script>