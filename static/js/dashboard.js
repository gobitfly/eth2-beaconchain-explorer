function createBlock(x, y) {
  use = document.createElementNS("http://www.w3.org/2000/svg", "use")
  use.setAttributeNS(null, "href", "#cube")
  use.setAttributeNS(null, "x", x)
  use.setAttributeNS(null, "y", y)
  return use
}

function appendBlocks(blocks) {
  $(".blue-cube g.move").each(function () {
    $(this).empty()
  })

  var cubes = document.querySelectorAll(".blue-cube g.move")
  for (var i = 0; i < blocks.length; i++) {
    var block = blocks[i]

    for (let i = 0; i < cubes.length; i++) {
      let cube = cubes[i]
      cube.appendChild(createBlock(block[0], block[1]))
    }
  }
  for (let i = 0; i < cubes.length; i++) {
    let cube = cubes[i]
    var use = document.createElementNS("http://www.w3.org/2000/svg", "use")
    use.setAttributeNS(null, "href", "#cube-small")
    use.setAttributeNS(null, "x", 129)
    use.setAttributeNS(null, "y", 56)
    cube.appendChild(use)
  }
}

var selectedBTNindex = null
var incomeChart = null
var incomeChartDefault = document.getElementById("balance-chart").innerHTML
var proposedChart = null
var proposedChartDefault = document.getElementById("proposed-chart").innerHTML
var summaryDefaultValue = "0.000"
var countdownIntervals = new Map()
var VALLIMIT = 280
var allIncomeLoaded = false

function hideValidatorHist() {
  if ($.fn.dataTable.isDataTable("#dash-validator-history-table")) {
    $("#dash-validator-history-table").DataTable().destroy()
  }

  $("#dash-validator-history-table").addClass("d-none")
  $("#dash-validator-history-art").removeClass("d-none")
  $("#dash-validator-history-art").addClass("d-flex")
  $("#dash-validator-history-index").text("")
  selectedBTNindex = null
}

function showValidatorHist(index) {
  if ($.fn.dataTable.isDataTable("#dash-validator-history-table")) {
    $("#dash-validator-history-table").DataTable().destroy()
  }

  $("#dash-validator-history-table").DataTable({
    processing: true,
    serverSide: true,
    lengthChange: false,
    ordering: false,
    searching: false,
    details: false,
    pagingType: "simple",
    pageLength: 10,
    ajax: dataTableLoader("/validator/" + index + "/history"),
    language: {
      searchPlaceholder: "Search by Epoch Number",
      search: "",
      paginate: {
        previous: '<i class="fas fa-chevron-left"></i>',
        next: '<i class="fas fa-chevron-right"></i>',
      },
    },
    columnDefs: [
      {
        targets: 1,
        createdCell: function (td, cellData, rowData, row, col) {
          $(td).css("width", "0px")
        },
      },
      {
        targets: 2,
        createdCell: function (td, cellData, rowData, row, col) {
          $(td).css("padding", "0px")
        },
      },
    ],
    drawCallback: function (settings) {
      formatTimestamps()
    },
  })
  $("#validator-history-table_wrapper > div:nth-child(3) > div:nth-child(1)").removeClass("col-md-5").removeClass("col-sm-12")
  $("#validator-history-table_wrapper > div:nth-child(3) > div:nth-child(2)").removeClass("col-md-7").removeClass("col-sm-12")
  $("#validator-history-table_wrapper > div:nth-child(3)").addClass("justify-content-center")
  $("#validator-history-table_paginate").attr("style", "padding-right: 0 !important")
  $("#validator-history-table_info").attr("style", "padding-top: 0;")
  $("#dash-validator-history-table").removeClass("d-none")
  $("#dash-validator-history-art").removeClass("d-flex")
  $("#dash-validator-history-art").addClass("d-none")
  $("#dash-validator-history-index").text(index)
  selectedBTNindex = index
  showSelectedValidator()
  updateValidatorInfo(index)
}

function toggleFirstrow() {
  $("#dashChartTabs a:first").tab("show")
  let id = $("#validators tbody>tr:nth-child(1)>td>button").attr("id")
  setTimeout(function () {
    $("#" + id).focus()
  }, 200)
}

function updateValidatorInfo(index) {
  fetch(`/validator/${index}/proposedblocks?draw=1&start=1&length=1`, {
    method: "GET",
  }).then((res) => {
    res.json().then((data) => {
      $("#blockCount span").text(data.recordsTotal)
    })
  })
  fetch(`/validator/${index}/attestations?draw=1&start=1&length=1`, {
    method: "GET",
  }).then((res) => {
    res.json().then((data) => {
      $("#attestationCount span").text(data.recordsTotal)
    })
  })
  fetch(`/validator/${index}/slashings?draw=1&start=1&length=1`, {
    method: "GET",
  }).then((res) => {
    res.json().then((data) => {
      var total = parseInt(data.recordsTotal)
      if (total > 0) {
        $("#slashingsCountDiv").removeClass("d-none")
        $("#slashingsCount span").text(total)
      } else {
        $("#slashingsCountDiv").addClass("d-none")
      }
    })
  })
  fetch(`/validator/${index}/effectiveness`, {
    method: "GET",
  }).then((res) => {
    res.json().then((data) => {
      setValidatorEffectiveness("effectiveness", data.effectiveness)
    })
  })
}

function getValidatorQueryString() {
  return window.location.href.slice(window.location.href.indexOf("?"), window.location.href.length)
}

var boxAnimationDirection = ""

