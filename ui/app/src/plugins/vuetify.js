import Vue from 'vue';
import Vuetify from 'vuetify/lib';
import 'vuetify/src/stylus/app.styl';

Vue.use(Vuetify, {
  customProperties: true,
  iconfont: 'fa', // select the Font Awesome iconfont for the entire app
});
