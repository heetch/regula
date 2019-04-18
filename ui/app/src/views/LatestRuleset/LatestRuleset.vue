<template>
  <v-container id="consult">
    <h1>{{path}}</h1>

    <v-layout>
      <v-flex md2 mt-5>
        <v-toolbar color="grey" dark>
          <v-toolbar-title>Parameters</v-toolbar-title>
        </v-toolbar>
        <v-card class="height-card scroll"
                            v-if="typeof(ruleset.signature) != 'undefined'"
>
          <v-card-text
            v-for="param in ruleset.signature.params"
            :key="param.name"
          >{{param.name}}: {{param.type}}</v-card-text>
        </v-card>
      </v-flex>

      <v-flex md2 ma-5>
        <v-toolbar color="grey" dark>
          <v-toolbar-title>Return type</v-toolbar-title>
        </v-toolbar>
        <v-card class="height-card"
                v-if="typeof(ruleset.signature) != 'undefined'"
                >
          <v-card-text>{{ruleset.signature.returnType}}</v-card-text>
        </v-card>
      </v-flex>

      <v-flex md2 ma-5>
        <v-toolbar color="grey" dark>
          <v-toolbar-title>Versions</v-toolbar-title>
        </v-toolbar>
        <v-card class="height-card scroll">
          <v-card-text v-for="version in ruleset.versions" :key="version">{{version}}</v-card-text>
        </v-card>
      </v-flex>

      <v-flex md3></v-flex>

      <v-flex md3 mt-5>
        <!-- change the link to edit page -->
        <router-link to="/rulesets/new">
          <v-btn dark color="primary">Edit</v-btn>
        </router-link>
      </v-flex>
    </v-layout>

    <!-- delegate rules to Rules component -->
    <Rules v-model="ruleset" :editMode="false"/>
  </v-container>
</template>

<script>
import axios from 'axios';
// import { Ruleset, Rule, Signature, Param } from '../NewRuleset/ruleset';
import Rules from '../NewRuleset/Rules.vue';

export default {
  components: {
    Rules,
  },

  props: {
    path: {
      type: String,
    },
  },

  data: () => ({
    ruleset: {},
  }),

  mounted() {
    this.fetchRuleset();
  },

  methods: {
    fetchRuleset() {
      const uri = '/ui/i/rulesets/' + this.path;
      return axios
        .get(uri)
        .then(({ data = {} }) => { this.ruleset = data; })
        .catch(error => error);
    },
  },
};
</script>

<style lang="scss" scoped>
#consult {
  .height-card {
    height: 200px;
  }

  .scroll {
    overflow-y: auto;
  }

  .rounded-card {
    border-radius: 50px;
  }
}
</style>
