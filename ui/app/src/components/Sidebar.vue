<template>
  <div id="sidebar">
    <v-treeview
      v-model="tree"
      :items="items"
      activatable
      item-key="name"
      open-on-click
    >
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
// turns a list of rulesets to a tree compatible with the TreeView component
export const rulesetsToTree = (rulesets = []) => {
  const tree = {};

  // turns all ruleset paths into a single tree.
  // ex: two rulesets a/b/c and a/b/d will produce the following tree
  // {a: {b: {c: [], d: []}}}
  rulesets.forEach(({ path }) => {
    let node = tree;

    path.split('/').forEach((chunk, idx, list) => {
      // check if the chunk already exists in the tree
      // to avoid overriding an existing node
      if (!Object.prototype.hasOwnProperty.call(node, chunk)) {
        node[chunk] = idx + 1 < list.length ? {} : [];
      }

      node = node[chunk];
    });
  });

  // walk is a private function that walks through a given object and returns
  // a tree compatible with the TreeView component
  const walk = (o = {}) =>
    Object.keys(o).map(k => ({
      name: k,
      ...(!Array.isArray(o[k]) && { children: walk(o[k]) }),
    }));

  return walk(tree);
};

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
      fetch('/ui/i/rulesets/')
        .then(stream => stream.json())
        .then(({ rulesets = [] }) => {
          this.items = rulesetsToTree(rulesets);
        })
        .catch(console.error);
    },
  },
};
</script>

<style lang="scss" scoped>
#sidebar {
  padding: 1em;
  overflow: auto;
  height: 100%;

  .new-ruleset a {
    text-decoration: none;
  }
}
</style>
