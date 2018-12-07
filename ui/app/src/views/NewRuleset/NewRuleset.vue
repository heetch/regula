<template>
  <v-container class="new-ruleset ">
    <h1 class="display-1">New ruleset</h1>

    <v-form>

      <Signature/>

      <Rules/>

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
};
</script>
