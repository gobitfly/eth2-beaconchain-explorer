var csrfToken = ""

const VALIDATOR_EVENTS = ['validator_attestation_missed', 'validator_proposal_missed', 'validator_proposal_submitted', 'validator_got_slashed']

const MONITORING_EVENTS = ['monitoring_machine_offline', 'monitoring_hdd_almostfull', 'monitoring_cpu_load']
const VALLIMIT = 100;
var indices = [];

function create_typeahead(input_container) {
  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.index
    },
    remote: {
      url: '/search/indexed_validators/%QUERY',
      wildcard: '%QUERY'
    }
  });
  var bhName = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.name
    },
    remote: {
      url: '/search/indexed_validators_by_name/%QUERY',
      wildcard: '%QUERY'
    }
  })
  var bhEth1Addresses = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.eth1_address
    },
    remote: {
      url: '/search/indexed_validators_by_eth1_addresses/%QUERY',
      wildcard: '%QUERY'
    }
  })
  $(input_container).typeahead(
    {
      minLength: 1,
      highlight: true,
      hint: false,
      autoselect: false
    },
    {
      limit: 5,
      name: 'validators',
      source: bhValidators,
      display: 'index',
      templates: {
        header: '<h5 class="font-weight-bold ml-3">Validators</h5>',
        suggestion: function(data) {
          return `<div class="font-weight-normal text-truncate high-contrast">${data.index}</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'addresses',
      source: bhEth1Addresses,
      display: 'address',
      templates: {
        header: '<h3>Validators by ETH1 Addresses</h3>',
        suggestion: function(data) {
          var len = data.validator_indices.length > VALLIMIT ? VALLIMIT+'+' : data.validator_indices.length 
          return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.eth1_address}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        }
      }
    },
    {
      limit: 5,
      name: 'name',
      source: bhName,
      display: 'name',
      templates: {
        header: '<h5 class="font-weight-bold ml-3">Validators by Name</h5>',
        suggestion: function(data) {
          var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + '+' : data.validator_indices.length
          return `<div class="font-weight-normal high-contrast" style="display: flex;"><div class="text-truncate" style="flex: 1 1 auto;">${data.name}</div><div style="max-width: fit-content; white-space: nowrap;">${len}</div></div>`
        }
      },
    });
  $(input_container).on('focus', function(e) {
    if (e.target.value !== "") {
      $(this).trigger($.Event('keydown', { keyCode: 40 }))
    }
  })
  // $(input_container).on('blur', function() {
  //   $(input_container).typeahead('val', '')
  // })
  $(input_container).on('input', function() {
    $('.tt-suggestion').first().addClass('tt-cursor')
  })
  $(input_container).on('typeahead:select', function(e, sug) {
    console.log(sug)
    $(input_container).typeahead('val', '')
    if (sug.eth1_address) {
      indices = sug.validator_indices;
      $(input_container).typeahead('val', sug.eth1_address)
      console.log(1)
    } else if(sug.name){
      indices = sug.validator_indices;
      $(input_container).typeahead('val', $(sug.name).text())
      console.log(2)
    } else{
      indices = [parseInt(sug.index)]
      $(input_container).typeahead('val', sug.index)
      console.log(3)
    }
    // $(input_container).attr('pk', sug.pubkey)
  })
}

function loadMonitoringData(data) {
  let mdata = []
  let id = 0
  for (let item of data) {
    for (let n of item.Notifications) {
      let nn = n.Notification.split(':')
      nn = nn[nn.length-1]
      let ns = nn.split('_')
      if (ns[0] === 'monitoring') {
        if (ns[1] === 'machine') {
          ns[1] = ns[2]
        }
        mdata.push({
          id: id,
          notification: ns[1],
          threshold: [n.Threshold, item],
          machine: item.Validator.Index,
          mostRecent: n.Timestamp,
          event: { pk: item.Validator.Pubkey, e: nn }
        })
        id += 1
      }
    }
  }

  if (mdata.length !== 0) {
    if ($('#monitoring-section-with-data').children().length === 0) {
      $('#monitoring-section-with-data').append(
        `<table class="table table-borderless table-hover" id="monitoring-notifications">
          <thead class="custom-table-head">
            <tr>
              <th scope="col" class="h6 border-bottom-0">Notification</th>
              <th scope="col" class="h6 border-bottom-0">Threshold</th>
              <th scope="col" class="h6 border-bottom-0">Machine</th>
              <th scope="col" class="h6 border-bottom-0">Most Recent</th>
              <th scope="col" class="h6 border-bottom-0"></th>
            </tr>
          </thead>
          <tbody></tbody>
        </table>`
      )
    }
  } else {
    $('#monitoring-section-empty').removeAttr('hidden')
  }

  let monitoringTable = $('#monitoring-notifications')

  monitoringTable.DataTable({
    language: {
      info: '_TOTAL_ entries',
      infoEmpty: 'No entries match',
      infoFiltered: '(from _MAX_ entries)',
      processing: 'Loading. Please wait...',
      search: '',
      searchPlaceholder: 'Search...',
      zeroRecords: 'No entries match'
    },
    processing: true,
    responsive: true,
    scroller: true,
    scrollY: 380,
    paging: true,
    data: mdata,
    rowId: 'id',
    initComplete: function(settings, json) {
      $('body').find('.dataTables_scrollBody').addClass('scrollbar')

      // click event to monitoring table edit button
      $('#monitoring-notifications #edit-monitoring-events').on('click', function(e) {
        $('#add-monitoring-validator-select').html("")
        for (let item of $('input.monitoring')) {
          $(item).prop('checked', false)
        }

        let ev = $(this).attr('event').split(',')
        for (let i of ev) {
          if (i.length > 0) {
            let t = i.split(':')
            for (let item of $('input.monitoring')) {
              let e = $(item).attr('event')
              if (e === t[0]) {
                $(item).prop('checked', true)
                let p = parseInt(parseFloat(t[1]) * 100)
                if (e.includes('_cpu_')) {
                  $('#cpu-input-range-val, #cpu-input-range').val(p)
                  $('#cpu-input-range').attr('style', `background-size: ${p}% 100%`)
                } else if (e.includes('_hdd_')) {
                  $('#hdd-input-range-val, #hdd-input-range').val(p)
                  $('#hdd-input-range').attr('style', `background-size: ${p}% 100%`)
                }
              }
            }
          }
        }
        $('#add-monitoring-validator-select').append(`<option value="${$(this).attr('pk')}">${$(this).attr("ind")}</option>`)
      });

      // click event to table remove button
      $('#monitoring-notifications #remove-btn').on('click', function(e) {
        $('#modaltext').text($(this).data('modaltext'))

        // set the row id 
        let rowId = $(this).parent().parent().attr('id')
        if (rowId === undefined) {
          rowId = 0
        }
        $('#confirmRemoveModal').attr('rowId', rowId)
        $('#confirmRemoveModal').attr('tablename', 'monitoring')
        $('#confirmRemoveModal').attr('pk', $(this).attr('pk'))
        $('#confirmRemoveModal').attr('event', $(this).attr('event'))
      });
    },
    columnDefs: [
      {
        targets: '_all',
        createdCell: function(td, cellData, rowData, row, col) {
          $(td).css('padding-top', '20px')
          $(td).css('padding-bottom', '20px')
        }
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: 'notification',
        render: function(data, type, row, meta) {
          return `<span class="badge badge-pill badge-light badge-custom-size font-weight-normal">${data.charAt(0).toUpperCase() + data.slice(1)}</span>`
        }
      },
      {
        targets: 1,
        responsivePriority: 3,
        data: 'threshold',
        render: function(data, type, row, meta) {
          if (type === 'display') {
            let e = ""
            for (let i of data[1].Notifications) {
              let nn = i.Notification.split(':')
              nn = nn[nn.length-1]
              let ns = nn.split('_')
              if (ns[0] === 'monitoring') {
                e += `${nn}:${i.Threshold},`
              }
            }

            // for machine offline event, there is no threshold value; we show N/A and hide the edit button
            // replaced (data[0] * 100).toFixed(2) with Math.trunc(data[0] * 100)
            return `
              <input type="text" class="form-control input-sm threshold_editable" title="Numbers in 1-100 range (including)" style="width: 60px; height: 30px;" hidden />
              <span class="threshold_non_editable">
                <span class="threshold_non_editable_text">${data[0] === "0" ? "N/A" : Math.trunc(data[0] * 100) + "%"}</span>
                <i 
                  class="fas fa-pen fa-xs text-muted i-custom ${data[0] === '0' ? 'd-none' : ''}" 
                  id="edit-monitoring-events" 
                  title="Click to edit" 
                  style="padding: .5rem; cursor: pointer;" 
                  data-toggle= "modal" 
                  data-target="#addMonitoringEventModal" 
                  pk="${data[1].Validator.Pubkey}" 
                  ind="${data[1].Validator.Index}" 
                  event="${e}"
                ></i>
              </span>`
          }
          return data[0]
        }
      },
      {
        targets: 2,
        responsivePriority: 2,
        data: 'machine',
        render: function(data, type, row, meta) {
          return `<span class="font-weight-bold"><i class="fas fa-male mr-2"></i><a style="padding: .25rem;" href="/validator/${data}">${data}</a></span>`
        }
      },
      {
        targets: 3,
        responsivePriority: 1,
        data: 'mostRecent',
        render: function(data, type, row, meta) {
          // for sorting and type checking use the original data (unformatted)
          if (type === 'sort' || type === 'type') {
            return data
          }
          if (parseInt(data) === 0) {
            return 'N/A'
          }
          return `<span class="heading-l4">${luxon.DateTime.fromMillis(data * 1000).toRelative({ style: "long" })}</span>`
        }
      },
      {
        targets: 4,
        orderable: false,
        responsivePriority: 3,
        data: 'event',
        render: function(data, type, row, meta) {
          return `<i class="fas fa-times fa-lg i-custom" pk="${data.pk}" event="${data.e}" id="remove-btn" title="Remove notification" style="padding: .5rem; color: var(--new-red); cursor: pointer;" data-toggle="modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>`
        }
      }
    ]
  })
}

