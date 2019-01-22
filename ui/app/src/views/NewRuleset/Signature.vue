<template>
  <!-- This component handles the form for the path and signature. -->
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
        v-model="value.path"
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
        v-for="(param, index) in value.signature.params"
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
            item-text="text"
            item-value="value"
          >
            <template
              slot="selection"
              slot-scope="{ item, index }"
            >
              <span>{{ item.value | capitalize}}</span>
            </template>
            <template
              slot="item"
              slot-scope="{ item, index }"
            >
              <span>{{ item.value | capitalize }} </span>
              <span class="grey--text caption">&nbsp;&nbsp;(e.g. {{ item.hint }})</span>
            </template>
          </v-select>
        </v-flex>

        <v-flex
          xs12
          sm1
          class="text-sm-center"
        >
          <v-btn
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
        v-model="value.signature.returnType"
        item-text="text"
        item-value="value"
      >
        <template
          slot="selection"
          slot-scope="{ item, index }"
        >
          <span>{{ item.value | capitalize }}</span>
        </template>
        <template
          slot="item"
          slot-scope="{ item, index }"
        >
          <span>{{ item.value | capitalize }} </span>
          <span class="grey--text caption">&nbsp;&nbsp;(e.g. {{ item.hint }})</span>
        </template>
      </v-select>
    </v-card-text>
  </v-card>
</template>


<script>
import { Ruleset, Param } from './ruleset';

export default {
  name: 'Signature',
  props: { value: Ruleset },

  data: () => ({
    // validation rules for the Path input. Match the API server regex
    pathRules: [
      v => !!v || 'Path is required',
      v => /^[a-z]+(?:[a-z0-9-/]?[a-z0-9])*$/.test(v) || 'Path must be valid',
    ],
    // validation rules for the param names. Match the API server regex
    paramNameRules: [
      v => !!v || 'Param name is required',
      v =>
        /^[a-z]+(?:[a-z0-9-]?[a-z0-9])*$/.test(v) || 'Param name must be valid',
    ],
    // List of allowed parameter types
    paramTypes: [
      { value: 'int64', hint: '123' },
      { value: 'float64', hint: '123.45' },
      { value: 'bool', hint: 'true' },
      { value: 'string', hint: 'foobar' },
    ],
    // List of allowed return types
    returnTypes: [
      { value: 'int64', hint: '123' },
      { value: 'float64', hint: '123.45' },
      { value: 'bool', hint: 'true' },
      { value: 'string', hint: 'foobar' },
      { value: 'JSON', hint: '{"foo": "bar"}' },
    ],
  }),

  methods: {
    // adds a param to the list when the user clicks on the + button.
    addParam() {
      this.value.signature.params.push(new Param());
    },

    // removes a param from the list when the user clicks on the - button.
    removeParam(index) {
      this.value.signature.params.splice(index, 1);
    },
  },
};
</script>

