<template>
  <div>
    <v-navigation-drawer
      fixed
      clipped
      v-model="drawer"
      app
    >
      <Sidebar />
    </v-navigation-drawer>

    <v-toolbar
        dense
        fixed
        clipped-left
        app
    >
      <v-toolbar-side-icon @click.stop="drawer = !drawer"></v-toolbar-side-icon>
      <v-toolbar-title class="headline">
        <span>Regula</span>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        flat
        href="https://github.com/heetch/regula"
        target="_blank"
      >
        <fa-icon :icon="['fab', 'github']" size="2x" />
      </v-btn>
      <v-btn
        flat
        href="https://regula.readthedocs.io/en/latest/"
        target="_blank"
      >
        <fa-icon :icon="['fas', 'book']" size="2x" />
      </v-btn>
    </v-toolbar>
  </div>
</template>

<script>
import { library } from '@fortawesome/fontawesome-svg-core';
import { faGithub } from '@fortawesome/free-brands-svg-icons';
import { faBook } from '@fortawesome/free-solid-svg-icons';
import Sidebar from './Sidebar';

library.add(faGithub, faBook);

export default {
  name: 'Navbar',
  components: {
    Sidebar,
  },
  data() {
    return {
      drawer: null,
    };
  },
};
</script>
