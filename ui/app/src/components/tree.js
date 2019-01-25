export const buildTree = (name = '', parents = [], rest = []) => ({
  name,
  path: [...parents, name].join('/'),
  ...(rest.length > 0 && {
    children: [buildTree(rest[0], [...parents, name], rest.slice(1))],
  }),
});

const mergeDupNodes = (trees = []) => {
  const sorted = trees.sort((a, b) => a.name.localeCompare(b.name));
  let i = 0;
  let prev = {};
  const list = [];

  while (i < sorted.length) {
    if (prev && sorted[i].name === prev.name) {
      prev.path = sorted[i].path || prev.path;
      prev.children = mergeDupNodes([...(prev.children || []), ...(sorted[i].children || [])]);
    } else {
      prev = sorted[i];
      list.push(sorted[i]);
    }

    i += 1;
  }

  return list;
};

export const rulesetsToTree = (rulesets = []) => {
  const trees = rulesets.map(({ path }) => {
    const chunks = path.split('/');
    return buildTree(chunks[0], [], chunks.slice(1));
  });

  return mergeDupNodes(trees);
};
