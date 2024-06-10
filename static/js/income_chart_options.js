function getIncomeChartOptions(clIncomeHistory, elIncomeHistory, title, height) {
  return {
    colors: ["#90ed7d", "#7cb5ec"],
    exporting: {
      scale: 1,
    },
    rangeSelector: {
      enabled: false,
    },
    chart: {
      type: "column",
      height: `${height}px"`,
      pointInterval: 24 * 3600 * 1000,
      events: {
        load: function () {
          $("#load-income-btn").removeClass("d-none")
        },
      },
    },
    title: {
      text: title,
    },
    credits: {
      enabled: false,
    },
    legend: {
      enabled: true,
    },
    plotOptions: {
      column: {
        stacking: "stacked",
        dataLabels: {
          enabled: false,
        },
        pointInterval: 24 * 3600 * 1000,
      },
      series: {
        turboThreshold: 10000,
      },
    },
    xAxis: {
      type: "datetime",
      range: 31 * 24 * 60 * 60 * 1000,
      labels: {
        formatter: function () {
          var epoch = timeToEpoch(this.value)
          var orig = this.axis.defaultLabelFormatter.call(this)
          return `${orig}<br/>Epoch ${epoch}`
        },
      },
    },
    yAxis: [
      {
        title: {
          text: `Income [${selectedCurrency}]`,
        },
        opposite: false,
        labels: {
          formatter: function () {
            return selectedCurrency === "ETH" ? trimToken(this.value) : trimCurrency(this.value)
          },
        },
      },
    ],
    series: [
      {
        name: "Execution Income",
        data: elIncomeHistory,
        showInNavigator: false,
        dataGrouping: {
          enabled: false,
        },
      },
      {
        name: "Consensus Income",
        data: clIncomeHistory,
        showInNavigator: true,
        dataGrouping: {
          enabled: false,
        },
      },
    ],
    tooltip: {
      split: false,
      shared: true,
      formatter: (tooltip) => {
        var text = ``
        var total = 0

        // time range for hovered point
        const ts = tooltip.chart.hoverPoints[0].x
        const startEpoch = timeToEpoch(ts)
        const timeForOneDay = 24 * 60 * 60 * 1000
        const endEpoch = timeToEpoch(ts + timeForOneDay) - 1
        const startDate = luxon.DateTime.fromMillis(ts)
        const endDate = luxon.DateTime.fromMillis(epochToTime(endEpoch + 1))
        text += `${startDate.toFormat("MMM-dd-yyyy HH:mm:ss")} - ${endDate.toFormat("MMM-dd-yyyy HH:mm:ss")}<br> Epochs ${startEpoch} - ${endEpoch}<br/>`

        // income
        for (var i = 0; i < tooltip.chart.hoverPoints.length; i++) {
          const value = tooltip.chart.hoverPoints[i].y
          const series = tooltip.chart.hoverPoints[i].series
          var iPrice = clPrice
          var iCurrency = clCurrency
          if (series.name == "Execution Income") {
            iPrice = elPrice
            iCurrency = elCurrency
          }

          text += `<span style="color:${tooltip.chart.hoverPoints[i].series.color}">\u25CF</span>  <b>${tooltip.chart.hoverPoints[i].series.name}:</b> ${getIncomeChartValueString(value, iCurrency, selectedCurrency, iPrice)}<br/>`
          total += value
        }

        // add total if hovered point contains rewards for both EL and CL
        if (tooltip.chart.hoverPoints.length > 1) {
          text += `<b>Total:</b> ${getIncomeChartValueString(total, selectedCurrency, selectedCurrency, clPrice)}`
        }

        return text
      },
    },
    responsive: {
      rules: [
        {
          condition: {
            callback: function () {
              return window.innerWidth >= 820
            },
          },
          chartOptions: {
            legend: {
              itemMarginTop: 7,
              itemMarginBottom: -7,
            },
          },
        },
      ],
    },
  }
}
