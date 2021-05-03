const VALLIMIT = 200
var currency = ""
// let validators = []

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
    })
    var bhEth1Addresses = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        identify: function (obj) {
            return obj.eth1_address
        },
        remote: {
            url: '/search/indexed_validators_by_eth1_addresses/%QUERY',
            wildcard: '%QUERY'
        }
    })
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
    })
    var bhGraffiti = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        identify: function (obj) {
            return obj.graffiti
        },
        remote: {
            url: '/search/indexed_validators_by_graffiti/%QUERY',
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
                header: '<h3>Validators</h3>',
                suggestion: function (data) {
                    return `<div class="text-monospace text-truncate high-contrast">${data.index}</div>`
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
                suggestion: function (data) {
                    var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + '+' : data.validator_indices.length
                    return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.eth1_address}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
                }
            }
        },
        {
            limit: 5,
            name: 'graffiti',
            source: bhGraffiti,
            display: 'graffiti',
            templates: {
                header: '<h3>Validators by Graffiti</h3>',
                suggestion: function (data) {
                    var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + '+' : data.validator_indices.length
                    return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.graffiti}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
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
                    var len = data.validator_indices.length > VALLIMIT ? VALLIMIT + '+' : data.validator_indices.length
                    return `<div class="text-monospace high-contrast" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.name}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
                }
            }
        }
    )

    $(input_container).on('focus', function (event) {
        if (event.target.value !== '') {
            $(this).trigger($.Event('keydown', { keyCode: 40 }))
        }
    })
    $(input_container).on('input', function () {
        $('.tt-suggestion').first().addClass('tt-cursor')
    })
    $(input_container).on('blur', function () {
        $(this).val('')
        $(input_container).typeahead('val', '')
    })
    $(input_container).on('typeahead:select', function (ev, sug) {
        validators = $('#validator-index-view').val().split(",");
        if (sug.validator_indices) {
            validators = validators.concat(sug.validator_indices);
        } else {
            validators.push(sug.index);
        }
        validators = Array.from(new Set(validators));
        if (validators.length > VALLIMIT){
            validators = validators.slice(0, VALLIMIT)
            alert(`No more than ${VALLIMIT} validators are allowed`)
        }
        $('#validator-index-view').val(validators);
        $(input_container).typeahead('val', '')
    })
}


function updateCurrencies(currencies, container) {
    for (item of currencies) {
        if (item === "ts") continue;
        $(container).append(`<option value="${item}">${item.toUpperCase()}</option>`);
    }

}

function getValidatorQueryString() {
    return window.location.href.slice(window.location.href.indexOf("?"), window.location.href.length)
}

function hideSpinner(){
    $("#loading-div").addClass("d-none")
    $("#loading-div").removeClass("d-flex")
}

function showTable(data){
    if (data.length > 0 && data[0].length === 6){
        currency = data[0][5].toUpperCase()
    }
    $('#tax-table').DataTable({
        processing: true,
        serverSide: false,
        ordering: true,
        searching: true,
        pagingType: 'full_numbers',
        pageLength: 100,
        lengthChange: false,
        data: data,
        dom: 'Bfrtip',
        buttons: [
            'copyHtml5',
            'excelHtml5',
            'csvHtml5',
            'pdfHtml5'
        ],
        drawCallback: function (settings) {
            hideSpinner()
            $("#form-div").addClass("d-none")
            $("#table-div").removeClass("d-none")
        },
        columnDefs: [
            {
                targets: 0,
                data: '0',
                "orderable": true,
                render: function (data, type, row, meta) {
                    return data
                }
            }, {
                targets: 1,
                data: '1',
                "orderable": true,
                render: function (data, type, row, meta) {
                    return parseFloat(data).toFixed(5)
                }
            }, {
                targets: 2,
                data: '2',
                "orderable": true,
                render: function (data, type, row, meta) {
                    return parseFloat(data).toFixed(5)
                }
            }, {
                targets: 3,
                data: '3',
                "orderable": false,
                render: function (data, type, row, meta) {
                    return `${currency} ${parseFloat(data).toFixed(5)}`
                }
            }, {
                targets: 4,
                data: '4',
                "orderable": false,
                render: function (data, type, row, meta) {
                   return `${currency} ${parseFloat(data).toFixed(5)}`
                }
            }, {
                targets: 5,
                data: '5',
                "orderable": false,
                visible: false,
                render: function (data, type, row, meta) {
                    return data.toUpperCase()
                }
            }]
    });
}

$(document).ready(function () {
    create_typeahead('.typeahead-validators');
    let qry = getValidatorQueryString()
    if (qry.length > 1){
        fetch(`/rewards/hist${qry}`,{
            method: "GET"
          }).then((res)=>{
            res.json().then((data)=>{
              showTable(data)
            })
          }).catch(()=>{
            alert("Failed to fetch the data :(")
            hideSpinner()      
          })
    }else{
        hideSpinner()
    }
})