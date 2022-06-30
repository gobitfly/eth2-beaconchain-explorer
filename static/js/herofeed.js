let number_of_doners = 0
let feedInterval = null
let loadedLocals = false
let donors = null
const numEtries = 16

function showDoner(addr, name, icon, msg) {
  if (msg === "" || msg === null || msg.includes("<") || msg.includes(">")) {
    msg = `<span style="font-style: italic;">Donated to beaconcha.in \u2764</span>`
  }
  let fullmsg = ""
  if (msg.length > 120) {
    fullmsg = msg
    msg = msg.slice(0, 120) + "..."
  }

  name = name.replace("<", "").replace(">", "")
  $("#hero-feed ul").prepend(`<li class="fade-in hover-shadow">
        <div class="d-flex flex-row" style="cursor: pointer;" onclick='window.open("https://gitcoin.co/grants/258/beaconchain-open-source-eth2-blockchain-explorer")'>
        <img src="${icon}"/> 
        <div class="d-flex flex-column usrdiv">
            <div class="d-flex flex-row">
            <span>${name}</span>
            </div>
            <span class="umsg" data-toggle="tooltip" title="${fullmsg}" >${msg}</span>
        </div>
        </div>
        </li>`)
}

function isDonerNew(donner) {
  if (donors !== null) {
    for (let oldItem of donors) {
      if (donner[0] === oldItem[0] && donner[1] === oldItem[1] && donner[2] === oldItem[2] && donner[3] === oldItem[3]) {
        return false
      }
    }
  }
  return true
}

function findNewDoner(data) {
  if (data.length > numEtries) {
    data = data.slice(0, numEtries)
  }

  for (let i = data.length - 1; i >= 0; i--) {
    if (isDonerNew(data[i])) {
      showDoner(data[i][0], data[i][1], data[i][2], data[i][3])
      number_of_doners++
      if (number_of_doners > numEtries) {
        $("#hero-feed ul>li:last").remove()
      }
    }
  }

  return data
}

function updateFeed() {
  $.ajax({
    url: "/gitcoinfeed",
    success: (data) => {
      let isLive = data.isLive
      if (!isLive && feedInterval !== null) {
        clearInterval(feedInterval)
        return
      }
      data = data.donors

      if (isLive) {
        $("#hero-feed").addClass("d-lg-flex fade-in-top")

        if (data.length > 0) {
          donors = findNewDoner(data)
          $("#hero-feed ul>li#gitcoinwaitmsg").remove()
          $(".hover-shadow").hover(
            function () {
              $(this).addClass("shadow color-shift-anim border-rounded")
            },
            function () {
              $(this).removeClass("shadow color-shift-anim border-rounded fade-in")
            }
          )
        } else {
          $("#hero-feed ul").html("")
          $("#hero-feed ul").prepend(`
                        <li id="gitcoinwaitmsg"><i class="far fa-clock mx-1"></i><span>Waiting for the next gitcoin round to start</span></li>
                    `)
        }
      }
    },
  })
}

$(document).ready(function () {
  updateFeed()
  feedInterval = setInterval(() => {
    if (document.hasFocus()) {
      updateFeed()
    }
  }, 2000)

  $(".donate-btn").on("click", () => {
    window.open("https://gitcoin.co/grants/258/beaconchain-open-source-eth2-blockchain-explorer")
  })
})
