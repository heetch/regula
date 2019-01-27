<template>
  <div id="sidebar">
    <v-treeview
      v-model="tree"
      :items="items"
      activatable
      item-key="name"
    >
      <template
        slot="label"
        slot-scope="{ item }"
      >
        <div
          class="v-treeview-node__label"
          @click="navigateToLatestRulesetPage(item)"
        >{{ item.name }}</div>
      </template>
    </v-treeview>
    <div class="new-ruleset mt-5">
      <router-link to="/rulesets/new">
        <v-btn
          fab
          dark
          color="primary"
        >
          <v-icon dark>mdi-plus</v-icon>
        </v-btn>
      </router-link>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import { rulesetsToTree } from './tree';

export default {
  name: 'Sidebar',
  data: () => ({
    tree: [],
    items: [],
  }),

  mounted() {
    this.fetchRulesets();
  },

  methods: {
    fetchRulesets() {
      axios
        .get('/ui/i/rulesets/')
        .then(({ data = {} }) => {
          const { rulesets = [] } = data;

          this.items = rulesetsToTree(rulesets);
        })
        .catch(console.error);
    },

    navigateToLatestRulesetPage(item) {
      this.$router.push(`/rulesets/${item.path}/latest`);
    },
  },
};
</script>

<style lang="scss" scoped>
#sidebar {
  padding: 1em;
  overflow: auto;
  height: 100%;

  a {
    text-decoration: none;
    color: rgba(0, 0, 0, 0.87);
  }

  .new-ruleset a {
    text-decoration: none;
  }
}
</style>
