function getCookie(cname) {
  var name = cname + "="
  var ca = document.cookie.split(";")
  for (var i = 0; i < ca.length; i++) {
    var c = ca[i]
    while (c.charAt(0) == " ") {
      c = c.substring(1)
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length)
    }
  }
  return ""
}

function renderTurnStile() {
  if (!window.turnstile || getCookie("turnstile") === "verified") return
  console.debug("renderTurnStile")

  if (window.isRequestingTurnstileToken) return
  window.isRequestingTurnstileToken = true

  window.turnstileWidgetId = turnstile.render("#turnstileModalContent", {
    sitekey: window.turnstileSiteKey,
    theme: "auto",
    callback: function (token) {
      window.turnstileToken = token
      console.log(`Challenge Success ${token}`)
      verifyTurnStileToken(() => {
        window.isRequestingTurnstileToken = false
      })
    },
    "before-interactive-callback": function () {
      $("#turnstileModal").modal("show")
    },
    "after-interactive-callback": function () {
      $("#turnstileModal").modal("hide")
    },
    "error-callback": function (error) {
      //https://developers.cloudflare.com/turnstile/reference/client-side-errors/
      console.log(`error callback called with ${error}`)
      window.isRequestingTurnstileToken = false
    },
    "expired-callback": function () {
      window.isRequestingTurnstileToken = false
    },
    "timeout-callback": function () {
      window.isRequestingTurnstileToken = false
    },
    "unsupported-callback": function () {
      console.log("unsupported callback called")
    },
  })
}

function _turnstileCb() {
  // onSiteLoad turnstile iframe loaded, show modal if cookie not set
  console.debug("_turnstileCb called")
  renderTurnStile()
}

function waitForTurnstileToken(cb) {
  if (window.turnstileSiteKey && !window.turnstileToken && !window.isRequestingTurnstileToken) {
    renderTurnStile()
  }
  if (window.turnstileSiteKey && !window.turnstileToken && getCookie("turnstile") !== "verified") {
    //we want it to match
    setTimeout(waitForTurnstileToken.bind(this, cb), 50) //wait 50 milliseconds then recheck
    return
  } else {
    cb && cb()
  }
}

function verifyTurnStileToken(cb) {
  fetch("/turnstile/verify", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      "X-TURNSTILE-TOKEN": window.turnstileToken,
    },
  })
    .then((res) => {
      cb && cb()
    })
    .catch((err) => {
      console.log("error verifying turnstile token", err)
    })
}
// resetting a widget causes it to rerender
// rerender only if the cookie is not present and there has been no request already for a token
function resetTurnstileToken() {
  if (getCookie("turnstile") !== "verified") {
    if (window.isRequestingTurnstileToken) return
    window.isRequestingTurnstileToken = true
    if (window.turnstile) {
      if (window.turnstileWidgetId) {
        window.turnstile.reset(window.turnstileWidgetId)
        window.turnstileToken = ""
      }
    }
  }
}
