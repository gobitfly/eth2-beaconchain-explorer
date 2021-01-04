// FAB toggle
function toggleFAB() {
  var fabContainer = document.querySelector('.fab-message')
  var fabButton = fabContainer.querySelector('.fab-message-button a')
  var fabToggle = document.getElementById('fab-message-toggle')
  fabContainer.classList.toggle('is-open')
  fabButton.classList.toggle('toggle-icon')
}
$(document).ready(function() {
  var fabContainer = document.querySelector('.fab-message')
  var messages = document.querySelector('.fab-message-content h3')
  if (messages) {
    fabContainer.style.display = 'initial'
  }
})

// Theme switch
function switchTheme(e) {
  var d1 = document.getElementById('app-theme');
  //checked is light
  if (e.target.checked) {
    d1.href = "/theme/css/beacon-light.min.css"
    document.documentElement.setAttribute('data-theme', 'light')
    localStorage.setItem('theme', 'light')

  } else { // dark theme
    d1.href = "/theme/css/beacon-dark.min.css"
    document.documentElement.setAttribute('data-theme', 'dark')
    localStorage.setItem('theme', 'dark')
  }
}
$('#toggleSwitch').on('change', switchTheme)

function hideInfoBanner(msg){
  localStorage.setItem('infoBannerStatus', msg)
  $('#infoBanner').attr('class', 'd-none')
}
// $("#infoBannerDissBtn").on('click', hideInfoBanner)

