let poolShare = {}
let poolchart = null
let totalDeposited = 0,
    totalIncome = 0,
    totalIperEth = 0;
var IDETH_SERIES = {}
var totalValidatorsPI = 0
var poolInfoTable = null
function getActive(poolValidators) {
    let active = 0
    let slashed = 0
    let pending = 0
    for (let item of poolValidators) {
        if (item.status === "active_online") {
            active++
        } else if (item.status === "slashed") {
            slashed++
        } else if (item.status === "pending") {
            pending++
        }
    }

    return [(active / poolValidators.length) * 100, {
        active: `<span style="font-size: 10px;" class="${active > 0 ? "text-success" : ''}"><i class="fas fa-male  mr-1"></i>${addCommas(active)}</span>`,
        slashed: `<span style="font-size: 10px;" class="${slashed > 0 ? "text-danger" : ''}"><i class="fas fa-user-slash fa-sm mx-1"></i>${addCommas(slashed)}</span>`,
        pending: `<span style="font-size: 10px;" class="${pending > 0 ? "text-info" : ''}"><i class="fas fa-male mr-1"></i>${addCommas(pending)}</span>`,
        total: `${addCommas(poolValidators.length)}`
    },
        slashed]
}

function updatePoolShare(arr) {
    // console.log(arr)
    for (let item of arr) {
        if (typeof item === 'object' && 'data' in item) {
            updatePoolShare(item.data)
        } else {
            poolShare[item[0]] = item[1]
            totalValidatorsPI += item[1]
        }
    }
}

function getValidatorCard(val) {
    if (val === undefined) return ""
    let color = ""
        note = `Activated on epoch <a href="/epoch/${val.activationepoch}">${val.activationepoch}</a>` 
        animate = ""
        tooltip = ''
    if (val.status === "slashed") {
        color = "text-danger"
        note = `Exit epoch <a href="/epoch/${val.exitepoch}">${val.exitepoch}</a>`
    }else if(val.status==="pending"){
        // color = "dark"
        note = `Will activate on epoch ${val.activationepoch}`
        animate = "fade-in-out"
    }else if (val.status==="active_offline"){
        tooltip = `data-toggle="tooltip" title="Validator is considered offline if there where no attestations in the last two epochs"`
    }
    return `
    <div class="col-sm-12 col-md-6 col-lg-4 col-xl-2 card shadow-sm p-2 m-1" style="min-height: 100px; max-height: 120px">
        <div class="d-flex flex-row justify-content-between">
            <a href="/validator/${val.validatorindex}"><i class="fas fa-male mr-1"></i> ${val.validatorindex}</a>
            <span ${tooltip} class="${color} ${animate}">${val.status.replace("_", " ").replace("offline", '<i class="fas fa-power-off fa-sm text-danger"></i>')}</span>
        </div>
        <hr/>
        <div class="d-flex flex-row justify-content-between">
            <span data-toggle="tooltip" title="31 Day Balance" style="font-size: 12px;">${(val.balance31d / 1e9).toFixed(4)} ETH</span>
            <span style="font-size: 12px;">${note}</span>
        </div>
    </div>
    `
}

function showPoolInfo(poolName, data) {
    $(".popupMain").html('')
    $("#poolPopUpTitle").html(`Displaying inactive validators of ${poolName}`)
    let data2Show = []
    for (let d of data) {
        if (d.status === "active_online") {
            continue
        }
        data2Show.push(d)
    }
    let data2ShowOnScroll = []
    if (data2Show.length > 100) {
        data2ShowOnScroll = data2Show.slice(100, data2Show.length - 1)
        data2Show = data2Show.slice(0, 100)
    } else if (data2Show.length === 0) {
        $(".popupMain").html(`All Validators are active <i class="fas fa-rocket"></i>`)
    }
    // console.log(data2Show[0], "data2Show")
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
    // console.log(tableData)
    for (let item of Object.keys(tableData)) {
        if (isNaN(parseInt(item, 10))) continue

        $("#" + tableData[item][2]).off("click")
        $("#" + tableData[item][2]).on("click", () => {
            showPoolInfo(tableData[item][0], tableData[item][7])
        })
        getPoolEffectiveness(tableData[item][2] + "eff", tableData[item][7])
        // getAvgCurrentStreak(tableData[item][2])
    }
}

