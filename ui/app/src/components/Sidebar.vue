<template>
  <div id="sidebar">
    <h2>Rulesets</h2>
    <v-treeview
      v-model="tree"
      :items="items"
      activatable
      item-key="name"
      open-on-click
    >
    </v-treeview>
  </div>
</template>

<script>
// turns a list of rulesets to a tree compatible with the TreeView component
export const rulesetsToTree = (rulesets = []) => {
  const tree = {};

  // turns each ruleset path in the form a/b/c into a
  // temporary tree of this form:
  // {a: {b: {c: []}}} and fills the tree variable
  rulesets.forEach(({ path }) => {
    let node = tree;

    path.split('/').forEach((chunk, idx, list) => {
      if (!Object.prototype.hasOwnProperty.call(node, chunk)) {
        node[chunk] = idx + 1 < list.length ? {} : [];
      }

      node = node[chunk];
    });
  });

  // walk is a private function that walks through a given object and returns
  // a tree compatible with the TreeView component
  const walk = (o = {}) => Object.keys(o)
    .map(k => ({ name: k, ...(!Array.isArray(o[k]) && { children: walk(o[k]) }) }));

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
      fetch('i/rulesets/')
        .then(stream => stream.json())
        .then(({ rulesets = [] }) => {
          this.items = rulesetsToTree(rulesets);
        })
        .catch(console.error);
    },
  },
};
</script>

<style scoped>
  #sidebar {
    padding: 1em;
    overflow: auto;
    height: 100%;
  }
</style>
