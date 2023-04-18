function createBlock(x, y) {
  use = document.createElementNS("http://www.w3.org/2000/svg", "use")
  // use.setAttributeNS(null, "style", `transform: translate(calc(${x} * var(--disperse-factor)), calc(${y} * var(--disperse-factor)));`)
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
    // use.setAttributeNS(null, "style", `transform: translate(calc(${x} * var(--disperse-factor)), calc(${y} * var(--disperse-factor)));`)
    use.setAttributeNS(null, "href", "#cube-small")
    use.setAttributeNS(null, "x", 129)
    use.setAttributeNS(null, "y", 56)
    cube.appendChild(use)
  }
}

var selectedBTNindex = null
var VALLIMIT = 280
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
    //pagingType: 'input', //not working
    pagingType: "simple",
    pageLength: 10,
    ajax: "/validator/" + index + "/history",
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
  $("#dash-validator-history-art").attr("class", "d-none")
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

  // searchInput.addEventListener('blur', function(ev) {
  //   var overview = document.getElementById('selected-validators-overview')
  //   // overview.classList.add('d-none')
  // })

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

function addValidatorUpdateUI() {
  $("#validators-tab").removeClass("disabled")
  $("#validator-art").attr("class", "d-none")
  $("#dash-validator-history-info").removeClass("d-none")
  $("#dash-validator-history-index-div").removeClass("d-none")
  $("#dash-validator-history-index-div").addClass("d-flex")
  // $('#selected-validators-input-button-val').removeClass('d-none')
  let anim = "goinboxanim"
  if (boxAnimationDirection === "out") anim = "gooutboxanim"

  $("#selected-validators-input-button-box").addClass("zoomanim")
  $("#selected-validators-input-button-val").addClass(anim)
  setTimeout(() => {
    // $('#selected-validators-input-button-val').addClass('d-none')
    $("#selected-validators-input-button-box").removeClass("zoomanim")
    $("#selected-validators-input-button-val").removeClass(anim)
  }, 1100)

  fetch(`/dashboard/data/effectiveness${getValidatorQueryString()}`, {
    method: "GET",
  }).then((res) => {
    res.json().then((data) => {
      let sum = 0.0
      for (let eff of data) {
        sum += eff
      }
      sum = sum / data.length
      setValidatorEffectiveness("validator-eff-total", sum)
    })
  })
  showProposedHistoryTable()
}

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
          return "<span>" + getRelativeTime(luxon.DateTime.fromMillis(data * 1000)) + "</span>"
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

// var proposedHistTableData = []
function showProposedHistoryTable() {
  // if (proposedHistTableData.length===0){
  fetch("/dashboard/data/proposalshistory" + getValidatorQueryString(), {
    method: "GET",
  }).then((res) => {
    res.json().then(function (data) {
      let proposedHistTableData = []
      for (let item of data) {
        proposedHistTableData.push([item[0], item[1], [item[2], item[3], item[4]]])
      }
      renderProposedHistoryTable(proposedHistTableData)
    })
  })
  // }else{
  //   renderProposedHistoryTable(proposedHistTableData)
  // }
}

function switchFrom(el1, el2, el3, el4) {
  $(el1).removeClass("proposal-switch-selected")
  $(el2).addClass("proposal-switch-selected")
  $(el3).addClass("d-none")
  $(el4).removeClass("d-none")
}

var firstSwitch = true

