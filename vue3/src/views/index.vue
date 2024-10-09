<template>
  <heroContainer></heroContainer>
  <postGenesis></postGenesis>
  <div class="row">
    <div class="col-lg-6 mt-3 pr-lg-2">
      <recentEpochs :updateIn="updateIn"></recentEpochs>
    </div>
    <div class="col-lg-6 mt-3 pr-lg-2">
      <recentBlocks :updateIn="updateIn"></recentBlocks>
    </div>
  </div>
  <div class="row">
    <div class="col-lg-12">
      <small style="position: absolute; right: 0" class="float-right pt-md-0 pt-2 m-2"> Next update in {{ updateIn }} s</small>
    </div>
  </div>
</template>

<script>
import { onServerPrefetch, onMounted, computed, defineComponent, ref, defineOptions } from "vue"
import { useState } from "@/stores/state.js"

import NetworkStatsChart from "@/components/index/networkStatsChart.vue"
import heroContainer from "@/components/index/heroContainer.vue"
import recentBlocks from "@/components/index/recentBlocks.vue"
import recentEpochs from "@/components/index/recentEpochs.vue"
import networkStats from "@/components/index/networkStats.vue"
import postGenesis from "@/components/index/postGenesis.vue"

export default defineComponent({
  setup() {
    const state = useState()
    const indexData = state.indexData
    const epochCompletePercent = state.epochCompletePercent
    const scheduledCount = state.scheduledCount
    const currentEpoch = state.currentEpoch
    const updateIn = ref(-1)

    return { indexData, scheduledCount, epochCompletePercent, scheduledCount, currentEpoch, updateIn }
  },
  asyncData() {
    const state = useState()
    return Promise.all([state.getIndexData(), state.getPageData()])
  },
  components: { heroContainer, recentBlocks, recentEpochs, networkStats, NetworkStatsChart, postGenesis },
  mounted() {
    const state = useState()
    state.getIndexData()
    setInterval(() => {
      this.tick()
    }, 1000)
  },
  methods: {
    tick() {
      const state = useState()
      if (this.updateIn <= 0) {
        state.getIndexData().then((res) => {
          this.updateIn = 10
        })
      } else {
        this.updateIn--
      }
    },
  },
})
</script>

<style>
@media (max-width: 960px) {
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
