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
        v-model="path"
        :rules="pathRules"
        label="Path"
        required
        hint="Must start with a letter and may contain letters,
                numbers, dashes or slashes. Must not end with a dash or slash."
      ></v-text-field>

      <!-- parameters  -->
      <h3 class="subheading mt-3">Parameters</h3>

      <v-layout
        row
        wrap
        v-for="(param, index) in params"
        :key="index"
      >
        <v-flex
          xs12
          sm6
          class="pr-2"
        >
          <v-text-field
            box
            :rules="paramNameRules"
            label="Name"
            v-model="param.name"
            required
            hint="Must start with a letter and may contain letters,
                  numbers or dashes. Must not end with a dash."
          ></v-text-field>
        </v-flex>

        <v-flex
          xs12
          sm5
          class="pr-2"
        >
          <v-select
            box
            :items="paramTypes"
            label="Type"
            v-model="param.type"
          ></v-select>
        </v-flex>

        <v-flex
          xs12
          sm1
          class="text-sm-center"
        >
          <v-btn
            v-if="index == 0"
            small
            fab
            color="error"
            disabled
            @click="removeParam(index)"
          >
            <v-icon dark>mdi-minus</v-icon>
          </v-btn>
          <v-btn
            v-if="index > 0"
            small
            fab
            color="error"
            @click="removeParam(index)"
          >
            <v-icon dark>mdi-minus</v-icon>
          </v-btn>
        </v-flex>
      </v-layout>

      <v-btn
        small
        fab
        color="secondary"
        class="ma-0 mt-2"
        @click="addParam"
      >
        <v-icon dark>mdi-plus</v-icon>
      </v-btn>

      <!-- return type -->
      <h3 class="subheading mt-4">Return type</h3>
      <v-select
        box
        :items="returnTypes"
        label="Type"
      ></v-select>
    </v-card-text>
  </v-card>
</template>


<script>
export default {
  name: 'Signature',

  data: () => ({
    path: '',
    pathRules: [
      v => !!v || 'Path is required',
      v => /^[a-z]+(?:[a-z0-9-/]?[a-z0-9])*$/.test(v) || 'Path must be valid',
    ],
    params: [{ name: '', type: '' }],
    paramNameRules: [
      v => !!v || 'Param name is required',
      v =>
        /^[a-z]+(?:[a-z0-9-]?[a-z0-9])*$/.test(v) || 'Param name must be valid',
    ],
    paramTypes: ['Int64', 'Float64', 'Bool', 'String'],
    returnTypes: ['Int64', 'Float64', 'Bool', 'String', 'Json'],
  }),

  methods: {
    addParam() {
      this.params.push({ name: '', type: '' });
    },

    removeParam(index) {
      this.params.splice(index, 1);
    },
  },
};
</script>

