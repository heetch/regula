// buildTree is a recursive function that takes a path split into chunks
// and returns a tree:
//
// For a given path 'a/b/c' it must return the following tree
// {name: 'a', children: [{ name: 'b', children: [ { name: 'c', path: 'a/b/c' } ] }] }
export const buildTree = (name = '', parents = [], rest = []) => ({
  name,
  ...(rest.length === 0 && { path: [...parents, name].join('/') }),
  ...(rest.length > 0 && {
    children: [buildTree(rest[0], [...parents, name], rest.slice(1))],
  }),
});

// mergeTrees takes a list of trees and merges them into a single one.
const mergeTrees = (trees = []) => {
  const sorted = trees.sort((a, b) => a.name.localeCompare(b.name));
  let i = 0;
  let prev = {};
  const list = [];

  while (i < sorted.length) {
    if (prev && sorted[i].name === prev.name) {
      const path = sorted[i].path || prev.path;
      if (path) {
        prev.path = path;
      }
      prev.children = mergeTrees([...(prev.children || []), ...(sorted[i].children || [])]);
    } else {
      prev = sorted[i];
      list.push(sorted[i]);
    }

    i += 1;
  }

  return list;
};

// rulesetsToTree takes a list of rulesets and returns a tree compatible
// with the Vuetify Treeview component format.
// In order to do that, it must split the paths, generate several trees
// then merge these trees into a single one and return it.
export const rulesetsToTree = (rulesets = []) => {
  const trees = rulesets.map(({ path }) => {
    const chunks = path.split('/');
    return buildTree(chunks[0], [], chunks.slice(1));
  });

  return mergeTrees(trees);
};
