const data = {
  monitoring: [
    {
      id: 1,
      notification: "CPU",
      threshold: 0.8,
      machine: "machine1",
      mostRecent: 1626078050
    },
    {
      id: 2,
      notification: "HDD",
      threshold: 0.8,
      machine: "machine1",
      mostRecent: 1625894270
    },
    {
      id: 3,
      notification: "Offline",
      threshold: null,
      machine: "machine1",
      mostRecent: 1625627930
    },
    {
      id: 4,
      notification: "CPU",
      threshold: 0.9,
      machine: "machine2",
      mostRecent: 1625721407
    },
    {
      id: 5,
      notification: "HDD",
      threshold: 0.9,
      machine: "machine2",
      mostRecent: 1625721527
    }
  ],
  network: [
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1622615148
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1622615145
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1625203607
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1625808407
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1625807927
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1625721527
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1625721407
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1625627930
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1625894270
    },
    {
      notification: "Finality issues",
      network: "Beaconchain",
      mostRecent: 1626078050
    }
  ]
};

var csrfToken = "";

const VALIDATOR_EVENTS = ["validator_attestation_missed","validator_balance_decreased",
                          "validator_proposal_missed","validator_proposal_submitted",
                          "validator_got_slashed"];
const MONITORING_EVENTS = ["monitoring_machine_offline", "monitoring_hdd_almostfull", "monitoring_cpu_load"];

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
  });

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
          return `<div class="font-weight-normal text-truncate high-contrast">${data.index}</div>`;
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
          var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + '+' : data.validator_indices.length;
          return `<div class="font-weight-normal high-contrast" style="display: flex;"><div class="text-truncate" style="flex: 1 1 auto;">${data.name}</div><div style="max-width: fit-content; white-space: nowrap;">${len}</div></div>`;
        }
    	}
  });

	$(input_container).on('focus', function(e) {
    if (e.target.value !== "") {
      $(this).trigger($.Event('keydown', {keyCode: 40}));
    }
  });

  $(input_container).on('input', function() {
    $('.tt-suggestion').first().addClass('tt-cursor');
  });

  $(input_container).on('typeahead:select', function(e, sug) {
    console.log(sug)
    $(input_container).val(sug.index);
    $(input_container).attr("pk", sug.pubkey);
  });
}

