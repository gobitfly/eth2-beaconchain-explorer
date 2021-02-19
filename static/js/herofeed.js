let number_of_doners = 0

function showDoner(addr, name, icon, msg) {
    if (msg === "" || msg === null) {
        msg = `<span style="font-style: italic;">Donated to beaconcha.in \u2764</span>`
    }
    if (msg.length > 120) {
        msg = msg.slice(0, 120) + "..."
    }
    $("#hero-feed ul").prepend(`<li class="fade-in">
        <div class="d-flex flex-row">
        <img src=${icon}"/> 
        <div class="d-flex flex-column usrdiv">
            <div class="d-flex flex-row">
            <span>${name}</span>
            <!--(<span class="pkey">${addr}</span>)-->
            </div>
            <span class="umsg">${msg}</span>
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

$(document).ready(function () {
    showLocallyStoredDoners()
    setInterval(() => {
        $.ajax({
            url: "/gitcoinfeed",
            success: (data) => {
                if (data.length > 0) {
                    $("#hero-feed").addClass("d-lg-flex")
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
    }, 2000)

    $(".donate-btn").on("click", () => {
        window.open("https://gitcoin.co/grants/258/beaconchain-open-source-eth2-blockchain-explorer")
    })
})