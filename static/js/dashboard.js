
function setTooltip(selector, message) {
  $(selector).tooltip('hide').attr('data-original-title', message)
  setTimeout(function(){
    $(selector)
      .tooltip('show');
  }, 50)
}

function hideTooltip(selector, message) {
  setTimeout(function () {
    $(selector).tooltip('hide')
      .attr('data-original-title', message)
  }, 1000);
}

function createBlock(x, y) {
  use = document.createElementNS("http://www.w3.org/2000/svg","use")
  // use.setAttributeNS(null, "style", `transform: translate(calc(${x} * var(--disperse-factor)), calc(${y} * var(--disperse-factor)));`)
  use.setAttributeNS(null, "href", "#cube")
  use.setAttributeNS(null, "x", x)
  use.setAttributeNS(null, "y", y)
  return use;
}

function appendBlocks(blocks) {

  var use = document.createElementNS("http://www.w3.org/2000/svg","use")
  // use.setAttributeNS(null, "style", `transform: translate(calc(${x} * var(--disperse-factor)), calc(${y} * var(--disperse-factor)));`)
  use.setAttributeNS(null, "href", "#cube-small")
  use.setAttributeNS(null, "x", 129)
  use.setAttributeNS(null, "y", 56)
  $("g.move").empty()

  for (var i = 0; i < blocks.length; i++) {
    var block = blocks[i];
    block = createBlock(block[0], block[1])
    document.querySelector('g.move').appendChild(block)
  }
  document.querySelector('g.move').appendChild(use)
}

