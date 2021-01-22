
function createBlock(x, y) {
  use = document.createElementNS("http://www.w3.org/2000/svg","use")
  // use.setAttributeNS(null, "style", `transform: translate(calc(${x} * var(--disperse-factor)), calc(${y} * var(--disperse-factor)));`)
  use.setAttributeNS(null, "href", "#cube")
  use.setAttributeNS(null, "x", x)
  use.setAttributeNS(null, "y", y)
  return use;
}

function appendBlocks(blocks) {
  $(".blue-cube g.move").each(function() {
    $(this).empty()
  })

  var cubes = document.querySelectorAll('.blue-cube g.move')
  for (var i = 0; i < blocks.length; i++) {
    var block = blocks[i];

    for (let i = 0; i < cubes.length; i++) {
      var cube = cubes[i];
      cube.appendChild(createBlock(block[0], block[1]))
    }    
  }
  for (let i = 0; i < cubes.length; i++) {
    var cube = cubes[i];
    var use = document.createElementNS("http://www.w3.org/2000/svg","use")
    // use.setAttributeNS(null, "style", `transform: translate(calc(${x} * var(--disperse-factor)), calc(${y} * var(--disperse-factor)));`)
    use.setAttributeNS(null, "href", "#cube-small")
    use.setAttributeNS(null, "x", 129)
    use.setAttributeNS(null, "y", 56)
    cube.appendChild(use)
  }
}

