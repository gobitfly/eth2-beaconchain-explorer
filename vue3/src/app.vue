<template>
  <banner></banner>
  <mainNavigation></mainNavigation>
  <main>
    <div class="container mt-2">
      <router-view v-slot="{ Component }">
        <component :is="Component" />
      </router-view>
    </div>
  </main>
  <myFooter id="footer"></myFooter>
</template>

<script setup>
import { ref, defineOptions } from "vue"

import "#/static/css/styles.css"
import "#/static/css/layout.css"
import "#/static/css/layout/banner.css"
import "#/static/css/fontawesome.min.css"
import "#/static/css/layout/toggle.css"
import "#/static/theme/css/beacon-light.min.css"

import banner from "@/components/banner.vue"
import myFooter from "@/components/footer.vue"
import mainNavigation from "@/components/mainNavigation.vue"
</script>

<style>
body,
#app {
  padding: 0;
  margin: 0;
}

#footer {
  margin-top: 3rem;
  margin-bottom: 4.2rem;
}
</style>
