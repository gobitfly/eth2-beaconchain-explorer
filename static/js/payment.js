  // In production, this should check CSRF, and not pass the session ID.
  // The customer ID for the portal should be pulled from the 
  // authenticated user on the server.
  var manageBillingForm = document.querySelectorAll('.manage-billing-form');
  for (let i = 0; i < manageBillingForm.length; i++) {
    manageBillingForm[i].addEventListener('submit', function(e) {
      console.log('submitting manage billing form')
      e.preventDefault();
      fetch('/stripe/customer-portal', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          returnURL: window.location.href
        }),
      })
        .then((response) => response.json())
        .then((data) => {
          window.location.href = data.url;
        })
        .catch((error) => {
          console.error('Error:', error);
        });
    });
  }