$(document).ready(function() {

  //bookmark button adds all validators in the dashboard to the watchlist
  $('#bookmark-button').on("click", function(event) {
    var tickIcon = $("<i class='fas fa-check' style='width:15px;'></i>")
    var spinnerSmall = $('<div class="spinner-border spinner-border-sm" role="status"><span class="sr-only">Loading...</span></div>')
    var bookmarkIcon = $("<i class='far fa-bookmark' style='width:15px;'></i>")
    var errorIcon = $("<i class='fas fa-exclamation' style='width:15px;'></i>")
    fetch('/dashboard/save', {
      method: "POST",
      // credentials: 'include',
      headers: {
        'Content-Type': 'application/json'
        // 'X-CSRF-Token': $("#bookmark-button").attr("csrf"),
      },
      body: JSON.stringify(state.validators),
    }).then(function(res) {
      console.log('response', res)
      if (res.status === 200 && !res.redirected) {
        // success
        console.log("success")
        $('#bookmark-button').empty().append(tickIcon)
        setTimeout(function() {
          $('#bookmark-button').empty().append(bookmarkIcon)
        }, 1000)
      } else if (res.redirected) {
        console.log('redirected!')
        $('#bookmark-button').attr("data-original-title", "Please login or sign up first.")
        $('#bookmark-button').tooltip('show')
        $('#bookmark-button').empty().append(errorIcon)
        setTimeout(function() {
          $('#bookmark-button').empty().append(bookmarkIcon)
          $('#bookmark-button').tooltip('hide')
          $('#bookmark-button').attr("data-original-title", "Save all to Watchlist")
        }, 2000)
      } else {
        // could not bookmark validators
        $('#bookmark-button').empty().append(errorIcon)
        setTimeout(function() {
          $('#bookmark-button').empty().append(bookmarkIcon)
        }, 2000)
      }
    }).catch(function(err) {
      $('#bookmark-button').empty().append(errorIcon)
      setTimeout(function() {
        $('#bookmark-button').empty().append(bookmarkIcon)
      }, 2000)
      console.log(err)
    })
  })
  var clearSearch = $('#clear-search')
  //'<i class="fa fa-copy"></i>'
  var copyIcon = $("<i class='fa fa-copy' style='width:15px'></i>")
  //'<i class="fas fa-check"></i>'
  var tickIcon = $("<i class='fas fa-check' style='width:15px;'></i>")

  clearSearch.on('click', function() {
    clearSearch.empty().append(tickIcon);
    setTimeout(function(){
      clearSearch.empty().append(copyIcon);
    }, 500)
  })

  var validatorsDataTable = window.vdt = $('#validators').DataTable({
    processing: true,
    serverSide: false,
    ordering: true,
    searching: true,
    paging: false,
    info: false,
    preDrawCallback: function() {
      // this does not always work.. not sure how to solve the staying tooltip
      try {
        $('#validators').find('[data-toggle="tooltip"]').tooltip('dispose')
      } catch (e) {}
    },
    drawCallback: function(settings) {
      $('#validators').find('[data-toggle="tooltip"]').tooltip()
    },
    order: [[1,'asc']],
    columnDefs: [
      {
        targets: 0,
        data: '0',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data
          return '<a href="/validator/' + data + '">0x' + data.substr(0, 8) + '...</a>'
        }
      },
      {
        targets: 1,
        data: '1',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data
          return '<a href="/validator/' + data + '">' + data + '</a>'
        }
      },
      {
        targets: 2,
        data: '2',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data ? data[0] : null
          return `${data[0]} (${data[1]})`
        }
      },
      {
        targets: 3,
        data: '3',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data ? data[0] : -1
          var d = data.split('_')
          var s = d[0].charAt(0).toUpperCase() + d[0].slice(1)
          if (d[1] === 'offline') 
            return `<span style="display:none">${d[1]}</span><span data-toggle="tooltip" data-placement="top" title="No attestation in the last 2 epochs">${s} <i class="fas fa-power-off fa-sm text-danger"></i></span>`
          if (d[1] === 'online')
            return `<span style="display:none">${d[1]}</span><span>${s} <i class="fas fa-power-off fa-sm text-success"></i></span>`
          return `<span>${s}</span>`
        }
      },
      {
        targets: 4,
        data: '4',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data ? data[0] : null
          if (data === null) return '-'
          return `<span data-toggle="tooltip" data-placement="top" title="${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})}">${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 5,
        data: '5',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data ? data[0] : null
          if (data === null) return '-'
          return `<span data-toggle="tooltip" data-placement="top" title="${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})}">${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 6,
        data: '6',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data ? data[0] : null
          if (data === null) return '-'
          return `<span data-toggle="tooltip" data-placement="top" title="${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})}">${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 7,
        data: '7',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data ? data[0] : null
          if (data === null) return 'No Attestation found'
          return `<span data-toggle="tooltip" data-placement="top" title="${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})}">${luxon.DateTime.fromMillis(data[1] * 1000).toRelative({ style: "short"})} (<a href="/block/${data[0]}">Block ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 8,
        data: '8',
        render: function(data, type, row, meta) {
          if (type == 'sort' || type == 'type') return data ? data[0] + data[1] : null
          return `<span data-toggle="tooltip" data-placement="top" title="${data[0]} executed / ${data[1]} missed"><span class="text-success">${data[0]}</span> / <span class="text-danger">${data[1]}</span></span>`
        }
      }
    ]
  })

  var bhValidators = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.index
    },
    remote: {
      url: '/search/indexed_validators/%QUERY',
      wildcard: '%QUERY'
    }
  })
  var bhEth1Addresses = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.eth1_address
    },
    remote: {
      url: '/search/indexed_validators_by_eth1_addresses/%QUERY',
      wildcard: '%QUERY'
    }
  })
  var bhName = new Bloodhound({
    datumTokenizer: Bloodhound.tokenizers.whitespace,
    queryTokenizer: Bloodhound.tokenizers.whitespace,
    identify: function(obj) {
      return obj.name
    },
    remote: {
      url: '/search/indexed_validators_by_name/%QUERY',
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
      url: '/search/indexed_validators_by_graffiti/%QUERY',
      wildcard: '%QUERY'
    }
  })

  $('.typeahead-dashboard').typeahead(
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
      display: 'index',
      templates: {
        header: '<h3>Validators</h3>',
        suggestion: function(data) {
          return `<div class="text-monospace text-truncate">${data.index}: ${data.pubkey}</div>`
        }
      }
    },
    {
      limit: 5,
      name: 'addresses',
      source: bhEth1Addresses,
      display: 'address',
      templates: {
        header: '<h3>Validators by ETH1 Addresses</h3>',
        suggestion: function(data) {
          var len = data.validator_indices.length > 100 ? '100+' : data.validator_indices.length 
          return `<div class="text-monospace" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.eth1_address}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        }
      }
    },
    {
      limit: 5,
      name: 'graffiti',
      source: bhGraffiti,
      display: 'graffiti',
      templates: {
        header: '<h3>Validators by Graffiti</h3>',
        suggestion: function(data) {
          var len = data.validator_indices.length > 100 ? '100+' : data.validator_indices.length 
          return `<div class="text-monospace" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.graffiti}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        }
      }
    },
    {
      limit: 5,
      name: 'name',
      source: bhName,
      display: 'name',
      templates: {
        header: '<h3>Validators by Name</h3>',
        suggestion: function(data) {
          var len = data.validator_indices.length > 100 ? '100+' : data.validator_indices.length 
          return `<div class="text-monospace" style="display:flex"><div class="text-truncate" style="flex:1 1 auto;">${data.name}</div><div style="max-width:fit-content;white-space:nowrap;">${len}</div></div>`
        }
      }
    }
  )
  $('.typeahead-dashboard').on('focus', function(event) {
    if (event.target.value !== '') {
      $(this).trigger($.Event('keydown', { keyCode: 40 }))
    }
  })
  $('.typeahead-dashboard').on('input', function() {
    $('.tt-suggestion').first().addClass('tt-cursor')
  })
  $('.typeahead-dashboard').on('typeahead:select', function(ev, sug) {
    if (sug.validator_indices) {
      addValidators(sug.validator_indices)
    } else {
      addValidator(sug.index)
    }
    $('.typeahead-dashboard').typeahead('val', '')
  })
  $('#pending').on('click', 'button', function() {
    var data = pendingTable.row($(this).parents('tr')).data()
    removeValidator(data[1])
  })
  $('#active').on('click', 'button', function() {
    var data = activeTable.row($(this).parents('tr')).data()
    removeValidator(data[1])
  })
  $('#ejected').on('click', 'button', function() {
    var data = ejectedTable.row($(this).parents('tr')).data()
    removeValidator(data[1])
  })
  $('#selected-validators').on('click', '.remove-validator', function() {
    removeValidator(this.parentElement.dataset.validatorIndex)
  })

  $('.multiselect-border input').on('focus', function(event) {
    $('.multiselect-border').addClass('focused')
  })
  $('.multiselect-border input').on('blur', function(event) {
    $('.multiselect-border').removeClass('focused')
  })

  $('#clear-search').on('click', function(event) {
    if(state) {
      state = setInitialState()
      localStorage.removeItem('dashboard_validators')
      window.location = "/dashboard"
    }
    // window.location = "/dashboard"
  })

  function setInitialState () {
    var _state = {}
    _state.validators = []
    _state.validatorsCount = {
      pending: 0,
      active: 0,
      ejected: 0,
      offline: 0
    }
    return _state;
  }

  var state = setInitialState()

  setValidatorsFromURL()
  renderSelectedValidators()
  updateState()

  function renderSelectedValidators() {
    var elHolder = document.getElementById('selected-validators')
    $('#selected-validators .item').remove()
    var elsItems = []
    for (var i = 0; i < state.validators.length; i++) {
      var v = state.validators[i]
      var elItem = document.createElement('li')
      elItem.classList = 'item'
      elItem.dataset.validatorIndex = v
      elItem.innerHTML = v + ' <i class="fas fa-times-circle remove-validator"></i>'
      elsItems.push(elItem)
    }
    elHolder.prepend(...elsItems)
  }

  function renderDashboardInfo() {
    var el = document.getElementById('dashboard-info')
    el.innerText = `Found ${state.validatorsCount.pending} pending, ${state.validatorsCount.active_online + state.validatorsCount.active_offline} active and ${state.validatorsCount.exited} exited validators`
  }

  function setValidatorsFromURL() {
    var usp = new URLSearchParams(window.location.search)
    var validatorsStr = usp.get('validators')
    if (!validatorsStr) {
      var validatorsStr = localStorage.getItem('dashboard_validators')
      if (validatorsStr) {
        state.validators = JSON.parse(validatorsStr)
        state.validators = state.validators.filter((v, i) => {
          v = escape(v)
          if (isNaN(parseInt(v))) return false
          return state.validators.indexOf(v) === i
        })
        state.validators.sort(sortValidators)
      } else {
        state.validators = []
      }
      return
    }
    state.validators = validatorsStr.split(',')
    state.validators = state.validators.filter((v, i) => {
      v = escape(v)
      if (isNaN(parseInt(v))) return false
      return state.validators.indexOf(v) === i
    })
    state.validators.sort(sortValidators)
    if (state.validators.length > 100) {
      state.validators = state.validators.slice(0,100)
      console.log("100 validators limit reached")
      alert('You can not add more than 100 validators to your dashboard')
    }
  }

  function addValidators(indices) {
    var limitReached = false
    indicesLoop:
    for (var j = 0; j < indices.length; j++) {
      if (state.validators.length >= 100) {
        limitReached = true
        break indicesLoop
      }
      var index = indices[j]+"" // make sure index is string
      for (var i = 0; i < state.validators.length; i++) {
        if (state.validators[i] === index)
          continue indicesLoop
      }
      state.validators.push(index)
    }
    state.validators.sort(sortValidators)
    renderSelectedValidators()
    updateState()
    if (limitReached) {
      console.log("100 validators limit reached")
      alert('You can not add more than 100 validators to your dashboard')
    }
  }

  function addValidator(index) {
    if (state.validators.length > 100) {
      alert('Too much validators, you can not add more than 100 validators to your dashboard!')
      return
    }
    index = index+"" // make sure index is string
    for (var i = 0; i < state.validators.length; i++) {
      if (state.validators[i] === index) return
    }
    state.validators.push(index)
    state.validators.sort(sortValidators)
    renderSelectedValidators()
    updateState()
  }

  function removeValidator(index) {
    for (var i = 0; i < state.validators.length; i++) {
      if (state.validators[i] === index) {
        state.validators.splice(i, 1)
        state.validators.sort(sortValidators)
        //removed last validator
        if(state.validators.length === 0) {
          state = setInitialState()
          localStorage.removeItem('dashboard_validators')
          window.location = "/dashboard"
          return
        } else {
          renderSelectedValidators()
          updateState()
        }
        return
      }
    }
  }

  function sortValidators(a, b) {
    var ai = parseInt(a)
    var bi = parseInt(b)
    return ai - bi
  }

  function addChange(selector, value) {
    if(selector !== undefined || selector !== null) {
      var element = document.querySelector(selector)
      if(element !== undefined) {
        // remove old
        element.classList.remove('decreased')
        element.classList.remove('increased')
        if(value < 0) {
          element.classList.add("decreased")
        } 
        if (value > 0) {
          element.classList.add("increased")
        }
      } else {
        console.error("Could not find element with selector", selector)
      }
    } else {
      console.error("selector is not defined", selector)
    }
  }

  function updateState() {
    // if(_range < xBlocks.length + 3 && _range !== -1) {

    //   appendBlocks(xBlocks.slice(_range, _range+3))
    //   _range = _range + 3;
    // } else if(_range !== -1) {
    //   _range = -1;
    // }

    localStorage.setItem('dashboard_validators', JSON.stringify(state.validators))
    if(state.validators.length) {
      console.log('length', state.validators)
      var qryStr = '?validators=' + state.validators.join(',')
      var newUrl = window.location.pathname + qryStr
      window.history.replaceState(null, 'Dashboard', newUrl)
    }
    var t0 = Date.now()
    if (state.validators && state.validators.length) {
      // if(state.validators.length >= 9) {
      //   appendBlocks(xBlocks)
      // } else {
      //   appendBlocks(xBlocks.slice(0, state.validators.length * 3 - 1))
      // }
      document.querySelector('#bookmark-button').style.visibility = "visible"
      document.querySelector('#copy-button').style.visibility = "visible"
      document.querySelector('#clear-search').style.visibility = "visible"

      $.ajax({
        url: '/dashboard/data/earnings' + qryStr,
        success: function(result) {
          var t1 = Date.now()
          console.log(`loaded earnings: fetch: ${t1-t0}ms`)
          if (!result) return
          // document.getElementById('stats').style.display = 'flex'
          var lastDay = (result.lastDay / 1e9 * exchangeRate).toFixed(4)
          var lastWeek = (result.lastWeek / 1e9 * exchangeRate).toFixed(4)
          var lastMonth = (result.lastMonth / 1e9 * exchangeRate).toFixed(4)
          var total = (result.total / 1e9 * exchangeRate).toFixed(4)

          console.log(total, exchangeRate, result.total)
          addChange("#earnings-day-header", lastDay)
          addChange("#earnings-week-header", lastWeek)
          addChange("#earnings-month-header", lastMonth)
          addChange("#earnings-total-header", total)

          document.querySelector('#earnings-day').innerHTML = (lastDay || '0.000') + " <span class='small text-muted'>" + currency + "</span>"
          document.querySelector('#earnings-week').innerHTML = (lastWeek || '0.000') + " <span class='small text-muted'>" + currency + "</span>"
          document.querySelector('#earnings-month').innerHTML = (lastMonth || '0.000') + " <span class='small text-muted'>" + currency + "</span>"
          document.querySelector('#earnings-total').innerHTML = (total || '0.000') + " <span class='small text-muted'>" + currency + "</span>"
        }
      })
      $.ajax({
        url: '/dashboard/data/validators' + qryStr,
        success: function(result) {
          var t1 = Date.now()
          console.log(`loaded validators-data: length: ${result.data.length}, fetch: ${t1-t0}ms`)
          if (!result || !result.data.length) {
            document.getElementById('validators-table-holder').style.display = 'none'
            return
          }
          // pubkey, idx, currbal, effbal, slashed, acteligepoch, actepoch, exitepoch
          // 0:pubkey, 1:idx, 2:[currbal,effbal], 3:state, 4:[actepoch,acttime], 5:[exit,exittime], 6:[wd,wdt], 7:[lasta,lastat], 8:[exprop,misprop]
          // console.log(`latestEpoch: ${result.latestEpoch}`)
          // var latestEpoch = result.latestEpoch
          state.validatorsCount.pending = 0
          state.validatorsCount.active_online = 0
          state.validatorsCount.active_offline = 0
          state.validatorsCount.slashing_online = 0
          state.validatorsCount.slashing_offline = 0
          state.validatorsCount.exiting_online = 0
          state.validatorsCount.exiting_offline = 0
          state.validatorsCount.exited = 0
          state.validatorsCount.slashed = 0
  
          for (var i=0; i<result.data.length; i++) {
            var v = result.data[i]
            var vIndex = v[1]
            var vState = v[3]
            if (!state.validatorsCount[vState]) state.validatorsCount[vState] = 0
            state.validatorsCount[vState]++
            var el = document.querySelector(`#selected-validators .item[data-validator-index="${vIndex}"]`)
            if (el) el.dataset.state = vState
          }
          validatorsDataTable.clear()

          validatorsDataTable.rows.add(result.data).draw()
  
          validatorsDataTable.column(6).visible(false)

          requestAnimationFrame(()=>{validatorsDataTable.columns.adjust().responsive.recalc()})

          document.getElementById('validators-table-holder').style.display = 'block'

          renderDashboardInfo()
        }
      })

    } else {
      document.querySelector('#copy-button').style.visibility = "hidden"
      document.querySelector('#bookmark-button').style.visibility = "hidden"
      document.querySelector('#clear-search').style.visibility = "hidden"
      // window.location = "/dashboard"
    }

    $('#copy-button')
    .attr('data-clipboard-text', window.location.href)

    renderCharts()
  }

  window.onpopstate = function(event) {
    setValidatorsFromURL()
    renderSelectedValidators()
    updateState()
  }
  window.addEventListener('storage', function(e) {
      var validatorsStr = localStorage.getItem('dashboard_validators')
      if (JSON.stringify(state.validators) === validatorsStr) {
        return
      }
      if (validatorsStr) {
        state.validators = JSON.parse(validatorsStr)
      } else {
        state.validators = []
      }
      state.validators = state.validators.filter((v, i) => state.validators.indexOf(v) === i)
      state.validators.sort(sortValidators)
      renderSelectedValidators()
      updateState()
  })

  function renderCharts() {
    var t0 = Date.now()
    // if (state.validators.length === 0) {
    //   document.getElementById('chart-holder').style.display = 'none'
    //   return
    // }
    document.getElementById('chart-holder').style.display = 'flex'
    if (state.validators && state.validators.length) {
      var qryStr = '?validators=' + state.validators.join(',')
      $.ajax({
        url: '/dashboard/data/balance' + qryStr,
        success: function(result) {
          var t1 = Date.now()
          var balance = new Array(result.length)
          var effectiveBalance = new Array(result.length)
          var validatorCount = new Array(result.length)
          var utilization = new Array(result.length)
          for (var i = 0; i < result.length; i++) {
            var res = result[i]
            validatorCount[i] = [res[0], res[1]]
            balance[i] = [res[0], res[2]]
            effectiveBalance[i] = [res[0], res[3]]
            utilization[i] = [res[0], res[3] / (res[1] * (32 * exchangeRate))]
          }
  
          var t2 = Date.now()
          createBalanceChart(effectiveBalance, balance, utilization)
          var t3 = Date.now()
          console.log(`loaded balance-data: length: ${result.length}, fetch: ${t1 - t0}ms, aggregate: ${t2 - t1}ms, render: ${t3 - t2}ms`)
        }
      })
      $.ajax({
        url: '/dashboard/data/proposals' + qryStr,
        success: function(result) {
          var t1 = Date.now()
          var t2 = Date.now()
          if (result && result.length) {
            createProposedChart(result)
          }
          var t3 = Date.now()
          console.log(`loaded proposal-data: length: ${result.length}, fetch: ${t1 - t0}ms, render: ${t3 - t2}ms`)
        }
      })

    }
  }
})

