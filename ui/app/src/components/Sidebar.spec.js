import { expect } from 'chai';
import Sidebar from '@/components/Sidebar.vue';

describe('Sidebar.vue', () => {
  it('renders props.msg when passed', () => {
    Sidebar.methods.rulesetsToTree({ rulesets: [{ path: 'a/b' }, { path: 'a/c' }, { path: 'a/d/e' }] });
    expect(Sidebar.data.items).to.equal([{
      name: 'a',
      children: {
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
      },
    }]);
  });
});
