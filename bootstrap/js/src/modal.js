/**
 * --------------------------------------------------------------------------
 * Bootstrap (v4.3.1): modal.js
 * Licensed under MIT (https://github.com/twbs/bootstrap/blob/master/LICENSE)
 * --------------------------------------------------------------------------
 */

import {
  getjQuery,
  TRANSITION_END,
  emulateTransitionEnd,
  getElementFromSelector,
  getTransitionDurationFromElement,
  isVisible,
  makeArray,
  reflow,
  typeCheckConfig
} from './util/index'
import Data from './dom/data'
import EventHandler from './dom/event-handler'
import Manipulator from './dom/manipulator'
import SelectorEngine from './dom/selector-engine'

/**
 * ------------------------------------------------------------------------
 * Constants
 * ------------------------------------------------------------------------
 */

const NAME = 'modal'
const VERSION = '4.3.1'
const DATA_KEY = 'bs.modal'
const EVENT_KEY = `.${DATA_KEY}`
const DATA_API_KEY = '.data-api'
const ESCAPE_KEYCODE = 27 // KeyboardEvent.which value for Escape (Esc) key

const Default = {
  backdrop: true,
  keyboard: true,
  focus: true,
  show: true
}

const DefaultType = {
  backdrop: '(boolean|string)',
  keyboard: 'boolean',
  focus: 'boolean',
  show: 'boolean'
}

const Event = {
  HIDE: `hide${EVENT_KEY}`,
  HIDE_PREVENTED: `hidePrevented${EVENT_KEY}`,
  HIDDEN: `hidden${EVENT_KEY}`,
  SHOW: `show${EVENT_KEY}`,
  SHOWN: `shown${EVENT_KEY}`,
  FOCUSIN: `focusin${EVENT_KEY}`,
  RESIZE: `resize${EVENT_KEY}`,
  CLICK_DISMISS: `click.dismiss${EVENT_KEY}`,
  KEYDOWN_DISMISS: `keydown.dismiss${EVENT_KEY}`,
  MOUSEUP_DISMISS: `mouseup.dismiss${EVENT_KEY}`,
  MOUSEDOWN_DISMISS: `mousedown.dismiss${EVENT_KEY}`,
  CLICK_DATA_API: `click${EVENT_KEY}${DATA_API_KEY}`
}

const ClassName = {
  SCROLLABLE: 'modal-dialog-scrollable',
  SCROLLBAR_MEASURER: 'modal-scrollbar-measure',
  BACKDROP: 'modal-backdrop',
  OPEN: 'modal-open',
  FADE: 'fade',
  SHOW: 'show',
  STATIC: 'modal-static'
}

const Selector = {
  DIALOG: '.modal-dialog',
  MODAL_BODY: '.modal-body',
  DATA_TOGGLE: '[data-toggle="modal"]',
  DATA_DISMISS: '[data-dismiss="modal"]',
  FIXED_CONTENT: '.fixed-top, .fixed-bottom, .is-fixed, .sticky-top',
  STICKY_CONTENT: '.sticky-top'
}

/**
 * ------------------------------------------------------------------------
 * Class Definition
 * ------------------------------------------------------------------------
 */

class Modal {
  constructor(element, config) {
    this._config = this._getConfig(config)
    this._element = element
    this._dialog = SelectorEngine.findOne(Selector.DIALOG, element)
    this._backdrop = null
    this._isShown = false
    this._isBodyOverflowing = false
    this._ignoreBackdropClick = false
    this._isTransitioning = false
    this._scrollbarWidth = 0
    Data.setData(element, DATA_KEY, this)
  }

  // Getters

  static get VERSION() {
    return VERSION
  }

  static get Default() {
    return Default
  }

  // Public

  toggle(relatedTarget) {
    return this._isShown ? this.hide() : this.show(relatedTarget)
  }

  show(relatedTarget) {
    if (this._isShown || this._isTransitioning) {
      return
    }

    if (this._element.classList.contains(ClassName.FADE)) {
      this._isTransitioning = true
    }

    const showEvent = EventHandler.trigger(this._element, Event.SHOW, {
      relatedTarget
    })

    if (this._isShown || showEvent.defaultPrevented) {
      return
    }

    this._isShown = true

    this._checkScrollbar()
    this._setScrollbar()

    this._adjustDialog()

    this._setEscapeEvent()
    this._setResizeEvent()

    EventHandler.on(this._element,
      Event.CLICK_DISMISS,
      Selector.DATA_DISMISS,
      event => this.hide(event)
    )

    EventHandler.on(this._dialog, Event.MOUSEDOWN_DISMISS, () => {
      EventHandler.one(this._element, Event.MOUSEUP_DISMISS, event => {
        if (event.target === this._element) {
          this._ignoreBackdropClick = true
        }
      })
    })

    this._showBackdrop(() => this._showElement(relatedTarget))
  }