function createBalanceChart(effective, balance, utilization, missedAttestations) {
  Highcharts.stockChart('balance-chart', {
    exporting: {
      scale: 1
    },
    rangeSelector: {
      enabled: false
    },
    chart: {
      type: 'line',
    },
    legend: {
      enabled: true
    },
    title: {
      text: 'Balance History for all Validators'
    },
    xAxis: {
      type: 'datetime',
      range: 7 * 24 * 60 * 60 * 1000,
      labels: {
        formatter: function(){
          var epoch = timeToEpoch(this.value)
          var orig = this.axis.defaultLabelFormatter.call(this)
          return `${orig}<br/>Epoch ${epoch}`
        }
      }
    },
    tooltip: {
      formatter: function(tooltip) {
        var orig = tooltip.defaultFormatter.call(this, tooltip)
        var epoch = timeToEpoch(this.x)
        orig[0] = `${orig[0]}<span style="font-size:10px">Epoch ${epoch}</span>`
        return orig
      }
    },
    yAxis: [
      {
        title: {
          text: 'Balance [' + currency + ']'
        },
        opposite: false,
        labels: {
          formatter: function() {
            return this.value.toFixed(0)
          },
          
        }
      },
      {
        title: {
          text: 'Validator Effectiveness'
        },
        labels: {
          formatter: function() {
            return (this.value * 100).toFixed(2) + '%'
          },

        },
        opposite: true
      }
    ],
    series: [
      {
        name: 'Balance',
        yAxis: 0,
        data: balance
      },
      {
        name: 'Effective Balance',
        yAxis: 0,
        step: true,
        data: effective
      },
      {
        name: 'Validator Effectiveness',
        yAxis: 1,
        data: utilization,
        tooltip: {
          pointFormatter: function() {
            return `<span style="color:${this.color}">●</span> ${this.series.name}: <b>${(this.y * 100).toFixed(2)}%</b><br/>`
          }
        }
      }
    ]
  })
}

