package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v72"
	portalsession "github.com/stripe/stripe-go/v72/billingportal/session"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/webhook"
)

// StripeCreateCheckoutSession creates a session to checkout api pricing subscription
func StripeCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	user := getUser(w, r)

	// check if a subscription exists
	subscription, err := db.GetUserSubscription(user.UserID)
	if err != nil {
		logger.Errorf("error retrieving user subscriptions %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	// don't let the user checkout another subscription
	// changing subscription is not yet supported
	if subscription.Active || subscription.SubscriptionID != nil {
		logger.Errorf("error there is an active subscription cannot create another one %v", err)
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, struct {
			ErrorData string `json:"error"`
		}{
			ErrorData: "could not create a new stripe session",
		})
		return
	}

	// get the product that the user wants to subscribe to
	var req struct {
		Price string `json:"priceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(utils.Config.Frontend.SiteDomain + "/user/settings"),
		CancelURL:  stripe.String(utils.Config.Frontend.SiteDomain + "/pricing"),
		// if the customer exists use the existing customer
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			// DefaultTaxRates: stripe.StringSlice([]string{
			// "txr_1HqcFcBiORp9oTlKnyNWVp4r",
			// "txr_1HqdWaBiORp9oTlKkij8L6dU",
			// }),
		},
		CustomerEmail: &subscription.Email,
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(req.Price),
				Quantity: stripe.Int64(1),
			},
		},
	}

	if subscription.CustomerID != nil {
		params.CustomerEmail = nil
		params.Customer = subscription.CustomerID
	}

	s, err := session.New(params)
	if err != nil {
		logger.WithError(err).Error("failed to create a new stripe checkout session")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, struct {
			ErrorData string `json:"error"`
		}{
			ErrorData: "could not create a new stripe session",
		})
		return
	}

	writeJSON(w, struct {
		SessionID string `json:"sessionId"`
	}{
		SessionID: s.ID,
	})
}

// StripeCustomerPortal redirects the user to the customer portal
func StripeCustomerPortal(w http.ResponseWriter, r *http.Request) {
	user := getUser(w, r)

	var req struct {
		ReturnURL string `json:"returnURL"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.WithError(err).Error("json.NewDecoder.Decode")
		return
	}

	var customerID string
	err := db.FrontendDB.Get(&customerID, `
	SELECT
		stripe_customerID
	FROM
		users
	WHERE
		users.id = $1
	`, user.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.WithError(err).Error("error could not retrieve stripe customer id")
		return
	}
	// The URL to which the user is redirected when they are done managing
	// billing in the portal.

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(req.ReturnURL),
	}
	ps, _ := portalsession.New(params)

	writeJSON(w, struct {
		URL string `json:"url"`
	}{
		URL: ps.URL,
	})
}

