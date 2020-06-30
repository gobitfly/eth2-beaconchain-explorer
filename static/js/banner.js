
// Update banner stats
//   data.lastProposedSlot number
//   data.currentSlot number
//   data.currentEpoch number
//   data.currentFinalizedEpoch number
//   data.finalityDelay number
//   data.syncing bool
function updateBanner() {
    fetch('/latestState').then(function(res) {
        return res.json()
    }).then(function(data) {
        // always visible
        var epochHandle = document.getElementById('banner-epoch-data')
        
        if (data.currentEpoch)
            epochHandle.textContent = data.currentEpoch;

        // always visible
        var slotHandle = document.getElementById('banner-slot-data')
        
        if(data.currentSlot)
            slotHandle.textContent = data.currentSlot;

        var finDelayDataHandle = document.getElementById('banner-fin-data')
        finDelayHtml = `
            <div id="banner-fin" class="info-item d-flex mr-3">
                <div class="info-item-header mr-1 text-warning">Finality</div>
                <div class="info-item-body text-warning">
                    <span id="banner-fin-data">${data.finalityDelay}</span>
                    <i class="fas fa-info-circle fa-sm" data-toggle="tooltip" title data-original-title="The last finalized epoch was ${data.finalityDelay} epochs ago."></i>
                </div>
            </div>
        `

        if (!finDelayDataHandle && data.finalityDelay > 3 && !data.syncing) {
            // create fin delay node
            document.getElementById('banner-slot').insertAdjacentHTML('afterend', finDelayHtml)
            $('#banner-fin i').tooltip('update')
        } else if (finDelayDataHandle && data.finalityDelay > 3 && !data.syncing) {
            // update fin delay node
            finDelayDataHandle.textContent = data.finalityDelay
            document.querySelector('#banner-fin i').setAttribute('data-original-title', `The last finalized epoch was ${data.finalityDelay} epochs ago.`)
            $('#banner-fin i').tooltip('update')
          } else {
            // delete fin delay node if it exists
            var findDelayHandle =  document.getElementById('banner-fin')
            if(findDelayHandle) findDelayHandle.remove();
        }
        if(data.syncing) {
            // remove fin delay if we are still syncing
            var findDelayHandle =  document.getElementById('banner-fin')
            if(findDelayHandle) findDelayHandle.remove();

            var bannerHandle = document.getElementById('banner-status')
            if(!bannerHandle) {
                var statusHtml = `
                <div id="banner-status" class="info-item d-flex mr-3">
                    <div class="info-item-header mr-1">Status</div>
                    <div class="info-item-body">Syncing</div>
                </div>
                `
                document.getElementById('banner').insertAdjacentHTML('beforeend', statusHtml)
            }
        } else {
            // delete sync if it exists otherwise do nothing
            var statusHandle = document.getElementById('banner-status')
            if(statusHandle) {
                statusHandle.remove()
            }
        }
    })
}