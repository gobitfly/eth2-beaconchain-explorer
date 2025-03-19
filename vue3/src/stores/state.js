import { defineStore } from "pinia"
import { getLaunchMetrics, getIndexData, getGitcoinfeed, getLatestState, getPageData } from "@/api/index"

export const useState = defineStore("state", {
  state() {
    return {
      launchMetrics: {},
      indexData: {},
      gitCoinfeed: {},
      latestState: {},
      pageData: {},
    }
  },
  actions: {
    async getPageData() {
      try {
        const pageData = await getPageData()
        if (pageData) {
          this.pageData = pageData
        }
      } catch (error) {
        console.log("getPageData error", error.message)
      }
    },
    async getLaunchMetrics() {
      try {
        const launchMetrics = await getLaunchMetrics()
        if (launchMetrics) {
          this.launchMetrics = launchMetrics
        }
      } catch (error) {
        console.log("getLaunchMetrics error", error.message)
      }
    },
    async getIndexData() {
      try {
        const indexData = await getIndexData()
        if (indexData) this.indexData = indexData
      } catch (error) {
        console.log("getIndexData error", error.message)
      }
    },
    async getGitcoinfeed() {
      try {
        const gitCoinfeed = await getGitcoinfeed()
        if (gitCoinfeed) this.gitCoinfeed = gitCoinfeed
      } catch (error) {
        console.log("getGitcoinfeed error", error.message)
      }
    },
    async getLatestState() {
      try {
        const latestState = await getLatestState()
        if (latestState) this.latestState = latestState
      } catch (error) {
        console.log("getLatestState error", error.message)
      }
    },
  },
  getters: {
    epochCompletePercent() {
      return ((32 - this.indexData.scheduled_count) / 32) * 100
    },
    scheduledCount() {
      return this.indexData.scheduled_count
    },
    EnteringValidators() {
      return this.indexData.entering_validators
    },
    currentEpoch() {
      return this.pageData.CurrentEpoch
    },
    participation: function () {
      if (this.indexData && this.indexData.epochs && this.indexData.epochs.length && this.indexData.epochs[0].globalparticipationrate !== 0) {
        return Math.round(this.indexData.epochs[0].globalparticipationrate * 1000) / 1000
      } else if (this.indexData && this.indexData.epochs && this.indexData.epochs.length > 1) {
        return Math.round(this.indexData.epochs[1].globalparticipationrate * 1000) / 1000
      } else {
        return 0
      }
    },
    chainGenesisTimestamp() {
      return this.pageData.ChainGenesisTimestamp
    },
    chainSecondsPerSlot() {
      return this.pageData.ChainSecondsPerSlot
    },
    chainSlotsPerEpoch() {
      return this.pageData.ChainSlotsPerEpoch
    },
  },
})