window.addEventListener("load", function () {
  var searchInput = document.querySelector("#selected-validators-input input.typeahead-dashboard")

  $("#selected-validators-input-button").on("click", function (ev) {
    var overview = document.getElementById("selected-validators-overview")
    overview.classList.toggle("d-none")
  })

  searchInput.addEventListener("focus", function (ev) {
    var overview = document.getElementById("selected-validators-overview")
    if (document.querySelector("#selected-validators-input-button > span").textContent) {
      overview.classList.remove("d-none")
    }
  })

  document.addEventListener("click", function (event) {
    var overview = document.getElementById("selected-validators-overview")
    var trgt = event.target
    let count = 0
    let match = false
    do {
      count++
      if (trgt.matches("li[data-validator-index]") || trgt.matches(".tt-suggestion") || trgt.matches(".tt-menu") || trgt.matches("#selected-validators-overview") || trgt.matches("#selected-validators-input input.typeahead-dashboard") || trgt.matches("#selected-validators-input-button")) {
        match = true
        break
      }
      if (count > 15) break
      trgt = trgt.parentNode
    } while (trgt && trgt.matches)
    if (!match) {
      overview.classList.add("d-none")
    }
  })
})

function showSelectedValidator() {
  setTimeout(function () {
    $("span[id^=dropdownMenuButton]").each(function (el, item) {
      if ($(item).attr("id") === "dropdownMenuButton" + selectedBTNindex) {
        $(item).addClass("bg-primary")
      } else {
        if (selectedBTNindex != null) {
          $(item).removeClass("bg-primary")
        }
      }
    })
  }, 100) //if deselected index is not clearing increase the time

  $(".hbtn").hover(
    function () {
      $(this).addClass("shadow")
    },
    function () {
      $(this).removeClass("shadow")
    }
  )
}

function showValidatorsInSearch(qty) {
  qty = parseInt(qty)
  let i = 0
  let l = []
  $("#selected-validators-input li:not(:last)").remove()
  $("#selected-validators.val-modal li").each(function (el, item) {
    if (i === qty) {
      return
    }
    l.push($(item).clone())
    i++
  })
  for (let i = 0; i < l.length; i++) {
    $("#selected-validators-input").prepend(l[l.length - (i + 1)])
  }
}

function renderProposedHistoryTable(data) {
  if ($.fn.dataTable.isDataTable("#proposals-table")) {
    $("#proposals-table").DataTable().destroy()
  }

  $("#proposals-table").DataTable({
    serverSide: false,
    data: data,
    processing: false,
    ordering: false,
    searching: true,
    pagingType: "full_numbers",
    lengthMenu: [10, 25, 50],
    preDrawCallback: function () {
      // this does not always work.. not sure how to solve the staying tooltip
      try {
        $("#proposals-table").find('[data-toggle="tooltip"]').tooltip("dispose")
      } catch (e) {
        console.error(e)
      }
    },
    drawCallback: function (settings) {
      $("#proposals-table").find('[data-toggle="tooltip"]').tooltip()
    },
    columnDefs: [
      {
        targets: 0,
        data: "0",
        render: function (data, type, row, meta) {
          return '<a href="/validator/' + data + '"><i class="fas fa-male fa-sm mr-1"></i>' + data + "</a>"
        },
      },
      {
        targets: 1,
        data: "1",
        render: function (data, type, row, meta) {
          // date and epochs
          const startEpoch = timeToEpoch(data * 1000)
          const startDate = luxon.DateTime.fromMillis(data * 1000)
          const timeForOneDay = 24 * 60 * 60 * 1000
          const endEpoch = timeToEpoch(data * 1000 + timeForOneDay) - 1
          const endDate = luxon.DateTime.fromMillis(epochToTime(endEpoch + 1))
          const tooltip = `${startDate.toFormat("MMM-dd-yyyy HH:mm:ss")} - ${endDate.toFormat("MMM-dd-yyyy HH:mm:ss")}<br> Epochs ${startEpoch} - ${endEpoch}<br/>`

          return `<span data-html="true" data-toggle="tooltip" data-placement="top" title="${tooltip}">${startDate.toFormat("yyyy-MM-dd")}</span>`
        },
      },
      {
        targets: 2,
        data: "2",
        render: function (data, type, row, meta) {
          return '<span class="text-success p-1">' + data[0] + "</span>/" + '<span class="text-danger p-1">' + data[1] + "</span>/" + '<span class="text-info p-1">' + data[2] + "</span>"
        },
      },
    ],
  })
}

function showProposedHistoryTable() {
  fetch("/dashboard/data/proposalshistory" + getValidatorQueryString(), {
    method: "GET",
  }).then((res) => {
    res.json().then(function (data) {
      let proposedHistTableData = []
      for (let item of data.data) {
        proposedHistTableData.push([item[0], item[1], [item[2], item[3], item[4]]])
      }
      renderProposedHistoryTable(proposedHistTableData)
    })
  })
}

function switchFrom(el1, el2, el3, el4) {
  $(el1).removeClass("proposal-switch-selected")
  $(el2).addClass("proposal-switch-selected")
  $(el3).addClass("d-none")
  $(el4).removeClass("d-none")
}

var firstSwitch = true

function initValidatorCountdown(validatorIndex, queueId, ts) {
  var now = Math.round(new Date().getTime() / 1000)
  var secondsLeft = ts - now
  setValidatorCountdown(validatorIndex, queueId, secondsLeft)

  if (!countdownIntervals.has(validatorIndex)) {
    countdownIntervals.set(
      validatorIndex,
      setInterval(function () {
        if (secondsLeft <= 0) {
          clearInterval(countdownIntervals.get(validatorIndex))
          return
        }

        secondsLeft -= 1
        setValidatorCountdown(validatorIndex, queueId, secondsLeft)
      }, 1000)
    )
  }
}