function createProposedChart(data) {
  var proposed = []
  var missed = []
  var orphaned = []
  data.map(d=>{
    if (d[1] == 1) proposed.push([d[0]*1000,1])
    else if (d[1] == 2) missed.push([d[0]*1000,1])
    else if (d[1] == 3) orphaned.push([d[0]*1000,1])
  })
  Highcharts.stockChart('proposed-chart', {
    chart: {
      type: 'column',
    },
    title: {
      text: 'Proposal History for all Validators',
    },
    legend: {
      enabled: true
    },
    colors: ["#7cb5ec", "#ff835c", "#e4a354", "#2b908f", "#f45b5b", "#91e8e1"],
    xAxis: {
      lineWidth: 0,
      tickColor: '#e5e1e1',
    },
    yAxis: [
      {
        title: {
          text: '# of Possible Proposals'
        },
        opposite: false
      }
    ],
    plotOptions: {
      column: {
        stacking: 'normal',
        dataGrouping: {
          enabled: true,
          forced: true,
          units: [['day', [1]]]
        }
      }
    },
    series: [
      {
        name: 'Proposed',
        color: '#7cb5ec',
        data: proposed
      },
      {
        name: 'Missed',
        color: '#ff835c',
        data: missed
      },
      {
        name: 'Orphaned',
        color: '#e4a354',
        data: orphaned
      }
    ],
    rangeSelector: {
      enabled: false
    }
  })
}
