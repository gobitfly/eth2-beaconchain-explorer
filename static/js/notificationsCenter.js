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
            validator: { index: 1, pubkey: "0xa1d1ad ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted", "Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1617427430 }
        },
        {
            validator: { index: 2, pubkey: "0xb2ff47 ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted"],
            mostRecent: { notification: "Balance decrease", timestamp: 1620030169 }
        },
        {
            validator: { index: 3, pubkey: "0x8e323f ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed"],
            mostRecent: { notification: "Proposals missed", timestamp: 1614687712 }
        },
        {
            validator: { index: 4, pubkey: "0xa62420 ..." },
            notifications: ["Attestations missed", "Balance decrease"],
            mostRecent: { notification: "Proposals submitted", timestamp: 1622557312 }
        },
        {
            validator: { index: 5, pubkey: "0xb2ce0f ..." },
            notifications: ["Attestations missed"],
            mostRecent: { notification: "Validator slashed", timestamp: 1619878855 }
        },
        {
            validator: { index: 6, pubkey: "0xa16c53 ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted", "Validator slashed"],
            mostRecent: { notification: "Validator slashed", timestamp: 1612617898 }
        },
        {
            validator: { index: 7, pubkey: "0xa25da1..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed", "Proposals submitted"],
            mostRecent: { notification: "Proposals submitted", timestamp: 1626006179 }
        },
        {
            validator: { index: 8, pubkey: "0x8078c7 ..." },
            notifications: ["Attestations missed", "Balance decrease", "Proposals missed"],
            mostRecent: { notification: "Proposals missed", timestamp: 1626031079 }
        },
        {
            validator: { index: 9, pubkey: "0xb016e3 ..." },
            notifications: ["Attestations missed", "Balance decrease"],
            mostRecent: { notification: "Balance decrease", timestamp: 1620677579 }
        },
        {
            validator: { index: 10, pubkey: "0x8efba2 ..." },
            notifications: ["Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1619878855 }
        },
        {
            validator: { index: 11, pubkey: "0x8efba3 ..." },
            notifications: ["Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1619878856 }
        },
        {
            validator: { index: 12, pubkey: "0x8efba4 ..." },
            notifications: ["Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1619878857 }
        },
        {
            validator: { index: 13, pubkey: "0x8efba5 ..." },
            notifications: ["Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1619878858 }
        },
        {
            validator: { index: 20, pubkey: "0x8efba6 ..." },
            notifications: ["Validator slashed"],
            mostRecent: { notification: "Attestations missed", timestamp: 1619878859 }
        }
    ]
};

var csrfToken = "";
var validators = null;

function create_typeahead(input_container) {
    var bhValidators = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        identify: function (obj) {
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
        identify: function (obj) {
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
                header: '<h3>Validators</h3>',
                suggestion: function (data) {
                    return `<div class="text-monospace text-truncate high-contrast">${data.index}</div>`;
                }
            }
        },
        {
            limit: 5,
            name: 'name',
            source: bhName,
            display: 'name',
            templates: {
                header: '<h3>Validators by Name</h3>',
                suggestion: function (data) {
                    var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + '+' : data.validator_indices.length;
                    return `<div class="text-monospace high-contrast" style="display: flex;"><div class="text-truncate" style="flex: 1 1 auto;">${data.name}</div><div style="max-width: fit-content; white-space: nowrap;">${len}</div></div>`;
                }
            }
        });

    $(input_container).on('focus', function (event) {
        if (event.target.value !== "") {
            $(this).trigger($.Event('keydown', { keyCode: 40 }));
        }
    });

    $(input_container).on('input', function () {
        $('.tt-suggestion').first().addClass('tt-cursor');
    });

    $(input_container).on('typeahead:select', function (ev, sug) {
        $(input_container).val(sug.index);
    });
}

function loadMonitoringData(data) {
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
        data: data,
        rowId: 'id',
        initComplete: function (settings, json) {
            $('body').find('.dataTables_scrollBody').addClass('scrollbar');

            // click event to monitoring table edit button
            $('#monitoring-notifications #edit-monitoring-events-btn').on('click', function (e) {
                e.stopPropagation();
                const threshold_editable_placeholder = $(this).parent().find('.threshold_non_editable_text').text().slice(0, -1);

                // close all other editable rows
                $('.threshold_editable').each(function () {
                    $(this).attr('hidden', true);
                    $(this).parent().find('.threshold_non_editable').css('display', 'inline-block');
                });

                $(this).parent().parent().find('.threshold_non_editable').css('display', 'none');
                $(this).parent().parent().find('.threshold_editable').removeAttr('hidden');
                $(this).parent().parent().find('.threshold_editable').attr('value', threshold_editable_placeholder);
            });

            // enter event to threshold input
            $('.threshold_editable').on('keypress', function (e) {
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
                        index = data.findIndex(function (item) {
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
            $('#monitoring-notifications #remove-btn').on('click', function (e) {
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
                    return '<span class="badge badge-pill badge-light badge-custom-size">' + data + '</span>';
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
                    return '<input type="text" class="form-control input-sm threshold_editable" title="Numbers in 1-100 range (including)" style="width: 60px; height: 30px;" hidden /><span class="threshold_non_editable"><span class="threshold_non_editable_text">' + data * 100 + '%</span> <i class="fas fa-pen fa-xs text-muted" id="edit-monitoring-events-btn" title="Click to edit" style="padding: .5rem; cursor: pointer;"></i></span>';
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
                defaultContent: '<i class="fas fa-times fa-lg" id="remove-btn" title="Remove notification" style="padding: .5rem; color: #f82e2e; cursor: pointer;" data-toggle= "modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>'
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
        initComplete: function (settings, json) {
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
    validators = data;
    let validatorsTable = $('#validators-notifications');
    // console.log('calling with', data);
    validatorsTable.DataTable({
        language: {
            info: '_TOTAL_ entries',
            infoEmpty: 'No entries match',
            infoFiltered: '(from _MAX_ entries)',
            processing: 'Loading. Please wait...',
            search: '',
            searchPlaceholder: 'Search...',
            select: {
                rows: {
                    _: '%d rows selected',
                    0: 'Click on a row to select it',
                    1: '1 row selected'
                }
            },
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
        initComplete: function (settings, json) {
            $('body').find('.dataTables_scrollBody').addClass('scrollbar');

            // click event to validators table edit button
            $('#validators-notifications #edit-validator-events-btn').on('click', function (e) {
                $('#manageNotificationsModal').attr('rowId', $(this).parent().parent().attr('id'));
            });

            // click event to remove button
            $('#validators-notifications #remove-btn').on('click', function (e) {
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
                data: 'Validator',
                render: function (data, type, row, meta) {
                    // for sorting and type checking use the original data (unformatted)
                    if (type === 'sort' || type === 'type') {
                        return data.Index;
                    }
                    return `<span class="font-weight-bold"><a href="/validator/${data.Index}"><i class="fas fa-male mr-1"></i>` + data.Index + '</a></span>' + `<a class="heading-l4 d-none d-sm-block mt-2" style="width: 5rem;" href="/validator/${data.Pubkey}">0x` + data.Pubkey.substring(0, 6) + '...</a>';
                }
            },
            {
                targets: 1,
                responsivePriority: 2,
                data: 'Notifications',
                render: function (data, type, row, meta) {
                    let notifications = '';
                    if (data.length === 0) {
                        return '<span>Not subscribed to any events</span>';
                    }
                    for (let notification of data) {
                        let badgeColor = '';
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
                    return '<div style="white-space: normal; max-width: 400px;">' + notifications + '</div>' + ' <i class="fas fa-pen fa-xs text-muted" id="edit-validator-events-btn" title="Manage the notifications you receive for the selected validator in the table" style="padding: .5rem; cursor: pointer;" data-toggle= "modal" data-target="#manageNotificationsModal"></i>';
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
                data: 'Notifications',
                render: function (data, type, row, meta) {
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
                defaultContent: '<i class="fas fa-times fa-lg" id="remove-btn" title="Remove validator" style="padding: .5rem; color: #f82e2e; cursor: pointer;" data-toggle= "modal" data-target="#confirmRemoveModal" data-modaltext="Are you sure you want to remove the entry?"></i>'
            }
        ],
        rowCallback: function (row, data, displayNum, displayIndex, dataIndex) {
            $(row).attr('title', 'Click the table row to select it or hold down the "Shift" key and click multiple rows to select them');
        },
        rowId: function (data, type, row, meta) {
            return data.Validator.Pubkey;
        }
    });
}

$(document).ready(function () {
    if (document.getElementsByName("CsrfField")[0] !== undefined) {
        csrfToken = document.getElementsByName('CsrfField')[0].value;
    }
    create_typeahead('.validator-typeahead');

    loadMonitoringData(data.monitoring);
    loadNetworkData(data.network);

    $(document).on('click', function (e) {
        // if click outside input while any threshold input visible, reset value and hide input
        if (e.target.className.indexOf('threshold_editable') < 0) {
            $('.threshold_editable').each(function () {
                $(this).attr('hidden', true);
                $(this).parent().find('.threshold_non_editable').css('display', 'inline-block');
            });
        }

        // remove selected class from rows on click outside
        if (!$('#validators-notifications').is(e.target) && $('#validators-notifications').has(e.target).length === 0 && !$('#manage-notifications-btn').is(e.target) && $('#manage-notifications-btn').has(e.target).length === 0) {
            $('#validators-notifications .selected').removeClass('selected');
        }
    });

    $('#remove-all-btn').on('click', function (e) {
        $('#modaltext').text($(this).data('modaltext'));
        $('#confirmRemoveModal').removeAttr('rowId');
        $('#confirmRemoveModal').attr('tablename', 'validators');
    });

    // click event to modal remove button
    $('#remove-button').on('click', function (e) {
        const rowId = $('#confirmRemoveModal').attr('rowId');
        const tablename = $('#confirmRemoveModal').attr('tablename');

        // if rowId also check tablename then delete row in corresponding data section
        // if no row id delete directly in correponding data section
        if (rowId !== undefined) {
            if (tablename === 'monitoring') {
                data.monitoring = data.monitoring.filter(function (item) {
                    return item.id.toString() !== rowId.toString();
                });
            }

            if (tablename === 'validators') {
                // console.log(rowId)
                $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Loading...</span></div>')
                fetch(`/validator/${rowId}/remove`, {
                    method: 'POST',
                    headers: { "X-CSRF-Token": csrfToken },
                    credentials: 'include',
                    body: { pubkey: `0x${rowId}` },
                }).then((res) => {

                    if (res.status == 200) {
                        $('#confirmRemoveModal').modal('hide');
                        window.location.reload(false)
                    } else {
                        alert("Error removing validator from Watchlist")
                        $('#confirmRemoveModal').modal('hide');
                    }
                    $(this).html('Remove')
                })

            }
        } else {
            if (tablename === 'validators') {
                // data.validators = [];
                $(this).html('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Loading...</span></div>')
                let pubkeys = []
                for (let item of validators){
                    pubkeys.push(item.Validator.Pubkey)
                }
                fetch(`/user/notifications-center/removeall`, {
                    method: 'POST',
                    headers: { "X-CSRF-Token": csrfToken },
                    credentials: 'include',
                    body: JSON.stringify(pubkeys),
                }).then((res) => {

                    if (res.status == 200) {
                        $('#confirmRemoveModal').modal('hide');
                        window.location.reload(false)
                    } else {
                        alert("Error removing All validator from Watchlist")
                        $('#confirmRemoveModal').modal('hide');
                    }
                    $(this).html('Remove')
                })
            }
        }

        if (tablename === 'monitoring') {
            $('#monitoring-notifications').DataTable().clear().destroy();
            loadMonitoringData(data.monitoring);
        }

        /* if (tablename === 'validators') {
    $('#validators-notifications').DataTable().clear().destroy();
    loadValidatorsData(data.validators);					
  } */
    });

    $('.range').on('input', function (e) {
        var target_id = $(this).data('target');
        var target = $(target_id);
        target.val($(this).val());
        if ($(this).attr('type') === 'range') {
            $(this).css('background-size', $(this).val() + '% 100%');
        } else {
            target.css('background-size', $(this).val() + '% 100%');
        }
    });

    $('#validators-notifications tbody').on('click', 'tr', function () {
        $(this).addClass('selected');
    });

    // on modal open after click event to validators table edit button
    $('#manageNotificationsModal').on('show.bs.modal', function (e) {
        // get the selected row (single row selected)
        let rowData = $('#validators-notifications').DataTable().row($('#' + $(this).attr('rowId'))).data();
        if (rowData) {
            console.log(rowData);
            $('#selected-validators-events-container').append(
                `<div id="validator-event-badge" class="d-inline-block badge badge-pill badge-light badge-custom-size mr-2 mb-2 font-weight-normal">
          Validator ${rowData.Validator.Index}
          <i class="fas fa-times ml-2" style="cursor: pointer;"></i>
        </div> `
            );

            rowData.Notifications.forEach(function (notification) {
                if (notification.Notification === 'validator_balance_decreased') {
                    $('[id^=validator_balance_decreased]').attr('checked', true);
                }
                if (notification.Notification === 'validator_attestation_missed') {
                    $('[id^=validator_attestation_missed]').attr('checked', true);
                }
                if (notification.Notification === 'validator_proposal_submitted') {
                    $('[id^=validator_proposal_submitted]').attr('checked', true);
                }
                if (notification.Notification === 'validator_proposal_missed') {
                    $('[id^=validator_proposal_missed]').attr('checked', true);
                }
                if (notification.Notification === 'validator_got_slashed') {
                    $('[id^=validator_got_slashed]').attr('checked', true);
                }
            });
        } else {
            // get the selected rows (mutiple rows selected)
            const rowsSelected = $('#validators-notifications').DataTable().rows('.selected').data();
            for (let i = 0; i < rowsSelected.length; i++) {
                $('#selected-validators-events-container').append(
                    `<div id="validator-event-badge" class="d-inline-block badge badge-pill badge-light badge-custom-size mr-2 mb-2 font-weight-normal">
            Validator ${rowsSelected[i].Validator.Index}
            <i class="fas fa-times ml-2" style="cursor: pointer;"></i>
          </div> `
                );
            }
        }
    });
    // on modal close
    $('#manageNotificationsModal').on('hide.bs.modal', function (e) {
        $(this).removeAttr('rowId');
        $('#selected-validators-events-container #validator-event-badge').remove();
        $('[id^=validator_balance_decreased]').attr('checked', false);
        $('[id^=validator_attestation_missed]').attr('checked', false);
        $('[id^=validator_proposal_submitted]').attr('checked', false);
        $('[id^=validator_proposal_missed]').attr('checked', false);
        $('[id^=validator_got_slashed]').attr('checked', false);

        // remove selected class from rows when modal closed
        $('#validators-notifications .selected').removeClass('selected');
    });
});
