var csrfToken = ""

const VALIDATOR_EVENTS = ["validator_attestation_missed", "validator_proposal_missed", "validator_proposal_submitted", "validator_got_slashed", "validator_synccommittee_soon", "validator_is_offline", "validator_withdrawal"]

// const MONITORING_EVENTS = ['monitoring_machine_offline', 'monitoring_hdd_almostfull', 'monitoring_cpu_load']

function create_typeahead(input_container) {
  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.index
    },
    remote: {
      url: "/search/validators/%QUERY",
      wildcard: "%QUERY",
    },
  })
  var bhName = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function (obj) {
      return obj.validator_indices.join(",")
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
  $(input_container).typeahead(
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
        header: '<h5 class="font-weight-bold ml-3">Validators</h5>',
        suggestion: function (data) {
          return `<div class="font-weight-normal text-truncate high-contrast">${data.index}: ${data.pubkey}</div>`
        },
      },
    },
    {
      limit: 5,
      name: "name",
      source: bhName,
      display: function (data) {
        return data && data.validator_indices && data.validator_indices.length ? data.validator_indices.join(",") : ""
      },
      templates: {
        header: '<h5 class="font-weight-bold ml-3">Validators by Name</h5>',
        suggestion: function (data) {
          var len = data.validator_indices.length > 10 ? 10 + "+" : data.validator_indices.length
          return `<div class="font-weight-normal high-contrast" style="display: flex;"><div class="text-truncate" style="flex: 1 1 auto;">${data.name}</div><div style="max-width: fit-content; white-space: nowrap;">${len}</div></div>`
        },
      },
    },
    {
      limit: 5,
      name: "addresses",
      source: bhEth1Addresses,
      display: function (data) {
        return data && data.validator_indices && data.validator_indices.length ? data.validator_indices.join(",") : ""
      },
      templates: {
        header: '<h5 class="font-weight-bold ml-3">ETH Address</h5>',
        suggestion: function (data) {
          var len = data.validator_indices.length > 10 ? 10 + "+" : data.validator_indices.length
          return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.eth1_address}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        },
      },
    },
    {
      limit: 5,
      name: "graffiti",
      source: bhGraffiti,
      display: function (data) {
        return data && data.validator_indices && data.validator_indices.length ? data.validator_indices.join(",") : ""
      },
      templates: {
        header: '<h5 class="font-weight-bold ml-3">Graffiti</h5>',
        suggestion: function (data) {
          var len = data.validator_indices.length > 10 ? 10 + "+" : data.validator_indices.length
          return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.graffiti}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        },
      },
    }
  )
  $(input_container).on("focus", function (e) {
    if (e.target.value !== "") {
      $(this).trigger($.Event("keydown", { keyCode: 40 }))
    }
  })
  $(input_container).on("input", function () {
    $(".tt-suggestion").first().addClass("tt-cursor")
  })
  // $(input_container).on("typeahead:select", function (e, sug) {

  //   // console.log('sug: ', sug.validator_indices, sug.validator_indices.length)
  //   // $(input_container).val('asdf')
  //   console.log('typeahead select', e, sug)

  //   if (sug.validator_indices && sug.validator_indices.length) {
  //     $(input_container).val(sug.validator_indices.join(','))
  //   }
  //   // else {
  //     // $(input_container).val(sug.index)
  //     // $(input_container).attr("pk", sug.pubkey)
  //     $("#add-validator-input").val("asdf")
  //   // }
  // })
}

