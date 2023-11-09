<template>
  <clientOnly>
    <highcharts :options="chartOptions"></highcharts>
  </clientOnly>
</template>

<script setup>
/* highcharts is globally importet only on client in entry-client.js */

import { computed } from "vue"
import { useState } from "@/stores/state.js"
import { timeToEpoch } from "@/utils.js"
import clientOnly from "@/components/clientOnly.vue"

const state = useState()

const creditsEnabled = state.indexData.slotVizData
const MainCurrency = state.indexData.MainCurrency
const StakedEtherChartData = state.indexData.staked_ether_chart_data
const ActiveValidatorsChartData = state.indexData.active_validators_chart_data

const config_frontend_maincurrency = "ETH"

const chartOptions = {
  credits: { enabled: creditsEnabled },
  rangeSelector: { enabled: false },
  navigator: { enabled: false },
  scrollbar: { enabled: false },
  chart: { type: "spline" },
  title: { text: "Network History" },
  subtitle: {},
  xAxis: {
    type: "datetime",
    range: 7 * 24 * 60 * 60 * 1000,
    labels: {
      formatter: function () {
        var epoch = timeToEpoch(this.value)
        var orig = this.axis.defaultLabelFormatter.call(this)
        return `${orig}<br/>Epoch ${epoch}`
      },
    },
  },
  yAxis: [
    {
      title: { text: "Balance" + MainCurrency },
      labels: {
        formatter: function () {
          return this.value.toFixed(0)
        },
      },
      opposite: false,
    },
    {
      title: { text: "Active Validators" },
      labels: {
        formatter: function () {
          return this.value.toFixed(0)
        },
      },
      opposite: true,
    },
  ],
  series: [
    {
      name: "Staked " + config_frontend_maincurrency,
      yAxis: 0,
      data: StakedEtherChartData,
    },
    {
      name: "Active Validators",
      yAxis: 1,
      data: ActiveValidatorsChartData,
    },
  ],
  legend: { enabled: true },
  tooltip: {
    valueDecimals: 0,
    formatter: function (tooltip) {
      var orig = tooltip.defaultFormatter.call(this, tooltip)
      var epoch = timeToEpoch(this.x)
      orig[0] = `${orig[0]}<span style="font-size:10px">Epoch ${epoch}</span>`
      return orig
    },
  },
}
</script>
