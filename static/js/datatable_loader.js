/**
 * function dataTableLoader(path)
 * used to create an ajax function for a DataTable.
 * This function is used to load data from the server for a DataTable.
 * It is debounced to avoid multiple requests to the server when the user clicks on the pagination buttons.
 * There is also a retry mechanism that retries the server request if the request fails.
 * @param path: Path to load the table data from the server
 */
function dataTableLoader(path) {
  const MAX_RETRIES = 5
  const DEBOUNCE_DELAY = 500
  const RETRY_DELAY = 1000

  let retries = 0
  let timeoutId

  const debounce = (func, delay) => {
    let debounceTimer
    return function () {
      const context = this
      const args = arguments
      clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => func.apply(context, args), delay)
    }
  }

  const doFetch = (tableData, callback) => {
    fetch(`${path}?${new URLSearchParams(tableData)}`, {headers: {'X-TURNSTILE-TOKEN': window.turnstileToken}})
      .then((response) => {
        if (!response.ok) {
          throw new Error(`Failed with status: ${response.status}`)
        }
        if (window.turnstile) {
          console.log("resetting turnstile")
          window.turnstile.reset(window.turnstileWidgetId);
          window.turnstileToken = "";
        }
        return response.json()
      })
      .then((data) => {
        callback(data)
      })
      .catch((err) => {
        console.log(err)
        if (retries < MAX_RETRIES) {
          retries++
          timeoutId = setTimeout(() => doFetch(tableData, callback), RETRY_DELAY * (retries + 1))
        } else {
          console.error("Failed to fetch data for path: ", path, "with error: ", err)
        }
      })
  }

  const fetchWithRetry = (tableData, callback) => {
    clearTimeout(timeoutId) // Clear any pending retries.
    retries = 0 // Reset retry count.

    function waitForTurnstileToken() {
      if(window.turnstileSiteKey && !window.turnstileToken) {//we want it to match
          setTimeout(waitForTurnstileToken, 50);//wait 50 milliseconds then recheck
          return;
      }
      doFetch(tableData, callback)
    }

    waitForTurnstileToken()
  }
  const debouncedFetchData = debounce(fetchWithRetry, DEBOUNCE_DELAY)

  return debouncedFetchData
}