function loadMonitoringData(data) {
  let mdata = []
  // let id = 0

  if (!data) {
    data = []
  }

  for (let i = 0; i < data.length; i++) {
    mdata.push({
      id: data[i].ID,
      notification: data[i].EventName,
      threshold: data[i].EventThreshold ? 1 - data[i].EventThreshold : data[i].EventThreshold,
      mostRecent: data[i].LastSent || 0,
      machine: data[i].EventFilter,
      event: {},
    })
  }
  // for (let item of data) {
  //   for (let n of item.Notifications) {
  //     let nn = n.Notification.split(':')
  //     nn = nn[nn.length - 1]
  //     let ns = nn.split('_')
  //     if (ns[0] === 'monitoring') {
  //       if (ns[1] === 'machine') {
  //         ns[1] = ns[2]
  //       }
  //       mdata.push({
  //         id: id,
  //         notification: ns[1],
  //         threshold: [n.Threshold, item],
  //         machine: item.Validator.Index,
  //         mostRecent: n.Timestamp,
  //         event: { pk: item.Validator.Pubkey, e: nn }
  //       })
  //       id += 1
  //     }
  //   }
  // }

  if (data.length !== 0) {
    if ($("#monitoring-section-with-data").children().length === 0) {
      $("#monitoring-section-with-data").append(
        `<table class="table table-borderless table-hover" id="monitoring-notifications">
          <thead class="custom-table-head">
            <tr>
              <th scope="col" class="h6 border-bottom-0">Notification</th>
              <th scope="col" class="h6 border-bottom-0">Threshold</th>
              <th scope="col" class="h6 border-bottom-0">Machine</th>
              <th scope="col" class="h6 border-bottom-0">Most Recent</th>
              <!-- <th scope="col" class="h6 border-bottom-0"></th> -->
            </tr>
          </thead>
          <tbody></tbody>
        </table>`
      )
    }
  } else {
    $("#monitoring-section-empty").removeAttr("hidden")
  }

  let monitoringTable = $("#monitoring-notifications")

  monitoringTable.DataTable({
    language: {
      info: "_TOTAL_ entries",
      infoEmpty: "No entries match",
      infoFiltered: "(from _MAX_ entries)",
      processing: "Loading. Please wait...",
      search: "",
      searchPlaceholder: "Search...",
      zeroRecords: "No entries match",
    },
    stateSave: true,
    processing: true,
    responsive: true,
    scroller: true,
    scrollY: 380,
    paging: true,
    data: mdata,
    rowId: "id",
    initComplete: function (settings, json) {
      $("body").find(".dataTables_scrollBody").addClass("scrollbar")

      // click event to monitoring table edit button
      $("#monitoring-notifications #edit-monitoring-events").on("click", function (e) {
        $("#add-monitoring-validator-select").html("")
        for (let item of $("input.monitoring")) {
          $(item).prop("checked", false)
        }

        let ev = $(this).attr("event").split(",")
        for (let i of ev) {
          if (i.length > 0) {
            let t = i.split(":")
            for (let item of $("input.monitoring")) {
              let e = $(item).attr("event")
              if (e === t[0]) {
                $(item).prop("checked", true)
                let p = parseInt(parseFloat(t[1]) * 100)
                if (e.includes("_cpu_")) {
                  $("#cpu-input-range-val, #cpu-input-range").val(p)
                  $("#cpu-input-range").attr("style", `background-size: ${p}% 100%`)
                } else if (e.includes("_hdd_")) {
                  $("#hdd-input-range-val, #hdd-input-range").val(p)
                  $("#hdd-input-range").attr("style", `background-size: ${p}% 100%`)
                }
              }
            }
          }
        }
        $("#add-monitoring-validator-select").append(`<option value="${$(this).attr("pk")}">${$(this).attr("ind")}</option>`)
      })

      // click event to table remove button
      $("#monitoring-notifications #remove-btn").on("click", function (e) {
        $("#modaltext").text($(this).data("modaltext"))

        // set the row id
        let rowId = $(this).parent().parent().attr("id")
        if (rowId === undefined) {
          rowId = 0
        }
        $("#confirmRemoveModal").attr("rowId", rowId)
        $("#confirmRemoveModal").attr("tablename", "monitoring")
        $("#confirmRemoveModal").attr("filter", $(this).attr("filter"))
        $("#confirmRemoveModal").attr("event", $(this).attr("event"))
      })
    },
    columnDefs: [
      {
        targets: "_all",
        createdCell: function (td, cellData, rowData, row, col) {
          $(td).css("padding-top", "20px")
          $(td).css("padding-bottom", "20px")
        },
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: "notification",
        render: function (data, type, row, meta) {
          data = data.replace(/^[a-zA-Z]+:/, "")
          // var monitoringEvents = {
          //   'monitoring_machine_offline': 'Machine Offline',
          // }
          // data = monitoringEvents[data]
          return `<span class="badge badge-pill badge-light badge-custom-size font-weight-normal">${data}</span>`
        },
      },
      {
        targets: 1,
        responsivePriority: 3,
        data: "threshold",
        render: function (data, type, row, meta) {
          // if (type === 'display') {
          //   let e = ""
          //   // for (let i of data[1].Notifications) {
          //   //   let nn = i.Notification.split(':')
          //   //   nn = nn[nn.length - 1]
          //   //   let ns = nn.split('_')
          //   //   if (ns[0] === 'monitoring') {
          //   //     e += `${nn}:${i.Threshold},`
          //   //   }
          //   // }

          //   // for machine offline event, there is no threshold value; we show N/A and hide the edit button
          //   // replaced (data[0] * 100).toFixed(2) with Math.trunc(data[0] * 100)
          //   return `
          //     <input type="text" class="form-control input-sm threshold_editable" title="Numbers in 1-100 range (including)" style="width: 60px; height: 30px;" hidden />
          //     <span class="threshold_non_editable">
          //       <span class="threshold_non_editable_text">${data[0] === "0" ? "N/A" : Math.trunc(data[0] * 100) + "%"}</span>
          //       <i
          //         class="fas fa-pen fa-xs text-muted i-custom ${data[0] === '0' ? 'd-none' : ''}"
          //         id="edit-monitoring-events"
          //         title="Click to edit"
          //         style="padding: .5rem; cursor: pointer;"
          //         data-toggle= "modal"
          //         data-target="#addMonitoringEventModal"
          //         pk="${data[1].Validator.Pubkey}"
          //         ind="${data[1].Validator.Index}"
          //         event="${e}"
          //       ></i>
          //     </span>`
          // }
          if (data) {
            return (data * 100).toFixed(0) + "%"
          } else {
            return "N/A"
          }
        },
      },
      {
        targets: 2,
        responsivePriority: 2,
        data: "machine",
        render: function (data, type, row, meta) {
          return `<span class="font-weight-bold"><i class="fas fa-server mr-2"></i></i>${data}</span>`
        },
      },
      {
        targets: 3,
        responsivePriority: 1,
        data: "mostRecent",
        render: function (data, type, row, meta) {
          // for sorting and type checking use the original data (unformatted)
          if (type === "sort" || type === "type") {
            return data
          }
          return `<span class="heading-l4">${getRelativeTime(getLuxonDateFromTimestamp(data)) || "N/A"}</span>`
        },
      },
      // {
      //   targets: 4,
      //   orderable: false,
      //   responsivePriority: 3,
      //   data: 'event',
      //   render: function (data, type, row, meta) {
      //     return `<i class="fas fa-times fa-lg i-custom" filter="${row.machine}" event="${row.notification}" id="remove-btn" title="Remove notification" style="padding: .5rem; color: var(--red); cursor: pointer;" data-toggle="modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>`
      //   }
      // }
    ],
  })
}