  hide(event) {
    if (event) {
      event.preventDefault()
    }

    if (!this._isShown || this._isTransitioning) {
      return
    }

    const hideEvent = EventHandler.trigger(this._element, Event.HIDE)

    if (hideEvent.defaultPrevented) {
      return
    }

    this._isShown = false
    const transition = this._element.classList.contains(ClassName.FADE)

    if (transition) {
      this._isTransitioning = true
    }

    this._setEscapeEvent()
    this._setResizeEvent()

    EventHandler.off(document, Event.FOCUSIN)

    this._element.classList.remove(ClassName.SHOW)

    EventHandler.off(this._element, Event.CLICK_DISMISS)
    EventHandler.off(this._dialog, Event.MOUSEDOWN_DISMISS)

    if (transition) {
      const transitionDuration = getTransitionDurationFromElement(this._element)

      EventHandler.one(this._element, TRANSITION_END, event => this._hideModal(event))
      emulateTransitionEnd(this._element, transitionDuration)
    } else {
      this._hideModal()
    }
  }

  dispose() {
    [window, this._element, this._dialog]
      .forEach(htmlElement => EventHandler.off(htmlElement, EVENT_KEY))

    /**
     * `document` has 2 events `Event.FOCUSIN` and `Event.CLICK_DATA_API`
     * Do not move `document` in `htmlElements` array
     * It will remove `Event.CLICK_DATA_API` event that should remain
     */
    EventHandler.off(document, Event.FOCUSIN)

    Data.removeData(this._element, DATA_KEY)

    this._config = null
    this._element = null
    this._dialog = null
    this._backdrop = null
    this._isShown = null
    this._isBodyOverflowing = null
    this._ignoreBackdropClick = null
    this._isTransitioning = null
    this._scrollbarWidth = null
  }

  handleUpdate() {
    this._adjustDialog()
  }

  // Private

  _getConfig(config) {
    config = {
      ...Default,
      ...config
    }
    typeCheckConfig(NAME, config, DefaultType)
    return config
  }

  _showElement(relatedTarget) {
    const transition = this._element.classList.contains(ClassName.FADE)
    const modalBody = SelectorEngine.findOne(Selector.MODAL_BODY, this._dialog)

    if (!this._element.parentNode ||
        this._element.parentNode.nodeType !== Node.ELEMENT_NODE) {
      // Don't move modal's DOM position
      document.body.appendChild(this._element)
    }

    this._element.style.display = 'block'
    this._element.removeAttribute('aria-hidden')
    this._element.setAttribute('aria-modal', true)

    if (this._dialog.classList.contains(ClassName.SCROLLABLE) && modalBody) {
      modalBody.scrollTop = 0
    } else {
      this._element.scrollTop = 0
    }

    if (transition) {
      reflow(this._element)
    }

    this._element.classList.add(ClassName.SHOW)

    if (this._config.focus) {
      this._enforceFocus()
    }

    const transitionComplete = () => {
      if (this._config.focus) {
        this._element.focus()
      }

      this._isTransitioning = false
      EventHandler.trigger(this._element, Event.SHOWN, {
        relatedTarget
      })
    }

    if (transition) {
      const transitionDuration = getTransitionDurationFromElement(this._dialog)

      EventHandler.one(this._dialog, TRANSITION_END, transitionComplete)
      emulateTransitionEnd(this._dialog, transitionDuration)
    } else {
      transitionComplete()
    }
  }

  _enforceFocus() {
    EventHandler.off(document, Event.FOCUSIN) // guard against infinite focus loop
    EventHandler.on(document, Event.FOCUSIN, event => {
      if (document !== event.target &&
          this._element !== event.target &&
          !this._element.contains(event.target)) {
        this._element.focus()
      }
    })
  }

  _setEscapeEvent() {
    if (this._isShown && this._config.keyboard) {
      EventHandler.on(this._element, Event.KEYDOWN_DISMISS, event => {
        if (event.which === ESCAPE_KEYCODE) {
          this._triggerBackdropTransition()
        }
      })
    } else {
      EventHandler.off(this._element, Event.KEYDOWN_DISMISS)
    }
  }

  _setResizeEvent() {
    if (this._isShown) {
      EventHandler.on(window, Event.RESIZE, () => this._adjustDialog())
    } else {
      EventHandler.off(window, Event.RESIZE)
    }
  }

