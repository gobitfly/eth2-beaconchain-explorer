import Axios from "axios"

let BACKEND_URL

//check if is on server
if (import.meta.env.SSR) BACKEND_URL = "http://localhost:8888"
else BACKEND_URL = window.location.origin

export const getLaunchMetrics = async () => {
  const { data } = await Axios.get(BACKEND_URL + "/mock/launchMetrics")
  return data
}

export const getIndexData = async () => {
  const { data } = await Axios.get(BACKEND_URL + "/mock/indexData")
  return data
}

export const getGitcoinfeed = async () => {
  const { data } = await Axios.get(BACKEND_URL + "/mock/gitcoinfeed")
  return data
}

export const getLatestState = async () => {
  const { data } = await Axios.get(BACKEND_URL + "/mock/latestState")
  return data
}

export const getPageData = async () => {
  const { data } = await Axios.get(BACKEND_URL + "/mock/getPageData")
  return data
}
