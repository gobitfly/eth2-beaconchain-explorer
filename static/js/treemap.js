;(function (window) {
  "use strict"

  function enc(s) {
    return encodeURIComponent(s || "")
  }

  // Inject minimal CSS to show pointer cursor on clickable treemap tiles across pages
  function ensureTreemapStyles() {
    var id = "beacon-treemap-styles"
    if (document.getElementById(id)) return
    var style = document.createElement("style")
    style.id = id
    style.type = "text/css"
    style.textContent = ".highcharts-point.clickable{cursor:pointer;}"
    if (document.head) document.head.appendChild(style)
  }
  ensureTreemapStyles()

  function buildPointsFromEntityRows(rows) {
    var OTHERS_THRESHOLD = 0.0025 // 0.01%
    var points = []
    // var othersShareSum = 0
    // var othersEffWeightedSum = 0

    if (!Array.isArray(rows)) return points

    for (var i = 0; i < rows.length; i++) {
      var item = rows[i] || {}
      var entity = item.entity || item.Entity || ""
      var eff = +item.efficiency || +item.Efficiency
      var share = +item.netShare || +item.NetShare || +item.net_share
      if (!entity || !isFinite(eff) || !isFinite(share)) continue

      if (share < OTHERS_THRESHOLD) {
        //   othersShareSum += share
        //   othersEffWeightedSum += eff * share
        continue
      }

      let path = "/entity/" + enc(entity) + "/-"
      points.push({
        name: entity,
        value: share,
        colorValue: eff,
        entity: entity,
        subEntity: "-",
        path: path,
        className: "clickable",
      })
    }

    // if (othersShareSum > 0) {
    //   var othersEff = othersEffWeightedSum > 0 ? othersEffWeightedSum / othersShareSum : null
    //   let path = "/entities"
    //   //points.push({ name: 'Others', value: othersShareSum, colorValue: othersEff, path: path, className: 'clickable'});
    // }

    return points
  }

  function renderTreemap(containerId, data, opts) {
    if (typeof Highcharts === "undefined" || !document.getElementById(containerId)) return

    var points = data
    if (!Array.isArray(points) || (points.length > 0 && (points[0].entity !== undefined || points[0].Entity !== undefined))) {
      points = buildPointsFromEntityRows(data)
    }

    var titleText = (opts && opts.title) || ""

    Highcharts.chart(containerId, {
      chart: { animation: false },
      colorAxis: {
        min: 0.98,
        max: 1,
        minColor: "#ffaa31",
        maxColor: "#60be60",
        visible: false,
      },
      plotOptions: {
        series: {
          point: {
            events: {
              click: function () {
                if (this && this.path) {
                  window.location.href = this.path
                }
              },
            },
          },
        },
        treemap: {
          dataLabels: {
            enabled: true,
            allowOverlap: true,
            style: { fontSize: "14px", fontWeight: "600", textOutline: "none" },
          },
        },
      },
      series: [{ type: "treemap", layoutAlgorithm: "squarified", clip: false, data: points }],
      title: { text: titleText },
      tooltip: {
        useHTML: true,
        pointFormatter: function () {
          var name = this.name || ""
          var sharePct = typeof this.value === "number" ? (this.value * 100).toFixed(2) + "%" : "N/A"
          var effPct = typeof this.colorValue === "number" ? (this.colorValue * 100).toFixed(2) + "%" : "N/A"
          return '<span style="font-size:11px">' + name + "</span><br/>" + "<span>Net share: <b>" + sharePct + "</b></span><br/>" + "<span>BeaconScore: <b>" + effPct + "</b></span>"
        },
      },
    })
  }

  function renderFromScript(containerId, scriptId, opts) {
    var el = document.getElementById(scriptId)
    var data = []
    if (el) {
      try {
        data = JSON.parse(el.textContent || "[]")
      } catch (e) {
        data = []
      }
    }
    renderTreemap(containerId, data, opts)
  }

  window.BeaconTreemap = {
    buildPointsFromEntityRows: buildPointsFromEntityRows,
    renderTreemap: renderTreemap,
    renderFromScript: renderFromScript,
  }
})(window)
