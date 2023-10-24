function applyTTFix() {
  $("button, a").on("mousedown", (evt) => {
    evt.preventDefault() // prevent setting the browser focus on all mouse buttons, which prevents tooltips from disapearing
  })
  truncateTooltip()
}

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

function setTs() {
  let timestamp = $("#timestamp")
  let unixTs = timestamp.attr("aria-ethereum-date")
  if (!unixTs) {
    unixTs = $("#unixTs").text()
  }
  var ts = luxon.DateTime.fromMillis(unixTs * 1000)
  let optionName = timestamp.attr("aria-timestamp-options")
  let selectedOption = document.querySelector(`input[name="${optionName}"]:checked`)?.value

  let text = ""
  switch (selectedOption) {
    case "local":
      text = ts.toFormat("MMM-dd-yyyy HH:mm:ss") + " UTC" + ts.toFormat("Z")
      break
    case "utc":
      text = ts.toUTC().toFormat("MMM-dd-yyyy hh:mm:ss a")
      break
    default:
      text = ts["ts"] / 1000
      break
  }

  timestamp.text(text)
}

function copyTs() {
  var text = $("#timestamp").text()
  navigator.clipboard.writeText(text)
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

function shortenAddress(address) {
  if (!address) {
    return ""
  }
  address = address.replace("0x", "0")
  return `0x${address.substr(0, 6)}...${address.substr(address.length - 6)}`
}

function hex2a(hexx) {
  var hex = hexx.toString() //force conversion
  var str = ""
  for (var i = 0; i < hex.length; i += 2) str += String.fromCharCode(parseInt(hex.substr(i, 2), 16))
  return str
}

var observeDOM = (function () {
  var MutationObserver = window.MutationObserver || window.WebKitMutationObserver

  return function (obj, callback) {
    if (!obj || obj.nodeType !== 1) return

    if (MutationObserver) {
      // define a new observer
      var mutationObserver = new MutationObserver(callback)

      // have the observer observe for changes in children
      mutationObserver.observe(obj, { childList: true, subtree: true })
      return mutationObserver
    }

    // browser support fallback
    else if (window.addEventListener) {
      obj.addEventListener("DOMNodeInserted", callback, false)
      obj.addEventListener("DOMNodeRemoved", callback, false)
    }
  }
})()

observeDOM(document.documentElement, applyTTFix)

/**
 * Listens to navigation events triggered by tab navigation and activates the related content.
 * The id of the content must have the following pattern: hashtag + "TabPanel"
 * This way we prevent the browser from jumping down to the tab content on intitial navigation
 * @param tabContainerId: Id if the parent container of the tab content's
 * @param tabBar: Id of the tabbar container
 * @param defaultTab: Id of the default content to be displayed (without the trailing TabPanel)
 **/
function activateTabbarSwitcher(tabContainerId, tabBar, defaultTab) {
  var lastTab = defaultTab
  const handleTabChange = (url) => {
    const split = url?.split("#")
    var selectedTab = defaultTab
    if (split?.length == 2) {
      selectedTab = split[1]
    }
    const container = $(`#${tabContainerId}`)
    if (!container) {
      return
    }

    container.find(".tab-pane.active").removeClass("active show")

    var someTabTriggerEl = container.find(`#${selectedTab}TabPanel`)
    if (!someTabTriggerEl.length) {
      someTabTriggerEl = container.find(`#${lastTab}TabPanel`)
      if (!someTabTriggerEl.length) {
        return
      }
    } else {
      lastTab = selectedTab
    }

    new bootstrap.Tab(someTabTriggerEl[0]).show()
  }
  const handleTabChangeClick = (event) => {
    var href = event.currentTarget.href
    if (href) {
      handleTabChange(href)
    }
  }

  window.addEventListener("DOMContentLoaded", function (ev) {
    handleTabChange(window.location.href)
  })

  if (window.navigation) {
    window.navigation.addEventListener("navigate", (event) => {
      if (!event.destination?.url || event.destination.url.split("#")[0] != event.srcElement?.currentEntry?.url?.split("#")[0]) {
        return
      }
      handleTabChange(event.destination?.url)
    })
  } else {
    //handle Firefox specific way as it does not support the navigation way
    const tabBarContainer = $(`#${tabBar}`)
    if (!tabBarContainer) {
      return
    }

    tabBarContainer.find(`.nav-link`).on("click", handleTabChangeClick)
  }

  handleTabChange(window.location.href)
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
  let requestNum = 10
  var timeWait = 0

  // used to overwrite Bloodhounds "transport._get" function which handles the rateLimitWait parameter
  // since we can't access the closure variable anymore (?)
  // copied and adapted from typeahead.js source code (https://github.com/twitter/typeahead.js/blob/master/dist/typeahead.bundle.js#L106)
  // -> so we have control via `timeWait` now
  var debounce = function (context, func) {
    var timeout, result

    return function () {
      var args = arguments,
        later = function () {
          timeout = null
          result = func.apply(context, args)
        }
      clearTimeout(timeout)
      timeout = setTimeout(later, timeWait)
      if (!timeout) {
        result = func.apply(context, args)
      }
      return result
    }
  }

  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.pubkey
    },
    remote: {
      url: "/search/validators/%QUERY",
      // use prepare hook to modify the rateLimitWait parameter on input changes
      // NOTE: we only need to do this for the first function because testing showed that queries are executed/queued in order
      // No need to update `timeWait` multiple times.
      prepare: function (_, settings) {
        var cur_query = $("input.typeahead.tt-input").val()
        timeWait = 4000 - Math.min(cur_query.length, 5) * 500
        // "wildcard" can't be used anymore, need to set query wildcard ourselves now
        settings.url = settings.url.replace("%QUERY", encodeURIComponent(cur_query))
        return settings
      },
      maxPendingRequests: requestNum,
    },
  })
  bhValidators.remote.transport._get = debounce(bhValidators.remote.transport, bhValidators.remote.transport._get)

  var bhEns = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj?.domain
    },
    remote: {
      url: "/search/ens/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
      transform: function (data) {
        return data?.address && data?.domain ? { data: { ...data } } : null
      },
    },
  })
  bhEns.remote.transport._get = debounce(bhEns.remote.transport, bhEns.remote.transport._get)

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
  bhSlots.remote.transport._get = debounce(bhSlots.remote.transport, bhSlots.remote.transport._get)

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
  bhBlocks.remote.transport._get = debounce(bhBlocks.remote.transport, bhBlocks.remote.transport._get)

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
  bhTransactions.remote.transport._get = debounce(bhTransactions.remote.transport, bhTransactions.remote.transport._get)

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
  bhGraffiti.remote.transport._get = debounce(bhGraffiti.remote.transport, bhGraffiti.remote.transport._get)

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
  bhEpochs.remote.transport._get = debounce(bhEpochs.remote.transport, bhEpochs.remote.transport._get)

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
  bhEth1Accounts.remote.transport._get = debounce(bhEth1Accounts.remote.transport, bhEth1Accounts.remote.transport._get)

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
  bhValidatorsByAddress.remote.transport._get = debounce(bhValidatorsByAddress.remote.transport, bhValidatorsByAddress.remote.transport._get)

  var bhPubkey = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.index
    },
    remote: {
      url: "/search/validators_by_pubkey/%QUERY",
      wildcard: "%QUERY",
      maxPendingRequests: requestNum,
    },
  })
  bhPubkey.remote.transport._get = debounce(bhPubkey.remote.transport, bhPubkey.remote.transport._get)

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
      name: "pubkeys",
      source: bhPubkey,
      display: "pubkey",
      templates: {
        header: '<h3 class="h5">Validators by Public Key</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate high-contrast">${data.pubkey}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "ens",
      source: bhEns,
      display: function (data) {
        return data?.address && data?.domain ? data.domain : null
      },
      templates: {
        header: '<h3 class="h5">Ens</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate"><a href="/ens/${data.domain}">${data.domain} Registration Overview</a></div>`
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
      display: (data) => data.address || data.name,
      templates: {
        header: '<h3 class="h5">Address</h3>',
        suggestion: function (data) {
          if (data.name) {
            return `
              <div class="d-flex justify-content-between">
                <div class="text-monospace text-truncate">${data.name}</div>
                <div class="text-monospace ml-1 d-flex">
                  ${shortenAddress(data.addres)}
                </div>
              </div>`
          }
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
    } else if (sug.pubkey !== undefined) {
      window.location = "/validator/" + sug.pubkey
    } else if (sug.epoch !== undefined) {
      window.location = "/epoch/" + sug.epoch
    } else if (sug.address !== undefined) {
      window.location = "/address/" + sug.address
    } else if (sug.eth1_address !== undefined) {
      window.location = "/validators/deposits?q=" + sug.eth1_address
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

$(document).on("inserted.bs.tooltip", function (event) {
  $("[aria-ethereum-date]").each(function () {
    formatAriaEthereumDate(this)
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

  var local = luxon.DateTime.fromMillis(dt * 1000)
  if (format === "FROMNOW") {
    $(elem).text(getRelativeTime(local))
    $(elem).attr("data-original-title", formatTimestampsTooltip(local))
    $(elem).attr("data-toggle", "tooltip")
  } else if (format === "LOCAL") {
    $(elem).text(local.toFormat("MMM-dd-yyyy HH:mm:ss") + " UTC" + local.toFormat("Z"))
    $(elem).attr("data-original-title", formatTimestampsTooltip(local))
    $(elem).attr("data-toggle", "tooltip")
  } else if (format === "TIMESTAMP") {
    setTs()
  } else {
    $(elem).text(local.toFormat(format))
  }
}

function truncateTooltip() {
  let nodes = $("[truncate-tooltip]")
  nodes.each((_, node) => {
    let title = ""
    if (node.scrollWidth > node.offsetWidth) {
      title = node.attributes["truncate-tooltip"].value
    }
    if (node.attributes["data-original-title"]?.value != title) {
      node.setAttribute("data-original-title", title)
      if (title !== "") {
        $(node).tooltip()
      }
    }
  })
}

function formatTimestamps(selStr) {
  var sel = $(document)
  if (selStr !== undefined) {
    sel = $(selStr)
  }
  sel.find(".timestamp").each(function () {
    var ts = $(this).data("timestamp")
    var local = luxon.DateTime.fromMillis(ts * 1000)

    $(this).text(getRelativeTime(local))
    $(this).attr("data-original-title", formatTimestampsTooltip(local))
  })

  if (sel.find('[data-toggle="tooltip"]').tooltip) {
    sel.find('[data-toggle="tooltip"]').tooltip()
  }
}

function formatTimestampsTooltip(local) {
  var toolTipFormat = "yyyy-MM-dd HH:mm:ss"
  var tooltip = local.toFormat(toolTipFormat)

  return tooltip
}

function getLuxonDateFromTimestamp(ts) {
  if (!ts) {
    return
  }

  // Parse Date depending on the format we get it
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

function trimPrice(value, decimals = 5) {
  if (value === undefined || value === null) {
    return ""
  }
  let parts = value.toString().split(".")
  return parts.length > 1 ? `${parts[0]}.${parts[1].substring(0, decimals)}` : parts[0]
}

function trimToken(value) {
  return trimPrice(value)
}

function trimCurrency(value) {
  return trimPrice(value, 2)
}

function getIncomeChartValueString(value, currency, priceCurrency, price) {
  if (currency == priceCurrency || (currency == "xDAI" && priceCurrency == "DAI")) {
    return `${trimToken(value)} ${currency}`
  }

  return `${trimToken(value / price)} ${currency} (${trimCurrency(value)} ${priceCurrency})`
}

$("[data-truncate-middle]").each(function (item) {
  truncateMiddle(this)
  addEventListener("resize", (event) => {
    truncateMiddle(this)
  })
  addEventListener("copy", (event) => {
    copyDots(event, this)
  })
})

// function for trimming an placing ellipsis in the middle when text is overflowing
function truncateMiddle(element) {
  element.innerHTML = element.getAttribute("data-truncate-middle")
  const parent = element.parentElement
  // get ratio of visible width to full width
  const ratio = parent.offsetWidth / parent.scrollWidth
  if (ratio < 1) {
    const removeCount = Math.ceil((parent.innerText.length * (1 - ratio)) / 2) + 1
    const originalText = element.getAttribute("data-truncate-middle")
    element.innerHTML = originalText.substr(0, originalText.length / 2 - removeCount) + "…" + originalText.substr(originalText.length / 2 + removeCount)
  }
}

// function for inserting correct text into clipboard when copying ellipsis of text truncated with 'truncateMiddle()'
function copyDots(event, element) {
  const selection = document.getSelection()
  if (selection.toString().includes("…")) {
    const originalText = element.getAttribute("data-truncate-middle")
    const diff = originalText.length - (element.innerText.length - 1)
    const replaceText = selection.toString().replace("…", originalText.substr(originalText.length / 2 - diff / 2, diff))
    event.clipboardData.setData("text/plain", replaceText)
    event.preventDefault()
  }
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
