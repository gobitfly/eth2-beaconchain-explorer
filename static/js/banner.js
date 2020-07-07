var bannerSearchIcon = document.getElementById('banner-search');
var bannerSearchInput = document.getElementById('banner-search-input');
var banner = document.getElementById('banner');
var bannerRight = document.querySelector('.info-banner-right')

bannerSearchIcon.addEventListener('click', function (event) {
  banner.classList.add('hide-mobile')
  bannerRight.classList.add('show-search')
  bannerSearchIcon.classList.add('show')
  bannerSearchInput.classList.add('show')
  bannerSearchInput.focus()
})

bannerSearchInput.addEventListener('blur', function () {
  bannerSearchIcon.classList.remove('show')
  bannerSearchInput.classList.remove('show')
  bannerRight.classList.remove('show-search')
  banner.classList.remove('hide-mobile')
})

function updateBanner() {
  fetch('/latestState').then(function (res) {
    return res.json()
  }).then(function (data) {
    // always visible
    var epochHandle = document.getElementById('banner-epoch-data')

    if (data.currentEpoch)
      epochHandle.textContent = data.currentEpoch;

    // always visible
    var slotHandle = document.getElementById('banner-slot-data')

    if (data.currentSlot)
      slotHandle.textContent = data.currentSlot;

    var finDelayDataHandle = document.getElementById('banner-fin-data')
    finDelayHtml = `
      <div id="banner-fin" class="info-item d-flex mr-3">
      <div class="info-item-header mr-1 text-warning">
        <span class="item-icon">
          <i class="fas fa-exclamation-triangle" data-toggle="tooltip" title="" data-original-title="The last finalized epoch was ${data.finalityDelay } epochs ago."></i>
        </span>
        <span class="item-text">
          Finality
        </span>
      </div>
      <div class="info-item-body text-warning">
        <span id="banner-fin-data">${data.finalityDelay }</span>
        <i class="fas fa-exclamation-triangle item-text" data-toggle="tooltip" title="" data-original-title="The last finalized epoch was ${data.finalityDelay } epochs ago."></i>
      </div>
    </div>
    `

    if (!finDelayDataHandle && data.finalityDelay > 3 && !data.syncing) {
      // create fin delay node
      document.getElementById('banner-slot').insertAdjacentHTML('afterend', finDelayHtml)
      $('#banner-fin i').each(function () {
        $(this).tooltip('update')
      })
    } else if (finDelayDataHandle && data.finalityDelay > 3 && !data.syncing) {
      // update fin delay node
      finDelayDataHandle.textContent = data.finalityDelay
      var icons = document.querySelectorAll('#banner-fin i')
      for (let i = 0; i < icons.length; i++) {
        const icon = icons[i];
        icon.setAttribute('data-original-title', `The last finalized epoch was ${ data.finalityDelay } epochs ago.`)
      }
      $('#banner-fin i').each(function () {
        $(this).tooltip('update')
      })
    } else {
      // delete fin delay node if it exists
      var findDelayHandle = document.getElementById('banner-fin')
      if (findDelayHandle) findDelayHandle.remove();
    }
    if (data.syncing) {
      // remove fin delay if we are still syncing
      var findDelayHandle = document.getElementById('banner-fin')
      if (findDelayHandle) findDelayHandle.remove();

      var bannerHandle = document.getElementById('banner-status')
      if (!bannerHandle) {
        var statusHtml = `
            <div id="banner-status" class="info-item d-flex mr-3">
              <div class="info-item-body">
                <i class="fas fa-sync" data-toggle="tooltip" title="" data-original-title="The explorer is currently syncing with the network"></i>
              </div>
            </div>
            `
        document.getElementById('banner').insertAdjacentHTML('beforeend', statusHtml)
      }
    } else {
      // delete sync if it exists otherwise do nothing
      var statusHandle = document.getElementById('banner-status')
      if (statusHandle) {
        statusHandle.remove()
      }
    }
  })
}
// update the banner every 12 seconds
setInterval(updateBanner, 12000)