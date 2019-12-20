/**
 * --------------------------------------------------------------------------
 * Bootstrap (v4.3.1): util/index.js
 * Licensed under MIT (https://github.com/twbs/bootstrap/blob/master/LICENSE)
 * --------------------------------------------------------------------------
 */

const MAX_UID = 1000000
const MILLISECONDS_MULTIPLIER = 1000
const TRANSITION_END = 'transitionend'

// Shoutout AngusCroll (https://goo.gl/pxwQGp)
const toType = obj => ({}.toString.call(obj).match(/\s([a-z]+)/i)[1].toLowerCase())

/**
 * --------------------------------------------------------------------------
 * Public Util Api
 * --------------------------------------------------------------------------
 */

const getUID = prefix => {
  do {
    prefix += ~~(Math.random() * MAX_UID) // "~~" acts like a faster Math.floor() here
  } while (document.getElementById(prefix))

  return prefix
}

const getSelector = element => {
  let selector = element.getAttribute('data-target')

  if (!selector || selector === '#') {
    const hrefAttr = element.getAttribute('href')

    selector = hrefAttr && hrefAttr !== '#' ? hrefAttr.trim() : null
  }

  return selector
}

const getSelectorFromElement = element => {
  const selector = getSelector(element)

  if (selector) {
    return document.querySelector(selector) ? selector : null
  }

  return null
}

const getElementFromSelector = element => {
  const selector = getSelector(element)

  return selector ? document.querySelector(selector) : null
}

const getTransitionDurationFromElement = element => {
  if (!element) {
    return 0
  }

  // Get transition-duration of the element
  let {
    transitionDuration,
    transitionDelay
  } = window.getComputedStyle(element)

  const floatTransitionDuration = parseFloat(transitionDuration)
  const floatTransitionDelay = parseFloat(transitionDelay)

  // Return 0 if element or transition duration is not found
  if (!floatTransitionDuration && !floatTransitionDelay) {
    return 0
  }

  // If multiple durations are defined, take the first
  transitionDuration = transitionDuration.split(',')[0]
  transitionDelay = transitionDelay.split(',')[0]

  return (parseFloat(transitionDuration) + parseFloat(transitionDelay)) * MILLISECONDS_MULTIPLIER
}

const triggerTransitionEnd = element => {
  const evt = document.createEvent('HTMLEvents')

  evt.initEvent(TRANSITION_END, true, true)
  element.dispatchEvent(evt)
}

const isElement = obj => (obj[0] || obj).nodeType

const emulateTransitionEnd = (element, duration) => {
  let called = false
  const durationPadding = 5
  const emulatedDuration = duration + durationPadding
  function listener() {
    called = true
    element.removeEventListener(TRANSITION_END, listener)
  }

  element.addEventListener(TRANSITION_END, listener)
  setTimeout(() => {
    if (!called) {
      triggerTransitionEnd(element)
    }
  }, emulatedDuration)
}

const typeCheckConfig = (componentName, config, configTypes) => {
  Object.keys(configTypes)
    .forEach(property => {
      const expectedTypes = configTypes[property]
      const value = config[property]
      const valueType = value && isElement(value) ?
        'element' :
        toType(value)

      if (!new RegExp(expectedTypes).test(valueType)) {
        throw new Error(
          `${componentName.toUpperCase()}: ` +
          `Option "${property}" provided type "${valueType}" ` +
          `but expected type "${expectedTypes}".`)
      }
    })
}

const makeArray = nodeList => {
  if (!nodeList) {
    return []
  }

  return [].slice.call(nodeList)
}

const isVisible = element => {
  if (!element) {
    return false
  }

  if (element.style && element.parentNode && element.parentNode.style) {
    const elementStyle = getComputedStyle(element)
    const parentNodeStyle = getComputedStyle(element.parentNode)

    return elementStyle.display !== 'none' &&
      parentNodeStyle.display !== 'none' &&
      elementStyle.visibility !== 'hidden'
  }

  return false
}

const findShadowRoot = element => {
  if (!document.documentElement.attachShadow) {
    return null
  }

  // Can find the shadow root otherwise it'll return the document
  if (typeof element.getRootNode === 'function') {
    const root = element.getRootNode()
    return root instanceof ShadowRoot ? root : null
  }

  if (element instanceof ShadowRoot) {
    return element
  }

  // when we don't find a shadow root
  if (!element.parentNode) {
    return null
  }

  return findShadowRoot(element.parentNode)
}

const noop = () => function () {}

const reflow = element => element.offsetHeight

const getjQuery = () => {
  const { jQuery } = window

  if (jQuery && !document.body.hasAttribute('data-no-jquery')) {
    return jQuery
  }

  return null
}

export {
  getjQuery,
  TRANSITION_END,
  getUID,
  getSelectorFromElement,
  getElementFromSelector,
  getTransitionDurationFromElement,
  triggerTransitionEnd,
  isElement,
  emulateTransitionEnd,
  typeCheckConfig,
  makeArray,
  isVisible,
  findShadowRoot,
  noop,
  reflow
}