function loadNetworkData(data) {
  let networkTable = $("#network-notifications")

  networkTable.DataTable({
    language: {
      info: "_TOTAL_ entries",
      infoEmpty: "No entries match",
      infoFiltered: "(from _MAX_ entries)",
      processing: "Loading. Please wait...",
      search: "",
      searchPlaceholder: "Search...",
      zeroRecords: "No entries match",
    },
    stateSave: true,
    processing: true,
    responsive: false,
    scroller: true,
    scrollY: 380,
    paging: true,
    data: data,
    initComplete: function (settings, json) {
      $("body").find(".dataTables_scrollBody").addClass("scrollbar")
    },
    columnDefs: [
      {
        targets: "_all",
        createdCell: function (td, cellData, rowData, row, col) {
          $(td).css("padding-top", "20px")
          $(td).css("padding-bottom", "20px")
        },
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: "Notification",
        render: function (data, type, row, meta) {
          return `<span class="badge badge-pill badge-light badge-custom-size font-weight-normal">${data}</span>`
        },
      },
      {
        targets: 1,
        responsivePriority: 2,
        data: "Network",
      },
      {
        targets: 2,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: `
          <div class="form-check">
        		<input class="form-check-input checkbox-custom-size" type="checkbox">
            <label class="form-check-label"></label>
          </div>`,
        visible: false,
      },
      {
        targets: 3,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: `
          <div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox">
            <label class="form-check-label"></label>
          </div>`,
        visible: false,
      },
      {
        targets: 4,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: `
          <div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox">
            <label class="form-check-label"></label>
          </div>`,
        visible: false,
      },
      {
        targets: 5,
        responsivePriority: 1,
        data: "Timestamp",
        render: function (data, type, row, meta) {
          if (type === "sort" || type === "type") {
            return data
          }
          return `<span class="heading-l4">${getRelativeTime(luxon.DateTime.fromMillis(data))}</span>`
        },
      },
    ],
    order: [[5, "desc"]],
  })
}

