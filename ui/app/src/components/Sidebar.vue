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
export const rulesetsToTree = (rulesets = []) => {
  const tree = {};

  rulesets.forEach(({ path }) => {
    let node = tree;

    path.split('/').forEach((chunk, idx, list) => {
      if (!Object.prototype.hasOwnProperty.call(node, chunk)) {
        node[chunk] = idx + 1 < list.length ? {} : [];
      }

      node = node[chunk];
    });
  });

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
      fetch('/i/rulesets/')
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