function loadNetworkData(data) {
  let networkTable = $('#network-notifications')

  networkTable.DataTable({
    language: {
      info: '_TOTAL_ entries',
      infoEmpty: 'No entries match',
      infoFiltered: '(from _MAX_ entries)',
      processing: 'Loading. Please wait...',
      search: '',
      searchPlaceholder: 'Search...',
      zeroRecords: 'No entries match'
    },
    processing: true,
    responsive: true,
    scroller: true,
    scrollY: 380,
    paging: true,
    data: data,
    initComplete: function(settings, json) {
      $('body').find('.dataTables_scrollBody').addClass('scrollbar')
    },
    columnDefs: [
      {
        targets: '_all',
        createdCell: function(td, cellData, rowData, row, col) {
          $(td).css('padding-top', '20px')
          $(td).css('padding-bottom', '20px')
        }
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: 'Notification',
        render: function(data, type, row, meta) {
          return `<span class="badge badge-pill badge-light badge-custom-size font-weight-normal">${data}</span>`
        }
      },
      {
        targets: 1,
        responsivePriority: 2,
        data: 'Network'
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
          visible: false
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
        visible: false
          
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
          visible: false
      },
      {
        targets: 5,
        responsivePriority: 1,
        data: 'Timestamp',
        render: function(data, type, row, meta) {
          if (type === 'sort' || type === 'type') {
            return data
          }
          return `<span class="heading-l4">${luxon.DateTime.fromMillis(data).toRelative({ style: "long" })}</span>`
        }
      }
    ],
    order: [[5, 'desc']]
  })
}