function loadValidatorsData(data) {
  let validatorsTable = $("#validators-notifications")
  validatorsTable.DataTable({
    language: {
      info: "_TOTAL_ entries",
      infoEmpty: "No entries match",
      infoFiltered: "(from _MAX_ entries)",
      processing: "Loading. Please wait...",
      search: "",
      searchPlaceholder: "Search...",
      zeroRecords: "No entries match",
      paginate: {
        previous: '<i class="fas fa-chevron-left"></i>',
        next: '<i class="fas fa-chevron-right"></i>',
      },
    },
    stateSave: true,
    processing: true,
    // responsive: true,
    paging: true,
    pagingType: "input",
    select: {
      items: "row",
      toggleable: true,
      // blurable: true,
    },
    fixedHeader: true,
    data: data,
    drawCallback: function (settings) {
      $('[data-toggle="tooltip"]').tooltip()

      // click event to validators table edit button
      $("#validators-notifications #edit-validator-events").on("click", function (e) {
        let row = $(this).parent().parent().parent()
        $("#ManageNotificationModal").attr("rowId", "")
        $("#ManageNotificationModal").attr("subscriptions", "")
        $("#ManageNotificationModal").attr("rowId", row.attr("id"))
        $("#ManageNotificationModal").attr("subscriptions", row.find("div[subscriptions]").attr("subscriptions"))
      })

      // click event to remove button
      $("#validators-notifications #remove-btn").on("click", function (e) {
        const rowId = $(this).parent().parent().parent().attr("id")
        $("#modaltext").text($(this).data("modaltext"))

        // set the row id
        $("#confirmRemoveModal").attr("rowId", rowId)
        $("#confirmRemoveModal").attr("tablename", "validators")
      })
    },
    initComplete: function (settings, json) {
      $("body").find(".dataTables_scrollBody").addClass("scrollbar")
    },
    columnDefs: [
      {
        targets: "_all",
        createdCell: function (td, cellData, rowData, row, col) {
          $(td).css("padding-top", "20px")
          $(td).css("padding-bottom", "20px")
        },
      },
      {
        targets: 0,
        responsivePriority: 2,
        data: null,
        render: function (data, type, row, meta) {
          if (type === "sort" || type === "type") {
            return row.Index
          }
          let datahref = `/validator/${row.Index || row.Pubkey}`
          return `<div class="d-flex align-items-center"><i style="flex 0 0 1rem" class="fas fa-male mr-2"></i><a style="flex: 1 1;" class="font-weight-bold no-highlight mx-2 d-flex flex-wrap" href=${datahref}><span>` + row.Index + `</span><span style="flex-basis: 100%;" class="heading-l4 d-none d-sm-inline-flex">0x` + row.Pubkey.substring(0, 6) + ` ...</span></a><span style="flex: 1 1 0;"></span><i style="flex: 0 0 1rem;" class="fa fa-copy text-muted d-none d-sm-inline p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="0x${row.Pubkey}"></i></div>`
        },
      },
      {
        targets: 1,
        responsivePriority: 2,
        data: null,
        render: function (data, type, row, meta) {
          if (type === "display") {
            if (!row.Notification) {
              row.Notification = []
            }
            let notifications = ""
            let hasItems = false
            let events = []
            row.Notification = row.Notification.sort((a, b) => (a.Notification > b.Notification ? 1 : -1))

            for (let i = 0; i < row.Notification.length; i++) {
              let n = row.Notification[i].Notification.split(":")
              n = n[n.length - 1]
              if (VALIDATOR_EVENTS.includes(n)) {
                events.push(n)
                hasItems = true
                let badgeColor = ""
                let textColor = ""
                switch (n) {
                  case "validator_attestation_missed":
                    badgeColor = "badge-light"
                    break
                  case "validator_proposal_submitted":
                    badgeColor = "badge-light"
                    break
                  case "validator_proposal_missed":
                    badgeColor = "badge-light"
                    break
                  case "validator_got_slashed":
                    badgeColor = "badge-light"
                    break
                  case "validator_synccommittee_soon":
                    badgeColor = "badge-light"
                    break
                  case "validator_is_offline":
                    badgeColor = "badge-light"
                    break
                  case "validator_withdrawal":
                    badgeColor = "badge-light"
                }
                notifications += `<span style="font-size: 12px; font-weight: 500;" class="badge badge-pill ${badgeColor} ${textColor} badge-custom-size mr-1 my-1">${n.replace("validator", "").replaceAll("_", " ")}</span>`
              }
            }
            if (!hasItems) {
              return `<div subscriptions=${events}>Not subscribed to any events</div>`
            }
            return `<div subscriptions=${events} style="white-space: normal;">${notifications}</div>`
          }
          return null
        },
      },
      {
        targets: 2,
        orderable: false,
        responsivePriority: 4,
        data: null,
        defaultContent: `
          <div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox">
            <label class="form-check-label"></label>
          </div>`,
        visible: false,
      },
      {
        targets: 3,
        orderable: false,
        responsivePriority: 4,
        data: null,
        render: function (data, type, row, meta) {
          // let status = data.length > 0 ? 'checked="true"' : ""
          let status = row.Notification.length > 0 ? '<i class="fas fa-check fa-lg"></i>' : ""
          return status
          /* `<div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="" ${status} disabled="true">
            <label class="form-check-label" for=""></label>
          </div>` */
        },
      },
      {
        targets: 4,
        orderable: false,
        responsivePriority: 4,
        data: null,
        defaultContent: `
        	<div class="form-check">
          	<input class="form-check-input checkbox-custom-size" type="checkbox">
            <label class="form-check-label"></label>
          </div>`,
        visible: false,
      },
      {
        targets: 5,
        responsivePriority: 1,
        data: null,
        render: function (data, type, row, meta) {
          let no_time = "N/A"
          if (row.Notification.length === 0) {
            return no_time
          }
          row.Notification.sort((a, b) => {
            return b.Timestamp - a.Timestamp
          })
          if (type === "sort" || type === "type") {
            return row.Notification[0].Timestamp
          }
          if (row.Notification[0].Timestamp === 0) {
            return no_time
          }
          return `<span class="badge badge-pill badge-light badge-custom-size mr-1 mr-sm-2">${
            row.Notification && row.Notification.length
              ? row.Notification[0].Notification.replace(/[a-zA-Z]+:/, "")
                  .replace("validator", "")
                  .replaceAll("_", " ")
              : "N/A"
          }</span><span class="heading-l4 d-block d-sm-inline-block mt-2 mt-sm-0">${getRelativeTime(luxon.DateTime.fromMillis(row.Notification[0].Timestamp * 1000)) || "N/A"}</span>`
        },
      },
      {
        targets: 6,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: '<div class="d-flex align-items-center"><i class="fas fa-pen fa-xs text-muted i-custom mx-2" id="edit-validator-events" title="Manage notifications for the selected validator(s)" style="padding: .5rem; cursor: pointer;" data-toggle= "modal" data-target="#ManageNotificationModal"></i><i class="fas fa-times fa-lg mx-2 i-custom" id="remove-btn" title="Remove validator" style="padding: .5rem; color: var(--red); cursor: pointer;" data-toggle= "modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i></div>',
      },
    ],
    rowId: function (data, type, row, meta) {
      return data ? data.Pubkey : null
    },
  })

  document.addEventListener("click", function (e) {
    let isWatchlistParent = false
    let limit = 0
    let tgt = e.target
    while (tgt) {
      if ($("#watchlist-container").is(tgt) || (tgt.classList && tgt.classList.contains("modal")) || (tgt.classList && tgt.classList.contains("page-link"))) {
        isWatchlistParent = true
        break
      }
      tgt = tgt.parentNode
      limit += 1
      if (limit > 50) {
        break
      }
    }
    if (!isWatchlistParent) {
      $("#validators-notifications").DataTable().rows().deselect()
    }
  })

  $("#selectAll-notifications-btn").on("click", () => {
    $("#validators-notifications").DataTable().rows().select()
  })

  // validatorsTable.on("select.dt", function (e, dt, type, indexes) {
  //   if (indexes && indexes.length) {
  //     document.getElementById("remove-selected-btn").removeAttribute("disabled")
  //     document.getElementById("manage-notifications-btn").removeAttribute("disabled")
  //     document.getElementById("selectAll-notifications-btn").setAttribute("disabled", true)
  //   }
  // })

  // validatorsTable.on("deselect.dt", function (e, dt, type, indexes) {
  //   if (indexes && indexes.length <= 1) {
  //     document.getElementById("remove-selected-btn").setAttribute("disabled", true)
  //     document.getElementById("manage-notifications-btn").setAttribute("disabled", true)
  //     document.getElementById("selectAll-notifications-btn").removeAttribute("disabled")
  //   }
  // })

  // show manage-notifications button and remove-all button only if there is data in the validator table
  // if (DATA.length !== 0) {
  //   $("#manage-notifications-btn").removeAttr("hidden")
  //   // $('#remove-all-btn').removeAttr('hidden')
  //   $("#view-dashboard").removeAttr("hidden")
  //   if ($(window).width() < 620) {
  //     $("#add-validator-btn-text").attr("hidden", true)
  //     $("#add-validator-btn-icon").removeAttr("hidden")
  //   } else {
  //     $("#add-validator-btn-text").removeAttr("hidden")
  //     $("#add-validator-btn-icon").attr("hidden", true)
  //   }
  //   // $(window).resize(function () {
  //   //   if ($(window).width() < 620) {
  //   //     $("#add-validator-btn-text").attr("hidden", true)
  //   //     $("#add-validator-btn-icon").removeAttr("hidden")
  //   //   } else {
  //   //     $("#add-validator-btn-text").removeAttr("hidden")
  //   //     $("#add-validator-btn-icon").attr("hidden", true)
  //   //   }
  //   // })
  // }
}

