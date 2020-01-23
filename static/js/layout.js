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
  if (e.target.checked) {
    document.documentElement.setAttribute('data-theme', 'light')
    localStorage.setItem('theme', 'light')
    document.getElementById('nav').classList.remove('navbar-dark')
    document.getElementById('nav').classList.add('navbar-light')
  } else {
    document.documentElement.setAttribute('data-theme', 'dark')
    document.getElementById('nav').classList.remove('navbar-light')
    document.getElementById('nav').classList.add('navbar-dark')
    localStorage.setItem('theme', 'dark')
  }
}
$('#toggleSwitch').on('change', switchTheme)

// typeahead
$(document).ready(function() {
  $('[data-toggle="tooltip"]').tooltip()

  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.index
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
      return obj.blockroot
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
      return obj.blockroot
    },
    remote: {
      url: '/search/epochs/%QUERY',
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
        header: '<h3>Validators</h3>',
        suggestion: function(data) {
          return `<div>${data.index}: ${data.pubkey.substring(0, 16)}…</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'blocks',
      source: bhBlocks,
      display: 'blockroot',
      templates: {
        header: '<h3>Blocks</h3>',
        suggestion: function(data) {
          return `<div>${data.slot}: ${data.blockroot.substring(0, 16)}…</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'epochs',
      source: bhEpochs,
      display: 'epoch',
      templates: {
        header: '<h3>Epochs</h3>',
        suggestion: function(data) {
          return `<div>${data.epoch}</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'graffiti',
      source: bhGraffiti,
      display: 'graffiti',
      templates: {
        header: '<h3>Graffiti</h3>',
        suggestion: function(data) {
          if (data.graffiti) {
            data.graffiti = data.graffiti.replace(/(^\")|(\"$)'/, '').trim()
            return `<div>${data.graffiti}</div>`
          } else {
            return `<div>${data.slot}<div>`
          }
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
  var searchIcon

  // $('input.typeahead').on('blur', function(input) {
  //   if (searchIcon)
  //     $(this)
  //       .parent()
  //       .parent()
  //       .append(searchIcon)
  // })

  // $('input.typeahead').on('focus', function(input) {
  //   if (searchIcon) searchIcon.detach()
  // })

  $('.typeahead').on('input', function(input) {
    var siblings = $(this)
      .parent()
      .siblings()
    if (siblings && siblings.length) {
      searchIcon = siblings
    }
    if (searchIcon)
      if (input.target.value !== '') {
        searchIcon.detach()
      } else {
        $(this)
          .parent()
          .parent()
          .append(searchIcon)
      }

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
    if (sug.blockroot !== undefined) {
      window.location = '/block/' + sug.blockroot
    } else if (sug.index !== undefined) {
      window.location = '/validator/' + sug.index
    } else if (sug.epoch !== undefined) {
      window.location = '/epoch/' + sug.epoch
    } else {
      console.log('invalid typeahead-selection', sug)
    }
  })
})

moment.locale((window.navigator.userLanguage || window.navigator.language).toLowerCase())
$('[aria-ethereum-date]').each(function(item) {
  var dt = $(this).attr('aria-ethereum-date')
  var format = $(this).attr('aria-ethereum-date-format')

  if (!format) {
    format = 'L LTS'
  }

  if (format === 'FROMNOW') {
    $(this).text(moment.unix(dt).fromNow())
  } else {
    $(this).text(moment.unix(dt).format(format))
  }
})

var indicator = $('#nav .nav-indicator')
var items = document.querySelectorAll('#nav .nav-item')
var selectedLi = indicator.parent()[0]
var navigated = false

function handleIndicator(el) {
  indicator.css({
    width: `${el.offsetWidth}px`,
    left: `${el.offsetLeft}px`,
    bottom: 0
  })
}

items.forEach(function(item, index) {
  item.addEventListener('click', el => {
    if (navigated === false) {
      indicator
        .css({
          width: `${selectedLi.offsetWidth}px`,
          left: `${selectedLi.offsetLeft}px`,
          bottom: 0
        })
        .detach()
        .appendTo('.navbar ul') //.appendTo(el.target)
    }
    navigated = true
    handleIndicator(item)
  })
})