function setValidatorCountdown(validatorIndex, queueId, secondsLeft) {
  let [seconds, minutes, hours, days] = [0, 0, 0, 0]
  if (secondsLeft > 0) {
    const duration = luxon.Duration.fromMillis(secondsLeft * 1000).shiftTo("days", "hours", "minutes", "seconds")

    seconds = duration.seconds
    minutes = duration.minutes
    hours = duration.hours
    days = duration.days
  }

  if (seconds < 10) {
    seconds = "0" + seconds
  }
  if (minutes < 10) {
    minutes = "0" + minutes
  }
  if (hours < 10) {
    hours = "0" + hours
  }
  if (days < 10) {
    days = "0" + days
  }

  var $element = $("#queue-" + validatorIndex)

  var tooltip = `
    <div>This validator is currently <span class="font-weight-bolder d-inline-block text-underlined">#${queueId}</span> in Queue.</div>
    <strong>${days} days ${hours} hr ${minutes} min ${seconds} sec</strong>`

  $element.attr("data-original-title", tooltip)

  if ($element.data("hover")) {
    $element.tooltip("show")
  }
}

function removeValidatorCountdown(validatorIndex) {
  if (countdownIntervals.has(validatorIndex)) {
    clearInterval(countdownIntervals.get(validatorIndex))
    countdownIntervals.delete(validatorIndex)
  }
}

