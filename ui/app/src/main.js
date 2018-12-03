import Vue from 'vue';
import './plugins/vuetify';
import './plugins/fetch';
import './plugins/font-awesome';
import App from './App';
import router from './router';

Vue.config.productionTip = false;

new Vue({
  router,
  render: h => h(App),
}).$mount('#app');
