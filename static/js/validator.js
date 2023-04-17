function setValidatorStatus(state, activationEpoch) {
  // deposited, deposited_valid, deposited_invalid, pending, active_online, active_offline, exiting_online, exiting_offline, slashing_online, slashing_offline, exited, slashed
  // we cans set elements to active, failed and done
  var status = state

  var depositToPending = document.querySelector(".validator__lifecycle-progress.validator__lifecycle-deposited")
  var pendingToActive = document.querySelector(".validator__lifecycle-progress.validator__lifecycle-pending")
  var activeToExited = document.querySelector(".validator__lifecycle-progress.validator__lifecycle-active")
  var depositedNode = document.getElementById("lifecycle-deposited")
  var pendingNode = document.getElementById("lifecycle-pending")
  var activeNode = document.getElementById("lifecycle-active")
  var exitedNode = document.getElementById("lifecycle-exited")

  var depositedDoneSet = new Set(["pending", "active_online", "active_offline", "exiting_online", "exiting_offline", "slashing_online", "slashing_offline", "exited", "slashed"])
  var pendingDoneSet = new Set(["active_online", "active_offline", "exiting_online", "exiting_offline", "slashing_online", "slashing_offline", "exited", "slashed"])
  var activeToExitSet = new Set(["exiting_online", "exiting_offline", "slashing_online", "slashing_offline"])
  var activeOnlineSet = new Set(["active_online", "exiting_online", "slashing_online"])
  var activeOfflineSet = new Set(["active_offline", "exiting_offline", "slashing_offline"])
  var activeDoneSet = new Set(["exited", "slashed"])

  if (depositedDoneSet.has(status)) {
    depositedNode.classList.add("done")
    depositToPending.classList.add("complete")
  }

  if (pendingDoneSet.has(status)) {
    pendingNode.classList.add("done")
    pendingToActive.classList.add("complete")
  }

  if (activeDoneSet.has(status)) {
    activeNode.classList.add("done")
    activeToExited.classList.add("complete")
  }

  if (activeToExitSet.has(status)) {
    activeToExited.classList.add("active")
    exitedNode.classList.add("active")
    if (status === "slashing_online" || status === "slashing_offline") exitedNode.classList.add("slashed")
  }

  if (activeOnlineSet.has(status)) {
    activeNode.classList.add("online")
  }

  if (activeOfflineSet.has(status)) {
    activeNode.classList.add("offline")
  }

  if (status === "slashed") {
    exitedNode.classList.add("failed")
  }

  if (status === "exited") {
    exitedNode.classList.add("done")
  }

  if (status === "deposited") {
    depositedNode.classList.add("active")
  }

  if (status === "deposited_valid") {
    depositedNode.classList.add("done")
    depositToPending.classList.add("active")
    pendingNode.classList.add("active")
  }

  if (status === "deposited_invalid") {
    depositedNode.classList.add("failed")
  }

  if (status === "pending") {
    if (activationEpoch > 100_000_000) {
      depositedNode.classList.add("done")
      depositToPending.classList.add("complete")
      pendingNode.classList.add("active")
    } else {
      pendingNode.classList.add("done")
      pendingToActive.classList.add("active")
      activeNode.classList.add("active")
    }
  }
}

function setupDashboardButtons(validatorIdx) {
  var validators = []
  $(document).ready(function () {
    updateDashboardButtons()
    $("#remove-from-dashboard-button").click(function () {
      validators = validators.filter(function (v, i, a) {
        if (v === validatorIdx) return false
        return true
      })
      validators.sort(sortValidators)
      localStorage.setItem("dashboard_validators", JSON.stringify(validators))
      $("#add-to-dashboard-button").show()
      $("#remove-from-dashboard-button").hide()
      $(this).tooltip("hide")
    })
    $("#add-to-dashboard-button").click(function () {
      validators.push(validatorIdx)
      validators.sort(sortValidators)
      localStorage.setItem("dashboard_validators", JSON.stringify(validators))
      $("#add-to-dashboard-button").hide()
      $("#remove-from-dashboard-button").show()
      $(this).tooltip("hide")
    })
  })
  window.addEventListener("storage", function (e) {
    // note: this fires only if storage changes within another tab
    updateDashboardButtons()
  })
  function updateDashboardButtons() {
    var validatorIsInDashboard = false
    var validatorsStr = localStorage.getItem("dashboard_validators")
    if (validatorsStr) {
      try {
        validators = JSON.parse(validatorsStr)
      } catch (e) {
        console.error("error parsing localStorage.dashboard_validators", e)
        validators = []
      }
    } else {
      validators = []
    }
    for (var i = 0; i < validators.length; i++) {
      if (validators[i] === validatorIdx) {
        validatorIsInDashboard = true
        break
      }
    }
    $("#remove-from-dashboard-button").hide()
    $("#add-to-dashboard-button").hide()
    if (validatorIsInDashboard) {
      $("#remove-from-dashboard-button").show()
    } else {
      $("#add-to-dashboard-button").show()
    }
  }
  function sortValidators(a, b) {
    var ai = parseInt(a)
    var bi = parseInt(b)
    return ai - bi
  }
}