function loadValidatorsData(data) {
  let validatorsTable = $('#validators-notifications')

  validatorsTable.DataTable({
    language: {
      info: '_TOTAL_ entries',
      infoEmpty: 'No entries match',
      infoFiltered: '(from _MAX_ entries)',
      processing: 'Loading. Please wait...',
      search: '',
      searchPlaceholder: 'Search...',
      zeroRecords: 'No entries match'
    },
    processing: true,
    responsive: true,
    paging: true,
    pagingType: 'first_last_numbers',
    select: {
      items: 'row',
      toggleable: true
    },
    fixedHeader: true,
    data: data,
    drawCallback: function(settings) {
      $('[data-toggle="tooltip"]').tooltip()
    },
    initComplete: function(settings, json) {
      $('body').find('.dataTables_scrollBody').addClass('scrollbar')

      // click event to validators table edit button
      $('#validators-notifications #edit-validator-events').on('click', function(e) {
        $('#manageNotificationsModal').attr('rowId', $(this).parent().parent().attr('id'))
      });

      // click event to remove button
      $('#validators-notifications #remove-btn').on('click', function(e) {
        const rowId = $(this).parent().parent().attr('id')
        $('#modaltext').text($(this).data('modaltext'))

        // set the row id 
        $('#confirmRemoveModal').attr('rowId', rowId)
        $('#confirmRemoveModal').attr('tablename', 'validators')
      });
    },
    columnDefs: [
      {
        targets: '_all',
        createdCell: function(td, cellData, rowData, row, col) {
          $(td).css('padding-top', '20px')
          $(td).css('padding-bottom', '20px')
        }
      },
      {
        targets: 0,
        responsivePriority: 2,
        data: 'Validator',
        render: function(data, type, row, meta) {
          if (type === 'sort' || type === 'type') {
            return data.Index
          }
          let datahref = `/validator/${data.Index || data.Pubkey}`
          return `<i class="fas fa-male mr-2"></i><a class="font-weight-bold" href=${datahref}>` + data.Index + `<span class="heading-l4 d-none d-sm-block mt-2">0x` + data.Pubkey.substring(0, 6) + ` ...</span></a><i class="fa fa-copy text-muted d-none d-sm-inline p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="0x${data.Pubkey}"></i>`
        }
      },
      {
        targets: 1,
        responsivePriority: 2,
        data: 'Notifications',
        render: function(data, type, row, meta) {
          if (type === 'display') {
            let notifications = ""
            let hasItems = false
            for (let notification of data) {
              let n = notification.Notification.split(':')
              n = n[n.length-1]
              if (VALIDATOR_EVENTS.includes(n)) { 
                hasItems = true
                let badgeColor = ""
                switch (n) {
                  case 'validator_attestation_missed':
                    badgeColor = 'badge-warning'
                    break
                  case 'validator_proposal_submitted':
                    badgeColor = 'badge-light'
                    break
                  case 'validator_proposal_missed':
                    badgeColor = 'badge-warning'
                    break
                  case 'validator_got_slashed':
                    badgeColor = 'badge-light'
                    break
                }
                notifications += `<span class="badge badge-pill ${badgeColor} badge-custom-size mr-1 my-1 font-weight-normal">${n.replace('validator', "").replaceAll('_', " ")}</span>`
              }
            }
            if (!hasItems) {
              return '<span>Not subscribed to any events</span><i class="d-block fas fa-pen fa-xs text-muted i-custom" id="edit-validator-events" title="Manage notifications for the selected validator(s)" style="width: 1.5rem; padding: .5rem; cursor: pointer;" data-toggle= "modal" data-target="#manageNotificationsModal"></i>'
            }
            return `<div style="white-space: normal; max-width: 400px;">${notifications}</div> <i class="fas fa-pen fa-xs text-muted i-custom" id="edit-validator-events" title="Manage notifications for the selected validator(s)" style="padding: .5rem; cursor: pointer;" data-toggle= "modal" data-target="#manageNotificationsModal"></i>`
          }
          return null
        }
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
        visible: false
      },
      {
        targets: 3,
        orderable: false,
        responsivePriority: 4,
        data: 'Notifications',
        render: function(data, type, row, meta) {
          // let status = data.length > 0 ? 'checked="true"' : ""
          let status = data.length > 0 ? '<i class="fas fa-check fa-lg"></i>' : ""
          return status
          /* `<div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="" ${status} disabled="true">
          	<label class="form-check-label" for=""></label>
          </div>` */
        }
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
        visible: false
      },
      {
        targets: 5,
        responsivePriority: 1,
        data: 'Notifications',
        render: function(data, type, row, meta) {
          let no_time = 'N/A'
          if (data.length === 0) {
            return no_time
          }
          data.sort((a, b) => {
            return b.Timestamp - a.Timestamp
          });
          if (type === 'sort' || type === 'type') {
            return data[0].Timestamp
          }
          if (data[0].Timestamp === 0) {
            return no_time
          }
          return `<span class="badge badge-pill badge-light badge-custom-size mr-1 mr-sm-2 font-weight-normal">${data[0].Notification.replace('validator', "").replaceAll('_', " ")}</span><span class="heading-l4 d-block d-sm-inline-block mt-2 mt-sm-0">${luxon.DateTime.fromMillis(data[0].Timestamp * 1000).toRelative({ style: "long" })}</span>`
        }
      },
      {
        targets: 6,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: '<i class="fas fa-times fa-lg i-custom" id="remove-btn" title="Remove validator" style="padding: .5rem; color: var(--new-red); cursor: pointer;" data-toggle= "modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>'
      }
    ],
    rowId: function(data, type, row, meta) {
      return data.Validator.Pubkey
    }
  })

  // show manage-notifications button and remove-all button only if there is data in the validator table
  if (DATA.length !== 0) {
    $('#manage-notifications-btn').removeAttr('hidden')
    $('#remove-all-btn').removeAttr('hidden')
    if ($(window).width() < 620) {
      $('#add-validator-btn-text').attr('hidden', true)
      $('#add-validator-btn-icon').removeAttr('hidden')
    } else {
      $('#add-validator-btn-text').removeAttr('hidden')
      $('#add-validator-btn-icon').attr('hidden', true)
    }
    $(window).resize(function() {
      if ($(window).width() < 620) {
        $('#add-validator-btn-text').attr('hidden', true)
        $('#add-validator-btn-icon').removeAttr('hidden')
      } else {
        $('#add-validator-btn-text').removeAttr('hidden')
        $('#add-validator-btn-icon').attr('hidden', true)
      }
    })
  }
}

