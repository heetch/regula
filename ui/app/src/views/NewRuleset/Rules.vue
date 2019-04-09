<template>
  <!-- This component handles the form for the rules.-->
  <v-card class="mt-3">
    <v-card-title primary-title>
      <div>
        <h3 class="title">Rules</h3>
      </div>
    </v-card-title>

    <v-card-text>
      <!-- rules  -->
      <div v-for="(rule, index) in value.rules" :key="index">
        <h3 class="subheading mb-3">Rule {{index + 1}}</h3>
        <v-layout row wrap>
          <v-flex xs12 sm7 class="pr-2">
            <editor
              v-model="rule.sExpr"
              lang="lisp"
              theme="tomorrow"
              width="100%"
              height="100"
              :options="editorOptions"
            ></editor>
          </v-flex>
          <v-flex xs12 sm4 class="pr-2">
            <v-text-field
              box
              v-if="value.signature.returnType !== 'JSON'"
              :rules="resultsRules"
              :type="returnTypeInputType"
              label="Result"
              required
              v-model="rule.returnValue"
              :disabled="editorOptions.readOnly"
            ></v-text-field>
            <v-textarea
              box
              v-if="value.signature.returnType === 'JSON'"
              :rules="resultsRules"
              label="Result"
              required
              v-model="rule.returnValue"
              :disabled="editorOptions.readOnly"
            ></v-textarea>
          </v-flex>
          <v-flex xs12 sm1 class="text-sm-center">
            <v-btn v-if="index == 0 && editMode" small fab color="error" disabled>
              <v-icon dark>mdi-minus</v-icon>
            </v-btn>
            <v-btn v-if="index > 0 && editMode" small fab color="error" @click="removeRule(index)">
              <v-icon dark>mdi-minus</v-icon>
            </v-btn>
          </v-flex>
        </v-layout>
      </div>
      <v-btn small fab color="secondary" class="ma-0 mt-2" @click="addRule">
        <v-icon dark>mdi-plus</v-icon>
      </v-btn>
    </v-card-text>
  </v-card>
</template>


<script>
import editor from 'vue2-ace-editor';
import 'brace/ext/language_tools';
import 'brace/mode/lisp';
import 'brace/theme/tomorrow';
import { Ruleset, Rule } from './ruleset';

export default {
  name: 'Rules',

  props: {
    value: Ruleset,
    editMode: {
      type: Boolean,
      default: true,
    },
  },

  data() {
    return {
      // validation rules for S-Expressions. Only check if they're not empty.
      codeRules: [v => !!v || 'Code is required'],
      // validation rules for return values. Only check if thery're not empty.
      resultsRules: [v => !!v || 'Result is required'],
      // editor customization
      editorOptions: {
        // true: disable the edit when used from the LatestRuleset component
        readOnly: !this.editMode,
        showGutter: false,
        showLineNumbers: false,
        highlightActiveLine: false,
        fontSize: '1.5em',
      },
    };
  },

  computed: {
    // select the right input type based on the selected ruleset return type.
    // JSON is handled separately in the component html.
    returnTypeInputType() {
      switch (this.value.signature.returnType) {
        case 'Int64':
          return 'number';
        case 'Float64':
          return 'number';
        default:
          return 'text';
      }
    },
  },

  components: {
    editor,
  },

  methods: {
    // add a new rule to the ruleset when a user clicks on the + button.
    addRule() {
      this.value.rules.push(new Rule());
    },

    // remove the selected ruleset from the ruleset when a user clicks on the - button.
    removeRule(index) {
      this.value.rules.splice(index, 1);
    },
  },
};
</script>