  _hideModal() {
    this._element.style.display = 'none'
    this._element.setAttribute('aria-hidden', true)
    this._element.removeAttribute('aria-modal')
    this._isTransitioning = false
    this._showBackdrop(() => {
      document.body.classList.remove(ClassName.OPEN)
      this._resetAdjustments()
      this._resetScrollbar()
      EventHandler.trigger(this._element, Event.HIDDEN)
    })
  }

  _removeBackdrop() {
    this._backdrop.parentNode.removeChild(this._backdrop)
    this._backdrop = null
  }

  _showBackdrop(callback) {
    const animate = this._element.classList.contains(ClassName.FADE) ?
      ClassName.FADE :
      ''

    if (this._isShown && this._config.backdrop) {
      this._backdrop = document.createElement('div')
      this._backdrop.className = ClassName.BACKDROP

      if (animate) {
        this._backdrop.classList.add(animate)
      }

      document.body.appendChild(this._backdrop)

      EventHandler.on(this._element, Event.CLICK_DISMISS, event => {
        if (this._ignoreBackdropClick) {
          this._ignoreBackdropClick = false
          return
        }

        if (event.target !== event.currentTarget) {
          return
        }

        this._triggerBackdropTransition()
      })

      if (animate) {
        reflow(this._backdrop)
      }

      this._backdrop.classList.add(ClassName.SHOW)

      if (!animate) {
        callback()
        return
      }

      const backdropTransitionDuration = getTransitionDurationFromElement(this._backdrop)

      EventHandler.one(this._backdrop, TRANSITION_END, callback)
      emulateTransitionEnd(this._backdrop, backdropTransitionDuration)
    } else if (!this._isShown && this._backdrop) {
      this._backdrop.classList.remove(ClassName.SHOW)

      const callbackRemove = () => {
        this._removeBackdrop()
        callback()
      }

      if (this._element.classList.contains(ClassName.FADE)) {
        const backdropTransitionDuration = getTransitionDurationFromElement(this._backdrop)
        EventHandler.one(this._backdrop, TRANSITION_END, callbackRemove)
        emulateTransitionEnd(this._backdrop, backdropTransitionDuration)
      } else {
        callbackRemove()
      }
    } else {
      callback()
    }
  }

  _triggerBackdropTransition() {
    if (this._config.backdrop === 'static') {
      const hideEvent = EventHandler.trigger(this._element, Event.HIDE_PREVENTED)
      if (hideEvent.defaultPrevented) {
        return
      }

      this._element.classList.add(ClassName.STATIC)
      const modalTransitionDuration = getTransitionDurationFromElement(this._element)
      EventHandler.one(this._element, TRANSITION_END, () => {
        this._element.classList.remove(ClassName.STATIC)
      })
      emulateTransitionEnd(this._element, modalTransitionDuration)
      this._element.focus()
    } else {
      this.hide()
    }
  }

  // ----------------------------------------------------------------------
  // the following methods are used to handle overflowing modals
  // ----------------------------------------------------------------------

  _adjustDialog() {
    const isModalOverflowing =
      this._element.scrollHeight > document.documentElement.clientHeight

    if (!this._isBodyOverflowing && isModalOverflowing) {
      this._element.style.paddingLeft = `${this._scrollbarWidth}px`
    }

    if (this._isBodyOverflowing && !isModalOverflowing) {
      this._element.style.paddingRight = `${this._scrollbarWidth}px`
    }
  }

  _resetAdjustments() {
    this._element.style.paddingLeft = ''
    this._element.style.paddingRight = ''
  }

  _checkScrollbar() {
    const rect = document.body.getBoundingClientRect()
    this._isBodyOverflowing = rect.left + rect.right < window.innerWidth
    this._scrollbarWidth = this._getScrollbarWidth()
  }

