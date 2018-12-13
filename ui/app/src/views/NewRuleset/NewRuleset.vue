<template>
  <v-container class="new-ruleset ">
    <h1 class="display-1">New ruleset</h1>

    <v-form
      ref="form"
      v-model="valid"
    >

      <Signature v-model="signature" />

      <Rules
        v-model="rules"
        :return-type="signature.returnType"
      />

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
    signature: {
      path: '',
      returnType: '',
      params: [],
    },
    rules: [{ sExpr: '(#true)', returnValue: '' }],
  }),

  methods: {
    submit() {
      if (this.$refs.form.validate()) {
        axios
          .put('/ui/i/rulesets', {
            path: this.signature.path,
            signature: {
              params: this.signature.params,
              returnType: this.signature.returnType,
            },
            rules: this.rules,
          })
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
