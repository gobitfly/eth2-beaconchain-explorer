const VALLIMIT = 200
const DECIMAL_POINTS_ETH = 6
const DECIMAL_POINTS_CURRENCY = 3
var csrfToken = ""
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
        if ($('#validator-index-view').val().charAt(0) === ",") {
            $('#validator-index-view').val($('#validator-index-view').val().slice(1))
        }
        $(input_container).typeahead('val', '')
    })
}


function updateCurrencies(currencies, container) {
    for (item of currencies) {
        if (item === "ts") continue;
        $("#"+container).append(`<option value="${item}">${item.toUpperCase()}</option>`);
    }

}

function getValidatorQueryString() {
    return window.location.href.slice(window.location.href.indexOf("?"), window.location.href.length)
}

function hideSpinner(){
    $("#loading-div").addClass("d-none")
    $("#loading-div").removeClass("d-flex")
}

function updateTotals(data){
    totalEth = 0.0
    totalCurrency = 0.0

    for(let item of data){
        totalEth+=parseFloat(item[2])
        totalCurrency+=parseFloat(item[4])
    }

    $("#total-income-eth-span").html(`ETH: <b>${(totalEth.toFixed(DECIMAL_POINTS_ETH))}</b>`)
    $("#total-income-currency-span").html(`${currency}: <b>${addCommas(totalCurrency.toFixed(DECIMAL_POINTS_CURRENCY))}</b>`)
    $("#totals-div").removeClass("d-none")
}

function addCommas(number) {
    return number.toString().replace(/,/g, "").replace(/\B(?=(\d{3})+(?!\d))/g, ",")
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
            $("#subscriptions-div").addClass("d-none")
            updateTotals(data)
            $(".dt-button").addClass("ml-2 ")
            $(".dt-button").attr("style", "border-radius: 20px; border-style: none; opacity: 0.9;")
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
                    return (parseFloat(data).toFixed(DECIMAL_POINTS_ETH))
                }
            }, {
                targets: 2,
                data: '2',
                "orderable": true,
                render: function (data, type, row, meta) {
                    return (parseFloat(data).toFixed(DECIMAL_POINTS_ETH))
                }
            }, {
                targets: 3,
                data: '3',
                "orderable": false,
                render: function (data, type, row, meta) {
                    return `${currency} ${addCommas(parseFloat(data).toFixed(DECIMAL_POINTS_CURRENCY))}`
                }
            }, {
                targets: 4,
                data: '4',
                "orderable": false,
                render: function (data, type, row, meta) {
                   return `${currency} ${addCommas(parseFloat(data).toFixed(DECIMAL_POINTS_CURRENCY))}`
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


function unSubUser(filter){
    // console.log(filter)
    fetch(`/user/rewards/unsubscribe?${filter}`, {
        method: 'POST',
        headers: {"X-CSRF-Token": csrfToken},
        credentials: 'include',
        body: "",
    }).then((res)=>{
        if (res.status == 200){
            res.json().then((data)=>{
                console.log(data.msg)
                window.location.reload(true)
            })              
        }
    })
}

function updateSubscriptionTable(data, container){
    if (data.length===0)return
    $('#'+container).DataTable({
        processing: true,
        serverSide: false,
        ordering: true,
        searching: true,
        pagingType: 'full_numbers',
        pageLength: 100,
        lengthChange: false,
        data: data,
        drawCallback: function(settings){
            $("#subscriptions-table-art").removeClass("d-flex").addClass("d-none")
            $("#subscriptions-table-div").removeClass("invisible")
        },
        columnDefs: [
            {
                targets: 0,
                data: '0',
                "orderable": true,
                render: function (data, type, row, meta) {
                    let date = data.split(" ")
                    if (date.length >=2){
                        return `${date[0]} ${date[1]}`
                    }
                    return data
                }
            }, {
                targets: 1,
                data: '1',
                "orderable": true,
                render: function (data, type, row, meta) {
                    return data.toUpperCase()
                }
            }, {
                targets: 2,
                data: '2',
                "orderable": false,
                render: function (data, type, row, meta) {
                    return `<textarea readonly style="height: 50px; width: 200px; overflow: auto; background-color: rgba(0, 0, 0, 0);" class="nice-scroll text-dark">${data}</textarea>`
                }
            }, {
                targets: 3,
                data: '3',
                "orderable": false,
                render: function (data, type, row, meta) {
                    return `
                        <div class="d-flex justify-content-center align-item-center">
                            <i class="fas fa-times text-danger" onClick='unSubUser("${data}")' style="cursor: pointer;"></i>
                        </div>
                        `
                }
            }]
    });
}

$(document).ready(function () {
    if (document.getElementsByName("CsrfField")[0]===undefined){
        console.error("Auth error")
    }else{
        csrfToken = document.getElementsByName("CsrfField")[0].value
    }

    if (localStorage.getItem("dashboard_validators").length){
        $('#validator-index-view').val(JSON.parse(localStorage.getItem("dashboard_validators")))
    }

    $('#validator-index-view').on("keyup", function(){
        $(this).val($(this).val().replace(/([a-zA-Z ])/g, ""))
    })


    $('input[id="datepicker"]').daterangepicker({
        pens: 'left',
        minDate: moment().subtract(365, 'days'), 
        maxDate: moment(),
        maxSpan: {
            'days': 365
        },
        ranges: {
            'This Month to date': [moment().startOf('month'), moment()],
            'Last Month to date': [moment().subtract(1, 'month').startOf('month'), moment()],
            'This Year to date': [moment().startOf('year'), moment()],
            'Last 365 days': [moment().subtract(365, 'days'), moment()],
         },
         locale: {
            format: 'DD/MM/YYYY'
        },
        singleDatePicker: true,
        alwaysShowCalendars: false
    }, function(start, end, label) {
        let end_d = moment()
        $("#days").val(end_d.diff(moment(start), 'days'))
    });
    
    create_typeahead('.typeahead-validators');
    let qry = getValidatorQueryString()
    // console.log(qry, qry.length)
    if (qry.length > 1){
        fetch(`/rewards/hist${qry}`,{
            method: "GET"
          }).then((res)=>{
              if (res.status !== 200){
                alert("Request failed :(")
                hideSpinner() 
              }
            res.json().then((data)=>{
              showTable(data)
            })
          }).catch(()=>{
            alert("Failed to fetch the data :(")
            hideSpinner()      
          })


        const urlParams = new URLSearchParams(window.location.search);
        const reqBody = JSON.stringify({validators: urlParams.get('validators'), currency: urlParams.get('currency')})
        // console.log(reqBody)
        if (urlParams.get('checkbox')==="on"){
            fetch(`/user/rewards/subscribe${qry}`, {
                method: 'POST',
                headers: {"X-CSRF-Token": csrfToken},
                credentials: 'include',
                body: reqBody,
            }).then((res)=>{
                if (res.status == 200){
                    res.json().then((data)=>{
                        console.log(data.msg)
                    })              
                }
            })
        }


    }else{
        hideSpinner()
    }
})