  _setScrollbar() {
    if (this._isBodyOverflowing) {
      // Note: DOMNode.style.paddingRight returns the actual value or '' if not set
      //   while $(DOMNode).css('padding-right') returns the calculated value or 0 if not set

      // Adjust fixed content padding
      makeArray(SelectorEngine.find(Selector.FIXED_CONTENT))
        .forEach(element => {
          const actualPadding = element.style.paddingRight
          const calculatedPadding = window.getComputedStyle(element)['padding-right']
          Manipulator.setDataAttribute(element, 'padding-right', actualPadding)
          element.style.paddingRight = `${parseFloat(calculatedPadding) + this._scrollbarWidth}px`
        })

      // Adjust sticky content margin
      makeArray(SelectorEngine.find(Selector.STICKY_CONTENT))
        .forEach(element => {
          const actualMargin = element.style.marginRight
          const calculatedMargin = window.getComputedStyle(element)['margin-right']
          Manipulator.setDataAttribute(element, 'margin-right', actualMargin)
          element.style.marginRight = `${parseFloat(calculatedMargin) - this._scrollbarWidth}px`
        })

      // Adjust body padding
      const actualPadding = document.body.style.paddingRight
      const calculatedPadding = window.getComputedStyle(document.body)['padding-right']

      Manipulator.setDataAttribute(document.body, 'padding-right', actualPadding)
      document.body.style.paddingRight = `${parseFloat(calculatedPadding) + this._scrollbarWidth}px`
    }

    document.body.classList.add(ClassName.OPEN)
  }

  _resetScrollbar() {
    // Restore fixed content padding
    makeArray(SelectorEngine.find(Selector.FIXED_CONTENT))
      .forEach(element => {
        const padding = Manipulator.getDataAttribute(element, 'padding-right')
        if (typeof padding !== 'undefined') {
          Manipulator.removeDataAttribute(element, 'padding-right')
          element.style.paddingRight = padding
        }
      })

    // Restore sticky content and navbar-toggler margin
    makeArray(SelectorEngine.find(`${Selector.STICKY_CONTENT}`))
      .forEach(element => {
        const margin = Manipulator.getDataAttribute(element, 'margin-right')
        if (typeof margin !== 'undefined') {
          Manipulator.removeDataAttribute(element, 'margin-right')
          element.style.marginRight = margin
        }
      })

    // Restore body padding
    const padding = Manipulator.getDataAttribute(document.body, 'padding-right')
    if (typeof padding === 'undefined') {
      document.body.style.paddingRight = ''
    } else {
      Manipulator.removeDataAttribute(document.body, 'padding-right')
      document.body.style.paddingRight = padding
    }
  }

  _getScrollbarWidth() { // thx d.walsh
    const scrollDiv = document.createElement('div')
    scrollDiv.className = ClassName.SCROLLBAR_MEASURER
    document.body.appendChild(scrollDiv)
    const scrollbarWidth = scrollDiv.getBoundingClientRect().width - scrollDiv.clientWidth
    document.body.removeChild(scrollDiv)
    return scrollbarWidth
  }

  // Static

  static jQueryInterface(config, relatedTarget) {
    return this.each(function () {
      let data = Data.getData(this, DATA_KEY)
      const _config = {
        ...Default,
        ...Manipulator.getDataAttributes(this),
        ...typeof config === 'object' && config ? config : {}
      }

      if (!data) {
        data = new Modal(this, _config)
      }

      if (typeof config === 'string') {
        if (typeof data[config] === 'undefined') {
          throw new TypeError(`No method named "${config}"`)
        }

        data[config](relatedTarget)
      } else if (_config.show) {
        data.show(relatedTarget)
      }
    })
  }

  static getInstance(element) {
    return Data.getData(element, DATA_KEY)
  }
}

/**
 * ------------------------------------------------------------------------
 * Data Api implementation
 * ------------------------------------------------------------------------
 */

EventHandler.on(document, Event.CLICK_DATA_API, Selector.DATA_TOGGLE, function (event) {
  const target = getElementFromSelector(this)

  if (this.tagName === 'A' || this.tagName === 'AREA') {
    event.preventDefault()
  }

  EventHandler.one(target, Event.SHOW, showEvent => {
    if (showEvent.defaultPrevented) {
      // only register focus restorer if modal will actually get shown
      return
    }

    EventHandler.one(target, Event.HIDDEN, () => {
      if (isVisible(this)) {
        this.focus()
      }
    })
  })

  let data = Data.getData(target, DATA_KEY)
  if (!data) {
    const config = {
      ...Manipulator.getDataAttributes(target),
      ...Manipulator.getDataAttributes(this)
    }

    data = new Modal(target, config)
  }

  data.show(this)
})

const $ = getjQuery()

/**
 * ------------------------------------------------------------------------
 * jQuery
 * ------------------------------------------------------------------------
 * add .modal to jQuery only if jQuery is present
 */
/* istanbul ignore if */
if ($) {
  const JQUERY_NO_CONFLICT = $.fn[NAME]
  $.fn[NAME] = Modal.jQueryInterface
  $.fn[NAME].Constructor = Modal
  $.fn[NAME].noConflict = () => {
    $.fn[NAME] = JQUERY_NO_CONFLICT
    return Modal.jQueryInterface
  }
}

export default Modal
