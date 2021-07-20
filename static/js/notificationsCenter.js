const data = {
    metrics: {
        validators: 5,
        notifications: 10,
        attestationsSubmitted: 1,
        attestationsMissed: 1,
        proposalsSubmitted: 2,
        proposalsMissed: 0
    },
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
        },
        {
            id: 6,
            notification: "Offline",
            threshold: null,
            machine: "machine2",
            mostRecent: 1625807927
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
    ],
    validators: [
        {
            id: 1,
            validator: { index: 1, pubkey: "0xa1d1ad ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted", "Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1617427430 }
        },
        {
            id: 2,
            validator: { index: 2, pubkey: "0xb2ff47 ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted"],
            mostRecent: { notification: "Balance decrease", timestamp: 1620030169 }
        },
        {
            id: 3,
            validator: { index: 3, pubkey: "0x8e323f ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed"],
            mostRecent: { notification: "Proposals missed", timestamp: 1614687712 }
        },
        {
            id: 4,
            validator: { index: 4, pubkey: "0xa62420 ..." },
            notifications: ["Attestations missed", "Balance decrease"],
            mostRecent: { notification: "Proposals submitted", timestamp: 1622557312 }
        },
        {
            id: 5,
            validator: { index: 5, pubkey: "0xb2ce0f ..." },
            notifications: ["Attestations missed"],
            mostRecent: { notification: "Validator slashed", timestamp: 1619878855 }
        },
        {
            id: 6,
            validator: { index: 6, pubkey: "0xa16c53 ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted", "Validator slashed"],
            mostRecent: { notification: "Validator slashed", timestamp: 1612617898 }
        },
        {
            id: 7,
            validator: { index: 7, pubkey: "0xa25da1..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted"],
            mostRecent: { notification: "Proposals submitted", timestamp: 1626006179 }
        },
        {
            id: 8,
            validator: { index: 8, pubkey: "0x8078c7 ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed"],
            mostRecent: { notification: "Proposals missed", timestamp: 1626031079 }
        },
        {
            id: 9,
            validator: { index: 9, pubkey: "0xb016e3 ..." },
            notifications: ["Attestations missed", "Balance decrease"],
            mostRecent: { notification: "Balance decrease", timestamp: 1620677579 }
        },
        {
            id: 10,
            validator: { index: 10, pubkey: "0x8efba2 ..." },
            notifications: ["Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1619878855 }
        }
    ]
};

// for (let key in data.metrics) {
//     if (data.metrics.hasOwnProperty(key)) document.getElementById(key).innerHTML = data.metrics[key];
// }

function loadMonitoringData(data) {
  let monitoringTable = $('#monitoring-notifications');

  monitoringTable.DataTable ({
    language: {
      info: '_TOTAL_ entries',
      infoEmpty: 'No entries match',
      infoFiltered: '(from _MAX_ entries)',
      processing: 'Loading. Please wait...',
      search: '',
      searchPlaceholder: 'Search...',
      zeroRecords: 'No entries match'
    },
    rowId: 'id',
    processing: true,
    responsive: true,
    scroller: true,
    scrollY: 380,
    paging: true,
    data: data,
    initComplete: function(settings, json) {
      $('body').find('.dataTables_scrollBody').addClass('scrollbar');

      // click event to table edit button
      $('#monitoring-notifications #edit-btn').on('click',  function(e) {
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
                newThreshold = parsed/100;
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
          alert('Enter an integer between 1 and 100')
      	}
    	}
  	});

  	// click event to table remove button
  	$('#monitoring-notifications #remove-btn').on('click',  function(e) {
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
      createdCell: function (td, cellData, rowData, row, col) {
        $(td).css('padding-top', '20px');
        $(td).css('padding-bottom', '20px');
    	}
    },
    {
      targets: 0,
      responsivePriority: 1,
      data: 'notification',
      render: function (data, type, row, meta) {
        return '<span class="badge badge-pill badge-primary badge-custom-size">' + data + '</span>';
      }
    },
    {
      targets: 1,
      responsivePriority: 3,
      data: 'threshold',
      render: function (data, type, row, meta) { 
        if (!data) {
          return '<span class="threshold_non_editable">N/A</span>';
        }
        return '<input type="text" class="form-control input-sm threshold_editable" title="Numbers in 1-100 range (including)" style="width: 60px; height: 30px;" hidden /><span class="threshold_non_editable"><span class="threshold_non_editable_text">' + data*100 + '%</span> <i class="fas fa-pen fa-xs text-muted" id="edit-btn" title="Click to edit" style="cursor: pointer;"></i></span>';
      }
    },
    {
      targets: 2,
      responsivePriority: 2,
    	data: 'machine'
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
        return `<span class="heading-l4">${luxon.DateTime.fromMillis(data * 1000).toRelative({ style: "long" })}</span>`;
      }
    },
    {
      targets: 4,
      orderable: false,
      responsivePriority: 3,
      data: null,
      defaultContent: '<i class="fas fa-times fa-lg" id="remove-btn" title="Remove notification" style="color: #f82e2e; cursor: pointer;" data-toggle= "modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>'
    }
  ]
	});
}

