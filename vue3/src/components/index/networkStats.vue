<template>
  <div class="col-md-4 responsive-border-right responsive-border-right-l">
    <div class="d-flex justify-content-between">
      <div class="p-2">
        <div class="text-secondary mb-0">Epoch</div>
        <h5 class="font-weight-normal mb-0"><span data-toggle="tooltip" data-placement="top" title="The most recent epoch" v-html="addCommas(indexData.current_epoch)"></span> / <span data-toggle="tooltip" data-placement="top" title="The most recent finalized epoch" v-html="addCommas(indexData.current_finalized_epoch)"></span></h5>
      </div>
      <div class="text-right p-2">
        <div class="text-secondary mb-0">Current Slot</div>
        <h5 class="font-weight-normal mb-0">
          <span data-toggle="tooltip" data-placement="top" title="The most recent slot" v-html="addCommas(indexData.current_slot)"></span>
        </h5>
      </div>
    </div>
  </div>
  <div class="col-md-4 responsive-border-right responsive-border-right-l">
    <div class="d-flex justify-content-between">
      <div class="p-2">
        <div class="text-secondary mb-0">Active Validators</div>
        <h5 class="font-weight-normal mb-0">
          <span data-toggle="tooltip" data-placement="top" title="The number of currently active validators" v-html="addCommas(indexData.active_validators)"></span>
        </h5>
      </div>
      <div class="text-right p-2">
        <template v-if="indexData.entering_validators">
          <div class="text-secondary mb-0"><span data-toggle="tooltip" data-placement="top" title="`Currently there are no pending Validators (churn limit is ${page.ValidatorsPerEpoch} per epoch or ${page.ValidatorsPerDay} per day with $(ActiveValidators) validators)`">Pending Validators</span></div>
        </template>
        <template v-else>
          <div class="text-secondary mb-0"><span data-toggle="tooltip" data-placement="top" title="It should take at least {{ indexData.NewDepositProcessAfter }} for a new deposit to be processed and an associated validator to be activated (churn limit is {{ indexData.ValidatorsPerEpoch }} per epoch or {{ indexData.ValidatorsPerDay }} per day with {{ indexData.active_validators }} validators)">Pending Validators</span></div>
        </template>
        <h5 class="font-weight-normal mb-0">
          <span data-toggle="tooltip" data-placement="top" title="The number of validators currently waiting to enter the active validator set" v-html="addCommas(indexData.entering_validators)"></span>
          / <span data-toggle="tooltip" data-placement="top" title="The number of validators currently waiting to exit the active validator set" v-html="addCommas(indexData.exiting_validators)"></span>
        </h5>
      </div>
    </div>
  </div>
  <div class="col-md-4">
    <div class="d-flex justify-content-between">
      <div class="p-2">
        <div class="text-secondary mb-0">Staked {{ config_frontend_maincurrency }}</div>
        <h5 class="font-weight-normal mb-0">
          <span data-toggle="tooltip" data-placement="top" title="The sum of all effective balances" v-html="addCommas(indexData.staked_ether)"></span>
        </h5>
      </div>
      <div class="text-right p-2">
        <div class="text-secondary mb-0">Average Balance</div>
        <h5 class="font-weight-normal mb-0">
          <span data-toggle="tooltip" data-placement="top" title="The average current balance of all validators staked" v-html="addCommas(indexData.average_balance)"></span>
        </h5>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onServerPrefetch, onMounted, computed } from "vue"
import { useState } from "@/stores/state.js"
import { addCommas } from "@/utils.js"

const state = useState()

const indexData = computed(() => state.indexData)

const config_frontend_maincurrency = "ETH"
</script>
