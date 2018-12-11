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
      params: [{ name: '', type: '' }],
    },
    rules: [{ code: '(#true)', result: '' }],
  }),

  methods: {
    submit() {
      if (this.$refs.form.validate()) {
        console.log(this.signature);
        console.log(this.rules);
      } else {
        console.log('Invalid form');
      }
    },
  },
};
</script>
