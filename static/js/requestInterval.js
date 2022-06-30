window.requestAnimFrame = (function () {
  return (
    window.requestAnimationFrame ||
    window.webkitRequestAnimationFrame ||
    window.mozRequestAnimationFrame ||
    window.oRequestAnimationFrame ||
    window.msRequestAnimationFrame ||
    function (/* function */ callback, /* DOMElement */ element) {
      window.setTimeout(callback, 1000 / 60)
    }
  )
})()

/**
 * Behaves the same as setInterval except uses requestAnimationFrame() where possible for better performance
 * @param {function} fn The callback function
 * @param {int} delay The delay in milliseconds
 */
window.requestInterval = function (fn, delay) {
  if (
    !window.requestAnimationFrame &&
    !window.webkitRequestAnimationFrame &&
    !(window.mozRequestAnimationFrame && window.mozCancelRequestAnimationFrame) && // Firefox 5 ships without cancel support
    !window.oRequestAnimationFrame &&
    !window.msRequestAnimationFrame
  )
    return window.setInterval(fn, delay)

  var start = new Date().getTime(),
    handle = new Object()

  function loop() {
    var current = new Date().getTime(),
      delta = current - start

    if (delta >= delay) {
      fn.call()
      start = new Date().getTime()
    }
    handle.value = requestAnimFrame(loop)
  }

  handle.value = requestAnimFrame(loop)
  return handle
}
