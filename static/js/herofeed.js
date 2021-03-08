let number_of_doners = 0
let feedInterval = null
let loadedLocals = false

function showDoner(addr, name, icon, msg) {
    if (msg === "" || msg === null || msg.includes("<") || msg.includes(">")) {
        msg = `<span style="font-style: italic;">Donated to beaconcha.in \u2764</span>`
    }
    let fullmsg = ""
    if (msg.length > 120) {
        fullmsg = msg
        msg = msg.slice(0, 120) + "..."
    }
    // msg = msg.replace("<", "^").replace("<", "^")
    name = name.replace("<", "").replace(">", "")
    $("#hero-feed ul").prepend(`<li class="fade-in">
        <div class="d-flex flex-row">
        <img src="${icon}" lowsrc="/img/logo.png"/> 
        <div class="d-flex flex-column usrdiv">
            <div class="d-flex flex-row">
            <span style="color: white;">${name}</span>
            </div>
            <span class="umsg" style="color: white;" data-toggle="tooltip" title="${fullmsg}" >${msg}</span>
        </div>
        </div>
        </li>`)
}

function showLocallyStoredDoners() {
    let donors = JSON.parse(localStorage.getItem("donors"))
    // console.log(donors)
    if (donors !== null) {
        for (let i = 1; i < donors.length; i++) {
            let index = donors.length - i
            // console.log(index, donors[index], donors[0])
            showDoner(donors[index][0], donors[index][1], donors[index][2], donors[index][3])
            number_of_doners++
        }
    }
    loadedLocals = true
}

function isDonerNew(donner) {
    let donors = JSON.parse(localStorage.getItem("donors"))
    if (donors !== null) {
        for (let oldItem of donors) {
            if (donner[0] === oldItem[0] && donner[1] === oldItem[1] && donner[2] === oldItem[2] && donner[3] === oldItem[3]) {
                return false
            }
        }
    }
    return true
}

function updateFeed() {
    $.ajax({
        url: "/gitcoinfeed",
        success: (data) => {
            let isLive = data.isLive
            if (!isLive && feedInterval !== null) {
                clearInterval(feedInterval)
            }
            data = data.donors
            if (data.length > 0) {
                if (!loadedLocals) showLocallyStoredDoners()
                $(".hero-image svg").addClass("hero-bg-blur")
                $("#hero-feed").addClass("d-lg-flex fade-in-top")
                for (let item of data) {
                    if (isDonerNew(item)) {
                        showDoner(item[0], item[1], item[2], item[3])
                        number_of_doners++
                        if (number_of_doners > 10) {
                            $("#hero-feed ul>li:last").remove()
                        }
                    }
                }

                localStorage.setItem("donors", JSON.stringify(data))
            }
        }
    })
}

$(document).ready(function () {
    updateFeed()
    feedInterval = setInterval(() => {
        updateFeed()
    }, 2000)

    $(".donate-btn").on("click", () => {
        window.open("https://gitcoin.co/grants/258/beaconchain-open-source-eth2-blockchain-explorer")
    })
})