// function remove_item_from_event_container(pubkey) {
//   for (let item of $('#selected-validators-events-container').find('span')) {
//     if (pubkey === $(item).attr('pk')) {
//       $(item).remove()
//       return
//     }
//   }
// }

$(function () {
  if (document.getElementsByName("CsrfField")[0] !== undefined) {
    csrfToken = document.getElementsByName("CsrfField")[0].value
  }

  create_typeahead(".validator-typeahead")
  // create_typeahead('.monitoring-typeahead')

  loadValidatorsData(DATA)
  loadMonitoringData(MONITORING)

  if (typeof NET !== "undefined" && NET && NET.Events_ts) {
    loadNetworkData(NET.Events_ts)
  }

  // $('#remove-all-btn').on('click', function (e) {
  //   $('#modaltext').text($(this).data('modaltext'))
  //   $('#confirmRemoveModal').removeAttr('rowId')
  //   $('#confirmRemoveModal').attr('tablename', 'validators')
  // })

  // click event to modal remove button
  $("#remove-button").on("click", function (e) {
    const rowId = $("#confirmRemoveModal").attr("rowId")
    const tablename = $("#confirmRemoveModal").attr("tablename")

    // if rowId also check tablename then delete row in corresponding data section
    // if no row id delete directly in corresponding data section
    if (rowId !== undefined) {
      if (tablename === "monitoring") {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
        fetch(`/user/notifications/unsubscribe?event=${encodeURI($("#confirmRemoveModal").attr("event"))}&filter=${encodeURI($("#confirmRemoveModal").attr("filter"))}`, {
          method: "POST",
          headers: { "X-CSRF-Token": csrfToken },
          credentials: "include",
          body: "",
        }).then((res) => {
          if (res.status == 200) {
            $("#confirmRemoveModal").modal("hide")
            window.location.reload()
          } else {
            alert("Error updating validators subscriptions")
            $("#confirmRemoveModal").modal("hide")
            window.location.reload()
          }
          $(this).html("Remove")
        })
      }

      if (tablename === "validators") {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
        fetch(`/validator/${rowId}/remove`, {
          method: "POST",
          headers: { "X-CSRF-Token": csrfToken },
          credentials: "include",
          body: { pubkey: `0x${rowId}` },
        }).then((res) => {
          if (res.status == 200) {
            $("#confirmRemoveModal").modal("hide")
            window.location.reload()
          } else {
            alert("Error removing validator from Watchlist")
            $("#confirmRemoveModal").modal("hide")
            window.location.reload()
          }
          $(this).html("Remove")
        })
      }
    } else {
      if (tablename === "validators") {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
        let pubkeys = []
        for (let item of DATA) {
          pubkeys.push(item.Validator.Pubkey)
        }
        fetch(`/user/notifications-center/removeall`, {
          method: "POST",
          headers: { "X-CSRF-Token": csrfToken },
          credentials: "include",
          body: JSON.stringify(pubkeys),
        }).then((res) => {
          if (res.status == 200) {
            $("#confirmRemoveModal").modal("hide")
            window.location.reload()
          } else {
            alert("Error removing all validators from Watchlist")
            $("#confirmRemoveModal").modal("hide")
            window.location.reload()
          }
          $(this).html("Remove")
        })
      }
    }

    if (tablename === "monitoring") {
      $("#monitoring-notifications").DataTable().clear().destroy()
      loadMonitoringData(DATA)
    }
  })

  $(".range").on("input", function (e) {
    const target_id = $(this).data("target")
    let target = $(target_id)
    target.val($(this).val())
    if ($(this).attr("type") === "range") {
      $(this).css("background-size", $(this).val() + "% 100%")
    } else {
      target.css("background-size", $(this).val() + "% 100%")
    }
  })

  // on modal open after click event to validators table edit button
  $("#ManageNotificationModal").on("show.bs.modal", function (e) {
    $("#ManageNotificationModal-form-content").show()
    $("#ManageNotificationModal button[type='submit']").prop("disabled", false)
    $("#ManageNotificationModal")
      .find('input[type="checkbox"]')
      .each(function () {
        $(this).prop("checked", false)
      })

    let rowID = $(this).attr("rowid")
    // console.log('row id: ', rowID)
    if (rowID) {
      // let inputs = document.querySelectorAll('#ManageNotificationModal input[type="checkbox"]')
      // let activeCount = 0
      let subscriptions = $(this).attr("subscriptions")
      document.getElementById("ManageNotificationModal-validators").value = rowID
      if (subscriptions) {
        subscriptions = subscriptions.split(",")
        for (let i = 0; i < subscriptions.length; i++) {
          let sub = subscriptions[i]
          $("#watchlist-selected-" + sub).prop("checked", true)
          // activeCount += 1
        }
      }

      // if (activeCount == inputs.length - 1) {
      //   $("#ManageNotificationModal-all").prop("checked", true)
      // }
    } else {
      let rowsSelected = $("#validators-notifications").DataTable().rows(".selected").data()
      if (rowsSelected && rowsSelected.length) {
        let valis = rowsSelected.map((row) => row.Pubkey).join(",")
        document.getElementById("ManageNotificationModal-validators").value = valis
        if (rowsSelected.length === 1) {
          document.getElementById("ManageNotificationModal-subtitle").innerHTML = `You've selected one validator from your watchlist`
        } else {
          document.getElementById("ManageNotificationModal-subtitle").innerHTML = `You've selected ${rowsSelected.length} validators from your watchlist`
        }
      } else {
        document.getElementById("ManageNotificationModal-subtitle").innerHTML = `No validators selected.`
        $("#ManageNotificationModal-form-content").hide()
        $("#ManageNotificationModal button[type='submit']").prop("disabled", true)
      }
    }
  })

  $("#ManageNotificationModal").on("hide.bs.modal", function (e) {
    $(this).removeAttr("rowid")
    $(this).removeAttr("subscriptions")
  })

  // select/deselect notification checkboxes for all events
  for (let event of $("#validator_all_events :input")) {
    $(event).on("click", function () {
      if ($(this).prop("checked")) {
        for (let item of VALIDATOR_EVENTS) {
          $(`#${item} input#${$(event).attr("id")}`).prop("checked", true)
        }
      } else {
        for (let item of VALIDATOR_EVENTS) {
          $(`#${item} input#${$(event).attr("id")}`).prop("checked", false)
        }
      }
    })
  }

  for (let event of $("#manage_all_events :input")) {
    $(event).on("click", function () {
      if ($(this).prop("checked")) {
        for (let item of VALIDATOR_EVENTS) {
          $(`#manage_${item} input#${$(event).attr("id")}`).prop("checked", true)
        }
      } else {
        for (let item of VALIDATOR_EVENTS) {
          $(`#manage_${item} input#${$(event).attr("id")}`).prop("checked", false)
        }
      }
    })
  }

  $("#add-monitoring-event-modal-button").on("click", function () {
    if (!MACHINES.length) {
      $("#add-validator-search-container > div:not(:first-child)").css("opacity", ".3")
      $("#add-monitoring-event").attr("disabled", true)
      $("#add-monitoring-validator-select").replaceWith(`
      <span>
        No machine found. Learn more about monitoring your validator and beacon node
        <a href="https://kb.beaconcha.in/beaconcha.in-explorer/mobile-app-less-than-greater-than-beacon-node">here</a>.
      </span>`)
    } else {
      $("#add-monitoring-validator-select").html("")
      for (let i = 0; i < MACHINES.length; i++) {
        if (MACHINES[i]) {
          $("#add-monitoring-validator-select").append(`<option value="${MACHINES[i]}">${MACHINES[i]}</option>`)
        }
      }
    }
  })

  $("#add-monitoring-event").on("click", function () {
    let events = []
    let filter = $("#add-monitoring-validator-select option:selected").val()
    for (let item of $("input.monitoring")) {
      if ($(item).prop("checked")) {
        let e = $(item).attr("event")
        let t = 0
        switch (e) {
          case "monitoring_cpu_load":
            t = parseFloat($("#cpu-input-range-val").val()) / 100
            break
          case "monitoring_hdd_almostfull":
            t = parseFloat($("#hdd-input-range-val").val()) / 100
            break
          default:
            t = 0
        }

        events.push({
          event_name: e,
          event_filter: filter,
          event_threshold: t,
        })
      }
    }
    fetch(`/user/notifications/bundled/subscribe`, {
      method: "POST",
      headers: { "X-CSRF-Token": csrfToken },
      credentials: "include",
      body: JSON.stringify(events),
    }).then((res) => {
      if (res.status == 200) {
        $("#ManageNotificationModal").modal("hide")
        window.location.reload()
      } else {
        alert("Error updating validators subscriptions")
        $("#ManageNotificationModal").modal("hide")
        window.location.reload()
      }
    })
  })

  $("#add-network-subscription").on("click", function () {
    if ($("#finalityIssues").prop("checked")) {
      $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Saving...</span></div>')
      fetch(`/user/notifications/subscribe?event=${$("#finalityIssues").attr("event")}&filter=0x${$("#finalityIssues").attr("event")}`, {
        method: "POST",
        headers: { "X-CSRF-Token": csrfToken },
        credentials: "include",
        body: "",
      }).then((res) => {
        if (res.status == 200) {
          $("#NetworkEventModal").modal("hide")
          window.location.reload()
        } else {
          alert("Error updating network subscriptions")
          $("#NetworkEventModal").modal("hide")
          window.location.reload()
        }
        $(this).html("Save")
      })
    } else {
      $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
      fetch(`/user/notifications/unsubscribe?event=${$("#finalityIssues").attr("event")}&filter=0x${$("#finalityIssues").attr("event")}`, {
        method: "POST",
        headers: { "X-CSRF-Token": csrfToken },
        credentials: "include",
        body: "",
      }).then((res) => {
        if (res.status == 200) {
          $("#NetworkEventModal").modal("hide")
          window.location.reload()
        } else {
          alert("Error updating network subscriptions")
          $("#NetworkEventModal").modal("hide")
          window.location.reload()
        }
        $(this).html("Save")
      })
    }
  })
})

// Sets a hidden input with the selected validators
$("#RemoveSelectedValidatorsModal").on("show.bs.modal", function (event) {
  $('#RemoveSelectedValidatorsModal button[type="submit"]').prop("disabled", false)
  let rowsSelected = $("#validators-notifications").DataTable().rows(".selected").data()
  if (rowsSelected && rowsSelected.length) {
    let valis = rowsSelected.map((row) => row.Pubkey).join(",")
    document.getElementById("RemoveSelectedValidatorsModal-input").value = valis
    if (rowsSelected.length === 1) {
      document.getElementById("RemoveSelectedValidatorsModal-modaltext").innerHTML = `You've selected one validator to remove from your watchlist`
    } else {
      document.getElementById("RemoveSelectedValidatorsModal-modaltext").innerHTML = `You've selected ${rowsSelected.length} validators to remove from your watchlist`
    }
  } else {
    document.getElementById("RemoveSelectedValidatorsModal-modaltext").innerHTML = `No validators selected.`
    $('#RemoveSelectedValidatorsModal button[type="submit"]').prop("disabled", true)
  }
})
