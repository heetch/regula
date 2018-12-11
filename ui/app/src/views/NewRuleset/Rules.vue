<template>
  <v-card class="mt-3">
    <v-card-title primary-title>
      <div>
        <h3 class="title">Rules</h3>
      </div>
    </v-card-title>

    <v-card-text>
      <!-- parameters  -->
      <div
        v-for="(rule, index) in value"
        :key="index"
      >
        <h3 class="subheading mb-3">Rule {{index + 1}}</h3>
        <v-layout
          row
          wrap
        >
          <v-flex
            xs12
            sm7
            class="pr-2"
          >
            <v-textarea
              box
              label="Code"
              :name="'rule-' + (index+1)"
              :rules="codeRules"
              v-model="rule.code"
            ></v-textarea>
          </v-flex>
          <v-flex
            xs12
            sm4
            class="pr-2"
          >
            <v-text-field
              box
              v-if="returnType != 'Json'"
              :rules="resultsRules"
              :type="returnTypeInputType"
              label="Result"
              required
              v-model="rule.result"
            ></v-text-field>
            <v-textarea
              box
              v-if="returnType == 'Json'"
              :rules="resultsRules"
              label="Result"
              required
              v-model="rule.result"
            ></v-textarea>
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
            >
              <v-icon dark>mdi-minus</v-icon>
            </v-btn>
            <v-btn
              v-if="index > 0"
              small
              fab
              color="error"
              @click="removeRule(index)"
            >
              <v-icon dark>mdi-minus</v-icon>
            </v-btn>
          </v-flex>
        </v-layout>
      </div>
      <v-btn
        small
        fab
        color="secondary"
        class="ma-0 mt-2"
        @click="addRule"
      >
        <v-icon dark>mdi-plus</v-icon>
      </v-btn>
    </v-card-text>
  </v-card>
</template>


<script>
export default {
  name: 'Rules',

  props: {
    value: Array,
    returnType: String,
  },

  data: () => ({
    codeRules: [v => !!v || 'Code is required'],
    resultsRules: [v => !!v || 'Result is required'],
  }),

  computed: {
    returnTypeInputType() {
      switch (this.returnType) {
        case 'String':
          return 'text';
        case 'Int64':
          return 'number';
        case 'Float64':
          return 'number';
        default:
          return 'text';
      }
    },
  },

  methods: {
    addRule() {
      this.value.push({ code: '(#true)', result: '' });
    },

    removeRule(index) {
      this.value.splice(index, 1);
    },
  },
};
</script>
