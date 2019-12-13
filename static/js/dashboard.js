$(document).ready(function() {
  var pendingTable = $('#pending')
    .DataTable({
      processing: true,
      serverSide: true,
      ordering: false,
      searching: true,
      ajax: '/dashboard/data/pending' + window.location.search,
      pagingType: 'full',
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
          targets: -1,
          data: null,
          orderable: false,
          defaultContent: '<button class="removeValidator">Remove</button>'
        }
      ]
    })
    .on('xhr.dt', function(e, settings, json, xhr) {
      // hide table if there are no data
      document.getElementById('pending-validators-table-holder').style.display = json.data.length ? 'block' : 'none'
    })
  var activeTable = $('#active')
    .DataTable({
      processing: true,
      serverSide: true,
      ordering: false,
      searching: true,
      ajax: '/dashboard/data/active' + window.location.search,
      pagingType: 'full',
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
          targets: -1,
          data: null,
          orderable: false,
          defaultContent: '<button class="removeValidator">Remove</button>'
        }
      ]
    })
    .on('xhr.dt', function(e, settings, json, xhr) {
      // hide table if there are no data
      document.getElementById('active-validators-table-holder').style.display = json.data.length ? 'block' : 'none'
    })
  var ejectedTable = $('#ejected')
    .DataTable({
      processing: true,
      serverSide: true,
      ordering: false,
      searching: true,
      ajax: '/dashboard/data/ejected' + window.location.search,
      pagingType: 'full',
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
          targets: -1,
          data: null,
          orderable: false,
          defaultContent: '<button class="removeValidator">Remove</button>'
        }
      ]
    })
    .on('xhr.dt', function(e, settings, json, xhr) {
      // hide table if there are no data
      document.getElementById('ejected-validators-table-holder').style.display = json.data.length ? 'block' : 'none'
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
      display: 'pubkey',
      templates: {
        // header: '<h3>Validators</h3>',
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
    $('.tt-suggestion')
      .first()
      .addClass('tt-cursor')
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
  // $('#validator-form').submit(function(event) {
  //   event.preventDefault()
  //   var search = $('#validator-form input').val()
  //   addValidator(search)
  // })

  var validators = []
  setValidatorsFromURL()
  renderSelectedValidators()
  renderCharts()

  function renderSelectedValidators() {
    var elHolder = document.getElementById('selected-validators')
    $('#selected-validators .item').remove()
    var elsItems = []
    for (var i = 0; i < validators.length; i++) {
      var v = validators[i]
      var elItem = document.createElement('li')
      elItem.classList = 'item'
      elItem.dataset.validatorIndex = v
      elItem.innerHTML = v + ' <i class="fas fa-times-circle remove-validator"></i>'
      elsItems.push(elItem)
    }
    elHolder.prepend(...elsItems)
  }

  function setValidatorsFromURL() {
    var usp = new URLSearchParams(window.location.search)
    var validatorsStr = usp.get('validators')
    if (!validatorsStr) {
      validators = []
      return
    }
    validators = validatorsStr.split(',')
  }

  function addValidator(index) {
    if (validators.length >= 100) {
      alert('Too much validators, you can not add more than 100 validators to your dashboard!')
      return
    }
    for (var i = 0; i < validators.length; i++) {
      if (validators[i] === index) return
    }
    validators.push(index)
    validators.sort(sortValidators)
    renderSelectedValidators()
    updateState()
  }

  function removeValidator(index) {
    for (var i = 0; i < validators.length; i++) {
      if (validators[i] === index) {
        validators.splice(i, 1)
        validators.sort(sortValidators)
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

  function updateState() {
    var qryStr = '?validators=' + validators.join(',')
    var newUrl = window.location.pathname + qryStr
    window.history.pushState(null, 'Dashboard', newUrl)
    pendingTable.ajax.url('/dashboard/data/pending' + qryStr)
    activeTable.ajax.url('/dashboard/data/active' + qryStr)
    ejectedTable.ajax.url('/dashboard/data/ejected' + qryStr)
    pendingTable.ajax.reload()
    activeTable.ajax.reload()
    ejectedTable.ajax.reload()
    renderCharts()
  }

  window.onpopstate = function(event) {
    setValidatorsFromURL()
    renderSelectedValidators()
    updateState()
  }

  function renderCharts() {
    if (validators.length === 0) {
      document.getElementById('balance-chart').style.display = 'none'
      document.getElementById('proposed-chart').style.display = 'none'
      return
    }
    document.getElementById('balance-chart').style.display = 'block'
    document.getElementById('proposed-chart').style.display = 'block'
    var qryStr = '?validators=' + validators.join(',')
    $.ajax({
      url: '/dashboard/data/balance' + qryStr,
      success: function(result) {
        var effective = result.effectiveBalanceHistory
        var balance = result.balanceHistory
        var utilization = []
        if (effective && effective.length && balance && balance.length) {
          var len = effective.length < balance.length ? effective.length : balance.length

          for (var i = 0; i < len; i++) {
            var numOfValidators = effective[i][2]
            var mostEffectiveBalance = numOfValidators * 3.2
            utilization.push([effective[i][0], effective[i][1] / mostEffectiveBalance])
          }
          createBalanceChart(effective, balance, utilization)
        }
      }
    })
    $.ajax({
      url: '/dashboard/data/proposals' + qryStr,
      success: function(result) {
        createProposedChart(result)
      }
    })
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
      animation: false,
      style: {
        fontFamily: 'Helvetica Neue", Helvetica, Arial, sans-serif'
      },
      backgroundColor: 'rgb(255, 255, 255)'
    },
    title: {
      text: 'Balance History for all Validators'
    },
    subtitle: {
      text: 'Source: beaconcha.in',
      style: {
        color: 'black'
      }
    },
    xAxis: {
      type: 'datetime',
      labels: {
        style: {
          color: 'black'
        }
      },
      range: 7 * 24 * 60 * 60 * 1000
    },
    yAxis: [
      {
        title: {
          text: 'Balance [ETH]',
          style: {
            color: '#26232780',
            'font-size': '0.8rem'
          }
        },
        opposite: false,
        labels: {
          formatter: function() {
            return this.value.toFixed(0)
          },
          style: {
            color: 'black'
          }
        }
      },
      {
        softMax: 1,
        softMin: 0,
        title: {
          text: 'Validator Effectiveness',
          style: {
            color: '#26232780',
            'font-size': '0.8rem'
          }
        },
        labels: {
          formatter: function() {
            return (this.value * 100).toFixed(0) + '%'
          },
          style: {
            color: 'black'
          }
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
    ],
    plotOptions: {
      line: {
        animation: false,
        lineWidth: 2.5
      }
    },
    legend: {
      enabled: true,
      layout: 'horizontal',
      align: 'center',
      verticalAlign: 'bottom',
      borderWidth: 0,
      itemStyle: {
        color: '#262327',
        'font-size': '0.8rem',
        'font-weight': 'lighter'
      },
      itemHoverStyle: {
        color: '#ff8723'
      }
    },
    credits: {
      enabled: false
    },
    navigator: {
      maskFill: '#1473e631',
      outlineColor: '#e5e1e1',
      handles: {
        backgroundColor: '#f5f3f3',
        borderColor: '#26232780'
      },
      xAxis: {
        gridLineColor: '#e5e1e1',
        labels: {
          style: {
            color: '#26232780'
          }
        }
      }
    },
    scrollbar: {
      barBackgroundColor: '#ebe7e7',
      barBorderWidth: 0,
      buttonArrowColor: '#262327',
      rifleColor: '#262327',
      buttonBackgroundColor: '#ebe7e7',
      buttonBorderColor: '#ebe7e7',
      trackBackgroundColor: '#f5f3f3',
      trackBorderColor: '#e5e1e180'
    },
    responsive: {
      rules: [
        {
          condition: {
            maxWidth: 760
          },
          chartOptions: {
            chart: {
              marginRight: 45
            },
            yAxis: [
              {
                title: {
                  text: null
                }
              },
              {
                title: {
                  text: null
                }
              }
            ]
          }
        }
      ]
    }
  })
}

function createProposedChart(data) {
  // if (!data || !data.length) return
  var proposed = data.map(d => [d.Day * 1000, d.Proposed])
  var missed = data.map(d => [d.Day * 1000, d.Missed])
  Highcharts.stockChart('proposed-chart', {
    exporting: {
      scale: 1
    },
    credits: {
      enabled: false
    },
    title: {
      text: 'Proposal History for all Validators'
    },
    subtitle: {
      text: 'Source: beaconcha.in',
      style: {
        color: 'black'
      }
    },
    chart: {
      type: 'column',
      animation: false,
      style: {
        fontFamily: 'Helvetica Neue", Helvetica, Arial, sans-serif'
      },
      backgroundColor: 'rgb(255, 255, 255)'
    },
    xAxis: {
      lineWidth: 0,
      tickColor: '#e5e1e1',
      labels: {
        style: {
          color: '#26232780'
        }
      },
      range: 7 * 24 * 3600 * 1000
    },
    yAxis: [
      {
        title: {
          text: '# of Possible Proposals',
          style: {
            color: '#26232780',
            'font-size': '0.8rem'
          }
        },
        labels: {
          style: {
            color: '#26232780'
          }
        },
        gridLineColor: '#e5e1e1',
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
    legend: {
      enabled: true,
      layout: 'horizontal',
      align: 'center',
      verticalAlign: 'bottom',
      borderWidth: 0,
      itemStyle: {
        color: '#262327',
        'font-size': '0.8rem',
        'font-weight': 'lighter'
      },
      itemHoverStyle: {
        color: '#ff8723'
      }
    },
    series: [
      {
        name: 'Proposed',
        data: proposed
      },
      {
        name: 'Missed',
        data: missed
      }
    ],
    rangeSelector: {
      enabled: false
    },
    navigator: {
      maskFill: '#1473e631',
      outlineColor: '#e5e1e1',
      handles: {
        backgroundColor: '#f5f3f3',
        borderColor: '#26232780'
      },
      xAxis: {
        gridLineColor: '#e5e1e1',
        labels: {
          style: {
            color: '#26232780'
          }
        }
      }
    },
    scrollbar: {
      barBackgroundColor: '#ebe7e7',
      barBorderWidth: 0,
      buttonArrowColor: '#262327',
      rifleColor: '#262327',
      buttonBackgroundColor: '#ebe7e7',
      buttonBorderColor: '#ebe7e7',
      trackBackgroundColor: '#f5f3f3',
      trackBorderColor: '#e5e1e180'
    },
    colors: ['#378ef0a0', '#fb8c13'],
    responsive: {
      rules: [
        {
          condition: {
            maxWidth: 760
          },
          chartOptions: {
            chart: {
              marginRight: 45
            },
            yAxis: [
              {
                title: {
                  text: null
                }
              },
              {
                title: {
                  text: null
                }
              }
            ]
          }
        }
      ]
    }
  })
}
