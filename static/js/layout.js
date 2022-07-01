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
    var utcDiff = ts["o"] / 60
    var hour = ts["c"]["hour"] - utcDiff
    var minute = ts["c"]["minute"]
    var second = ts["c"]["second"]
    var periode = ""
    if (hour <= 12) {
      periode = " AM"
    } else {
      hour -= 12
      periode = " PM"
    }
    if (hour.toString().length == 1) {
      hour = "0" + hour
    }
    if (minute.toString().length == 1) {
      minute = "0" + minute
    }
    if (second.toString().length == 1) {
      second = "0" + second
    }
    $("#timestamp").text(ts.toFormat("MMM-dd-yyyy") + " " + hour + ":" + minute + ":" + second + periode)
  }
}

function setLocal() {
  if ($("#optionUtc").is(":checked") || $("#optionTs").is(":checked")) {
    var unixTs = $("#unixTs").text()
    var ts = luxon.DateTime.fromMillis(unixTs * 1000)
    var hour = ts["c"]["hour"]
    var minute = ts["c"]["minute"]
    var second = ts["c"]["second"]
    if (hour.toString().length == 1) {
      hour = "0" + hour
    }
    if (minute.toString().length == 1) {
      minute = "0" + minute
    }
    if (second.toString().length == 1) {
      second = "0" + second
    }
    $("#timestamp").text(ts.toFormat("MMM-dd-yyyy") + " " + hour + ":" + minute + ":" + second + " UTC + " + ts["o"] / 60 + "h")
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

// typeahead
$(document).ready(function () {
  formatTimestamps() // make sure this happens before tooltips
  $('[data-toggle="tooltip"]').tooltip()

  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.pubkey
    },
    remote: {
      url: "/search/validators/%QUERY",
      wildcard: "%QUERY",
    },
  })

  var bhBlocks = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.slot
    },
    remote: {
      url: "/search/blocks/%QUERY",
      wildcard: "%QUERY",
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
    },
  })

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
      display: "blockroot",
      templates: {
        header: '<h3 class="h5">Blocks</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">${data.slot}: ${data.blockroot}</div>`
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
          return `<div class="text-monospace text-truncate">${data.slot}: ${data.txhash}</div>`
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
        header: '<h3 class="h5">ETH1 Addresses</h3>',
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate">0x${data.address}</div>`
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
    if (sug.slot !== undefined) {
      if (sug.txhash !== undefined) window.location = "/block/" + sug.slot + "#transactions"
      else window.location = "/block/" + sug.slot
    } else if (sug.index !== undefined) {
      if (sug.index === "deposited") window.location = "/validator/" + sug.pubkey
      else window.location = "/validator/" + sug.index
    } else if (sug.epoch !== undefined) {
      window.location = "/epoch/" + sug.epoch
    } else if (sug.address !== undefined) {
      window.location = "/validators/eth1deposits?q=" + sug.address
    } else if (sug.graffiti !== undefined) {
      // sug.graffiti is html-escaped to prevent xss, we need to unescape it
      var el = document.createElement("textarea")
      el.innerHTML = sug.graffiti
      window.location = "/blocks?q=" + encodeURIComponent(el.value)
    } else {
      console.log("invalid typeahead-selection", sug)
    }
  })
})

$("[aria-ethereum-date]").each(function (item) {
  var dt = $(this).attr("aria-ethereum-date")
  var format = $(this).attr("aria-ethereum-date-format")

  if (!format) {
    format = "ff"
  }

  if (format === "FROMNOW") {
    $(this).text(luxon.DateTime.fromMillis(dt * 1000).toRelative({ style: "short" }))
    $(this).attr("title", luxon.DateTime.fromMillis(dt * 1000).toFormat("ff"))
    $(this).attr("data-toggle", "tooltip")
  } else if (format === "LOCAL") {
    var local = luxon.DateTime.fromMillis(dt * 1000)
    var utc = local.toUTC()
    var utcHour = utc["c"]["hour"]
    var localHour = local["c"]["hour"]
    var localMinute = local["c"]["minute"]
    var localSecond = local["c"]["minute"]
    var diff = localHour - utcHour
    var utcDiff = ""
    if (diff < 0) {
      utcDiff = " UTC - " + diff * -1
    } else {
      utcDiff = " UTC + " + diff
    }
    if (localHour.toString().length == 1) {
      localHour = "0" + localHour
    }
    if (localMinute.toString().length == 1) {
      localMinute = "0" + localMinute
    }
    if (localSecond.toString().length == 1) {
      localSecond = "0" + localSecond
    }

    $(this).text(local.toFormat("MMM-dd-yyyy") + " " + localHour + ":" + localMinute + ":" + localSecond + utcDiff + "h")
    $(this).attr("title", luxon.DateTime.fromMillis(dt * 1000).toFormat("ff"))
    $(this).attr("data-toggle", "tooltip")
  } else {
    $(this).text(luxon.DateTime.fromMillis(dt * 1000).toFormat(format))
  }
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

// Javascript to enable link to tab
var url = document.location.toString()
if (url.match("#")) {
  $('.nav-tabs a[href="#' + url.split("#")[1] + '"]').tab("show")
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
    $(this).text(tsLuxon.toRelative({ style: "short" }))
  })
  sel.find('[data-toggle="tooltip"]').tooltip()
}

function addCommas(number) {
  return number
    .toString()
    .replace(/,/g, "")
    .replace(/\B(?=(\d{3})+(?!\d))/g, ",")
}
