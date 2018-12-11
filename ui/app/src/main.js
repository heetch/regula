import Vue from 'vue';
import '@mdi/font/css/materialdesignicons.css';

import './plugins/vuetify';
import App from './App';
import router from './router';

Vue.config.productionTip = false;

new Vue({
  router,
  render: h => h(App),
}).$mount('#app');