function remove_item_from_event_container(pubkey) {
  for (let item of $('#selected-validators-events-container').find('span')) {
    if (pubkey === $(item).attr('pk')) {
      $(item).remove()
      return
    }
  }
}

$(document).ready(function() {
  if (document.getElementsByName('CsrfField')[0] !== undefined) {
    csrfToken = document.getElementsByName('CsrfField')[0].value
  }

  create_typeahead('.validator-typeahead')
  // create_typeahead('.monitoring-typeahead')

  loadValidatorsData(DATA)
  loadMonitoringData(DATA)
  loadNetworkData(NET.Events_ts)

  $(document).on('click', function(e) {
    // remove selected class from rows on click outside
    if (!$('#validators-notifications').is(e.target) && $('#validators-notifications').has(e.target).length === 0 && !$('#manage-notifications-btn').is(e.target) && $('#manage-notifications-btn').has(e.target).length === 0) {
      $('#validators-notifications .selected').removeClass('selected')
    }
  })

  $('#remove-all-btn').on('click', function(e) {
    $('#modaltext').text($(this).data('modaltext'))
    $('#confirmRemoveModal').removeAttr('rowId')
    $('#confirmRemoveModal').attr('tablename', 'validators')
  })

  // click event to modal remove button
  $('#remove-button').on('click', function(e) {
    const rowId = $('#confirmRemoveModal').attr('rowId')
    const tablename = $('#confirmRemoveModal').attr('tablename')

    // if rowId also check tablename then delete row in corresponding data section
    // if no row id delete directly in correponding data section
    if (rowId !== undefined) {
      if (tablename === 'monitoring') {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
        fetch(`/user/notifications/unsubscribe?event=${$('#confirmRemoveModal').attr('event')}&filter=0x${$('#confirmRemoveModal').attr('pk')}`, {
          method: 'POST',
          headers: { "X-CSRF-Token": csrfToken },
          credentials: 'include',
          body: ""
        }).then(res => {
          if (res.status == 200) {
            $('#confirmRemoveModal').modal('hide')
            window.location.reload()
          } else {
            alert('Error updating validators subscriptions')
            $('#confirmRemoveModal').modal('hide')
            window.location.reload()
          }
          $(this).html('Remove')
        })
      }

      if (tablename === 'validators') {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
        fetch(`/validator/${rowId}/remove`, {
          method: 'POST',
          headers: { "X-CSRF-Token": csrfToken },
          credentials: 'include',
          body: { pubkey: `0x${rowId}` }
        }).then(res => {
          if (res.status == 200) {
            $('#confirmRemoveModal').modal('hide')
            window.location.reload()
          } else {
            alert('Error removing validator from Watchlist')
            $('#confirmRemoveModal').modal('hide')
            window.location.reload()
          }
          $(this).html('Remove')
        })
      }
    } else {
      if (tablename === 'validators') {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
        let pubkeys = []
        for (let item of DATA) {
          pubkeys.push(item.Validator.Pubkey)
        }
        fetch(`/user/notifications-center/removeall`, {
          method: 'POST',
          headers: { "X-CSRF-Token": csrfToken },
          credentials: 'include',
          body: JSON.stringify(pubkeys)
        }).then(res => {
          if (res.status == 200) {
            $('#confirmRemoveModal').modal('hide')
            window.location.reload()
          } else {
            alert('Error removing all validators from Watchlist')
            $('#confirmRemoveModal').modal('hide')
            window.location.reload()
          }
          $(this).html('Remove')
        })
      }
    }

    if (tablename === 'monitoring') {
      $('#monitoring-notifications').DataTable().clear().destroy()
      loadMonitoringData(DATA)
    }
  })

  $('.range').on('input', function(e) {
    const target_id = $(this).data('target')
    let target = $(target_id)
    target.val($(this).val())
    if ($(this).attr('type') === 'range') {
      $(this).css('background-size', $(this).val() + '% 100%')
    } else {
      target.css('background-size', $(this).val() + '% 100%')
    }
  })

  // on modal open after click event to validators table edit button
  $('#manageNotificationsModal').on('show.bs.modal', function(e) {
    // get the selected row (single row selected)
    let rowData = $('#validators-notifications').DataTable().row($('#' + $(this).attr('rowId'))).data()
    if (rowData && rowData.Validator) {
      $('#selected-validators-events-container').append(
        `<span id="validator-event-badge" class="d-inline-block badge badge-pill badge-light badge-custom-size mr-2 mb-2 font-weight-normal" pk=${rowData.Validator.Pubkey}>
        		Validator ${rowData.Validator.Index}
          	<i class="fas fa-times ml-2" style="cursor: pointer;" title="Remove from selected validators" onclick="remove_item_from_event_container('${rowData.Validator.Pubkey}')"></i>
        </span>`
      )

      for (let event of $('#manage_all_events :input')) {
        for (let item of rowData.Notifications) {
          let n = item.Notification.split(':')
          n = n[n.length-1]
          $(`#manage_${n} input#${$(event).attr('id')}`).prop('checked', true)
        }
      }
    } else {
      // get the selected rows (mutiple rows selected)
      const rowsSelected = $('#validators-notifications').DataTable().rows('.selected').data()
      if (rowsSelected && rowsSelected.length) {
        for (let i = 0; i < rowsSelected.length; i++) {
          $('#selected-validators-events-container').append(
            `<span id="validator-event-badge" class="d-inline-block badge badge-pill badge-light badge-custom-size mr-2 mb-2 font-weight-normal" pk=${rowsSelected[i].Validator.Pubkey}>
              Validator ${rowsSelected[i].Validator.Index}
              <i class="fas fa-times ml-2" style="cursor: pointer;" onclick="remove_item_from_event_container('${rowsSelected[i].Validator.Pubkey}')"></i>
            </span>`
          )
        }
      } else {
        $('#selected-validators-events-container').prev('span').text('ℹ️ No validators selected')
        $('#selected-validators-events-container').html('<span>Select validators from the table. Hold down <kbd>Ctrl</kbd> to select multiple rows.</span>')
        $('#update-subs-button').attr('disabled', '')
      }
    }1234
  })

  // on modal close
  $('#manageNotificationsModal').on('hide.bs.modal', function(e) {
    $(this).removeAttr('rowId')
    $('#selected-validators-events-container #validator-event-badge').remove()
    for (let event of $('#manage_all_events :input')) {
      for (let item of VALIDATOR_EVENTS) {
        $(`#manage_${item} input#${$(event).attr('id')}`).prop('checked', false)
      }
      $(event).prop('checked', false)
    }

    $('[id^=all_events]').attr('checked', false)

    // remove selected class from rows when modal closed
    $('#validators-notifications .selected').removeClass('selected')
  });

  function get_validator_manage_sub_events() {
    let events = []
    for (let item of VALIDATOR_EVENTS) {
      events.push({
        event: item,
        email: $(`#manage_${item} :input#email`).prop('checked'),
        push: $(`#manage_${item} :input#push`).prop('checked'),
        web: $(`#manage_${item} :input#web`).prop('checked')
      })
    }
    return events
  }

  $('#update-subs-button').on('click', function() {
    let bc = $(this).html()
    $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Saving...</span></div>')
    let pubkeys = []
    for (let item of $('#selected-validators-events-container').find('span')) {
      pubkeys.push($(item).attr('pk'))
    }
    if (pubkeys.length === 0) {
      $(this).html(bc)
      return
    }
    let events = get_validator_manage_sub_events();
    fetch(`/user/notifications-center/updatesubs`, {
      method: 'POST',
      headers: { "X-CSRF-Token": csrfToken },
      credentials: 'include',
      body: JSON.stringify({ pubkeys: pubkeys, events: events })
    }).then(res => {
      if (res.status == 200) {
        $('#manageNotificationsModal').modal('hide')
        window.location.reload()
      } else {
        alert('Error updating validators subscriptions')
        $('#manageNotificationsModal').modal('hide')
        window.location.reload()
      }
      $(this).html(bc)
    })
  })

  function get_validator_sub_events() {
    let events = []
    for (let item of VALIDATOR_EVENTS) {
      events.push({
        event: item,
        email: $(`#${item} :input#email`).prop('checked'),
        push: $(`#${item} :input#push`).prop('checked'),
        web: $(`#${item} :input#web`).prop('checked')
      });
    }
    return events
  }

  $('#add-validator-button').on('click', function() {
    try {
      // let index = parseInt($('#add-validator-input').val())
      let events = get_validator_sub_events()
      if (indices.length>0) {
        let bc = $(this).html()
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Saving...</span></div>')
        fetch(`/user/notifications-center/validatorsub`, {
          method: 'POST',
          headers: { "X-CSRF-Token": csrfToken },
          credentials: 'include',
          body: JSON.stringify({ indices: indices, events: events })
        }).then(res => {
          if (res.status == 200) {
            $('#addValidatorModal').modal('hide')
            window.location.reload()
          } else {
            alert('Error adding validators to Watchlist')
            $('#addValidatorModal').modal('hide')
            window.location.reload()
          }
          $(this).html(bc)
        });
      }
    } catch {
      alert('Invalid Validator Index') 
    }
  })

  // select/deselect notification checkboxes for all events
  for (let event of $('#validator_all_events :input')) {
    $(event).on('click', function() {
      if ($(this).prop('checked')) {
        for (let item of VALIDATOR_EVENTS) {
          $(`#${item} input#${$(event).attr('id')}`).prop('checked', true)
        }
      } else {
        for (let item of VALIDATOR_EVENTS) {
          $(`#${item} input#${$(event).attr('id')}`).prop('checked', false)
        }
      }
    })
  }

  for (let event of $('#manage_all_events :input')) {
    $(event).on('click', function() {
      if ($(this).prop('checked')) {
        for (let item of VALIDATOR_EVENTS) {
          $(`#manage_${item} input#${$(event).attr('id')}`).prop('checked', true)
        }
      } else {
        for (let item of VALIDATOR_EVENTS) {
          $(`#manage_${item} input#${$(event).attr('id')}`).prop('checked', false)
        }
      }
    })
  }

  $('#add-monitoring-event-modal-button').on('click', function() {
    $('#add-monitoring-validator-select').html('')
    for (let item of DATA.sort((a, b) => { return a.Validator.Index - b.Validator.Index })) {
      $('#add-monitoring-validator-select').append(`<option value="${item.Validator.Pubkey}">${item.Validator.Index}</option>`)
    }
  })

  $('#add-monitoring-event').on('click', function() {
    let pubkey = $('#add-monitoring-validator-select option:selected').val()
    events = []
    for (let item of $('input.monitoring')) {
      if ($(item).prop('checked')) {
        let e = $(item).attr('event')
        let t = 0
        switch (e) {
          case 'monitoring_cpu_load':
            t = parseFloat($('#cpu-input-range-val').val()) / 100
            break
          case 'monitoring_hdd_almostfull':
            t = parseFloat($('#hdd-input-range-val').val()) / 100
            break
          default:
            t = 0
        }
        events.push({
          event: e,
          email: true,
          threshold: t
        })
      }
    }
    fetch(`/user/notifications-center/monitoring/updatesubs`, {
      method: 'POST',
      headers: { "X-CSRF-Token": csrfToken },
      credentials: 'include',
      body: JSON.stringify({ pubkeys: [pubkey], events: events })
    }).then(res => {
      if (res.status == 200) {
        $('#manageNotificationsModal').modal('hide')
        window.location.reload()
      } else {
        alert('Error updating validators subscriptions')
        $('#manageNotificationsModal').modal('hide')
        window.location.reload()
      }
    })
  })

  $('#add-network-subscription').on('click', function() {
    if ($('#finalityIssues').prop('checked')) {
      $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Saving...</span></div>')
      fetch(`/user/notifications/subscribe?event=${$('#finalityIssues').attr('event')}&filter=0x${$('#finalityIssues').attr('event')}`, {
        method: 'POST',
        headers: { "X-CSRF-Token": csrfToken },
        credentials: 'include',
        body: ""
      }).then(res => {
        if (res.status == 200) {
          $('#addNetworkEventModal').modal('hide')
          window.location.reload()
        } else {
          alert('Error updating network subscriptions')
          $('#addNetworkEventModal').modal('hide')
          window.location.reload()
        }
        $(this).html('Save')
      })
    } else {
      $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing...</span></div>')
      fetch(`/user/notifications/unsubscribe?event=${$('#finalityIssues').attr('event')}&filter=0x${$('#finalityIssues').attr('event')}`, {
        method: 'POST',
        headers: { "X-CSRF-Token": csrfToken },
        credentials: 'include',
        body: ""
      }).then(res => {
        if (res.status == 200) {
          $('#addNetworkEventModal').modal('hide')
          window.location.reload()
        } else {
          alert('Error updating network subscriptions')
          $('#addNetworkEventModal').modal('hide')
          window.location.reload()
        }
        $(this).html('Save')
      })
    }
  })
})
