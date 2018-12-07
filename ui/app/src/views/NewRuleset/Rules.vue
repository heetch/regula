<template>
    <v-card class="mt-3">

        <v-card-title primary-title>
          <div>
            <h3 class="title">Rules</h3>
          </div>
        </v-card-title>

        <v-card-text>
            <!-- parameters  -->
          <h3 class="subheading mb-3">Rule 1</h3>
          <v-layout row wrap>
            <v-flex xs12 sm8 class="pr-2">
              <v-textarea
                box
                label="Code"
                name="rule-1"
                value="(#true)"
              ></v-textarea>
            </v-flex>
            <v-flex xs12 sm4 class="pr-2">
              <v-text-field
                box
                :rules="pathRules"
                label="Result"
                required
              ></v-text-field>
            </v-flex>
          </v-layout>
          <v-btn small fab color="secondary" class="ma-0 mt-2">
            <v-icon dark>mdi-plus</v-icon>
          </v-btn>

        </v-card-text>

    </v-card>
</template>

<script>
export default {
  name: 'Rules',
};
</script>
