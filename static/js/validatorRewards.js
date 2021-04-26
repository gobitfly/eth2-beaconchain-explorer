const VALLIMIT = 200
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
            console.log(sug.validator_indices)
            validators = validators.concat(sug.validator_indices);
        } else {
            console.log(sug.index)
            validators.push(sug.index);
        }
        validators = Array.from(new Set(validators));
        $('#validator-index-view').val(validators);
        $(input_container).typeahead('val', '')
    })
}


function updateCurrencies(currencies, container){
    for (item of currencies){
        if(item==="ts")continue;
        $(container).append(`<option>${item.toUpperCase()}</option>`);
    }

}


$(document).ready(function () {
    create_typeahead('.typeahead-validators');


})