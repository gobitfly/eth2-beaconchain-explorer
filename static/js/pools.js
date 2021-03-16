function getActive(poolValidators){
    let active = 0
    for (let item of poolValidators){
        if (item.status==="active_online"){
            active++
        }
    }

    return [100.0 - ((1.0-(active/poolValidators.length))*100), `${active}/${poolValidators.length}`]
}

function getValidatorCard(val){
    if (val===undefined) return ""
    let bg = "danger"
    if (val.status==="active_online"){
        bg = "success"
    }
    return `
    <div class="col-sm-12 col-md-6 col-lg-4 col-xl-2 card shadow-sm p-2 m-1" style="min-height: 100px; max-height: 100px">
        <div class="d-flex flex-row justify-content-between">
            <a href="/validator/${val.validatorindex}"><i class="fas fa-male mr-1"></i> ${val.validatorindex}</a>
            <span class="text-${bg}">${val.status.replace("_", " ")}</span>
        </div>
        <hr/>
        <span data-toggle="tooltip" title="31 Day Balance">${(val.balance31d/1e9).toFixed(4)} ETH</span>
    </div>
    `
}

function showPoolInfo(data){
    $(".popupMain").off('scroll')
    $(".popupMain").html("")
    let data2Show = data
    let data2ShowOnScroll = []
    if (data.length > 100){
        data2Show = data.slice(0, 100)
        data2ShowOnScroll = data.slice(100, data.length-1)
    }
    for (let item of data2Show){
        $(".popupMain").append(getValidatorCard(item))
    }
    $("#poolPopUP").removeClass("d-none")
    $('html, body').animate({
        scrollTop: $("body").offset().top
    }, 1500);
    if (data2ShowOnScroll.length > 0){
        $(".popupMain").on('scroll', (e)=>{
            var elem = $(e.currentTarget);
            if (elem[0].scrollHeight - elem.scrollTop() <= elem.outerHeight()){
                for (let i = 0; i<100; i++){
                    $(".popupMain").append(getValidatorCard(data2ShowOnScroll.shift()))
                }
            }
        })
    }
}

function addHandlers(tableData){
    for (let item of tableData){
        $("#"+item[2]).on("click", ()=>{
            showPoolInfo(item[5])
        })
    }
}

function updateTableType(){
    if($(window).width() > 1444){
        $("#poolTable").addClass("table")
        $("#poolTable").removeClass("table-responsive")
    }else{
        $("#poolTable").removeClass("table")
        $("#poolTable").addClass("table-responsive")
    }
}

function randerTable(tableData){
    $('#staking-pool-table').DataTable({
        processing: true,
        serverSide: false,
        ordering: true,
        searching: false,
        pagingType: 'full_numbers',
        data: tableData,
        preDrawCallback: function() {
            try {
                $('#staking-pool-table').find('[data-toggle="tooltip"]').tooltip('dispose')
            } catch (e) {}
        },
        drawCallback: function(settings) {
            $('#staking-pool-table').find('[data-toggle="tooltip"]').tooltip()
            $(".hover-shadow").hover(function(){
                $(this).addClass("shadow");
                }, function(){
                $(this).removeClass("shadow");
            });
            addHandlers(tableData)
            updateTableType()
        },
        columnDefs: [
                {
                    targets: 0,
                    data: '0',
                    "orderable": true
                }, {
                    targets: 1,
                    data: '1',
                    "orderable": true
                }, {
                    targets: 2,
                    data: '2',
                    "orderable": false
                }, {
                    targets: 3,
                    data: '3',
                    "orderable": true,
                    render: function(data, type, row, meta) {
                        return data
                    }
                }, {
                    targets: 4,
                    data: '4',
                    "orderable": true,
                    render: function(data, type, row, meta) {
                        function getIncomeStats(){
                            return `
                                <div>
                                    <span>Last Day: ${(data.lastDay/1e9).toFixed(4)}</span>
                                    <span>Last Week: ${(data.lastWeek/1e9).toFixed(4)}</span>
                                    <span>Last Month: ${(data.lastMonth/1e9).toFixed(4)}</span>
                                </div>
                                `
                        }
                        return `<span data-toggle="tooltip" title="${getIncomeStats()}" data-html="true">
                                ${parseInt(data.total/1e9)}
                                </span>
                                `
                    }
                }, {
                    targets: 5,
                    data: '5',
                    render: function(data, type, row, meta) {
                        let info = getActive(data)
                        let bg = "bg-success"
                        let fg = "white"
                        if (parseInt(info[0]) < 60){
                            bg = "bg-danger"
                            fg = "black"
                        }
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
                }
            ],
        order: [[3,'desc']],
    })
}

function updateEthSupply(){
    let circulatingETH = parseInt(ETH_SUPPLY.result / 1e18);
    let stakedEth = STAKED_ETH;
    let eth = parseFloat(stakedEth.split(" ")[0].replace(",", "").replace(",", ""));
    let progress = ((eth/circulatingETH)*100).toFixed(2);
    $(".staked-progress").width(progress);
    $(".staked-progress, #staked-percent").html(`${progress}%`);
    $("#ethCsupply").html(circulatingETH.toString().replace(/,/g, "").replace(/\B(?=(\d{3})+(?!\d))/g, ",") +" ETH")
}

$(document).ready(function () {
    $(window).on("resize", ()=>{
        updateTableType()
    })
    $("#poolPopUpBtn").on("click", ()=>{$("#poolPopUP").addClass("d-none")})
    
    updateEthSupply()
    
    let tableData = []
    for (let el of POOL_INFO){
        tableData.push([el.name, el.category, el.address, parseInt(el.poolIncome.totalDeposits/1e9), el.poolIncome, el.poolInfo])
    }

    randerTable(tableData)
})