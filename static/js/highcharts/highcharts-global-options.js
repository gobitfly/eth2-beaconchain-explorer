// Highcharts global options
Highcharts.setOptions({
  time: {
    useUTC: false,
  },
  credits: {
    enabled: true,
    href: "https://beaconcha.in",
    text: "beaconcha.in",
    style: {
      color: "var(--body-color)",
    },
  },
  exporting: {
    scale: 1,
    enabled: false,
  },
  title: {
    style: {
      color: "var(--font-color)",
    },
  },
  subtitle: {
    style: {
      color: "var(--font-color)",
    },
  },
  chart: {
    animation: false,
    style: {
      fontFamily: 'Helvetica Neue", Helvetica, Arial, sans-serif',
      color: "var(--body-color)",
      fontSize: "12px",
    },
    backgroundColor: "var(--bg-color)",
  },
  colors: ["#7cb5ec", "#90ed7d", "#f7a35c", "#D87F98", "#FFF3EC", "#A99F9F", "#2b908f", "#f45b5b", "#91e8e1", "#A0AC53", "#5F53AC", "#8085e9", "#f15c80", "#e4d354"],
  legend: {
    enabled: true,
    layout: "horizontal",
    align: "center",
    verticalAlign: "bottom",
    borderWidth: 0,
    itemStyle: {
      color: "var(--body-color)",
      "font-size": "0.8rem",
      // 'font-weight': 'lighter'
    },
    itemHoverStyle: {
      color: "var(--primary)",
    },
  },
  xAxis: {
    ordinal: false,
    labels: {
      style: {
        color: "var(--font-color)",
      },
    },
  },
  yAxis: {
    title: {
      style: {
        color: "var(--font-color)",
        // 'font-size': '0.8rem'
      },
    },
    labels: {
      style: {
        color: "var(--body-color)",
        "font-size": "0.8rem",
      },
    },
    gridLineColor: "var(--border-color-transparent)",
  },
  navigation: {
    menuStyle: {
      border: "1px solid var(--border-color-transparent)",
      background: "var(--bg-color-nav)",
      padding: "1px 0",
      "box-shadow": "var(--bg-color-nav) 2px 2px 5px",
    },
    buttonOptions: {
      symbolStroke: "var(--body-color)",
      symbolFill: "var(--body-color)",
      theme: {
        fill: "var(--bg-color)",
        states: {
          hover: {
            fill: "var(--primary)",
          },
          select: {
            fill: "var(--primary)",
          },
        },
      },
    },
    menuItemStyle: {
      padding: "0.5em 1em",
      color: "var(--transparent-font-color)",
      background: "none",
      fontSize: "11px/14px",
      transition: "background 250ms, color 250ms",
    },
    menuItemHoverStyle: { background: "var(--primary)", color: "var(--font-color)" },
  },
  navigator: {
    enabled: true,
    maskFill: "var(--mask-fill-color)",
    outlineColor: "var(--border-color)",
    handles: {
      backgroundColor: "var(--bg-color-nav)",
      borderColor: "var(--transparent-font-color)",
    },
    xAxis: {
      gridLineColor: "var(--border-color)",
    },
  },
  scrollbar: {
    barBackgroundColor: "var(--bg-color-nav)",
    barBorderWidth: 0,
    buttonArrowColor: "var(--font-color)",
    rifleColor: "var(--dark)",
    buttonBackgroundColor: "var(--bg-color-nav)",
    buttonBorderColor: "var(--bg-color-transparent)",
    trackBackgroundColor: "var(--bg-color)",
    trackBorderColor: "var(--border-color-transparent)",
  },
  // responsive: {
  //   rules: [
  //     {
  //       condition: {
  //         maxWidth: 590
  //       },
  //       chartOptions: {
  //         chart: {
  //           marginRight: 80
  //         },
  //         yAxis: [
  //           {
  //             title: {
  //               text: null
  //             }
  //           },
  //           {
  //             title: {
  //               text: null
  //             }
  //           }
  //         ]
  //       }
  //     }
  //   ]
  // },
  plotOptions: {
    line: {
      animation: false,
      lineWidth: 2.5,
    },
    column: {
      dataGrouping: {
        approximation: "sum",
      },
    },
  },
})
