<template>
  <div style="position: relative" class="card mt-3 index-stats">
    <template v-if="!indexData.ShowSyncingMessage">
      <div style="position: absolute; border-bottom-left-radius: 0; border-bottom-right-radius: 0; font-size: 0.7rem; height: 0.8rem" class="progress w-100" data-placement="bottom" :title="'This epoch is ' + epochCompletePercent + '% complete'">
        <div :style="'width:' + epochCompletePercent + '%;padding: 0 .3rem;'" class="progress-bar bg-secondary" role="progressbar" :aria-valuenow="scheduledCount" aria-valuemin="0" aria-valuemax="32">
          <span v-if="scheduledCount > 0">{{ scheduledCount }} / 32 slots left in epoch {{ currentEpoch }}</span>
          <span v-else>{{ currentEpoch }} epoch complete</span>
        </div>
      </div>
    </template>
    <div class="card-header pt-3">
      <div class="row">
        <networkStats></networkStats>
      </div>
    </div>
    <div class="card-body">
      <networkStatsChart></networkStatsChart>
    </div>
  </div>
</template>

<script setup>
import { useState } from "@/stores/state.js"
import networkStats from "@/components/index/networkStats.vue"
import NetworkStatsChart from "@/components/index/networkStatsChart.vue"

const state = useState()
const indexData = state.indexData
const epochCompletePercent = state.epochCompletePercent
const scheduledCount = state.scheduledCount
const currentEpoch = state.currentEpoch
</script>
