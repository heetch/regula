import { expect } from 'chai';
import { rulesetsToTree } from '@/components/Sidebar.vue';

describe('Sidebar.vue', () => {
  it('builds tree correctly', () => {
    const items = rulesetsToTree([{ path: 'a/b' }, { path: 'a/c' }, { path: 'a/d/e' }]);
    expect(items).to.eql([{
      name: 'a',
      children: [
        {
          name: 'b',
        },
        {
          name: 'c',
        },
        {
          name: 'd',
          children: [
            {
              name: 'e',
            },
          ],
        },
      ],
    }]);
  });
});