function loadMonitoringData(data) {
    // console.log(data)
  let mdata = []
  let id = 0
  for (let item of data){
      for (let n of item.Notifications){
          let ns = n.Notification.split("_")
          if (ns[0]==="monitoring"){
              mdata.push({
                id: id,
                notification: ns[1],
                threshold: n.Threshold,
                machine: item.Validator.Index,
                mostRecent: n.Timestamp
              })
              id+=1;
          }
      }
  }
//   console.log(mdata)
  let monitoringTable = $('#monitoring-notifications');

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
      $('body').find('.dataTables_scrollBody').addClass('scrollbar');
      
      // click event to monitoring table edit button
      $('#monitoring-notifications #edit-monitoring-events-btn').on('click', function(e) {
        e.stopPropagation();
        const threshold_editable_placeholder = $(this).parent().find('.threshold_non_editable_text').text().slice(0, -1);

        // close all other editable rows
        $('.threshold_editable').each(function() {
          $(this).attr('hidden', true);
          $(this).parent().find('.threshold_non_editable').css('display', 'inline-block');
        });

        $(this).parent().parent().find('.threshold_non_editable').css('display', 'none');
        $(this).parent().parent().find('.threshold_editable').removeAttr('hidden');
        $(this).parent().parent().find('.threshold_editable').attr('value', threshold_editable_placeholder);
      });

      // enter event to threshold input
      $('.threshold_editable').on('keypress', function(e) {
        if (e.which == 13) {
          const rowId = $(this).parent().parent().attr('id');
          let newThreshold = $(this).val();

          // validate input
          let isValid = false;
          if (isNaN(newThreshold) == false) {
            const parsed = parseInt(newThreshold, 10);
            if (isNaN(parsed)) {
              isValid = false;
            } else {
              if (parsed > 0 && parsed <= 100) {
                newThreshold = parsed / 100;
                isValid = true;
              } else {
                isValid = false;
              }
            }
          } else {
            isValid = false;
          }

          if (isValid) {
            index = data.findIndex(function(item) {
              return item.id.toString() === rowId.toString();
            });

            data[index].threshold = newThreshold;

            // destroy and reload table after edit
            $('#monitoring-notifications').DataTable().clear().destroy();
            loadMonitoringData(data);
        	} else {
          alert('Enter an integer between 1 and 100');
        	}
      	}
      });

      // click event to table remove button
      $('#monitoring-notifications #remove-btn').on('click', function(e) {
				$('#modaltext').text($(this).data('modaltext'));

				// set the row id 
				const rowId = $(this).parent().parent().attr('id');
				$('#confirmRemoveModal').attr('rowId', rowId);
				$('#confirmRemoveModal').attr('tablename', 'monitoring');
      });
    },
    columnDefs: [
      {
      	targets: '_all',
        createdCell: function(td, cellData, rowData, row, col) {
          $(td).css('padding-top', '20px');
          $(td).css('padding-bottom', '20px');
        }
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: 'notification',
        render: function(data, type, row, meta) {
          return '<span class="badge badge-pill badge-light badge-custom-size">' + data + '</span>';
        }
      },
      {
        targets: 1,
        responsivePriority: 3,
        data: 'threshold',
        render: function(data, type, row, meta) {
          if (!data) {
            return '<span class="threshold_non_editable">N/A</span>';
          }
          return '<input type="text" class="form-control input-sm threshold_editable" title="Numbers in 1-100 range (including)" style="width: 60px; height: 30px;" hidden /><span class="threshold_non_editable"><span class="threshold_non_editable_text">' + (data * 100).toFixed(2) + '%</span> <i class="fas fa-pen fa-xs text-muted i-custom" id="edit-monitoring-events-btn" title="Click to edit" style="padding: .5rem; cursor: pointer;"></i></span>';
        }
      },
      {
        targets: 2,
        responsivePriority: 2,
        data: 'machine',
        render: function(data, type, row, meta) {
            return `<span class="font-weight-bold"><i class="fas fa-male mr-1"></i><a style="padding: .25rem;" href="/validator/${data}">${data}</a></span>`
        }
      },
      {
        targets: 3,
        responsivePriority: 1,
        data: 'mostRecent',
        render: function (data, type, row, meta) {
        	// for sorting and type checking use the original data (unformatted)
          if (type === 'sort' || type === 'type') {
            return data;
          }

          if (parseInt(data)===0) return "N/a"
          return `<span class="heading-l4">${luxon.DateTime.fromMillis(data * 1000).toRelative({ style: "long" })}</span>`;
        }
    	},
      {
        targets: 4,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: '<i class="fas fa-times fa-lg i-custom" id="remove-btn" title="Remove notification" style="padding: .5rem; color: var(--new-red); cursor: pointer;" data-toggle= "modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>'
      }
    ],
  });
}

function loadNetworkData(data) {
  let networkTable = $('#network-notifications');

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
      $('body').find('.dataTables_scrollBody').addClass('scrollbar');
    },
    columnDefs: [
      {
        targets: '_all',
          createdCell: function(td, cellData, rowData, row, col) {
            $(td).css('padding-top', '20px');
            $(td).css('padding-bottom', '20px');
          }
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: 'notification',
        render: function(data, type, row, meta) {
          return '<span class="badge badge-pill badge-light badge-custom-size">' + data + '</span>';
      	}
      },
      {
        targets: 1,
        responsivePriority: 2,
        data: 'network'
      },
      {
        targets: 2,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: `
          <div class="form-check">
        		<input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="">
            <label class="form-check-label" for=""></label>
          </div>`
      },
      {
        targets: 3,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: `
          <div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="">
            <label class="form-check-label" for=""></label>
          </div>`
      },
      {
        targets: 4,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: `
          <div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="">
            <label class="form-check-label" for=""></label>
          </div>`
      },
      {
        targets: 5,
        responsivePriority: 1,
        data: 'mostRecent',
        render: function(data, type, row, meta) {
          // for sorting and type checking use the original data (unformatted)
          if (type === 'sort' || type === 'type') {
            return data;
          }
          return `<span class="heading-l4">${luxon.DateTime.fromMillis(data * 1000).toRelative({ style: "long" })}</span>`;
        }
    	}
    ]
  });
}