$(document).ready(function () {
  $("#rewards-button").on("click", () => {
    localStorage.setItem("load_dashboard_validators", true)
    window.location.href = "/rewards"
  })

  $(".proposal-switch").on("click", () => {
    if ($(".switch-chart").hasClass("proposal-switch-selected")) {
      if (firstSwitch) {
        showProposedHistoryTable()
        firstSwitch = false
      }
      switchFrom(".switch-chart", ".switch-table", "#proposed-chart", "#proposed-table-div")
    } else if ($(".switch-table").hasClass("proposal-switch-selected")) {
      switchFrom(".switch-table", ".switch-chart", "#proposed-table-div", "#proposed-chart")
    }
  })

  $("#validators").on("page.dt", function () {
    showSelectedValidator()
  })
  //bookmark button adds all validators in the dashboard to the watchlist
  $("#bookmark-button").on("click", function (event) {
    var tickIcon = $("<i class='fas fa-check' style='width:15px;'></i>")
    var bookmarkIcon = $("<i class='far fa-bookmark' style='width:15px;'></i>")
    var errorIcon = $("<i class='fas fa-exclamation' style='width:15px;'></i>")
    var validatorIndices = state.validators.filter((v) => {
      return !isValidatorPubkey(v)
    })
    fetch("/dashboard/save", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(validatorIndices),
    })
      .then(function (res) {
        console.log("response", res)
        if (res.status === 200 && !res.redirected) {
          // success
          console.log("success")
          $("#bookmark-button").empty().append(tickIcon)
          setTimeout(function () {
            $("#bookmark-button").empty().append(bookmarkIcon)
          }, 1000)
        } else if (res.redirected) {
          console.log("redirected!")
          $("#bookmark-button").attr("data-original-title", "Please login or sign up first.")
          $("#bookmark-button").tooltip("show")
          $("#bookmark-button").empty().append(errorIcon)
          setTimeout(function () {
            $("#bookmark-button").empty().append(bookmarkIcon)
            $("#bookmark-button").tooltip("hide")
            $("#bookmark-button").attr("data-original-title", "Save all to Watchlist")
          }, 2000)
        } else {
          // could not bookmark validators
          $("#bookmark-button").empty().append(errorIcon)
          setTimeout(function () {
            $("#bookmark-button").empty().append(bookmarkIcon)
          }, 2000)
        }
      })
      .catch(function (err) {
        $("#bookmark-button").empty().append(errorIcon)
        setTimeout(function () {
          $("#bookmark-button").empty().append(bookmarkIcon)
        }, 2000)
        console.log(err)
      })
  })

  $(document).on("mouseenter", ".hoverCheck[data-track=hover]", function () {
    $(this).data("hover", true)
  })

  $(document).on("mouseleave", ".hoverCheck[data-track=hover]", function () {
    $(this).data("hover", false)
  })

  var clearSearch = $("#clear-search")
  var copyIcon = $("<i class='fa fa-copy' style='width:15px'></i>")
  var tickIcon = $("<i class='fas fa-check' style='width:15px;'></i>")

  clearSearch.on("click", function () {
    clearSearch.empty().append(tickIcon)
    setTimeout(function () {
      clearSearch.empty().append(copyIcon)
    }, 500)
  })
  $.fn.DataTable.ext.pager.numbers_length = 5
  var validatorsDataTable = (window.vdt = $("#validators").DataTable({
    processing: true,
    serverSide: false,
    searching: true,
    stateSave: true,
    stateSaveCallback: function (settings, data) {
      data.start = 0
      localStorage.setItem("DataTables_" + settings.sInstance, JSON.stringify(data))
    },
    stateLoadCallback: function (settings) {
      return JSON.parse(localStorage.getItem("DataTables_" + settings.sInstance))
    },
    pageLength: 10,
    pagingType: "full_numbers",
    scrollY: "503px",
    info: false,
    language: {
      search: "",
      searchPlaceholder: "Search...",
      paginate: {
        previous: '<i class="fas fa-chevron-left"></i>',
        next: '<i class="fas fa-chevron-right"></i>',
      },
    },
    dom: "<'row'<'col-sm-12 col-md-6 filter-by-status'><'col-sm-12 col-md-6'f>>" + "<'row'<'col-sm-12'tr>>" + "<'row'<'col-sm-12 col-md-5'l><'col-sm-12 col-md-7'p>>",
    preDrawCallback: function () {
      // this does not always work.. not sure how to solve the staying tooltip
      try {
        $("#validators").find('[data-toggle="tooltip"]').tooltip("dispose")
      } catch (e) {
        console.error(e)
      }
    },
    drawCallback: function (settings) {
      formatTimestamps()
      $("#validators").find('[data-toggle="tooltip"]').tooltip()
    },
    order: [[1, "asc"]],
    columnDefs: [
      // Pubkey
      {
        targets: 0,
        data: "0",
        createdCell: function (td, cellData, rowData, row, col) {
          $(td).css("display", "flex")
          $(td).css("align-items", "center")
          $(td).css("justify-content", "space-between")
        },
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") {
            return data
          }
          return `<a href="/validator/${data}">0x${data.substr(0, 8)}...</a><i class="fa fa-copy text-muted p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="0x${data}"></i>`
        },
      },
      // Index
      {
        targets: 1,
        data: "1",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data
          if (isNaN(parseInt(data))) {
            return `<span class="m-0 p-2">${data}</span>`
          } else {
            return `<span class="m-0 p-2 hbtn" id="dropdownMenuButton${data}" style="cursor: pointer;" onclick="showValidatorHist('${data}')">${data}</span>`
          }
        },
      },
      // Current balance / Effective balance
      {
        targets: 2,
        data: "2",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          return `${data[0]}`
        },
      },
      // Index / State / Queue ahead / Estimated activation ts
      {
        targets: 3,
        data: "3",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : -1
          var d = data[1].split("_")
          var s = d[0].charAt(0).toUpperCase() + d[0].slice(1)

          if (d[0] === "pending" && d[1] !== "deposited") {
            initValidatorCountdown(data[0], data[2], data[3])
            return `<span class="hoverCheck" data-track='hover' id="queue-${data[0]}" data-html="true" data-toggle="tooltip" data-placement="top">${s} (#<span>${data[2]}</span>)</span>`
          }
          if (d[1] === "offline") return `<span style="display:none">${d[1]}</span><span data-toggle="tooltip" data-placement="top" title="No attestation in the last 2 epochs">${s} <i class="fas fa-power-off fa-sm text-danger"></i></span>`
          if (d[1] === "online") return `<span style="display:none">${d[1]}</span><span>${s} <i class="fas fa-power-off fa-sm text-success"></i></span>`
          return `<span>${s}</span>`
        },
      },
      // Activation epoch / Activation ts
      {
        targets: 4,
        visible: false,
        data: "4",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          if (data === null) return "-"
          return `<span data-toggle="tooltip" data-placement="top" title="${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))}">${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        },
      },
      // Exit epoch / Exit ts
      {
        targets: 5,
        visible: false,
        data: "5",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          if (data === null) return "-"
          return `<span data-toggle="tooltip" data-placement="top" title="${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))}">${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        },
      },
      // Withdrawable epoch / Withdrawable ts
      {
        targets: 6,
        data: "6",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          if (data === null) return "-"
          return `<span data-toggle="tooltip" data-placement="top" title="${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))}">${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        },
      },
      // Last attestation / Last attestation ts
      {
        targets: 7,
        data: "7",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          if (data === null) return "No Attestation found"
          return `${data[1]}`
        },
      },
      // Executed proposals / Missed proposals
      {
        targets: 8,
        data: "8",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] + data[1] : null
          return `<span data-toggle="tooltip" data-placement="top" title="${data[0]} executed / ${data[1]} missed"><span class="text-success">${data[0]}</span> / <span class="text-danger">${data[1]}</span></span>`
        },
      },
      // Performance last 7d
      {
        targets: 9,
        data: "9",
        render: function (data, type, row, meta) {
          return data
        },
      },
      // Deposit address
      {
        targets: 10,
        orderable: false,
        data: function (data) {
          return data[10]
        },
        visible: false, // hidden column for filtering only
        render: function (data, type) {
          if (type == "filter") return data
          return null
        },
      },
    ],
  }))

  function create_validators_typeahead(input_container_selector, table_selector) {
    var bhEth1Addresses = new Bloodhound({
      datumTokenizer: Bloodhound.tokenizers.whitespace,
      queryTokenizer: Bloodhound.tokenizers.whitespace,
      identify: function (obj) {
        return obj.eth1_address
      },
      remote: {
        url: "/search/indexed_validators_by_eth1_addresses/%QUERY",
        wildcard: "%QUERY",
      },
    })
    $(input_container_selector).typeahead(
      {
        minLength: 1,
        highlight: true,
        hint: false,
        autoselect: false,
      },
      {
        limit: 5,
        name: "addresses",
        source: bhEth1Addresses,
        display: function (data) {
          return data?.eth1_address || ""
        },
        templates: {
          header: '<h5 class="font-weight-bold ml-3">ETH Address</h5>',
          suggestion: function (data) {
            var len = data.validator_indices.length > 10 ? 10 + "+" : data.validator_indices.length
            return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">0x${data.eth1_address}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
          },
        },
      }
    )
    $(input_container_selector).on("focus", function (e) {
      if (e.target.value !== "") {
        $(this).trigger($.Event("keydown", { keyCode: 40 }))
      }
    })
    $(input_container_selector).on("input", function () {
      $(".tt-suggestion").first().addClass("tt-cursor")
    })
    $(input_container_selector).bind("typeahead:select", function (ev, suggestion) {
      if (suggestion?.eth1_address) {
        $(table_selector).DataTable().search(suggestion.eth1_address)
        $(table_selector).DataTable().draw()
      }
    })
  }
  create_validators_typeahead("input[aria-controls='validators']", "#validators")

  var timeWait = 0
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
      return obj.index
    },
    remote: {
      url: "/search/indexed_validators/%QUERY",
      // use prepare hook to modify the rateLimitWait parameter on input changes
      // NOTE: we only need to do this for the first function because testing showed that queries are executed/queued in order
      // No need to update `timeWait` multiple times.
      prepare: function (_, settings) {
        var cur_query = $(".typeahead-dashboard").val()
        timeWait = 4000 - Math.min(cur_query.length, 5) * 500
        // "wildcard" can't be used anymore, need to set query wildcard ourselves now
        settings.url = settings.url.replace("%QUERY", encodeURIComponent(cur_query))
        return settings
      },
    },
  })
  bhValidators.remote.transport._get = debounce(bhValidators.remote.transport, bhValidators.remote.transport._get)
  var bhPubkey = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.index
    },
    remote: {
      url: "/search/validators_by_pubkey/%QUERY",
      wildcard: "%QUERY",
    },
  })
  bhPubkey.remote.transport._get = debounce(bhPubkey.remote.transport, bhPubkey.remote.transport._get)
  var bhEth1Addresses = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.eth1_address
    },
    remote: {
      url: "/search/indexed_validators_by_eth1_addresses/%QUERY",
      wildcard: "%QUERY",
    },
  })
  bhEth1Addresses.remote.transport._get = debounce(bhEth1Addresses.remote.transport, bhEth1Addresses.remote.transport._get)
  var bhName = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.name
    },
    remote: {
      url: "/search/indexed_validators_by_name/%QUERY",
      wildcard: "%QUERY",
    },
  })
  bhName.remote.transport._get = debounce(bhName.remote.transport, bhName.remote.transport._get)
  var bhGraffiti = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.graffiti
    },
    remote: {
      url: "/search/indexed_validators_by_graffiti/%QUERY",
      wildcard: "%QUERY",
    },
  })
  bhGraffiti.remote.transport._get = debounce(bhGraffiti.remote.transport, bhGraffiti.remote.transport._get)

  $(".typeahead-dashboard").typeahead(
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
      display: "index",
      templates: {
        header: "<h3>Validators</h3>",
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate high-contrast">${data.index}: ${data.pubkey}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "pubkeys",
      source: bhPubkey,
      display: "pubkey",
      templates: {
        header: "<h3>Validators by Public Key</h3>",
        suggestion: function (data) {
          return `<div class="text-monospace text-truncate high-contrast">${data.pubkey}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "addresses",
      source: bhEth1Addresses,
      display: "address",
      templates: {
        header: "<h3>Validators by ETH Addresses</h3>",
        suggestion: function (data) {
          var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + "+" : data.validator_indices.length
          return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.eth1_address}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        },
      },
    },
    {
      limit: 5,
      name: "graffiti",
      source: bhGraffiti,
      display: "graffiti",
      templates: {
        header: "<h3>Validators by Graffiti</h3>",
        suggestion: function (data) {
          var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + "+" : data.validator_indices.length
          return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.graffiti}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        },
      },
    },
    {
      limit: 5,
      name: "name",
      source: bhName,
      display: "name",
      templates: {
        header: "<h3>Validators by Name</h3>",
        suggestion: function (data) {
          var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + "+" : data.validator_indices.length
          return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.name}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        },
      },
    }
  )
  $(".typeahead-dashboard").on("focus", function (event) {
    if (event.target.value !== "") {
      $(this).trigger($.Event("keydown", { keyCode: 40 }))
    }
  })
  $(".typeahead-dashboard").on("input", function () {
    $(".tt-suggestion").first().addClass("tt-cursor")
  })
  $(".typeahead-dashboard").on("blur", function () {
    $(this).val("")
    $(".typeahead-dashboard").typeahead("val", "")
  })
  $(".typeahead-dashboard").on("typeahead:select", function (ev, sug) {
    if (sug.validator_indices) {
      addValidators(sug.validator_indices)
    } else if (sug.index != null) {
      addValidator(sug.index)
    } else {
      addValidator("0x" + sug.pubkey)
    }
    boxAnimationDirection = "in"
    $(".typeahead-dashboard").typeahead("val", "")
  })

  $("#pending").on("click", "button", function () {
    var data = pendingTable.row($(this).parents("tr")).data()
    removeValidator(data[1])
  })
  $("#active").on("click", "button", function () {
    var data = activeTable.row($(this).parents("tr")).data()
    removeValidator(data[1])
  })
  $("#ejected").on("click", "button", function () {
    var data = ejectedTable.row($(this).parents("tr")).data()
    removeValidator(data[1])
  })
  $("#selected-validators").on("click", ".remove-validator", function () {
    removeValidator(this.parentElement.dataset.validatorIndex)
  })
  $("#selected-validators-input").on("click", ".remove-validator", function () {
    removeValidator(this.parentElement.dataset.validatorIndex)
  })

  $(".multiselect-border input").on("focus", function (event) {
    $(".multiselect-border").addClass("focused")
  })
  $(".multiselect-border input").on("blur", function (event) {
    $(".multiselect-border").removeClass("focused")
  })

  $("#clear-search").on("click", function (event) {
    if (state) {
      state = setInitialState()
      localStorage.removeItem("dashboard_validators")
      window.location = "/dashboard"
      selectedBTNindex = null
    }
  })

  function setInitialState() {
    var _state = {}
    _state.validators = []
    _state.validatorsCount = {
      pending: 0,
      active: 0,
      ejected: 0,
      offline: 0,
    }
    return _state
  }

  var state = setInitialState()

  setValidatorsFromURL()
  renderSelectedValidators()
  updateState()

  function isValidatorPubkey(identifier) {
    return identifier.startsWith("0x") && identifier.length === 98
  }

  function firstValidatorWithIndex() {
    return state.validators.find((v) => !isValidatorPubkey(v))
  }

  function renderSelectedValidators() {
    if (state.validators.length > VALLIMIT) return
    var elHolder = document.getElementById("selected-validators")
    $("#selected-validators .item").remove()
    $("#selected-validators hr").remove()
    var elsItems = []
    for (var i = 0; i < state.validators.length; i++) {
      if (i % 25 === 0 && i !== 0) {
        var hr = document.createElement("hr")
        hr.classList.add("w-100")
        hr.classList.add("my-1")
        elsItems.push(hr)
      }
      var v = state.validators[i]
      var elItem = document.createElement("li")
      elItem.classList = "item"
      elItem.dataset.validatorIndex = v
      var validatorDisplay = v
      if (isValidatorPubkey(v)) {
        validatorDisplay = v.slice(0, 6) + "..." + v.slice(-4)
      }
      elItem.innerHTML = '<i class="fas fa-times-circle remove-validator"></i> <span>' + validatorDisplay + "</span>"
      elsItems.push(elItem)
    }
    elHolder.prepend(...elsItems)
  }

  function addValidatorUpdateUI() {
    $("#validators-tab").removeClass("disabled")
    $("#validator-art").attr("class", "d-none")

    if (firstValidatorWithIndex() !== undefined) {
      $("#dash-validator-history-info").removeClass("d-none")
      $("#dash-validator-history-index-div").removeClass("d-none")
      $("#dash-validator-history-index-div").addClass("d-flex")

      fetch(`/dashboard/data/effectiveness${getValidatorQueryString()}`, {
        method: "GET",
      }).then((res) => {
        res.json().then((data) => {
          if (Object.keys(data).length === 0) {
            return
          }
          let sum = 0.0
          for (let eff of data) {
            sum += eff
          }
          sum = sum / data.length
          setValidatorEffectiveness("validator-eff-total", sum)
        })
      })

      showProposedHistoryTable()
    } else {
      $("#dash-validator-history-info").addClass("d-none")
      $("#dash-validator-history-index-div").removeClass("d-flex")
      $("#dash-validator-history-index-div").addClass("d-none")

      $("#validator-eff-total").html(summaryDefaultValue)
      renderProposedHistoryTable([])
    }

    let anim = "goinboxanim"
    if (boxAnimationDirection === "out") anim = "gooutboxanim"

    $("#selected-validators-input-button-box").addClass("zoomanim")
    $("#selected-validators-input-button-val").addClass(anim)
    setTimeout(() => {
      $("#selected-validators-input-button-box").removeClass("zoomanim")
      $("#selected-validators-input-button-val").removeClass(anim)
    }, 1100)
  }

  function renderDashboardInfo() {
    var el = document.getElementById("dashboard-info")
    var depositedText = ""
    if (state.validatorsCount.deposited > 0) {
      depositedText = `, ${state.validatorsCount.deposited} deposited`
    }
    var slashedText = ""
    if (state.validatorsCount.slashed > 0) {
      slashedText = `, ${state.validatorsCount.slashed} slashed`
    }
    el.innerText = `${state.validatorsCount.active_online + state.validatorsCount.active_offline} active (${state.validatorsCount.active_online} online, ${state.validatorsCount.active_offline} offline)${depositedText}, ${state.validatorsCount.pending} pending, ${state.validatorsCount.exited + state.validatorsCount.slashed} exited validators (${state.validatorsCount.exited} voluntary${slashedText})`

    if (state.validators.length > 0) {
      showSelectedValidator()
      addValidatorUpdateUI()

      firstValidatorWithHistory = firstValidatorWithIndex()
      if (firstValidatorWithHistory === undefined) {
        hideValidatorHist()
      } else if (selectedBTNindex !== firstValidatorWithHistory) {
        // don't query if not necessary)
        showValidatorHist(firstValidatorWithHistory)
      }
    } else {
      $("#validatorModal").modal("hide")
    }

    if (state.validators.length > 0) {
      $("#selected-validators-input-button").removeClass("d-none")
      $("#selected-validators-input-button").addClass("d-flex")
      $("#selected-validators-input-button span").html(state.validators.length)
    } else {
      $("#selected-validators-input-button").removeClass("d-flex")
      $("#selected-validators-input-button").addClass("d-none")
    }
  }

  function setValidatorsFromURL() {
    var usp = new URLSearchParams(window.location.search)
    var validatorsStr = usp.get("validators")
    if (!validatorsStr) {
      validatorsStr = localStorage.getItem("dashboard_validators")
      if (validatorsStr) {
        state.validators = JSON.parse(validatorsStr)
        state.validators = state.validators.filter((v, i) => {
          v = escape(v)
          if (isNaN(parseInt(v))) return false
          return state.validators.indexOf(v) === i
        })
        state.validators.sort(sortValidators)
      } else {
        state.validators = []
      }
      return
    }
    state.validators = validatorsStr.split(",")
    state.validators = state.validators.filter((v, i) => {
      v = escape(v)
      if (isNaN(parseInt(v))) return false
      return state.validators.indexOf(v) === i
    })
    state.validators.sort(sortValidators)

    if (state.validators.length > VALLIMIT) {
      state.validators = state.validators.slice(0, VALLIMIT)
      console.log(`${VALLIMIT} validators limit reached`)
      handleLimitHit()
    }
  }

  function addValidators(indices) {
    var overview = document.getElementById("selected-validators-overview")
    if (state.validators.length === 0) {
      overview.classList.remove("d-none")
    }
    var limitReached = false
    indicesLoop: for (var j = 0; j < indices.length; j++) {
      if (state.validators.length >= VALLIMIT) {
        limitReached = true
        break indicesLoop
      }
      var index = indices[j] + "" // make sure index is string
      for (var i = 0; i < state.validators.length; i++) {
        if (state.validators[i] === index) continue indicesLoop
      }
      state.validators.push(index)
    }

    if (limitReached) {
      console.log(`${VALLIMIT} validators limit reached`)
      handleLimitHit()
    }
    state.validators.sort(sortValidators)
    renderSelectedValidators()
    updateState()
  }

  function handleLimitHit() {
    if (VALLIMIT == 300) {
      // user is already at the top tier, no need to advertise it to them
      alert(`Sorry, too many validators! You can not currently add more than ${VALLIMIT} validators to your dashboard.`)
    } else {
      if (window.confirm(`With your current premium level, you can not add more than ${VALLIMIT} validators to your dashboard.\n\nBy upgrading to the Whale Tier, this limit gets raised to 280 validators!`)) {
        window.location.href = "/premium"
      }
    }
  }

  function addValidator(index) {
    var overview = document.getElementById("selected-validators-overview")
    if (state.validators.length === 0) {
      overview.classList.remove("d-none")
    }
    if (state.validators.length >= VALLIMIT) {
      handleLimitHit()
      return
    }
    index = index + "" // make sure index is string
    for (var i = 0; i < state.validators.length; i++) {
      if (state.validators[i] === index) return
    }
    state.validators.push(index)
    state.validators.sort(sortValidators)
    renderSelectedValidators()
    updateState()
  }

  function removeValidator(index) {
    boxAnimationDirection = "out"
    for (var i = 0; i < state.validators.length; i++) {
      if (state.validators[i] === index) {
        state.validators.splice(i, 1)
        state.validators.sort(sortValidators)
        //removed last validator
        if (state.validators.length === 0) {
          state = setInitialState()
          localStorage.removeItem("dashboard_validators")
          window.location = "/dashboard"
          return
        } else {
          removeValidatorCountdown(parseInt(index))
          renderSelectedValidators()
          updateState()
        }
        return
      }
    }
  }

  function sortValidators(a, b) {
    var ai = parseInt(a)
    var bi = parseInt(b)

    return ai - bi
  }

  function updateState() {
    if (state.validators.length > VALLIMIT) {
      return
    }
    localStorage.setItem("dashboard_validators", JSON.stringify(state.validators))
    window.dispatchEvent(new CustomEvent("dashboard_validators_set"))

    if (state.validators.length) {
      var qryStr = "?validators=" + state.validators.join(",")
      if (window.location.search != qryStr) {
        var newUrl = window.location.pathname + qryStr + window.location.hash
        window.history.replaceState(null, "Dashboard", newUrl)
      }
    }
    var t0 = Date.now()
    if (state.validators && state.validators.length) {
      document.querySelector("#copy-button").style.visibility = "visible"
      document.querySelector("#clear-search").style.visibility = "visible"

      $.ajax({
        url: "/dashboard/data/validators" + qryStr,
        success: function (result) {
          var t1 = Date.now()
          console.log(`loaded validators-data: length: ${result.data.length}, fetch: ${t1 - t0}ms`)
          if (!result || !result.data.length) {
            document.getElementById("validators-table-holder").style.display = "none"
            return
          }
          // pubkey, idx, currbal, effbal, slashed, acteligepoch, actepoch, exitepoch
          // 0:pubkey, 1:idx, 2:[currbal,effbal], 3:state, 4:[actepoch,acttime], 5:[exit,exittime], 6:[wd,wdt], 7:[lasta,lastat], 8:[exprop,misprop]
          state.validatorsCount.deposited = 0
          state.validatorsCount.pending = 0
          state.validatorsCount.active_online = 0
          state.validatorsCount.active_offline = 0
          state.validatorsCount.slashing_online = 0
          state.validatorsCount.slashing_offline = 0
          state.validatorsCount.exiting_online = 0
          state.validatorsCount.exiting_offline = 0
          state.validatorsCount.exited = 0
          state.validatorsCount.slashed = 0

          for (var i = 0; i < result.data.length; i++) {
            var v = result.data[i]
            var vIndex = v[1]
            var vState = v[3][1]
            if (!state.validatorsCount[vState]) state.validatorsCount[vState] = 0
            state.validatorsCount[vState]++
            var el = document.querySelector(`#selected-validators .item[data-validator-index="${vIndex}"]`)
            if (el) el.dataset.state = vState
          }
          validatorsDataTable.clear()

          validatorsDataTable.rows.add(result.data).draw()

          validatorsDataTable.column(6).visible(false)

          requestAnimationFrame(() => {
            validatorsDataTable.columns.adjust().responsive.recalc()
          })

          document.getElementById("validators-table-holder").style.display = "block"

          renderDashboardInfo()
        },
      })

      if (firstValidatorWithIndex() !== undefined) {
        document.querySelector("#rewards-button").style.visibility = "visible"
        document.querySelector("#bookmark-button").style.visibility = "visible"

        $.ajax({
          url: "/dashboard/data/earnings" + qryStr,
          success: function (result) {
            var t1 = Date.now()
            console.log(`loaded earnings: fetch: ${t1 - t0}ms`)
            if (!result) return

            document.querySelector("#earnings-day").innerHTML = result.lastDayFormatted || summaryDefaultValue
            document.querySelector("#earnings-week").innerHTML = result.lastWeekFormatted || summaryDefaultValue
            document.querySelector("#earnings-month").innerHTML = result.lastMonthFormatted || summaryDefaultValue
            document.querySelector("#earnings-total").innerHTML = result.totalFormatted || summaryDefaultValue
            $("#earnings-total").find('[data-toggle="tooltip"]').tooltip()
            document.querySelector("#balance-total").innerHTML = result.totalBalance || summaryDefaultValue
            $("#balance-total span:first").removeClass("text-success").removeClass("text-danger")
            $("#balance-total span:first").html($("#balance-total span:first").html().replace("+", ""))
          },
        })
      } else {
        document.querySelector("#rewards-button").style.visibility = "hidden"
        document.querySelector("#bookmark-button").style.visibility = "hidden"

        document.querySelector("#earnings-day").innerHTML = summaryDefaultValue
        document.querySelector("#earnings-week").innerHTML = summaryDefaultValue
        document.querySelector("#earnings-month").innerHTML = summaryDefaultValue
        document.querySelector("#earnings-total").innerHTML = summaryDefaultValue
        document.querySelector("#balance-total").innerHTML = summaryDefaultValue
      }
    } else {
      document.querySelector("#copy-button").style.visibility = "hidden"
      document.querySelector("#rewards-button").style.visibility = "hidden"
      document.querySelector("#bookmark-button").style.visibility = "hidden"
      document.querySelector("#clear-search").style.visibility = "hidden"
    }

    $("#copy-button").attr("data-clipboard-text", window.location.href)

    if (state.validators && firstValidatorWithIndex() !== undefined) {
      renderCharts()
    } else {
      hideCharts()
    }
  }

  window.onpopstate = function (event) {
    setValidatorsFromURL()
    renderSelectedValidators()
    updateState()
  }
  window.addEventListener("storage", function (e) {
    var validatorsStr = localStorage.getItem("dashboard_validators")
    if (JSON.stringify(state.validators) === validatorsStr) {
      return
    }
    if (validatorsStr) {
      state.validators = JSON.parse(validatorsStr)
    } else {
      state.validators = []
    }
    state.validators = state.validators.filter((v, i) => state.validators.indexOf(v) === i)
    state.validators.sort(sortValidators)
    renderSelectedValidators()
    updateState()
  })

  function hideCharts() {
    hideIncomeChart()
    hideProposedChart()
  }

  function hideIncomeChart() {
    if (incomeChart) {
      incomeChart.destroy()
      incomeChart = null
    }
    document.getElementById("balance-chart").innerHTML = incomeChartDefault
  }

  function hideProposedChart() {
    if (proposedChart) {
      proposedChart.destroy()
      proposedChart = null
    }
    document.getElementById("proposed-chart").innerHTML = proposedChartDefault
  }

  function renderCharts() {
    var t0 = Date.now()
    var qryStr = "?validators=" + state.validators.join(",")
    $.ajax({
      url: "/dashboard/data/allbalances" + qryStr + "&days=31",
      success: function (result) {
        var t1 = Date.now()
        createIncomeChart(result.consensusChartData, result.executionChartData)
        var t2 = Date.now()
        console.log(`loaded balance-data: length: ${result.length}, fetch: ${t1 - t0}ms, render: ${t2 - t1}ms`)
        allIncomeLoaded = false
        $("#load-income-btn").removeClass("d-none")
      },
    })
    $.ajax({
      url: "/dashboard/data/proposals" + qryStr,
      success: function (result) {
        var t1 = Date.now()
        if (result && result.length) {
          createProposedChart(result)
        } else {
          var chart = $("#proposed-chart").highcharts()
          if (chart !== undefined) {
            hideProposedChart()
          }
        }
        var t2 = Date.now()
        console.log(`loaded proposal-data: length: ${result.length}, fetch: ${t1 - t0}ms, render: ${t2 - t1}ms`)
      },
    })
  }

  $("#load-income-btn").on("click", () => {
    if (allIncomeLoaded || incomeChart == null) {
      return
    }
    allIncomeLoaded = true

    const url = "/dashboard/data/allbalances?validators=" + state.validators.join(",")
    $("#load-income-btn").text("Loading...")
    fetch(url)
      .then((response) => {
        if (!response.ok) {
          throw new Error("Network response was not ok.")
        }
        return response.json()
      })
      .then((data) => {
        createIncomeChart(data.consensusChartData, data.executionChartData)
        $("#load-income-btn").addClass("d-none")
      })
      .catch((error) => {
        console.error(error)
        alert("Error loading income data. Please try again.")
        allIncomeLoaded = false
      })
      .finally(() => {
        $("#load-income-btn").text("Show all rewards")
      })
  })
})

