<template>
    <v-card class="mt-3">

        <v-card-title primary-title>
          <div>
            <h3 class="title">Signature</h3>
          </div>
        </v-card-title>

        <v-card-text>
          <!-- path -->
          <v-text-field
            box
            :rules="pathRules"
            label="Path"
            required
            hint="Must start with a letter and contain letters,
              numbers, dashes or slashes. Must not end with a dash or slash."
          ></v-text-field>

          <!-- parameters  -->
          <h3 class="subheading mt-3">Parameters</h3>
          <v-layout row wrap>
            <v-flex xs12 sm6 class="pr-2">
              <v-text-field
                box
                :rules="paramNameRules"
                label="Name"
                required
                hint="Must start with a letter and contain letters,
                  numbers or dashes. Must not end with a dash."
              ></v-text-field>
            </v-flex>
            <v-flex xs12 sm6 class="pr-2">
              <v-select
                box
                :items="paramTypes"
                label="Type"
              ></v-select>
            </v-flex>
          </v-layout>
          <v-btn small fab color="secondary" class="ma-0 mt-2">
            <v-icon dark>mdi-plus</v-icon>
          </v-btn>

          <!-- return type -->
          <h3 class="subheading mt-4">Return type</h3>
          <v-select
            box
            :items="paramTypes"
            label="Type"
          ></v-select>

        </v-card-text>

      </v-card>
</template>

<script>
export default {
  name: 'Signature',
};
</script>