// typeahead
$(document).ready(function() {
  formatTimestamps() // make sure this happens before tooltips
  $('[data-toggle="tooltip"]').tooltip()

  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.pubkey
    },
    remote: {
      url: '/search/validators/%QUERY',
      wildcard: '%QUERY'
    }
  })

  var bhBlocks = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.slot
    },
    remote: {
      url: '/search/blocks/%QUERY',
      wildcard: '%QUERY'
    }
  })

  var bhGraffiti = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.graffiti
    },
    remote: {
      url: '/search/graffiti/%QUERY',
      wildcard: '%QUERY'
    }
  })

  var bhEpochs = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.epoch
    },
    remote: {
      url: '/search/epochs/%QUERY',
      wildcard: '%QUERY'
    }
  })

  var bhEth1Accounts = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.account
    },
    remote: {
      url: '/search/eth1_addresses/%QUERY',
      wildcard: '%QUERY'
    }
  })

  $('.typeahead').typeahead(
    {
      minLength: 1,
      highlight: true,
      hint: false,
      autoselect: false
    },
    {
      limit: 5,
      name: 'validators',
      source: bhValidators,
      display: 'pubkey',
      templates: {
        header: '<h3 class="h5">Validators</h3>',
        suggestion: function(data) {
          return `<div class="text-monospace text-truncate">${data.index}: ${data.pubkey}</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'blocks',
      source: bhBlocks,
      display: 'blockroot',
      templates: {
        header: '<h3 class="h5">Blocks</h3>',
        suggestion: function(data) {
          return `<div class="text-monospace text-truncate">${data.slot}: ${data.blockroot}</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'epochs',
      source: bhEpochs,
      display: 'epoch',
      templates: {
        header: '<h3 class="h5">Epochs</h3>',
        suggestion: function(data) {
          return `<div>${data.epoch}</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'addresses',
      source: bhEth1Accounts,
      display: 'address',
      templates: {
        header: '<h3 class="h5">ETH1 Addresses</h3>',
        suggestion: function(data) {
          return `<div class="text-monospace text-truncate">0x${data.address}</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'graffiti',
      source: bhGraffiti,
      display: 'graffiti',
      templates: {
        header: '<h3 class="h5">Blocks by Graffiti</h3>',
        suggestion: function(data) {
          return `<div class="text-monospace" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.graffiti}</div><div style="max-width:fit-content;white-space:nowrap;">${data.count}</div></div>`
        }
      }
    }
  )

  $('.typeahead').on('focus', function(event) {
    if (event.target.value !== '') {
      $(this).trigger(
        $.Event('keydown', {
          keyCode: 40
        })
      )
    }
  })

  $('.typeahead').on('input', function(input) {

    $('.tt-suggestion')
      .first()
      .addClass('tt-cursor')
  })

  $('.tt-menu').on('mouseenter', function() {
    $('.tt-suggestion')
      .first()
      .removeClass('tt-cursor')
  })

  $('.tt-menu').on('mouseleave', function() {
    $('.tt-suggestion')
      .first()
      .addClass('tt-cursor')
  })

  $('.typeahead').on('typeahead:select', function(ev, sug) {
    if (sug.slot !== undefined) {
      window.location = '/block/' + sug.slot
    } else if (sug.index !== undefined) {
      if (sug.index === 'deposited')
        window.location = '/validator/' + sug.pubkey
      else 
        window.location = '/validator/' + sug.index
    } else if (sug.epoch !== undefined) {
      window.location = '/epoch/' + sug.epoch
    } else if (sug.address !== undefined) {
      window.location = '/validators/eth1deposits?q=' + sug.address
    } else if (sug.graffiti !== undefined) {
      // sug.graffiti is html-escaped to prevent xss, we need to unescape it
      var el = document.createElement('textarea')
      el.innerHTML = sug.graffiti
      window.location = '/blocks?q=' + encodeURIComponent(el.value)
    } else {
      console.log('invalid typeahead-selection', sug)
    }
  })
})

$('[aria-ethereum-date]').each(function(item) {
  var dt = $(this).attr('aria-ethereum-date')
  var format = $(this).attr('aria-ethereum-date-format')

  if (!format) {
    format = 'ff'
  }

  if (format === 'FROMNOW') {
    $(this).text(luxon.DateTime.fromMillis(dt * 1000).toRelative({ style: "short"}))
    $(this).attr('title', luxon.DateTime.fromMillis(dt * 1000).toFormat("ff"))
    $(this).attr('data-toggle', 'tooltip')
  } else {
    $(this).text(luxon.DateTime.fromMillis(dt * 1000).toFormat(format))
  }
})



$(document).ready(function() {
    var clipboard = new ClipboardJS('[data-clipboard-text]');
    clipboard.on('success', function (e) {
      var title = $(e.trigger).attr("data-original-title")
      $(e.trigger).tooltip('hide')
      .attr('data-original-title', "Copied!")
      .tooltip('show');

        setTimeout(function () {
        $(e.trigger).tooltip('hide')
        .attr('data-original-title', title)
      }, 1000);
    });

    clipboard.on('error', function (e) {
      var title = $(e.trigger).attr("data-original-title")
      $(e.trigger).tooltip('hide')
      .attr('data-original-title', "Failed to Copy!")
      .tooltip('show');

    setTimeout(function () {
      $(e.trigger).tooltip('hide')
      .attr('data-original-title', title)
    }, 1000);
  });
})

// With HTML5 history API, we can easily prevent scrolling!
$('.nav-tabs a').on('shown.bs.tab', function (e) {
    if (history.replaceState) {
        history.pushState(null, null, e.target.hash);
    } else {
        window.location.hash = e.target.hash; //Polyfill for old browsers
    }
})

// Javascript to enable link to tab
var url = document.location.toString();
if (url.match('#')) {
    $('.nav-tabs a[href="#'+url.split('#')[1]+'"]').tab('show') ;
}

function formatTimestamps(selStr) {
  var sel = $(document)
  if (selStr !== undefined) {
    sel = $(selStr)
  }
  sel.find('.timestamp').each(function(){
    var ts = $(this).data('timestamp')
    var tsLuxon = luxon.DateTime.fromMillis(ts * 1000)
    $(this).attr("data-original-title", tsLuxon.toFormat("ff"))
    $(this).text(tsLuxon.toRelative({ style: "short"}))
  })
  sel.find('[data-toggle="tooltip"]').tooltip()
}
