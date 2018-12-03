import Vue from 'vue';
import './plugins/vuetify';
import './plugins/fetch';
import './plugins/font-awesome';
import App from './App';

Vue.config.productionTip = false;

new Vue({
  render: h => h(App),
}).$mount('#app');
