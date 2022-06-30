// In production, this should check CSRF, and not pass the session ID.
// The customer ID for the portal should be pulled from the
// authenticated user on the server.
var manageBillingForm = document.querySelectorAll(".manage-billing-form")
for (let i = 0; i < manageBillingForm.length; i++) {
  manageBillingForm[i].addEventListener("submit", function (e) {
    e.preventDefault()
    var token = ""
    if (document.getElementsByName("CsrfField") && document.getElementsByName("CsrfField").length) {
      token = document.getElementsByName("CsrfField")[0].value
    }
    fetch("/user/stripe/customer-portal", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": token,
      },
      credentials: "include",
      body: JSON.stringify({
        returnURL: window.location.href,
      }),
    })
      .then((response) => response.json())
      .then((data) => {
        window.location.href = data.url
      })
      .catch((error) => {
        console.error("Error:", error)
      })
  })
}
