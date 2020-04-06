<template>
  <v-app id="inspire">
    <v-content>
      <v-container class="fill-height" fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card class="elevation-12">
              <v-toolbar color="primary" dark flat>
                <v-toolbar-title>Login</v-toolbar-title>
              </v-toolbar>
              <v-card-text>
                <v-form>
                  <v-text-field
                    label="admin"
                    name="login"
                    prepend-icon="mdi-account"
                    type="text"
                    disabled
                  />

                  <v-text-field
                    v-model="token"
                    label="Password (Token)"
                    name="password"
                    prepend-icon="mdi-lock"
                    type="password"
                  />
                </v-form>
              </v-card-text>
              <v-card-actions>
                <v-spacer />
                <v-btn color="primary" @click="login">Login</v-btn>
              </v-card-actions>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
import AuthService from "@/services/AuthService.js";

export default {
  data() {
    return {
      token: ""
    };
  },
  methods: {
    async login() {
      try {
        const response = await AuthService.login(this.token);
        console.log(response);
      } catch (error) {
        console.log(error);
      }
    }
  }
};
</script>