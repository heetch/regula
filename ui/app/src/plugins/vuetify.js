import Vue from 'vue';
import Vuetify from 'vuetify/lib';
import 'vuetify/src/stylus/app.styl';

Vue.use(Vuetify, {
  customProperties: true,
  iconfont: 'mdi', // select the Material Design Icons iconfont for the entire app
});