function loadValidatorsData(data) {
  // console.log(data);
  // validators = data;

  let validatorsTable = $('#validators-notifications');
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
      toggleable: false
    },
    fixedHeader: true,
    data: data,
    initComplete: function(settings, json) {
      $('body').find('.dataTables_scrollBody').addClass('scrollbar');

      // click event to validators table edit button
      $('#validators-notifications #edit-validator-events-btn').on('click', function(e) {
        $('#manageNotificationsModal').attr('rowId', $(this).parent().parent().attr('id'));
      });

      // click event to remove button
      $('#validators-notifications #remove-btn').on('click', function(e) {
        const rowId = $(this).parent().parent().attr('id');
        $('#modaltext').text($(this).data('modaltext'));

        // set the row id 
        $('#confirmRemoveModal').attr('rowId', rowId);
        $('#confirmRemoveModal').attr('tablename', 'validators');
      });
    },
    columnDefs: [
      {
        targets: '_all',
        createdCell: function(td, cellData, rowData, row, col) {
          $(td).css('padding-top', '20px');
          $(td).css('padding-bottom', '20px');
        }
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: 'Validator',
        render: function(data, type, row, meta) {
          // for sorting and type checking use the original data (unformatted)
          if (type === 'sort' || type === 'type') {
            return data.Index;
          }
          return `<span class="font-weight-bold"><i class="fas fa-male mr-1"></i><a style="padding: .25rem;" href="/validator/${data.Index}">` + data.Index + '</a></span>' + `<a class="heading-l4 d-none d-sm-block mt-2" style="width: 5rem;" href="/validator/${data.Pubkey}">0x` + data.Pubkey.substring(0, 6) + '...</a>';
        }
      },
      {
        targets: 1,
        responsivePriority: 2,
        data: 'Notifications',
        render: function(data, type, row, meta) {
          if (type === 'display') {
          let notifications = "";
          let hasItems = false;
          for (let notification of data) {
            // console.log(notification.Notification, VALIDATOR_EVENTS, VALIDATOR_EVENTS.includes(notification.Notification));
            if (VALIDATOR_EVENTS.includes(notification.Notification)) {
              hasItems=true;
              let badgeColor = "";
              switch (notification.Notification) {
                case 'validator_balance_decreased':
                  badgeColor = 'badge-light';
                  break;
                case 'validator_attestation_missed':
                  badgeColor = 'badge-warning';
                  break;
                case 'validator_proposal_submitted':
                  badgeColor = 'badge-light';
                  break;
                case 'validator_proposal_missed':
                  badgeColor = 'badge-warning';
                  break;
                case 'validator_got_slashed':
                  badgeColor = 'badge-light';
                  break;
              }
              notifications += '<span class="badge badge-pill ' + badgeColor + ' badge-custom-size mr-1 my-1">' + notification.Notification.replaceAll("_", " ") + '</span>';
            }
          }
          if (!hasItems) {
            return '<span>Not subscribed to any events</span><i class="d-block fas fa-pen fa-xs text-muted i-custom" id="edit-validator-events-btn" title="Manage the notifications you receive for the selected validator in the table" style="width: 1.5rem; padding: .5rem; cursor: pointer;" data-toggle= "modal" data-target="#manageNotificationsModal"></i>';
          }
          return '<div style="white-space: normal; max-width: 400px;">' + notifications + '</div>' + ' <i class="fas fa-pen fa-xs text-muted i-custom" id="edit-validator-events-btn" title="Manage the notifications you receive for the selected validator in the table" style="padding: .5rem; cursor: pointer;" data-toggle= "modal" data-target="#manageNotificationsModal"></i>';
        }
        return null;
        }
      },
      {
        targets: 2,
        orderable: false,
        responsivePriority: 4,
        data: null,
        defaultContent: `
          <div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="">
            <label class="form-check-label" for=""></label>
          </div>`,
        visible: false
      },
      {
        targets: 3,
        orderable: false,
        responsivePriority: 4,
        data: 'Notifications',
        render: function(data, type, row, meta) {
            let status = data.length > 0 ? 'checked="true"' : "";
            // console.log(data, data.length, data.length>0, status);
            return `
        	<div class="form-check">
            <input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="" ${status} disabled="true">
          	<label class="form-check-label" for=""></label>
          </div>`
        } 
      },
      {
        targets: 4,
        orderable: false,
        responsivePriority: 4,
        data: null,
        defaultContent: `
        	<div class="form-check">
          	<input class="form-check-input checkbox-custom-size" type="checkbox" value="" id="">
            <label class="form-check-label" for=""></label>
          </div>`,
        visible: false
      },
      {
        targets: 5,
        responsivePriority: 1,
        data: 'Notifications',
        render: function(data, type, row, meta) {
					// for sorting and type checking use the original data (unformatted)
					// data = data.Notifications
					let no_time = 'N/A';
          if (data.length === 0) {
            return no_time;
          }
          data.sort((a, b) => {
            return a.age - b.age;
          });
          if (type === 'sort' || type === 'type') {
            return data[0].Timestamp;
          }
          if (data[0].Timestamp === 0) {
            return no_time;
          }
          return '<span class="badge badge-pill badge-light badge-custom-size mr-1 mr-sm-3">' + data[0].Notification + '</span>' + `<span class="heading-l4 d-block d-sm-inline-block mt-2 mt-sm-0">${luxon.DateTime.fromMillis(data[0].Timestamp * 1000).toRelative({ style: "long" })}</span>`;
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
    rowCallback: function(row, data, displayNum, displayIndex, dataIndex) {
      $(row).attr('title', 'Click the table row to select it or hold down CTRL and click multiple rows to select them');
    },
    rowId: function(data, type, row, meta) {
      return data.Validator.Pubkey;
    }
  });

  // show manage-notifications button and remove-all button only if there is data in the validator table
  if (DATA.length !== 0) {
    $('#manage-notifications-btn').removeAttr('hidden');
    $('#remove-all-btn').removeAttr('hidden');
  }
}