$(document).ready(function () {
  $("button").on("mousedown", (evt) => {
    evt.preventDefault() // prevent setting the browser focus on all mouse buttons, which prevents tooltips from disapearing
  })
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
    // var spinnerSmall = $('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Loading...</span></div>')
    var bookmarkIcon = $("<i class='far fa-bookmark' style='width:15px;'></i>")
    var errorIcon = $("<i class='fas fa-exclamation' style='width:15px;'></i>")
    fetch("/dashboard/save", {
      method: "POST",
      // credentials: 'include',
      headers: {
        "Content-Type": "application/json",
        // 'X-CSRF-Token': $("#bookmark-button").attr("csrf"),
      },
      body: JSON.stringify(state.validators),
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

  var clearSearch = $("#clear-search")
  //'<i class="fa fa-copy"></i>'
  var copyIcon = $("<i class='fa fa-copy' style='width:15px'></i>")
  //'<i class="fas fa-check"></i>'
  var tickIcon = $("<i class='fas fa-check' style='width:15px;'></i>")

  clearSearch.on("click", function () {
    clearSearch.empty().append(tickIcon)
    setTimeout(function () {
      clearSearch.empty().append(copyIcon)
    }, 500)
  })

  var validatorsDataTable = (window.vdt = $("#validators").DataTable({
    processing: true,
    serverSide: false,
    ordering: true,
    lengthChange: false,
    searching: true,
    pagingType: "full_numbers",
    lengthMenu: [10, 25, 50],
    info: false,
    language: {
      search: "",
      searchPlaceholder: "Search...",
    },
    preDrawCallback: function () {
      // this does not always work.. not sure how to solve the staying tooltip
      try {
        $("#validators").find('[data-toggle="tooltip"]').tooltip("dispose")
      } catch (e) {
        console.error(e)
      }
    },
    drawCallback: function (settings) {
      $("#validators").find('[data-toggle="tooltip"]').tooltip()
    },
    order: [[1, "asc"]],
    columnDefs: [
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
          // return '<a href="/validator/' + data + '">0x' + data.substr(0, 8) + '...</a>'
          return `<a href="/validator/${data}">0x${data.substr(0, 8)}...</a><i class="fa fa-copy text-muted p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="0x${data}"></i>`
        },
      },
      {
        targets: 1,
        data: "1",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data
          // return '<a href="/validator/' + data + '">' + data + '</a>'
          return `<span class="m-0 p-2 hbtn" id="dropdownMenuButton${data}" style="cursor: pointer;" onclick="showValidatorHist('${data}')">
                      ${data}
                  </span>
                 `
        },
      },
      {
        targets: 2,
        data: "2",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          return `${data[0]}`
        },
      },
      {
        targets: 3,
        data: "3",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : -1
          var d = data.split("_")
          var s = d[0].charAt(0).toUpperCase() + d[0].slice(1)
          if (d[1] === "offline") return `<span style="display:none">${d[1]}</span><span data-toggle="tooltip" data-placement="top" title="No attestation in the last 2 epochs">${s} <i class="fas fa-power-off fa-sm text-danger"></i></span>`
          if (d[1] === "online") return `<span style="display:none">${d[1]}</span><span>${s} <i class="fas fa-power-off fa-sm text-success"></i></span>`
          return `<span>${s}</span>`
        },
      },
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
      {
        targets: 6,
        data: "6",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          if (data === null) return "-"
          return `<span data-toggle="tooltip" data-placement="top" title="${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))}">${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        },
      },
      {
        targets: 7,
        data: "7",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] : null
          if (data === null) return "No Attestation found"
          return `<span>${getRelativeTime(luxon.DateTime.fromMillis(data[1] * 1000))}</span>`
        },
      },
      {
        targets: 8,
        data: "8",
        render: function (data, type, row, meta) {
          if (type == "sort" || type == "type") return data ? data[0] + data[1] : null
          return `<span data-toggle="tooltip" data-placement="top" title="${data[0]} executed / ${data[1]} missed"><span class="text-success">${data[0]}</span> / <span class="text-danger">${data[1]}</span></span>`
        },
      },
    ],
  }))

  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.index
    },
    remote: {
      url: "/search/indexed_validators/%QUERY",
      wildcard: "%QUERY",
    },
  })
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
    } else {
      addValidator(sug.index)
    }
    boxAnimationDirection = "in"
    // addValidatorUpdateUI()
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
    // window.location = "/dashboard"
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
      elItem.innerHTML = '<i class="fas fa-times-circle remove-validator"></i> <span>' + v + "</span>"
      elsItems.push(elItem)
    }
    elHolder.prepend(...elsItems)
  }

  function renderDashboardInfo() {
    var el = document.getElementById("dashboard-info")
    var slashedText = ""
    if (state.validatorsCount.slashed > 0) {
      slashedText = `, ${state.validatorsCount.slashed} slashed`
    }
    el.innerText = `${state.validatorsCount.active_online + state.validatorsCount.active_offline} active (${state.validatorsCount.active_online} online, ${state.validatorsCount.active_offline} offline), ${state.validatorsCount.pending} pending, ${state.validatorsCount.exited + state.validatorsCount.slashed} exited validators (${state.validatorsCount.exited} voluntary${slashedText})`

    if (state.validators.length > 0) {
      showSelectedValidator()
      addValidatorUpdateUI()
      if (selectedBTNindex != state.validators[0]) {
        // don't query if not necessary
        showValidatorHist(state.validators[0])
      }
      // showValidatorsInSearch(3)
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
    // if (state.validators.length >= VALLIMIT) {
    //   alert(`You can not add more than ${VALLIMIT} validators to your dashboard`)
    //   return
    // }
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

  // function addChange(selector, value) {
  //   if(selector !== undefined || selector !== null) {
  //     var element = document.querySelector(selector)
  //     if(element !== undefined) {
  //       // remove old
  //       element.classList.remove('decreased')
  //       element.classList.remove('increased')
  //       if(value < 0) {
  //         element.classList.add("decreased")
  //       }
  //       if (value > 0) {
  //         element.classList.add("increased")
  //       }
  //     } else {
  //       console.error("Could not find element with selector", selector)
  //     }
  //   } else {
  //     console.error("selector is not defined", selector)
  //   }
  // }

  function updateState() {
    // if(_range < xBlocks.length + 3 && _range !== -1) {

    //   appendBlocks(xBlocks.slice(_range, _range+3))
    //   _range = _range + 3;
    // } else if(_range !== -1) {
    //   _range = -1;
    // }
    if (state.validators.length > VALLIMIT) {
      // alert(`Too many validators, you can not add more than ${VALLIMIT} validators to your dashboard!`)
      return
    }
    localStorage.setItem("dashboard_validators", JSON.stringify(state.validators))
    window.dispatchEvent(new CustomEvent("dashboard_validators_set"))

    if (state.validators.length) {
      // console.log('length', state.validators)
      var qryStr = "?validators=" + state.validators.join(",")
      var newUrl = window.location.pathname + qryStr
      window.history.replaceState(null, "Dashboard", newUrl)
    }
    var t0 = Date.now()
    if (state.validators && state.validators.length) {
      // if(state.validators.length >= 9) {
      //   appendBlocks(xBlocks)
      // } else {
      //   appendBlocks(xBlocks.slice(0, state.validators.length * 3 - 1))
      // }
      document.querySelector("#rewards-button").style.visibility = "visible"
      document.querySelector("#bookmark-button").style.visibility = "visible"
      document.querySelector("#copy-button").style.visibility = "visible"
      document.querySelector("#clear-search").style.visibility = "visible"

      $.ajax({
        url: "/dashboard/data/earnings" + qryStr,
        success: function (result) {
          var t1 = Date.now()
          console.log(`loaded earnings: fetch: ${t1 - t0}ms`)
          if (!result) return

          document.querySelector("#earnings-day").innerHTML = result.lastDayFormatted || "0.000"
          document.querySelector("#earnings-week").innerHTML = result.lastWeekFormatted || "0.000"
          document.querySelector("#earnings-month").innerHTML = result.lastMonthFormatted || "0.000"
          document.querySelector("#earnings-total").innerHTML = result.totalFormatted || "0.000"
          $("#earnings-total").find('[data-toggle="tooltip"]').tooltip()
          document.querySelector("#balance-total").innerHTML = result.totalBalance || "0.000"
          $("#balance-total span:first").removeClass("text-success").removeClass("text-danger")
          $("#balance-total span:first").html($("#balance-total span:first").html().replace("+", ""))
          // addChange("#earnings-total-change", result.total)
        },
      })
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
          // console.log(`latestEpoch: ${result.latestEpoch}`)
          // var latestEpoch = result.latestEpoch
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
            var vState = v[3]
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
    } else {
      document.querySelector("#copy-button").style.visibility = "hidden"
      document.querySelector("#rewards-button").style.visibility = "hidden"
      document.querySelector("#bookmark-button").style.visibility = "hidden"
      document.querySelector("#clear-search").style.visibility = "hidden"
      // window.location = "/dashboard"
    }

    $("#copy-button").attr("data-clipboard-text", window.location.href)

    renderCharts()
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

  function renderCharts() {
    var t0 = Date.now()
    // if (state.validators.length === 0) {
    //   document.getElementById('chart-holder').style.display = 'none'
    //   return
    // }
    // document.getElementById('chart-holder').style.display = 'flex'
    if (state.validators && state.validators.length) {
      var qryStr = "?validators=" + state.validators.join(",")
      $.ajax({
        url: "/dashboard/data/allbalances" + qryStr,
        success: function (result) {
          var t1 = Date.now()
          // let prevDayIncome = 0
          // let prevDay = null
          // let prevIncome = 0
          // for (var i = 0; i < result.length; i++) {
          //   var res = result[i]

          //   let day = new Date(res[0])
          //   if (prevDay===null) prevDay=day
          //   // balance[i] = [res[0], res[2]-(i===0 ? res[2] : prevBalance)]
          //   prevDayIncome+=res[2]-(i===0 ? res[2] : prevIncome)
          //   prevIncome = res[2]
          //   // console.log(day!==prevDay, day, prevDay, res[0])
          //   if (day.getDay()!==prevDay.getDay()){
          //     income.push([day.getTime(), prevDayIncome])
          //     prevDayIncome = 0
          //     prevDay=day
          //   }
          // }

          var t2 = Date.now()
          createBalanceChart(result.consensusChartData, result.executionChartData)
          var t3 = Date.now()
          console.log(`loaded balance-data: length: ${result.length}, fetch: ${t1 - t0}ms, aggregate: ${t2 - t1}ms, render: ${t3 - t2}ms`)
        },
      })
      $.ajax({
        url: "/dashboard/data/proposals" + qryStr,
        success: function (result) {
          var t1 = Date.now()
          var t2 = Date.now()
          if (result && result.length) {
            createProposedChart(result)
          }
          var t3 = Date.now()
          console.log(`loaded proposal-data: length: ${result.length}, fetch: ${t1 - t0}ms, render: ${t3 - t2}ms`)
        },
      })
    }
  }
})

function createBalanceChart(income, executionIncomeHistory) {
  executionIncomeHistory = executionIncomeHistory || []
  // console.log("u", utilization)
  Highcharts.stockChart("balance-chart", {
    exporting: {
      scale: 1,
    },
    rangeSelector: {
      enabled: false,
    },
    chart: {
      type: "column",
      height: "500px",
      pointInterval: 24 * 3600 * 1000,
    },
    legend: {
      enabled: true,
    },
    title: {
      text: "Daily Income for all Validators",
    },
    navigator: {
      series: {
        data: income,
        color: "#7cb5ec",
      },
    },
    plotOptions: {
      column: {
        stacking: "stacked",
        dataLabels: {
          enabled: false,
        },
        pointInterval: 24 * 3600 * 1000,
        // pointIntervalUnit: 'day',
        dataGrouping: {
          forced: true,
          units: [["day", [1]]],
        },
      },
    },
    xAxis: {
      type: "datetime",
      range: 31 * 24 * 60 * 60 * 1000,
      labels: {
        formatter: function () {
          var epoch = timeToEpoch(this.value)
          var orig = this.axis.defaultLabelFormatter.call(this)
          return `${orig}<br/>Epoch ${epoch}`
        },
      },
    },
    tooltip: {
      formatter: function (tooltip) {
        var orig = tooltip.defaultFormatter.call(this, tooltip)
        var epoch = timeToEpoch(this.x)
        orig[0] = `${orig[0]}<span style="font-size:10px">Epoch ${epoch}</span>`
        if (currency !== "ETH") {
          orig[1] = `<span style="color:${this.points[0].color}">●</span> Daily Income: <b>${this.y.toFixed(2)}</b><br/>`
        }
        return orig
      },
      dateTimeLabelFormats: {
        day: "%A, %b %e, %Y",
        minute: "%A, %b %e",
        hour: "%A, %b %e",
      },
    },
    yAxis: [
      {
        title: {
          text: "Income [" + currency + "]",
        },
        opposite: false,
        labels: {
          formatter: function () {
            if (currency !== "ETH") {
              return this.value.toFixed(2)
            }
            return this.value.toFixed(5)
          },
        },
      },
    ],
    series: [
      {
        name: "Daily Consensus Income",
        data: income,
        index: 2,
      },
      {
        name: "Daily Execution Income",
        data: executionIncomeHistory,
        index: 1,
      },
    ],
  })
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
  Highcharts.stockChart("proposed-chart", {
    chart: {
      type: "column",
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
}
