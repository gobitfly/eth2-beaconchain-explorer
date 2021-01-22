var bannerContainer=document.querySelector(".info-banner-container"),bannerSearch=document.querySelector(".info-banner-search"),bannerSearchIcon=document.getElementById("banner-search"),bannerSearchInput=document.getElementById("banner-search-input");bannerSearch.addEventListener("click",function(){bannerContainer.classList.add("searching"),bannerSearchInput.focus()}),bannerSearchInput.addEventListener("blur",function(){bannerContainer.classList.remove("searching")});function updateBanner(){fetch("/latestState").then(function(res){return res.json()}).then(function(data){var epochHandle=document.getElementById("banner-epoch-data");data.currentEpoch&&(epochHandle.textContent=data.currentEpoch);var ethPriceHandle=document.getElementById("banner-eth-price-data");data.ethPrice&&(ethPriceHandle.innerHTML="$"+data.ethPrice);var slotHandle=document.getElementById("banner-slot-data");data.currentSlot&&(slotHandle.textContent=data.currentSlot);var finDelayDataHandle=document.getElementById("banner-fin-data");if(finDelayHtml=`
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
    `,!finDelayDataHandle&&data.finalityDelay>3&&!data.syncing)document.getElementById("banner-slot").insertAdjacentHTML("afterend",finDelayHtml),$("#banner-fin i").each(function(){$(this).tooltip("update")});else if(finDelayDataHandle&&data.finalityDelay>3&&!data.syncing){finDelayDataHandle.textContent=data.finalityDelay;var icons=document.querySelectorAll("#banner-fin i");for(let i=0;i<icons.length;i++){const icon=icons[i];icon.setAttribute("data-original-title",`The last finalized epoch was ${data.finalityDelay} epochs ago.`)}$("#banner-fin i").each(function(){$(this).tooltip("update")})}else{var findDelayHandle=document.getElementById("banner-fin");findDelayHandle&&findDelayHandle.remove()}if(data.syncing){var findDelayHandle=document.getElementById("banner-fin");findDelayHandle&&findDelayHandle.remove();var bannerHandle=document.getElementById("banner-status");if(!bannerHandle){var statusHtml=`
            <div id="banner-status" class="info-item d-flex mr-3">
              <div class="info-item-body">
                <i class="fas fa-sync" data-toggle="tooltip" title="" data-original-title="The explorer is currently syncing with the network"></i>
              </div>
            </div>
            `;document.getElementById("banner-stats").insertAdjacentHTML("beforeend",statusHtml)}}else{var statusHandle=document.getElementById("banner-status");statusHandle&&statusHandle.remove()}})}setInterval(updateBanner,12e3);