// StripeWebhook receive events from stripe webhook service
func StripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.WithError(err).Error("error failed to read body for StripeWebhook")
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), utils.Config.Frontend.Stripe.Webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.WithError(err).Error("error constructing webhook stripe signature event")
		return
	}

	switch event.Type {
	case "customer.created":
		var customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}
		if customer.Email != "" {
			err = db.UpdateStripeCustomer(customer.Email, customer.ID)
			if err != nil {
				logger.WithError(err).Error("error could not update user with a stripe customerID")
				http.Error(w, "Internal server error", 503)
				return
			}
		} else {
			logger.Error("error no email provided when creating stripe customer")
			http.Error(w, "Internal server error", 503)
			return
		}

	case "customer.deleted":
		var customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}

		err = db.UpdateRemoveStripeCustomer(customer.ID)
		if err != nil {
			logger.WithError(err).Error("error could not update user with a stripe customerID")
			http.Error(w, "Internal server error", 503)
			return
		}

	case "checkout.session.completed":
		// Payment is successful and the subscription is created.
		// You should provision the subscription.
		// inform the user that the payment is being processed
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}

		// if session.Customer.Email != "" {
		// 	err = db.UpdateStripeCustomer(session.Customer.Email, session.Customer.ID)
		// 	if err != nil {
		// 		logger.WithError(err).Error("error could not update user with a stripe customerID")
		// 		http.Error(w, "Internal server error", 503)
		// 		return
		// 	}
		// } else {
		// 	logger.Error("the session object does not have a customer email", session, session.Customer)
		// 	http.Error(w, "Internal server error", 503)
		// 	return
		// }

	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}

		if subscription.Items == nil {
			logger.WithError(err).Error("error creating subscription no items found", subscription)
			http.Error(w, "Internal server error", 503)
			return
		}

		if len(subscription.Items.Data) == 0 {
			logger.WithError(err).Error("error creating subscription no items found", subscription)
			http.Error(w, "Internal server error", 503)
			return
		}

		err = db.UpdateAddSubscription(subscription.Customer.ID, subscription.Items.Data[0].Price.ID, subscription.ID)
		if err != nil {
			logger.WithError(err).Error("error updating user with subscription", event.Data.Object)
			http.Error(w, "Internal server error", 503)
			return
		}

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}

		if subscription.Items == nil {
			logger.WithError(err).Error("error creating subscription no items found", subscription)
			http.Error(w, "Internal server error", 503)
			return
		}

		if len(subscription.Items.Data) == 0 {
			logger.WithError(err).Error("error creating subscription no items found", subscription)
			http.Error(w, "Internal server error", 503)
			return
		}

		priceID := subscription.Items.Data[0].Price.ID

		currPriceID, err := db.GetUserPriceID(subscription.Customer.ID)
		if err != nil && err != sql.ErrNoRows {
			logger.WithError(err).Error("error retrieving customers priceID ", subscription.Customer.ID)
			http.Error(w, "Internal server error", 503)
			return
		}

		err = db.UpdateAddSubscription(subscription.Customer.ID, priceID, subscription.ID)
		if err != nil {
			logger.WithError(err).Error("error updating user with subscription", event.Data.Object)
			http.Error(w, "Internal server error", 503)
			return
		}

		if currPriceID != nil && *currPriceID != priceID {
			EmailCustomerAboutPlanChange(subscription.Customer.Email)
		}

	case "customer.subscription.deleted":
		// delete customer token
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}

		err = db.UpdateRemoveSubscription(subscription.Customer.ID)
		if err != nil {
			logger.WithError(err).Error("error updating user to remove subscription", event.Data.Object)
			http.Error(w, "Internal server error", 503)
			return
		}

	// inform the user when the subscription will expire
	case "invoice.paid":
		// Continue to provision the subscription as payments continue to be made.
		// Store the status in your database and check when a user accesses your service.
		// This approach helps you avoid hitting rate limits.
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = db.UpdateActivateSubsciption(invoice.Customer.ID)
		if err != nil {
			logger.WithError(err).Error("error failed to activate subscription for customer", invoice.Customer.ID)
			http.Error(w, "Internal server error", 503)
			return
		}

	case "invoice.payment_failed":
		// The payment failed or the customer does not have a valid payment method.
		// The subscription becomes past_due. Notify your customer and send them to the
		// customer portal to update their payment information.
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "Internal server error", 503)
			return
		}
		EmailCustomerAboutFailedPayment(invoice.CustomerEmail)
	default:
		return
		// unhandled event type
	}
}

func EmailCustomerAboutFailedPayment(email string) {
	msg := fmt.Sprintf("Payment processing failed. Could not provision your API key. Please contact support at support@beaconcha.in.")
	// escape html
	msg = template.HTMLEscapeString(msg)
	err := mail.SendMail(email, "Failed Payment", msg)
	if err != nil {
		logger.Errorf("error sending failed payment mail: %v", err)
		return
	}
}

func EmailCustomerAboutPlanChange(email string) {
	msg := fmt.Sprintf("You have successfully changed your payment plan")
	// escape html
	msg = template.HTMLEscapeString(msg)
	err := mail.SendMail(email, "Payment Plan Change", msg)
	if err != nil {
		logger.Errorf("error sending order fulfillment email: %v", err)
		return
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.WithError(err).Error("error failed to writeJSON NewEncoder")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		logger.WithError(err).Error("error failed to writeJSON io.Copy")
		return
	}
}
