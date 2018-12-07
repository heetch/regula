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
        <router-link to="/">Regula</router-link>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn
        flat
        icon
        href="https://github.com/heetch/regula"
        target="_blank"
      >
        <v-icon>mdi-github-circle</v-icon>
      </v-btn>
      <v-btn
        flat
        icon
        href="https://regula.readthedocs.io/en/latest/"
        target="_blank"
      >
        <v-icon>mdi-book-open-outline</v-icon>
      </v-btn>
    </v-toolbar>
  </div>
</template>

<script>
import Sidebar from '@/components/Sidebar.vue';

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

<style lang="scss" scoped>
.headline {
  a {
    text-decoration: none;
    color: rgba(0, 0, 0, 0.87);

    &:visited {
      color: rgba(0, 0, 0, 0.87);
    }
  }
}
</style>