function remove_item_from_event_container(pubkey) {
  for (let item of $('#selected-validators-events-container').find('span')) {
    if (pubkey === $(item).attr('pk')) {
      $(item).remove();
      return;
    }
  }
}

$(document).ready(function() {
  if (document.getElementsByName('CsrfField')[0] !== undefined) {
    csrfToken = document.getElementsByName('CsrfField')[0].value;
  }
  create_typeahead('.validator-typeahead');
  // create_typeahead('.monitoring-typeahead');


  loadValidatorsData(DATA);
  loadMonitoringData(DATA)
  loadNetworkData(data.network);

  $(document).on('click', function(e) {
    // if click outside input while any threshold input visible, reset value and hide input
    if (e.target.className.indexOf('threshold_editable') < 0) {
      $('.threshold_editable').each(function() {
        $(this).attr('hidden', true);
        $(this).parent().find('.threshold_non_editable').css('display', 'inline-block');
      });
    }

    // remove selected class from rows on click outside
    if (!$('#validators-notifications').is(e.target) && $('#validators-notifications').has(e.target).length === 0 && !$('#manage-notifications-btn').is(e.target) && $('#manage-notifications-btn').has(e.target).length === 0) {
      $('#validators-notifications .selected').removeClass('selected');
    }
  });

  $('#remove-all-btn').on('click', function(e) {
    $('#modaltext').text($(this).data('modaltext'));
    $('#confirmRemoveModal').removeAttr('rowId');
    $('#confirmRemoveModal').attr('tablename', 'validators');
  });

  // click event to modal remove button
  $('#remove-button').on('click', function(e) {
    const rowId = $('#confirmRemoveModal').attr('rowId');
    const tablename = $('#confirmRemoveModal').attr('tablename');

    // if rowId also check tablename then delete row in corresponding data section
  	// if no row id delete directly in correponding data section
    if (rowId !== undefined) {
      if (tablename === 'monitoring') {
        data.monitoring = data.monitoring.filter(function(item) {
          return item.id.toString() !== rowId.toString();
        });
      }

      if (tablename === 'validators') {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing validator...</span></div>');
        fetch(`/validator/${rowId}/remove`, {
          method: 'POST',
          headers: { "X-CSRF-Token": csrfToken },
          credentials: 'include',
          body: { pubkey: `0x${rowId}` }
        }).then(res => {
          if (res.status == 200) {
            $('#confirmRemoveModal').modal('hide');
            window.location.reload(false);
          } else {
            alert('Error removing validator from Watchlist');
            $('#confirmRemoveModal').modal('hide');
            window.location.reload();
          }
          $(this).html('Removed');
        });
      }
    } else {
    	if (tablename === 'validators') {
        $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Removing all...</span></div>');
        let pubkeys = [];
        for (let item of DATA) {
          pubkeys.push(item.Validator.Pubkey);
        }
        fetch(`/user/notifications-center/removeall`, {
          method: 'POST',
          headers: { "X-CSRF-Token": csrfToken },
          credentials: 'include',
          body: JSON.stringify(pubkeys)
        }).then(res => {
          if (res.status == 200) {
            $('#confirmRemoveModal').modal('hide');
            window.location.reload(false);
          } else {
            alert('Error removing all validators from Watchlist');
            $('#confirmRemoveModal').modal('hide');
            window.location.reload();
          }
          $(this).html('Removed');
        });
      }
    }

    // remove selected class from rows on click outside
    if (!$('#validators-notifications').is(e.target) && $('#validators-notifications').has(e.target).length === 0 && !$('#manage-notifications-btn').is(e.target) && $('#manage-notifications-btn').has(e.target).length === 0) {
      $('#validators-notifications .selected').removeClass('selected');
    }
    if (tablename === 'monitoring') {
      $('#monitoring-notifications').DataTable().clear().destroy();
      loadMonitoringData(data.monitoring);
    }
  });

  $('.range').on('input', function(e) {
    const target_id = $(this).data('target');
    let target = $(target_id);
    target.val($(this).val());
    if ($(this).attr('type') === 'range') {
      $(this).css('background-size', $(this).val() + '% 100%');
    } else {
      target.css('background-size', $(this).val() + '% 100%');
    }
  });

  $('#validators-notifications tbody').on('click', 'tr', function() {
    $(this).addClass('selected');
  });

  // on modal open after click event to validators table edit button
  $('#manageNotificationsModal').on('show.bs.modal', function(e) {
    // get the selected row (single row selected)
    let rowData = $('#validators-notifications').DataTable().row($('#' + $(this).attr('rowId'))).data();
    // console.log(rowData)
    if (rowData) {
      $('#selected-validators-events-container').append(
        `<span id="validator-event-badge" class="d-inline-block badge badge-pill badge-light badge-custom-size mr-2 mb-2 font-weight-normal" pk=${rowData.Validator.Pubkey}>
        		Validator ${rowData.Validator.Index}
          	<i class="fas fa-times ml-2" style="cursor: pointer;" onclick="remove_item_from_event_container('${rowData.Validator.Pubkey}')"></i>
        </span> `
      );

      for(let event of $('#manage_all_events :input')) { 
        for (let item of rowData.Notifications) {
          $(`#manage_${item.Notification} input#${$(event).attr('id')}`).prop('checked', true);
        }
      }
    } else {
      // get the selected rows (mutiple rows selected)
      const rowsSelected = $('#validators-notifications').DataTable().rows('.selected').data();
      for (let i = 0; i < rowsSelected.length; i++) {
        $('#selected-validators-events-container').append(
          `<span id="validator-event-badge" class="d-inline-block badge badge-pill badge-light badge-custom-size mr-2 mb-2 font-weight-normal" pk=${rowsSelected[i].Validator.Pubkey}>
            Validator ${rowsSelected[i].Validator.Index}
            <i class="fas fa-times ml-2" style="cursor: pointer;" onclick="remove_item_from_event_container('${rowsSelected[i].Validator.Pubkey}')"></i>
          </span> `
        );
      }
    }
  });

  // on modal close
	$('#manageNotificationsModal').on('hide.bs.modal', function(e) {
  	$(this).removeAttr('rowId');
    $('#selected-validators-events-container #validator-event-badge').remove();
    for(let event of $('#manage_all_events :input')) { 
      for (let item of VALIDATOR_EVENTS) {
        $(`#manage_${item} input#${$(event).attr('id')}`).prop('checked', false);
      }
      $(event).prop('checked', false);
    }

    $('[id^=all_events]').attr('checked', false);

    // remove selected class from rows when modal closed
    $('#validators-notifications .selected').removeClass('selected');
  });

  function get_validator_manage_sub_events() {
    let events = [];  
    for (let item of VALIDATOR_EVENTS) {
      events.push({
        event: item,
        email: $(`#manage_${item} :input#email`).prop('checked'),
        push: $(`#manage_${item} :input#push`).prop('checked'),
        web: $(`#manage_${item} :input#web`).prop('checked')
      });
    }
    return events;
  }

  $('#update-subs-btn').on('click', function() {
    let pubkeys = [];
    for (let item of $('#selected-validators-events-container').find('span')) {
      pubkeys.push($(item).attr('pk'));
    }
    if (pubkeys.length === 0) {
      return;
    }
    let events = get_validator_manage_sub_events();
    fetch(`/user/notifications-center/updatesubs`, {
      method: 'POST',
      headers: { "X-CSRF-Token": csrfToken },
      credentials: 'include',
      body: JSON.stringify({ pubkeys: pubkeys, events: events })
    }).then(res => {
      if (res.status == 200) {
        $('#manageNotificationsModal').modal('hide');
        window.location.reload(false);
      } else {
        alert('Error updating validators subscriptions');
        $('#manageNotificationsModal').modal('hide');
        window.location.reload();
      }
    }); 
  })

  // all events checkboxes (push, email, web)
  $('#all_events_push').on('click', function() {
    $('[id$=push]').attr('checked', $(this).is(':checked'));
  });
  $('#all_events_email').on('change', function() {
    $('[id$=email]').attr('checked', $(this).is(':checked'));
  });
  $('#all_events_web').on('change', function() {
    $('[id$=web]').attr('checked', $(this).is(':checked'));
  });

  // customize tables tooltips
  $('#validators-notifications').DataTable().$('tr').tooltip({ width: 5 });


  function get_validator_sub_events() {
    let events = [];  
    for (let item of VALIDATOR_EVENTS) {
      events.push({
        event: item,
        email: $(`#${item} :input#email`).prop('checked'),
        push: $(`#${item} :input#push`).prop('checked'),
        web: $(`#${item} :input#web`).prop('checked')
        });
      }
    return events;
  }

  $('#add-validator-button').on('click', function() {
    // console.log("clicked", $("#validator_attestation_missed>input#email"));
    try {
      let index = parseInt($('#add-validator-input').val());
      let events = get_validator_sub_events();
      // console.log(events);
      if (!isNaN(index)) {
        fetch(`/user/notifications-center/validatorsub`, {
          method: 'POST',
          headers: { "X-CSRF-Token": csrfToken },
          credentials: 'include',
          body: JSON.stringify({ pubkey: $('#add-validator-input').attr('pk'), events: events })
        }).then(res => {
            if (res.status == 200) {
              $('#addValidatorModal').modal('hide');
              window.location.reload(false);
            } else {
              alert('Error adding validators to Watchlist');
              $('#addValidatorModal').modal('hide');
              window.location.reload();
            }
          }); 
        }
    } catch {
      alert('Invalid Validator Index!');
      // console.log(validators, $("#add-validator-input").val());    
    }
  });

  //select/deselect notification checkboxes for all events
  for(let event of $('#validator_all_events :input')) { 
    $(event).on('click', function() {
      if ($(this).prop('checked')) {
        for (let item of VALIDATOR_EVENTS) {
          $(`#${item} input#${$(event).attr('id')}`).prop('checked', true);
        }
      } else {
        for (let item of VALIDATOR_EVENTS) {
          $(`#${item} input#${$(event).attr('id')}`).prop('checked', false);
        }
      }
    });
   }

   for(let event of $('#manage_all_events :input')) { 
    $(event).on('click', function() {
        if ($(this).prop('checked')) {
          for (let item of VALIDATOR_EVENTS) {
            $(`#manage_${item} input#${$(event).attr('id')}`).prop('checked', true);
          }
        } else {
           for (let item of VALIDATOR_EVENTS) {
            $(`#manage_${item} input#${$(event).attr("id")}`).prop('checked', false);
          }
        }
    });
  }

  $('#add-monitoring-event-modal-btn').on('click', function() {
    for (let item of DATA.sort((a,b) => { return a.Validator.Index-b.Validator.Index })) {
      $('#add-monitoring-validator-select').append(`<option value="${item.Validator.Pubkey}">${item.Validator.Index}</option>`);
    }
  });

  $('#add-monitoring-event-btn').on('click', function() {
    let pubkey = $('#add-monitoring-validator-select option:selected').val();
    events = [];
    for (let item of $('input.monitoring')) {
      if ($(item).prop('checked')) {
        // console.log($(item), $(item).attr("event"), $(item).prop("event"));
        let e = $(item).attr('event');
        let t = 0;
        switch(e) {
          case 'monitoring_cpu_load':
            t = parseFloat($('#cpu-input-range-val').val())/100;
            break;
          case 'monitoring_hdd_almostfull':
            t = parseFloat($('#hdd-input-range-val').val())/100;
            break;
          default:
            t=0;
        }
        events.push({
          event: e,
          email: true,
          threshold: t
        });
      }
    }
    // console.log(pubkey, events);
    fetch(`/user/notifications-center/monitoring/updatesubs`, {
      method: 'POST',
      headers: { "X-CSRF-Token": csrfToken },
      credentials: 'include',
      body: JSON.stringify({ pubkeys: [pubkey], events: events })
    }).then(res => {
      if (res.status == 200) {
        $('#manageNotificationsModal').modal('hide');
        window.location.reload(false);
      } else {
        alert('Error updating validators subscriptions');
        $('#manageNotificationsModal').modal('hide');
        window.location.reload();
      }
    }); 
  });
});
