import { expect } from 'chai';
import { buildTree, rulesetsToTree } from '@/components/tree';

describe('tree.js buildTree', () => {
  it('builds tree correctly', () => {
    expect(buildTree('a', [], [])).to.eql({
      name: 'a',
      path: 'a',
    });

    expect(buildTree('b', ['a'], [])).to.eql({
      name: 'b',
      path: 'a/b',
    });

    expect(buildTree('b', ['a'], ['c', 'd'])).to.eql({
      name: 'b',
      children: [
        {
          name: 'c',
          children: [{ name: 'd', path: 'a/b/c/d' }],
        },
      ],
    });
  });
});

describe('tree.js rulesetsToTree', () => {
  it('builds tree correctly', () => {
    const items = rulesetsToTree([
      {
        path: 'a/b',
      },
      {
        path: 'a/c',
      },
      {
        path: 'a/c/z',
      },
      {
        path: 'a/d/e',
      },
      {
        path: 'a/d/f',
      },
    ]);
    expect(items).to.eql([
      {
        name: 'a',
        children: [
          {
            name: 'b',
            path: 'a/b',
          },
          {
            name: 'c',
            path: 'a/c',
            children: [
              {
                name: 'z',
                path: 'a/c/z',
              },
            ],
          },
          {
            name: 'd',
            children: [
              {
                name: 'e',
                path: 'a/d/e',
              },
              {
                name: 'f',
                path: 'a/d/f',
              },
            ],
          },
        ],
      },
    ]);
  });
});
