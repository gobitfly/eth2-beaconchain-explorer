let tableData = []

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
        console.log("data2Show len", data2Show.length, data2ShowOnScroll.length, data2Show, data2ShowOnScroll[0])
    }
    for (let item of data2Show){
        $(".popupMain").append(getValidatorCard(item))
    }
    $("#poolPopUP").removeClass("d-none")
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

function addHandlers(){
    for (let item of tableData){
        $("#"+item[2]).on("click", ()=>{
            // showPoolPopUP(item[2])
            showPoolInfo(item[4])
        })
    }
}

$(document).ready(function () {
    $("#poolPopUpBtn").on("click", ()=>{$("#poolPopUP").addClass("d-none")})
    // $(".popupMain").bind('scroll', detectBottom);
    fetch("https://api.etherscan.io/api?module=stats&action=ethsupply&apikey=")
    .then(res => res.json())
    .then(data => {
            let circulatingETH = parseInt(data.result / 1e18);
            let stakedEth = STAKED_ETH;
            let eth = parseFloat(stakedEth.split(" ")[0].replace(",", "").replace(",", ""));
            let progress = ((eth/circulatingETH)*100).toFixed(2);
            $(".staked-progress").width(progress);
            $(".staked-progress, #staked-percent").html(`${progress}%`);
            $("#ethCsupply").html(circulatingETH.toString().replace(/,/g, "").replace(/\B(?=(\d{3})+(?!\d))/g, ",") +" ETH")
        })
    .catch((error) => {
        alert("Page may not function properly if you are using adblock or other types of communication blocking software")
    });
    
    
    for (let el of POOL_INFO){
        tableData.push([el.name, el.category, el.address, el.deposit, el.poolInfo])
    }

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
            addHandlers()
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
                    "orderable": true
                }, {
                    targets: 4,
                    data: '4',
                    render: function(data, type, row, meta) {
                        // console.log(type, row, meta)
                        let info = getActive(data)
                        return `<span data-toggle="tooltip" title="${info[0].toFixed(2)}% of validators are active in this pool"
                                    style="cursor: pointer;" class="hover-shadow p-1"
                                    id="${row[2]}">${info[1]}</span>`
                    }
                }
            ],
        order: [[3,'desc']],
    })
})