function createIncomeChart(income, executionIncomeHistory) {
  executionIncomeHistory = executionIncomeHistory || []
  const incomeChartOptions = getIncomeChartOptions(income, executionIncomeHistory, "Daily Income for all Validators", 627)
  incomeChart = Highcharts.stockChart("balance-chart", incomeChartOptions)
}

function createProposedChart(data) {
  var proposed = []
  var missed = []
  var orphaned = []
  data.map((d) => {
    if (d[1] == 1) proposed.push([d[0] * 1000, 1])
    else if (d[1] == 2) missed.push([d[0] * 1000, 1])
    else if (d[1] == 3) orphaned.push([d[0] * 1000, 1])
  })
  proposedChart = Highcharts.stockChart("proposed-chart", {
    chart: {
      type: "column",
      height: "630px",
    },
    title: {
      text: "Proposal History for all Validators",
    },
    legend: {
      enabled: true,
    },
    colors: ["#7cb5ec", "#ff835c", "#e4a354", "#2b908f", "#f45b5b", "#91e8e1"],
    xAxis: {
      lineWidth: 0,
      tickColor: "#e5e1e1",
    },
    yAxis: [
      {
        title: {
          text: "# of Possible Proposals",
        },
        opposite: false,
      },
    ],
    plotOptions: {
      column: {
        stacking: "normal",
        dataGrouping: {
          enabled: true,
          forced: true,
          units: [["day", [1]]],
        },
      },
    },
    series: [
      {
        name: "Proposed",
        color: "#7cb5ec",
        data: proposed,
      },
      {
        name: "Missed",
        color: "#ff835c",
        data: missed,
      },
      {
        name: "Missed (Orphaned)",
        color: "#e4a354",
        data: orphaned,
      },
    ],
    rangeSelector: {
      enabled: false,
    },
  })
  $(".proposal-switch").show()
}
