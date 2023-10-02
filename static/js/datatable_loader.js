/**
 * function dataTableLoader(path, param, dataSrc)
 * used to create an ajax function for a DataTable.
 * This function is used to load data from the server for a DataTable.
 * It is debounced to avoid multiple requests to the server when the user clicks on the pagination buttons.
 * There is also a retry mechanism that retries the server request if the request fails.
 * @param path: Path to load the table data from the server
 * @param param: Additional parameters to be passed to the server
 * @param dataSrc: Function to be called on the data returned from the server
 */
function dataTableLoader(path, param, dataSrc) {
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
    fetch(`${path}?${$.param(tableData)}&${param}`)
      .then((response) => {
        if (!response.ok) {
          throw new Error(`Failed with status: ${response.status}`)
        }
        return response.json()
      })
      .then((data) => {
        if (typeof dataSrc === "function") {
          data = dataSrc({ data })
        }
        callback(data)
      })
      .catch((err) => {
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
    doFetch(tableData, callback)
  }
  const debouncedFetchData = debounce(fetchWithRetry, DEBOUNCE_DELAY)

  return debouncedFetchData
}
