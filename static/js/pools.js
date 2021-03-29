let poolShare = {}
let totalValidators = 0

function getActive(poolValidators) {
    let active = 0
    let slashed = 0
    for (let item of poolValidators) {
        if (item.status === "active_online") {
            active++
        }else if (item.status === "slashed"){
            slashed++
        }
    }

    let view = `<i class="fas fa-male ${active>0?"text-success":""} fa-sm mr-1"></i> ${addCommas(active)} <i class="fas fa-user-slash ${slashed>0?"text-danger":""} fa-sm mx-1"></i> ${addCommas(slashed)} / ${addCommas(poolValidators.length)}`

    return [(active / poolValidators.length) * 100, view, slashed]
}

function getValidatorCard(val) {
    if (val === undefined) return ""
    let bg = "danger"
    if (val.status === "active_online") {
        bg = "success"
    }
    return `
    <div class="col-sm-12 col-md-6 col-lg-4 col-xl-2 card shadow-sm p-2 m-1" style="min-height: 100px; max-height: 100px">
        <div class="d-flex flex-row justify-content-between">
            <a href="/validator/${val.validatorindex}"><i class="fas fa-male mr-1"></i> ${val.validatorindex}</a>
            <span class="text-${bg}">${val.status.replace("_", " ")}</span>
        </div>
        <hr/>
        <span data-toggle="tooltip" title="31 Day Balance">${(val.balance31d / 1e9).toFixed(4)} ETH</span>
    </div>
    `
}

function showPoolInfo(data) {
    $(".popupMain").html("")
    let data2Show = []
    for (let d of data) {
        if (d.status === "active_online") {
            continue
        }
        data2Show.push(d)
    }
    let data2ShowOnScroll = []
    if (data2Show.length > 100) {
        data2ShowOnScroll = data2Show.slice(100, data.length - 1)
        data2Show = data2Show.slice(0, 100)
    } else if (data2Show.length === 0) {
        $(".popupMain").html(`All Validators are active <i class="fas fa-rocket"></i>`)
    }

    for (let item of data2Show) {
        $(".popupMain").append(getValidatorCard(item))
    }

    $("#poolPopUP").removeClass("d-none")
    $('html, body').animate({
        scrollTop: $("body").offset().top
    }, 1500);
    if (data2ShowOnScroll.length > 0) {
        $(".popupMain").off('scroll')
        $(".popupMain").on('scroll', (e) => {
            var elem = $(e.currentTarget);
            // console.log((elem[0].scrollHeight - elem.scrollTop() <= elem.outerHeight()), elem[0].scrollHeight, elem.scrollTop(), elem.outerHeight())
            if ((elem[0].scrollHeight - elem.scrollTop()) - 10 <= elem.outerHeight()) {
                for (let i = 0; i < 100; i++) {
                    $(".popupMain").append(getValidatorCard(data2ShowOnScroll.shift()))
                }
            }
        })
    }
}

function addHandlers(tableData) {
    for (let item of Object.keys(tableData)) {
        if (isNaN(parseInt(item, 10))) continue

        $("#" + tableData[item][2]).off("click")
        $("#" + tableData[item][2]).on("click", () => {
            showPoolInfo(tableData[item][6])
        })
        getPoolEffectiveness(tableData[item][2] + "eff", tableData[item][6])
        getAvgCurrentStreak(tableData[item][2])
    }
}

function updateTableType() {
    $("#staking-pool-table_wrapper div.row:last").addClass("mt-4")
}