function makeTotalVisisble(id) {
    $("#" + id).removeClass("d-none")
    $("#" + id).addClass("tableTotalTop shadow")

}

function updateTableType() {
    $("#staking-pool-table_wrapper div.row:last").addClass("mt-4")
    $("#tableDepositTotal").html(addCommas(totalDeposited))
    $("#tableIncomeTotal").html(addCommas(totalIncome))
    $("#tableIpDTotal").html((totalIperEth / POOL_INFO.length).toFixed(5))
    $("#tableValidatorsTotal").html(addCommas(TOTAL_VALIDATORS))
}

function updateEthSupply() {
    let circulatingETH = parseInt(ETH_SUPPLY.result / 1e18);
    let stakedEth = STAKED_ETH;
    let eth = parseFloat(stakedEth.split(" ")[0].replace(",", "").replace(",", ""));
    totalDeposited = eth
    let progress = ((eth / circulatingETH) * 100).toFixed(2);
    $(".staked-progress").width(progress);
    $("#staked-percent").html(`${progress}%`);
    $("#ethCsupply").html(addCommas(circulatingETH) + " ETH")
}

function getPoolEffectiveness(id, data) {
    let load = async () => {
        let query = "?validators="
        for (let item of data.slice(0, 100)) {
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

function randerTable(tableData) {
    poolInfoTable = $('#staking-pool-table').DataTable({
        processing: true,
        serverSide: false,
        ordering: true,
        searching: true,
        pagingType: 'first_last_numbers',
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
                    if (data === "") { return "Unknown" }
                    return data
                }
            }, {
                targets: 1,
                data: '1',
                "orderable": true,
                render: function (data, type, row, meta) {
                    if (data === "" || data === null) { return "Unknown" }
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
                    let val = ((parseFloat(poolShare[data]) / totalValidatorsPI) * 100).toFixed(3)

                    if (type === 'display') {
                        if (isNaN(val)) return "0.00%"
                        return `${val}%`
                    }

                    if (isNaN(val)) return 0
                    return poolShare[data]
                }
            }, {
                targets: 4,
                data: '4',
                "orderable": true,
                render: function (data, type, row, meta) {
                    if (type === 'display') {
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
                    if (type === 'display') {
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
                    let ipd = parseInt(data.earningsInPeriod) / parseInt(data.earningsInPeriodBalance)
                    if (isNaN(ipd)) {
                        ipd = 0
                    }
                    if (type === 'display') {
                        return `<span data-toggle="tooltip" title="Calculated based on active validators between epochs ${data.epochStart} <-> ${data.epochEnd}. 
                                                                    Total income of selected validators in this period was ~${addCommas((parseInt(data.earningsInPeriod) / 1e9).toFixed(3))} ETH and total balance was ~${addCommas((parseInt(data.earningsInPeriodBalance) / 1e9).toFixed(1))} ETH">
                            ${parseFloat(ipd).toFixed(5)}
                        </span>`
                    }

                    return ipd
                }
            }, {
                targets: 7,
                data: '7',
                "orderable": true,
                render: function (data, type, row, meta) {
                    let info = getActive(data)
                    let bg = "bg-success"
                    let fg = "white"
                    if (parseInt(info[0]) < 60) {
                        bg = "bg-danger"
                        fg = "black"
                    }
                    if (type === 'display') {
                        return `
                            <div id="${row[2]}" style="cursor: pointer; border-radius: 10px; border-style: none;" class="d-flex flex-column hover-shadow" style="height: 100%;" 
                                                data-toggle="tooltip" title="${info[0].toFixed(2)}% of validators are active in this pool">
                                <div class="d-flex justify-content-between">
                                    ${info[1].active} ${info[1].slashed}
                                </div>
                                <div class="progress" style="height: 3px;">    
                                    <div class="progress-bar progress-bar-success ${bg}" 
                                        role="progressbar" aria-valuenow="${parseInt(info[0])}"
                                        aria-valuemin="0" aria-valuemax="100" style="width: ${parseInt(info[0])}%; color: ${fg};" >
                                    </div>
                                </div>
                                <div class="d-flex justify-content-center">
                                    <span style="font-size: 10px;">${info[1].total}</span>
                                </div>
                            </div>
                            `
                    }
                    return info[2]
                }
            }, {
                targets: 8,
                data: '8',
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
                targets: 9,
                data: '9',
                "orderable": false,
                visible: false,
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
        order: [[5, 'desc']],
    })
}

function showChartSwitch(chart) {
    $("#uncheckAllSeriesbtn").remove()
    $("#returnOriginalSeriesbtn").remove()
    chart.renderer.text('<i id="uncheckAllSeriesbtn" style="cursor: pointer; font-size: 20px;" class="fas fa-eye-slash"></i>',
        chart.chartWidth - 50, 22, true)
        .attr({ zIndex: 3 })
        .on('click', function () {
            let option = $('#uncheckAllSeriesbtn').hasClass("text-primary")
            let series = chart.series;
            for (i = 0; i < chart.series.length; i++) {
                series[i].setVisible(option, option);
            }
            chart.redraw();

            if (option) {
                $('#uncheckAllSeriesbtn').removeClass("text-primary")
            } else {
                $('#uncheckAllSeriesbtn').addClass("text-primary")
            }
        })
        .add();

    chart.renderer.text('<i id="returnOriginalSeriesbtn" style="cursor: pointer; font-size: 20px;" class="fas fa-long-arrow-alt-left"></i>',
        chart.chartWidth - 80, 22, true)
        .attr({ zIndex: 3 })
        .on('click', function () {
            // let option = $('#uncheckAllSeries').hasClass("text-primary")
            updateChartSeries(IDETH_SERIES.mainSeries, null)
            $("#returnOriginalSeriesbtn").removeClass("text-primary")
            $('#uncheckAllSeriesbtn').removeClass("text-primary")
        })
        .add();

    if (parseInt(localStorage.getItem("chartWelcomeAnimatoin")) !== 1) {
        switchCharts()
        setTimeout(() => {
            switchCharts()
            localStorage.setItem("chartWelcomeAnimatoin", 1)
        }, 2000)
    }
}

function updateChartSeries(pseries, name) {
    while (poolchart.series.length > 0) {
        poolchart.series[0].remove(false);
    }

    for (let item of pseries) {
        if (item.name.includes(name) || name === null) {
            // console.log(item.name, name)
            poolchart.addSeries(item)
        }
    }

    poolchart.redraw()
}

function randerChart(dataSeries) {
    // console.log(dataSeries)
    poolchart = Highcharts.chart('poolsIDChart', {
        chart: {
            height: 500,
            type: 'line',
            zoomType: 'y'
        },
        title: {
            text: 'Income Per Deposited ETH',
            x: -20, //center
            useHTML: true
        },
        xAxis: {
            title: {
                text: 'EPOCH'
            }
        },
        yAxis: {
            title: {
                text: 'ETH'
            },
            labels: {
                format: '{value:.5f}'
            }
        },
        tooltip: {
            animation: true,
            shared: true,
            useHTML: true,
            formatter: function (tooltip) {
                return this.points.reduce(function (s, point) {
                    return s + `<tr>
                                    <td><span style="color: ${point.series.color};">\u25CF</span></td>
                                    <td>${point.series.name}</td> 
                                    <td><b>${point.y.toFixed(5)}</b></td> 
                                    <td>ETH</td>
                                </tr>`;
                }, `<div style="font-weight:bold; text-align:center;">${this.x}</div><table>`) + '</table>';
            },
        },
        legend: {
            layout: 'vertical',
            align: 'right',
            verticalAlign: 'middle',
            borderWidth: 0,
            showInLegend: false,
            useHTML: true,
            navigation: {
                activeColor: 'var(--primary)',
                animation: true,
                arrowSize: 12,
                inactiveColor: '#CCC',
                style: {
                    fontWeight: 'bold',
                    color: 'var(--dark)',
                    fontSize: '12px'
                }
            }
        },
        series: dataSeries.mainSeries,
        plotOptions: {
            series: {
                cursor: 'pointer',
                events: {
                    click: function (event) {
                        updateChartSeries(dataSeries.drillSeries, this.name)
                        $("#returnOriginalSeriesbtn").addClass("text-primary")
                    }
                }
            }
        },
        responsive: {
            rules: [{
                condition: {
                    maxWidth: 500
                },
                chartOptions: {
                    legend: {
                        align: 'center',
                        verticalAlign: 'bottom',
                        layout: 'horizontal'
                    },
                    yAxis: {
                        labels: {
                            align: 'left',
                            x: 0,
                            y: -5
                        },
                        title: {
                            text: null
                        }
                    },
                    subtitle: {
                        text: null
                    },
                    credits: {
                        enabled: false
                    }
                }
            }]
        }
    }, showChartSwitch)
}

function switchCharts() {
    if ($(".chart-pi").hasClass("d-none")) {
        $(".chart-line").addClass("d-none")
        $(".chart-pi").removeClass("d-none")
        $("div.chart-switch-btn i:first").addClass("text-primary")
        $("div.chart-switch-btn i:first").removeClass("text-dark")
        $("div.chart-switch-btn i:last").addClass("text-dark")
        $("div.chart-switch-btn i:last").removeClass("text-primary")
    } else if ($(".chart-line").hasClass("d-none")) {
        $(".chart-pi").addClass("d-none")
        $(".chart-line").removeClass("d-none")
        $("div.chart-switch-btn i:last").addClass("text-primary")
        $("div.chart-switch-btn i:last").removeClass("text-dark")
        $("div.chart-switch-btn i:first").addClass("text-dark")
        $("div.chart-switch-btn i:first").removeClass("text-primary")
    }
}

$(document).ready(function () {
    $("#poolPopUpBtn").on("click", () => { $("#poolPopUP").addClass("d-none") })

    updatePoolShare(drill.series)

    updateEthSupply()


    let tableData = []
    for (let el of POOL_INFO) {
        // totalDeposited += parseInt(el.poolIncome.totalDeposits / 1e9)
        totalIncome += parseInt(el.poolIncome.total / 1e9)
        let ipd = parseInt(el.poolIncome.earningsInPeriod) / parseInt(el.poolIncome.earningsInPeriodBalance)
        isNaN(ipd) ? ipd = 0 : ipd;
        totalIperEth += ipd

        if (el.name === "" && IS_MAINNET) continue;

        tableData.push([el.name, el.category, el.address, el.name,
        parseInt(el.poolIncome.totalDeposits / 1e9), el.poolIncome,
        el.poolIncome,
        el.poolInfo, el.address, el.address])
    }

    randerTable(tableData)
    // console.log(tableData[0], "tableData")

    fetch(`/pools/chart/income_per_eth`, {
        method: "GET"
    }).then((res) => {
        if (res.status !== 200) {
            alert("Chart Request failed :(")
        }
        res.json().then((data) => {
            IDETH_SERIES = data
            randerChart(IDETH_SERIES)
        })
    }).catch(() => {
        alert("Failed to fetch the chart data :(")
    })

    $(".chart-switch-btn").on("click", () => {
        switchCharts()
    })

    $("#totalmsg").html(`"Total Income" and "Average Income Per Deposited ETH" are based on top 100 pools by number of validators`)
    $(window).on('resize', function () {
        showChartSwitch(poolchart)
    })
})