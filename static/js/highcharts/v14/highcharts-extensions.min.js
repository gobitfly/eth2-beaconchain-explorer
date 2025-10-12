/* Highcharts v14 extensions used by the entities overview page.
 * Adds support for colorAxis.title options by wrapping ColorAxis.setOptions.
 */
(function (H) {
  if (!H || !H.ColorAxis || !H.wrap) return;
  H.wrap(H.ColorAxis && H.ColorAxis.prototype, 'setOptions', function (proceed, userOptions) {
    proceed.apply(this, Array.prototype.slice.call(arguments, 1));
    if (userOptions && userOptions.title) {
      this.options.title = H.merge({ style: {} }, userOptions.title);
    }
  });
})(Highcharts);
