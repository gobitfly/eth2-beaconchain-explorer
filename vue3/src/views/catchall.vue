<template lang="pug">
.container
  router-link(to="/") 
    h1 back to INDEX {{ updateIn }}
   
  
</template>

<script setup>
import { onServerPrefetch, onMounted, computed, defineComponent, ref, defineOptions } from "vue"
import { useState } from "@/stores/state.js"

const state = useState()
const indexData = state.indexData
const epochCompletePercent = state.epochCompletePercent
const scheduledCount = state.scheduledCount
const currentEpoch = state.currentEpoch
const updateIn = ref(-1)

defineOptions({
  asyncData: () => {
    const state = useState()
    return Promise.all([state.getIndexData(), state.getPageData()])
  },
})
</script>

<style>
.container display flex flex-wrap wrap @media (max-width: 960px) {
  .index-stats {
    font-size: 0.9rem;
  }

  .index-stats h5,
  .card-title {
    font-size: 1rem !important;
  }
}

[v-cloak] {
  visibility: hidden;
}

.responsive-border-right {
  border-right-color: rgb(222, 226, 230);
  border-right-style: solid;
  border-right-width: 1px;
}

@media (max-width: 767px) {
  .responsive-border-right-l {
    border: hidden;
  }
}
</style>
