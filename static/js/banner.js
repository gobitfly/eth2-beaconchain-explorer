var bannerContainer = document.querySelector(".info-banner-container")
var bannerSearch = document.querySelector(".info-banner-search")
var bannerSearchIcon = document.getElementById("banner-search")
var bannerSearchInput = document.getElementById("banner-search-input")

bannerSearch.addEventListener("click", function () {
  bannerContainer.classList.add("searching")
  bannerSearchInput.focus()
})

bannerSearchInput.addEventListener("blur", function () {
  bannerContainer.classList.remove("searching")
})

function getCookie(cname) {
  var name = cname + "="
  var ca = document.cookie.split(";")
  for (var i = 0; i < ca.length; i++) {
    var c = ca[i]
    while (c.charAt(0) == " ") {
      c = c.substring(1)
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length)
    }
  }
  return ""
}

function updateBanner() {
  fetch("/latestState")
    .then(function (res) {
      return res.json()
    })
    .then(function (data) {
      // always visible
      var epochHandle = document.getElementById("banner-epoch-data")

      if (data.currentEpoch) epochHandle.innerHTML = addCommas(data.currentEpoch)

      var ethPriceHandle = document.getElementById("banner-eth-price-data")

      let userCurrency = getCookie("currency")

      switch (userCurrency) {
        case "AUD":
          if (data.audRoundPrice && data.audTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>A$ </span>" + "<span class='k-formatted-price'>" + data.audTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.audRoundPrice) + "</span>")
          }
          break
        case "CAD":
          if (data.cadRoundPrice && data.cadTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>C$ </span>" + "<span class='k-formatted-price'>" + data.cadTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.cadRoundPrice) + "</span>")
          }
          break
        case "CNY":
          if (data.cnyRoundPrice && data.cnyTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>¥ </span>" + "<span class='k-formatted-price'>" + data.cnyTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.cnyRoundPrice) + "</span>")
          }
          break
        case "EUR":
          if (data.eurRoundPrice && data.eurTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>€ </span>" + "<span class='k-formatted-price'>" + data.eurTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.eurRoundPrice) + "</span>")
          }
          break
        case "GBP":
          if (data.gbpRoundPrice && data.gbpTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>£ </span>" + "<span class='k-formatted-price'>" + data.gbpTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.gbpRoundPrice) + "</span>")
          }
          break
        case "JPY":
          if (data.jpyRoundPrice && data.jpyTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>¥ </span>" + "<span class='k-formatted-price'>" + data.jpyTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.jpyRoundPrice) + "</span>")
          }
          break
        case "RUB":
          if (data.rubRoundPrice && data.rubTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>₽ </span>" + "<span class='k-formatted-price'>" + data.rubTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.rubRoundPrice) + "</span>")
          }
          break
        default:
          if (data.usdRoundPrice && data.usdTruncPrice) {
            return (ethPriceHandle.innerHTML = "<span class='currency-symbol'>$ </span>" + "<span class='k-formatted-price'>" + data.usdTruncPrice + "</span>" + "<span class='price'>" + addCommas(data.usdRoundPrice) + "</span>")
          }
      }

      // always visible
      var slotHandle = document.getElementById("banner-slot-data")

      if (data.currentSlot) slotHandle.innerHTML = addCommas(data.currentSlot)

      var finDelayDataHandle = document.getElementById("banner-fin-data")
      finDelayHtml = `
      <div id="banner-fin" class="info-item d-flex mr-3">
      <div class="info-item-header mr-1 text-warning">
        <span class="item-icon">
          <i class="fas fa-exclamation-triangle" data-toggle="tooltip" title="" data-original-title="The last finalized epoch was ${data.finalityDelay} epochs ago."></i>
        </span>
        <span class="item-text">
          Finality
        </span>
      </div>
      <div class="info-item-body text-warning">
        <span id="banner-fin-data">${data.finalityDelay}</span>
        <i class="fas fa-exclamation-triangle item-text" data-toggle="tooltip" title="" data-original-title="The last finalized epoch was ${data.finalityDelay} epochs ago."></i>
      </div>
    </div>
    `

      if (!finDelayDataHandle && data.finalityDelay > 3 && !data.syncing) {
        // create fin delay node
        document.getElementById("banner-slot").insertAdjacentHTML("afterend", finDelayHtml)
        $("#banner-fin i").each(function () {
          $(this).tooltip("update")
        })
      } else if (finDelayDataHandle && data.finalityDelay > 3 && !data.syncing) {
        // update fin delay node
        finDelayDataHandle.textContent = data.finalityDelay
        var icons = document.querySelectorAll("#banner-fin i")
        for (let i = 0; i < icons.length; i++) {
          const icon = icons[i]
          icon.setAttribute("data-original-title", `The last finalized epoch was ${data.finalityDelay} epochs ago.`)
        }
        $("#banner-fin i").each(function () {
          $(this).tooltip("update")
        })
      } else {
        // delete fin delay node if it exists
        let findDelayHandle = document.getElementById("banner-fin")
        if (findDelayHandle) findDelayHandle.remove()
      }
      if (data.syncing) {
        // remove fin delay if we are still syncing
        let findDelayHandle = document.getElementById("banner-fin")
        if (findDelayHandle) findDelayHandle.remove()

        var bannerHandle = document.getElementById("banner-status")
        if (!bannerHandle) {
          var statusHtml = `
          <a data-toggle="tooltip" title="The explorer is currently syncing with the network" id="banner-status" style="white-space: nowrap;" class="mr-2" href="/"><i class="fas fa-sync"></i> <span>|</span></a>
            `
          document.getElementById("banner-home").remove()
          document.getElementById("banner-stats").insertAdjacentHTML("afterbegin", statusHtml)
        }
      } else {
        // delete sync if it exists otherwise do nothing
        var statusHandle = document.getElementById("banner-status")
        if (statusHandle) {
          var homeHtml = `
          <a id="banner-home" style="white-space: nowrap;" class="mr-2" href="/"><i class="fas fa-home"></i> <span>|</span></a>
          `
          statusHandle.remove()
          document.getElementById("banner-stats").insertAdjacentHTML("afterbegin", homeHtml)
        }
      }
    })
}

// update the banner every 12 seconds
setInterval(updateBanner, 12000)