function randerTable(tableData) {
    $('#staking-pool-table').DataTable({
        processing: true,
        serverSide: false,
        ordering: true,
        searching: true,
        pagingType: 'full_numbers',
        data: tableData,
        lengthMenu: [10, 25],
        preDrawCallback: function () {
            try {
                $('#staking-pool-table').find('[data-toggle="tooltip"]').tooltip('dispose')
            } catch (e) { }
        },
        drawCallback: function (settings) {
            $('#staking-pool-table').find('[data-toggle="tooltip"]').tooltip()
            $(".hover-shadow").hover(function () {
                $(this).addClass("shadow");
            }, function () {
                $(this).removeClass("shadow");
            });
            let api = this.api();
            let curData = api.rows({ page: 'current' }).data()
            // console.log(curData, typeof(curData), Object.keys(curData) ,api.row(this).data());
            addHandlers(curData)
            updateTableType()

        },
        columnDefs: [
            {
                targets: 0,
                data: '0',
                "orderable": true,
                render: function (data, type, row, meta) {
                    if (data===""){return "Unknown"}
                    return data
                }
            }, {
                targets: 1,
                data: '1',
                "orderable": true,
                render: function (data, type, row, meta) {
                    if (data==="" || data===null){return "Unknown"}
                    return data
                }
            }, {
                targets: 2,
                data: '2',
                "orderable": false,
                render: function (data, type, row, meta) {
                    return `<a href="/validators/eth1deposits?q=0x${data}">0x${data.slice(0, 7)}...</a>`
                }
            }, {
                targets: 3,
                data: '3',
                "orderable": true,
                render: function (data, type, row, meta) {
                    let val = parseFloat((poolShare[data]/totalValidators)*100).toFixed(2)
                    
                    if(type === 'display') {
                        if (isNaN(val)) return "Unknown"
                        return `${parseFloat((poolShare[data]/totalValidators)*100).toFixed(2)}%`
                    }

                    if (isNaN(val)) return 0
                    return poolShare[data]
                }
            },{
                targets: 4,
                data: '4',
                "orderable": true,
                render: function (data, type, row, meta) {
                    if(type === 'display') {
                        return addCommas(data)
                    }

                    return data
                }
            }, {
                targets: 5,
                data: '5',
                "orderable": true,
                render: function (data, type, row, meta) {
                    function getIncomeStats() {
                        return `
                                Last Day: ${addCommas(parseInt(data.lastDay / 1e9))}
                                Last Week: ${addCommas(parseInt(data.lastWeek / 1e9))}
                                Last Month: ${addCommas(parseInt(data.lastMonth / 1e9))}
                                `
                    }
                    if(type === 'display') {
                        return `<span data-toggle="tooltip" title="${getIncomeStats()}" data-html="true">
                                ${addCommas(parseInt(data.total / 1e9))}
                                </span>
                                `
                    }

                    return parseInt(data.total / 1e9)
                }
            }, {
                targets: 6,
                data: '6',
                "orderable": true,
                render: function (data, type, row, meta) {
                    let info = getActive(data)
                    let bg = "bg-success"
                    let fg = "white"
                    if (parseInt(info[0]) < 60) {
                        bg = "bg-danger"
                        fg = "black"
                    }
                    if(type === 'display') {
                        return `
                            <div id="${row[2]}" style="cursor: pointer;" class="d-flex flex-column hover-shadow" style="height: 100%;" 
                                                data-toggle="tooltip" title="${info[0].toFixed(2)}% of validators are active in this pool">
                                <div class="d-flex justify-content-center">
                                    ${info[1]}
                                </div>
                                <div class="progress" style="height: 3px;">    
                                    <div class="progress-bar progress-bar-success ${bg}" 
                                        role="progressbar" aria-valuenow="${parseInt(info[0])}"
                                        aria-valuemin="0" aria-valuemax="100" style="width: ${parseInt(info[0])}%; color: ${fg};" >
                                    </div>
                                </div>
                            </div>
                            `
                    }
                    return info[2]
                }
            }, {
                targets: 7,
                data: '7',
                "orderable": false,
                render: function (data, type, row, meta) {
                    return `
                            <div id="${data}eff" data-toggle="tooltip" data-original-title="Average Attestation Eff. of Top 200 validators (highest balance)">
                                <div class="spinner-grow spinner-grow-sm text-primary" role="status">
                                    <span class="sr-only">Loading...</span>
                                </div>
                            </div>
                        `
                }

            }, {
                targets: 8,
                data: '8',
                "orderable": false,
                render: function (data, type, row, meta) {
                    return `
                            <div id="${data}streak">
                                <div class="spinner-grow spinner-grow-sm text-primary" role="status">
                                    <span class="sr-only">Loading...</span>
                                </div>
                            </div>
                        `
                }

            }
        ],
        order: [[4, 'desc']],
    })
}

function addCommas(number) {
    return number.toString().replace(/,/g, "").replace(/\B(?=(\d{3})+(?!\d))/g, ",")
}

function updateEthSupply() {
    let circulatingETH = parseInt(ETH_SUPPLY.result / 1e18);
    let stakedEth = STAKED_ETH;
    let eth = parseFloat(stakedEth.split(" ")[0].replace(",", "").replace(",", ""));
    let progress = ((eth / circulatingETH) * 100).toFixed(2);
    $(".staked-progress").width(progress);
    $("#staked-percent").html(`${progress}%`);
    $("#ethCsupply").html(addCommas(circulatingETH) + " ETH")
}

function getPoolEffectiveness(id, data) {
    let load = async () => {
        let query = "?validators="
        for (let item of data.slice(0, 200)) {
            query += item.validatorindex + ","
        }
        query = query.substring(0, query.length - 1)
        // console.log(query)
        let resp = await fetch(`/dashboard/data/effectiveness${query}`)
        resp = await resp.json()
        let eff = 0.0
        for (let incDistance of resp) {
            if (incDistance === 0.0) {
                continue
            }
            eff += (1.0 / incDistance) * 100.0
        }
        eff = eff / resp.length
        setValidatorEffectiveness(id, eff)
    }

    load().then(() => { })
}

function getAvgCurrentStreak(id) {
    fetch("/pools/streak/current?pool=0x" + id)
        .then((resp) => {
            resp.json()
                .then((data) => {
                    $(`#${id}streak`).html(addCommas(parseInt(data)))
                })
        }).catch((err) => {
            console.log(err)
            $(`#${id}streak`).html("N/a")
        })
}

function updatePoolShare (arr){
    // console.log(arr)
    for (let item of arr){
        if (typeof item === 'object' && 'data' in item){
            updatePoolShare(item.data)
        }else {
            poolShare[item[0]] = item[1]
            totalValidators += item[1]
        }
    }
}

$(document).ready(function () {
    // $(window).on("resize", () => {
    //     updateTableType()
    // })
    $("#poolPopUpBtn").on("click", () => { $("#poolPopUP").addClass("d-none") })

    updateEthSupply()

    updatePoolShare(drill.series)

    let tableData = []
    for (let el of POOL_INFO) {
        tableData.push([el.name, el.category, el.address, el.name,
        parseInt(el.poolIncome.totalDeposits / 1e9),
        el.poolIncome, el.poolInfo,
        el.address, el.address])
    }

    randerTable(tableData)

})