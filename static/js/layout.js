//We want to prevent the intial page scroll to tab anchors
function stopInitialScrollEvent(event) {
  event.preventDefault()
  event.stopImmediatePropagation()
  event.stopPropagation()
  window.scrollTo(0, 0)
}
window.addEventListener("scroll", stopInitialScrollEvent)
window.addEventListener("load", function (event) {
  window.removeEventListener("scroll", stopInitialScrollEvent)
})

// FAB toggle
function toggleFAB() {
  var fabContainer = document.querySelector(".fab-message")
  var fabButton = fabContainer.querySelector(".fab-message-button a")
  // var fabToggle = document.getElementById('fab-message-toggle')
  fabContainer.classList.toggle("is-open")
  fabButton.classList.toggle("toggle-icon")
}
$(document).ready(function () {
  var fabContainer = document.querySelector(".fab-message")
  var messages = document.querySelector(".fab-message-content h3")
  if (messages) {
    fabContainer.style.display = "initial"
  }
})

// Theme switch
function switchTheme(e) {
  var d1 = document.getElementById("app-theme")
  //checked is light
  if (e.target.checked) {
    d1.href = "/theme/css/beacon-light.min.css"
    document.documentElement.setAttribute("data-theme", "light")
    localStorage.setItem("theme", "light")
  } else {
    // dark theme
    d1.href = "/theme/css/beacon-dark.min.css"
    document.documentElement.setAttribute("data-theme", "dark")
    localStorage.setItem("theme", "dark")
  }
}
$("#toggleSwitch").on("change", switchTheme)

function hideInfoBanner(msg) {
  localStorage.setItem("infoBannerStatus", msg)
  $("#infoBanner").attr("class", "d-none")
}
// $("#infoBannerDissBtn").on('click', hideInfoBanner)

function setValidatorEffectiveness(elem, eff) {
  if (elem === undefined) return
  eff = parseInt(eff)
  if (eff >= 100) {
    $("#" + elem).html(`<span class="text-success"> ${eff}% - Perfect <i class="fas fa-grin-stars"></i>`)
  } else if (eff > 80) {
    $("#" + elem).html(`<span class="text-success"> ${eff}% - Good <i class="fas fa-smile"></i></span>`)
  } else if (eff > 60) {
    $("#" + elem).html(`<span class="text-warning"> ${eff}% - Fair <i class="fas fa-meh"></i></span>`)
  } else {
    $("#" + elem).html(`<span class="text-danger"> ${eff}% - Bad <i class="fas fa-frown"></i></span>`)
  }
}

function setUtc() {
  if ($("#optionLocal").is(":checked") || $("#optionTs").is(":checked")) {
    var unixTs = $("#unixTs").text()
    var ts = luxon.DateTime.fromMillis(unixTs * 1000)
    $("#timestamp").text(ts.toUTC().toFormat("MMM-dd-yyyy hh:mm:ss a"))
  }
}

function setLocal() {
  if ($("#optionUtc").is(":checked") || $("#optionTs").is(":checked")) {
    var unixTs = $("#unixTs").text()
    var ts = luxon.DateTime.fromMillis(unixTs * 1000)
    $("#timestamp").text(ts.toFormat("MMM-dd-yyyy HH:mm:ss") + " UTC" + ts.toFormat("Z"))
  }
}

function setTs() {
  var unixTs = $("#unixTs").text()
  var utc = luxon.DateTime.fromMillis(unixTs * 1000)
  $("#timestamp").text(utc["ts"] / 1000)
}

function copyTs() {
  var text = $("#timestamp").text()
  tsArr = text.split(" ")
  if (tsArr.length > 1) {
    navigator.clipboard.writeText(tsArr[0] + " " + tsArr[1])
  } else {
    navigator.clipboard.writeText(tsArr[0])
  }
}

function viewHexDataAs(id, type) {
  var extraDataHex = $(`#${id}`).attr("aria-hex-data")
  if (!extraDataHex) {
    return
  }

  if (type === "hex") {
    $(`#${id}`).text(extraDataHex)
  } else {
    try {
      var r = decodeURIComponent(extraDataHex.replace(/\s+/g, "").replace(/[0-9a-f]{2}/g, "%$&"))
      $(`#${id}`).text(r.replace("0x", ""))
    } catch (e) {
      $(`#${id}`).text(hex2a(extraDataHex.replace("0x", "")))
    }
  }
}

function hex2a(hexx) {
  var hex = hexx.toString() //force conversion
  var str = ""
  for (var i = 0; i < hex.length; i += 2) str += String.fromCharCode(parseInt(hex.substr(i, 2), 16))
  return str
}

