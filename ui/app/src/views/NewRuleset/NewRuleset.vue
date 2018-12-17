
<template>
  <!-- This page displays the form to create a new ruleset -->
  <v-container class="new-ruleset ">
    <h1 class="display-1">New ruleset</h1>

    <v-form
      ref="form"
      v-model="valid"
    >

      <!-- delegate path and signature form to Signature component -->
      <Signature v-model="ruleset" />

      <!-- delegate rules form to Rules component -->
      <Rules v-model="ruleset" />

      <div class="text-xs-right mt-3">
        <v-btn
          :disabled="!valid"
          @click="submit"
          color="primary"
        >
          create new ruleset
        </v-btn>
      </div>
    </v-form>

  </v-container>
</template>

<script>
import axios from 'axios';
import { Ruleset, Rule } from './ruleset';
import Signature from './Signature.vue';
import Rules from './Rules.vue';

export default {
  name: 'NewRuleset',

  components: {
    Signature,
    Rules,
  },

  data: () => ({
    valid: false,
    ruleset: new Ruleset({
      // start with one rule with default values
      rules: [new Rule()],
    }),
  }),

  methods: {
    submit() {
      if (this.$refs.form.validate()) {
        axios
          .put('/ui/i/rulesets', this.ruleset)
          // temporary error handling until we implement the backend endpoint
          .then(console.log)
          .catch(console.error);
      } else {
        console.log('Invalid form');
      }
    },
  },
};
</script>
