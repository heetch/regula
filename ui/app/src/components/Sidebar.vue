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
export default {
  name: 'Sidebar',
  data: () => ({
    tree: [],
    items: [
      {
        name: 'a',
      },
      {
        name: 'b',
      },
      {
        name: 'c',
        children: [
          {
            name: 'd',
            children: [{
              name: 'e',
            }],
          },
        ],
      },
    ],
  }),
};
</script>

<style scoped>
  #sidebar {
    padding: 1em;
    overflow: auto;
    height: 100%;
  }
</style>