function loadNetworkData(data) {
  let networkTable = $('#network-notifications');

  networkTable.DataTable ({
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
        createdCell: function (td, cellData, rowData, row, col) {
      		$(td).css('padding-top', '20px');
          $(td).css('padding-bottom', '20px');
    		}
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: 'notification',
        render: function (data, type, row, meta) {
          return '<span class="badge badge-pill badge-primary badge-custom-size">' + data + '</span>';
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
        render: function (data, type, row, meta) {
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
  let validatorsTable = $('#validators-notifications');

  validatorsTable.DataTable ({
    language: {
      info: '_TOTAL_ entries',
      // TODO: place at the bottom of the container
      infoEmpty: 'No entries match',
      infoFiltered: '(from _MAX_ entries)',
      processing: 'Loading. Please wait...',
      search: '',
      searchPlaceholder: 'Search...',
      zeroRecords: 'No entries match'
    },
    rowId: 'id',
    processing: true,
    responsive: true,
    // scroller: true,
    // scrollY: 610,
    paging: true,
    // TODO: place at the bottom of the container
    pagingType: 'first_last_numbers',
    select: {
      items: 'row',
      blurable: true,
      className: 'row-selected'
    },
    fixedHeader: true,
    data: data,
    initComplete: function(settings, json) {
      $('body').find('.dataTables_scrollBody').addClass('scrollbar');
            
      // click event to remove button
      $('#validators-notifications #remove-btn').on('click',  function(e) {
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
        createdCell: function (td, cellData, rowData, row, col) {
      		$(td).css('padding-top', '20px');
          $(td).css('padding-bottom', '20px');
    		}
      },
      {
        targets: 0,
        responsivePriority: 1,
        data: 'validator',
        render: function (data, type, row, meta) {
          return '<span>' + data.index + '</span>' + '<span class="heading-l4 d-none d-sm-block mt-2">' + data.pubkey + '</span>';
        }
    	},
      {
        targets: 1,
        responsivePriority: 2,
        data: 'notifications',
        render: function (data, type, row, meta) {
          let notifications = '';
          for (let notification in data) {
            notifications += '<span class="badge badge-pill badge-primary badge-custom-size mr-1 my-1">' + data[notification] + '</span>';
          }
          // TODO: add functionality for edit button
          return '<div style="white-space: normal; max-width: 400px;">' + notifications + '</div>' + ' <i class="fas fa-pen fa-xs text-muted" id="edit-btn" title="Click to edit validator notifications" style="cursor: pointer;"></i>';
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
          </div>`
      },
      {
        targets: 3,
        orderable: false,
        responsivePriority: 4,
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
        responsivePriority: 4,
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
        render: function (data, type, row, meta) {
          // for sorting and type checking use the original data (unformatted)
          if (type === 'sort' || type === 'type'){
            return data.timestamp;
          }
          return '<span class="badge badge-pill badge-primary badge-custom-size mr-1 mr-sm-3">' + data.notification + '</span>' + `<span class="heading-l4 d-block d-sm-inline-block mt-2 mt-sm-0">${luxon.DateTime.fromMillis(data.timestamp * 1000).toRelative({ style: "long" })}</span>`;
        }
    	},
      {
      	targets: 6,
        orderable: false,
        responsivePriority: 3,
        data: null,
        defaultContent: '<i class="fas fa-times fa-lg" id="remove-btn" title="Remove validator" style="color: #f82e2e; cursor: pointer;" data-toggle= "modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>'
    	}
    ]
  });
}

$(document).ready(function () {
  loadMonitoringData(data.monitoring);
  loadNetworkData(data.network);
  loadValidatorsData(data.validators);

  $(document).on('click', function(e) {
    // if click outside input while any threshold input visible, reset value and hide input
    if (e.target.className.indexOf('threshold_editable') < 0){
      $('.threshold_editable').each(function() {
        $(this).attr('hidden', true);
      	$(this).parent().find('.threshold_non_editable').css('display', 'inline-block');
      });
    }
  });

  $('#remove-all-btn').on('click',  function(e) {
    $('#modaltext').text($(this).data('modaltext'));
    $('#confirmRemoveModal').removeAttr('rowId');
    $('#confirmRemoveModal').attr('tablename', 'validators');
  });

  // click event to modal remove button
  $('#remove-button').on('click',  function(e) {
    const rowId = $('#confirmRemoveModal').attr('rowId');
    const tablename = $('#confirmRemoveModal').attr('tablename');
        
    // if rowId also check tablename then delete row in corresponding data section
    // if no row id delete directly in correponding data section
    if (rowId !== undefined) {
    	if (tablename === 'monitoring') {
        data.monitoring = data.monitoring.filter(function( item ) {
          return item.id.toString() !== rowId.toString();
        });
      }

      if (tablename === 'validators') {
        data.validators = data.validators.filter(function( item ) {
          return item.id.toString() !== rowId.toString();
        });						
      }
    } else {
      if (tablename === 'validators') {
        data.validators = [];
      }
    }

    if (tablename === 'monitoring') {
      $('#monitoring-notifications').DataTable().clear().destroy();
    	loadMonitoringData(data.monitoring);
    }

  	if (tablename === 'validators') {
      $('#validators-notifications').DataTable().clear().destroy();
      loadValidatorsData(data.validators);					
    }
  });

  $('.range').on('input', function(event) {
    var target_id = $(this).data('target');
    var target = $(target_id);
    target.val($(this).val());
    if ($(this).attr('type') === 'range') {
      $(this).css('background-size', $(this).val() + '% 100%');
    } else {
      target.css('background-size', $(this).val() + '% 100%');
    }
    
    console.log(target, target.val())
    // target.cs('background-size', `${$(this).val()}%`);
    // target.css('background-size', $(this).val() + '%');
  });
});