$(document).ready(function() {
/* 
  [121, 48],
  [121, 24],
  [121, 0],
  [100, 60],
  [100, 36],
  [100, 12],
  [142, 60],
  [142, 36],
  [142, 12],
  [163, 72],
  [163, 48],
  [163, 24],
  [79, 72],
  [79, 48],
  [79, 24],
  [121, 72],
  [121, 48],
  [121, 24],
  [100, 84],
  [100, 60],
  [100, 36],
  [142, 84],
  [142, 60],
  [142, 36],
  [121, 96],
  [121, 72],
  [129, 56]
*/
 var xBlocks = [
  [121, 48],
  [121, 24],
  [121, 0],
  [100, 60],
  [100, 36],
  [100, 12],
  [142, 60],
  [142, 36],
  [142, 12],
  [163, 72],
  [163, 48],
  [163, 24],
  [79, 72],
  [79, 48],
  [79, 24],
  [121, 72],
  [121, 48],
  [121, 24],
  [100, 84],
  [100, 60],
  [100, 36],
  [142, 84],
  [142, 60],
  [142, 36],
  [121, 96],
  [121, 72],
  // [129, 56]
 ]



  var clipboard = new ClipboardJS('#copy-button');

  var copyButton = $('#copy-button')
  var clearSearch = $('#clear-search')
  //'<i class="fa fa-copy"></i>'
  copyIcon = $("<i class='fa fa-copy' style='width:15px'></i>")
  //'<i class="fas fa-check"></i>'
  tickIcon = $("<i class='fas fa-check' style='width:15px;'></i>")


  clipboard.on('success', function (e) {
    copyButton.empty().append(tickIcon);

    setTooltip('#copy-button', 'Link Copied')
    hideTooltip('#copy-button', 'Copy Link to Dashboard')

    setTimeout(function(){
      copyButton.empty().append(copyIcon);
    }, 500)
  });

  clipboard.on('error', function (e) {
    setTooltip('#copy-button', 'Failed to copy Dashboard Link!');
    hideTooltip('#copy-button');
  });

  copyButton.on('click', function() {
    
  })

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
          return '<a href="/validator/' + data + '">0x' + data.substr(0, 8) + '...</a>'
        }
      },
      {
        targets: 1,
        data: '1',
        render: function(data, type, row, meta) {
          return '<a href="/validator/' + data + '">' + data + '</a>'
        }
      },
      {
        targets: 2,
        data: '2',
        render: function(data, type, row, meta) {
          return `${data[0]} (${data[1]})`
        }
      },
      {
        targets: 3,
        data: '3',
        render: function(data, type, row, meta) {
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
          if (data === null) 
            return '-'
          return `<span data-toggle="tooltip" data-placement="top" title="${moment.unix(data[1]).format()}">${moment.unix(data[1]).fromNow()} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 5,
        data: '5',
        render: function(data, type, row, meta) {
          if (data === null) 
            return '-'
          return `<span data-toggle="tooltip" data-placement="top" title="${moment.unix(data[1]).format()}">${moment.unix(data[1]).fromNow()} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 6,
        data: '6',
        render: function(data, type, row, meta) {
          if (data === null) 
            return '-'
          return `<span data-toggle="tooltip" data-placement="top" title="${moment.unix(data[1]).format()}">${moment.unix(data[1]).fromNow()} (<a href="/epoch/${data[0]}">Epoch ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 7,
        data: '7',
        render: function(data, type, row, meta) {
          if (data === null)
            return 'No Attestation found'
          return `<span data-toggle="tooltip" data-placement="top" title="${moment.unix(data[1]).format()}">${moment.unix(data[1]).fromNow()} (<a href="/block/${data[0]}">Block ${data[0]}</a>)</span>`
        }
      },
      {
        targets: 8,
        data: '8',
        render: function(data, type, row, meta) {
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
      url: '/search/validators/%QUERY',
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
        suggestion: function(data) {
          return `<div>${data.index}: ${data.pubkey.substring(0, 16)}…</div>`
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
    addValidator(sug.index)
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
        state.validators = state.validators.filter((v, i) => state.validators.indexOf(v) === i)
        state.validators.sort(sortValidators)
      } else {
        state.validators = []
      }
      return
    }
    state.validators = validatorsStr.split(',')
    state.validators = state.validators.filter((v, i) => state.validators.indexOf(v) === i)
    state.validators.sort(sortValidators)
  }

  function addValidator(index) {
    if (state.validators.length >= 100) {
      alert('Too much validators, you can not add more than 100 validators to your dashboard!')
      return
    }
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
        renderSelectedValidators()
        updateState()
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
    var qryStr = '?validators=' + state.validators.join(',')
    var newUrl = window.location.pathname + qryStr
    window.history.pushState(null, 'Dashboard', newUrl)
    var t0 = Date.now()
    if(state.validators && state.validators.length) {
      if(state.validators.length >= 9) {
        appendBlocks(xBlocks)
      } else {
        appendBlocks(xBlocks.slice(0, state.validators.length * 3 - 1))
      }
      document.querySelector('#copy-button').style.visibility = "visible"
      document.querySelector('#clear-search').style.visibility = "visible"

      $.ajax({
        url: '/dashboard/data/earnings' + qryStr,
        success: function(result) {
          var t1 = Date.now()
          console.log(`loaded earnings: fetch: ${t1-t0}ms`)
          if (!result) return
          // document.getElementById('stats').style.display = 'flex'
          var lastDay = (result.lastDay/1e9).toFixed(4) 
          var lastWeek = (result.lastWeek/1e9).toFixed(4)
          var lastMonth = (result.lastMonth/1e9).toFixed(4)
          var total = (result.total/1e9).toFixed(4)
  
          addChange("#earnings-day-header", lastDay)
          addChange("#earnings-week-header", lastWeek)
          addChange("#earnings-month-header", lastMonth)
          addChange("#earnings-total-header", total)
  
  
  
          document.querySelector('#earnings-day').innerText = lastDay || '0.000'
          document.querySelector('#earnings-week').innerText = lastWeek || '0.000'
          document.querySelector('#earnings-month').innerText = lastMonth || '0.000'
          document.querySelector('#earnings-total').innerText = total || '0.000'
  //         document.querySelector('#stats-earnings .stats-box-body').innerText = `total: ${(result.total/1e9).toFixed(4)} ETH
  // 1 day: ${(result.lastDay/1e9).toFixed(4)} ETH
  // 7 days: ${(result.lastWeek/1e9).toFixed(4)} ETH
  // 31 days: ${(result.lastMonth/1e9).toFixed(4)} ETH`
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
          console.log(`latestEpoch: ${result.latestEpoch}`)
          var latestEpoch = result.latestEpoch
          state.validatorsCount.pending = 0
          state.validatorsCount.active_online = 0
          state.validatorsCount.active_offline = 0
          state.validatorsCount.slashing_online = 0
          state.validatorsCount.slashing_offline = 0
          state.validatorsCount.exiting_online = 0
          state.validatorsCount.exiting_offline = 0
          state.validatorsCount.exited  = 0
  
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
          console.log('search',validatorsDataTable.columns().search())
          validatorsDataTable.rows.add(result.data).draw()
  
          validatorsDataTable.column(6).visible(false)
  
          requestAnimationFrame(()=>{validatorsDataTable.columns.adjust().responsive.recalc()})
  
          // document.getElementById('stats').style.display = 'flex'
          // document.querySelector('#stats-validators-status').innerText = `showing ${state.validatorsCount.pending} pending, ${state.validatorsCount.active_online} active, `
  //         `pending:  ${state.validatorsCount.pending}
  // active:   ${state.validatorsCount.active_online} / ${state.validatorsCount.active_offline}
  // slashing: ${state.validatorsCount.slashing_online} / ${state.validatorsCount.slashing_offline}
  // exiting:  ${state.validatorsCount.exiting_online} / ${state.validatorsCount.exiting_offline}
  // exited:   ${state.validatorsCount.exited}`
  
          document.getElementById('validators-table-holder').style.display = 'block'
          renderDashboardInfo()
        }
      })

    } else {
      document.querySelector('#copy-button').style.visibility = "hidden"
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
    if(state.validators && state.validators.length) {
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
            utilization[i] = [res[0], res[3] / (res[1] * 3.2)]
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

function createBalanceChart(effective, balance, utilization) {
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
          text: 'Balance [ETH]'
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
  var proposed = data.map(d => [d.Day * 1000, d.Proposed])
  var missed = data.map(d => [d.Day * 1000, d.Missed])
  var orphaned = data.map(d => [d.Day * 1000, d.Orphaned])
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
      range: 7 * 24 * 3600 * 1000
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