// typeahead
$(document).ready(function () {
  // format timestamps within tooltip titles
  formatTimestamps() // make sure this happens before tooltips
  if ($('[data-toggle="tooltip"]').tooltip) {
    $('[data-toggle="tooltip"]').tooltip()
  }

  // set maxParallelRequests to number of datasets queried in each search
  // make sure this is set in every one bloodhound object
  let requestNum = 8

  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.pubkey
    },
    remote: {
      url: "/search/validators/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  var bhSlots = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.slot
    },
    remote: {
      url: "/search/slots/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  var bhBlocks = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.block
    },
    remote: {
      url: "/search/blocks/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  var bhTransactions = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.txhash
    },
    remote: {
      url: "/search/transactions/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  var bhGraffiti = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.graffiti
    },
    remote: {
      url: "/search/graffiti/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  var bhEpochs = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.epoch
    },
    remote: {
      url: "/search/epochs/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  var bhEth1Accounts = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.account
    },
    remote: {
      url: "/search/eth1_addresses/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  var bhValidatorsByAddress = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.eth1_address
    },
    remote: {
      url: "/search/count_indexed_validators_by_eth1_address/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })

  // before adding datasets make sure requestNum is set to the correct value
  $(".typeahead").typeahead(
    {
      minLength: 1,
      highlight: true,
      hint: false,
      autoselect: false,
    },
    {
      limit: 5,
      name: "validators",
      source: bhValidators,
      display: "pubkey",
      templates: {
        header: '<h3 class="h5">Validators</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">${data.index}: ${data.pubkey}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "blocks",
      source: bhBlocks,
      display: "hash",
      templates: {
        header: '<h3 class="h5">Blocks</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">${data.block}: ${data.hash}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "slots",
      source: bhSlots,
      display: "blockroot",
      templates: {
        header: '<h3 class="h5">Slots</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">${data.slot}: 0x${data.blockroot}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "transactions",
      source: bhTransactions,
      display: "txhash",
      templates: {
        header: '<h3 class="h5">Transactions</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">0x${data.txhash}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "epochs",
      source: bhEpochs,
      display: "epoch",
      templates: {
        header: '<h3 class="h5">Epochs</h3>',
        suggestion: function (data) {
          return `<div>${data.epoch}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "addresses",
      source: bhEth1Accounts,
      display: "address",
      templates: {
        header: '<h3 class="h5">Address</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">0x${data.address}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "validators-by-address",
      source: bhValidatorsByAddress,
      display: "eth1_address",
      templates: {
        header: '<h3 class="h5">Validators by Address</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">${data.count}: 0x${data.eth1_address}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "graffiti",
      source: bhGraffiti,
      display: "graffiti",
      templates: {
        header: '<h3 class="h5">Blocks by Graffiti</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.graffiti}</div><div style="max-width:fit-content;white-space:nowrap;">${data.count}</div></div>`
        },
      },
    }
  )

  $(".typeahead").on("focus", function (event) {
    if (event.target.value !== "") {
      $(this).trigger(
        $.Event("keydown", {
          keyCode: 40,
        })
      )
    }
  })

  $(".typeahead").on("input", function (input) {
    $(".tt-suggestion").first().addClass("tt-cursor")
  })

  $(".tt-menu").on("mouseenter", function () {
    $(".tt-suggestion").first().removeClass("tt-cursor")
  })

  $(".tt-menu").on("mouseleave", function () {
    $(".tt-suggestion").first().addClass("tt-cursor")
  })

  $(".typeahead").on("typeahead:select", function (ev, sug) {
    if (sug.txhash !== undefined) {
      window.location = "/tx/" + sug.txhash
    } else if (sug.block !== undefined) {
      window.location = "/block/" + sug.block
    } else if (sug.slot !== undefined) {
      window.location = "/slot/" + sug.slot
    } else if (sug.index !== undefined) {
      if (sug.index === "deposited") window.location = "/validator/" + sug.pubkey
      else window.location = "/validator/" + sug.index
    } else if (sug.epoch !== undefined) {
      window.location = "/epoch/" + sug.epoch
    } else if (sug.address !== undefined) {
      window.location = "/address/" + sug.address
    } else if (sug.eth1_address !== undefined) {
      window.location = "/validators/initiated-deposits?q=" + sug.eth1_address
    } else if (sug.graffiti !== undefined) {
      // sug.graffiti is html-escaped to prevent xss, we need to unescape it
      var el = document.createElement("textarea")
      el.innerHTML = sug.graffiti
      window.location = "/slots?q=" + encodeURIComponent(el.value)
    } else {
      console.log("invalid typeahead-selection", sug)
    }
  })
})

$("[aria-ethereum-date]").each(function () {
  formatAriaEthereumDate(this)
})

$("[aria-ethereum-duration]").each(function () {
  formatAriaEthereumDuration(this)
})

$(document).ready(function () {
  var clipboard = new ClipboardJS("[data-clipboard-text]")
  clipboard.on("success", function (e) {
    var title = $(e.trigger).attr("data-original-title")
    $(e.trigger).tooltip("hide").attr("data-original-title", "Copied!").tooltip("show")

    setTimeout(function () {
      $(e.trigger).tooltip("hide").attr("data-original-title", title)
    }, 1000)
  })

  clipboard.on("error", function (e) {
    var title = $(e.trigger).attr("data-original-title")
    $(e.trigger).tooltip("hide").attr("data-original-title", "Failed to Copy!").tooltip("show")

    setTimeout(function () {
      $(e.trigger).tooltip("hide").attr("data-original-title", title)
    }, 1000)
  })
})

// With HTML5 history API, we can easily prevent scrolling!
$(".nav-tabs a").on("shown.bs.tab", function (e) {
  if (history.replaceState) {
    history.pushState(null, null, e.target.hash)
  } else {
    window.location.hash = e.target.hash //Polyfill for old browsers
  }
})

$(".nav-pills a").on("shown.bs.tab", function (e) {
  if (history.replaceState) {
    history.pushState(null, null, e.target.hash)
  } else {
    window.location.hash = e.target.hash //Polyfill for old browsers
  }
})

// Javascript to enable link to tab
var url = document.location.toString()
if (url.match("#")) {
  $('.nav-tabs a[href="#' + url.split("#")[1] + '"]').tab("show")
  $('.nav-pills a[href="#' + url.split("#")[1] + '"]').tab("show")
}

function formatAriaEthereumDate(elem) {
  var dt = $(elem).attr("aria-ethereum-date")
  var format = $(elem).attr("aria-ethereum-date-format")

  if (!format) {
    format = "ff"
  }

  if (format === "FROMNOW") {
    $(elem).text(getRelativeTime(luxon.DateTime.fromMillis(dt * 1000)))
    $(elem).attr("title", luxon.DateTime.fromMillis(dt * 1000).toFormat("ff"))
    $(elem).attr("data-toggle", "tooltip")
  } else if (format === "LOCAL") {
    var local = luxon.DateTime.fromMillis(dt * 1000)
    $(elem).text(local.toFormat("MMM-dd-yyyy HH:mm:ss") + " UTC" + local.toFormat("Z"))
    $(elem).attr("title", luxon.DateTime.fromMillis(dt * 1000).toFormat("ff"))
    $(elem).attr("data-toggle", "tooltip")
  } else {
    $(elem).text(luxon.DateTime.fromMillis(dt * 1000).toFormat(format))
  }
}

function formatTimestamps(selStr) {
  var sel = $(document)
  if (selStr !== undefined) {
    sel = $(selStr)
  }
  sel.find(".timestamp").each(function () {
    var ts = $(this).data("timestamp")
    var tsLuxon = luxon.DateTime.fromMillis(ts * 1000)
    $(this).attr("data-original-title", tsLuxon.toFormat("ff"))

    $(this).text(getRelativeTime(tsLuxon))
  })

  if (sel.find('[data-toggle="tooltip"]').tooltip) {
    sel.find('[data-toggle="tooltip"]').tooltip()
  }
}

function getLuxonDateFromTimestamp(ts) {
  if (!ts) {
    return
  }

  // Parse Date depanding on the format we get it
  if (`${ts}`.includes("T")) {
    if (ts === "0001-01-01T00:00:00Z") {
      return
    } else {
      return luxon.DateTime.fromISO(ts)
    }
  } else {
    let parsedDate = parseInt(ts)
    if (parsedDate === 0 || isNaN(parsedDate)) {
      return
    }
    return luxon.DateTime.fromMillis(parsedDate * 1000)
  }
}

function getRelativeTime(tsLuxon) {
  if (!tsLuxon) {
    return
  }
  var prefix = ""
  var suffix = ""
  if (tsLuxon.diffNow().milliseconds > 0) {
    prefix = "in "
  } else {
    // inverse the difference of the timestamp (3 seconds into the past becomes 3 seconds into the future)
    var now = luxon.DateTime.utc()
    tsLuxon = luxon.DateTime.fromSeconds(now.ts / 10e2 - tsLuxon.diffNow().milliseconds / 10e2)
    suffix = " ago"
  }
  var duration = tsLuxon.diffNow(["days", "hours", "minutes", "seconds"])
  const formattedDuration = formatLuxonDuration(duration)
  return `${prefix}${formattedDuration}${suffix}`
}

function formatAriaEthereumDuration(elem) {
  const attr = $(elem).attr("aria-ethereum-duration")
  const duration = luxon.Duration.fromMillis(attr).shiftTo("days", "hours", "minutes", "seconds")
  $(elem).text(formatLuxonDuration(duration))
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

function addCommas(number) {
  return number
    .toString()
    .replace(/,/g, "")
    .replace(/\B(?=(\d{3})+(?!\d))/g, "<span class='thousands-separator'></span>")
}

function getIncomeChartValueString(value, currency, ethPrice) {
  if (this.currency === "ETH") {
    return `${value.toFixed(5)} ETH`
  }

  return `${(value / ethPrice).toFixed(5)} ETH (${value.toFixed(2)} ${currency})`
}

$("[data-tooltip-date=true]").each(function (item) {
  let titleObject = $($.parseHTML($(this).attr("title")))
  titleObject.find("[aria-ethereum-date]").each(function () {
    formatAriaEthereumDate(this)
  })
  titleObject.find("[aria-ethereum-duration]").each(function () {
    formatAriaEthereumDuration(this)
  })
  $(this).attr("title", titleObject.prop("outerHTML"))
})
