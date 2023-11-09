import { useState } from "@/stores/state.js"
import { DateTime } from "luxon"

const state = useState()

export function addCommas(number) {
  return number
    .toString()
    .replace(/,/g, "")
    .replace(/\B(?=(\d{3})+(?!\d))/g, "<span class='thousands-separator'></span>")
}
export function slotToTime(slot) {
  var gts = state.chainGenesisTimestamp
  var sps = state.chainSecondsPerSlot
  return (gts + slot * sps) * 1000
}

export function epochToTime(epoch) {
  var gts = state.chainGenesisTimestamp
  var sps = state.chainSecondsPerSlot
  var spe = state.chainSlotsPerEpoch
  return (gts + epoch * sps * spe) * 1000
}

export function timeToEpoch(ts) {
  var gts = state.chainGenesisTimestamp
  var sps = state.chainSecondsPerSlot
  var spe = state.chainSlotsPerEpoch
  var slot = Math.floor((ts / 1000 - gts) / sps)
  var epoch = Math.floor(slot / spe)
  if (epoch < 0) return 0
  return epoch
}

export function timeToSlot(ts) {
  var gts = state.chainGenesisTimestamp
  var sps = state.chainSecondsPerSlot
  var spe = state.chainSlotsPerEpoch
  var slot = Math.floor((ts / 1000 - gts) / sps)
  if (slot < 0) return 0
  return slot
}

function formatTimestampsTooltip(local) {
  var toolTipFormat = "yyyy-MM-dd HH:mm:ss"
  var tooltip = local.toFormat(toolTipFormat)
  return tooltip
}

export function getRelativeTime(tsLuxon) {
  if (!tsLuxon) {
    return
  }
  var prefix = ""
  var suffix = ""
  if (tsLuxon.diffNow().milliseconds > 0) {
    prefix = "in "
  } else {
    // inverse the difference of the timestamp (3 seconds into the past becomes 3 seconds into the future)
    var now = DateTime.utc()
    tsLuxon = DateTime.fromSeconds(now.ts / 10e2 - tsLuxon.diffNow().milliseconds / 10e2)
    suffix = " ago"
  }
  var duration = tsLuxon.diffNow(["days", "hours", "minutes", "seconds"])
  const formattedDuration = formatLuxonDuration(duration)
  return `${prefix}${formattedDuration}${suffix}`
}
export function fromNow(date) {
  return getRelativeTime(DateTime.fromISO(date))
}

function formatLuxonDuration(duration) {
  var daysPart = Math.round(duration.days)
  var hoursPart = Math.round(duration.hours)
  var minutesPart = Math.round(duration.minutes)
  var secondsPart = Math.round(duration.seconds)
  if (daysPart === 0 && hoursPart === 0 && minutesPart === 0 && secondsPart === 0) {
    return `0 secs`
  }
  var sDays = daysPart === 1 ? "" : "s"
  var sHours = hoursPart === 1 ? "" : "s"
  var sMinutes = minutesPart === 1 ? "" : "s"
  var sSeconds = secondsPart === 1 ? "" : "s"
  var parts = []
  if (daysPart !== 0) {
    parts.push(`${daysPart} day${sDays}`)
  }
  if (hoursPart !== 0) {
    parts.push(`${hoursPart} hr${sHours}`)
  }
  if (minutesPart !== 0) {
    parts.push(`${minutesPart} min${sMinutes}`)
  }
  if (secondsPart !== 0 && parts.length == 0) {
    parts.push(`${secondsPart} sec${sSeconds}`)
  }
  if (parts.length === 1) {
    return `${parts[0]}`
  } else if (parts.length > 1) {
    return `${parts[0]} ${parts[1]}`
  } else {
    return `${duration.days}days  ${duration.hours}hrs ${duration.minutes}mins ${duration.seconds}secs`
  }
}

export function timestampTooltip(date) {
  return formatTimestampsTooltip(DateTime.fromISO(date))
}

export function formatAmount() {
  //....
}
