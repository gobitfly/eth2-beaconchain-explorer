/**
 * function dataTableLoader(path)
 * used to create an axja function for a DataTable
 * if there is currently a request pending and a new one comes in then the old one will be canceled
 * and the new one will start loading after a timeout
 * if during that timeout a new request comes in the it's request will not be sent to the server
 * and the new request will start its timout ....
 * There is also a retry mechanism that retries (for up to 5 tries ) the server request if the request fails.
 * @param path: Path to load the table data from the server
 * */
function dataTableLoader(path) {
  var currentLoader

  var createLoader = (tableData, callback, instant) => {
    var isLoading = true

    var start = async (counter) => {
      if (!instant || counter) {
        await new Promise((resolve) =>
          setTimeout(async () => {
            resolve()
          }, 1000 + 1000 * counter)
        )
      }
      if (!isLoading) {
        return
      }
      try {
        const response = await fetch(`${path}?${new URLSearchParams(tableData)}`)
        const result = await response.json()
        if (!isLoading) {
          return
        }
        isLoading = false
        callback(result)
      } catch (e) {
        if (counter < 5) {
          console.warn(`Could not load [${path}] data will try again`)
          start(++counter)
        } else {
          console.error(`Could not load [${path}] data after 5 tries`, e)
        }
      }
    }
    start(0)
    return {
      stop: () => (isLoading = false),
      isLoading: () => isLoading,
    }
  }

  return (tableData, callback) => {
    if (currentLoader?.isLoading()) {
      currentLoader.stop()
    }
    var nextLoader = createLoader(tableData, callback, !currentLoader?.isLoading)
    currentLoader = nextLoader
  }
}
