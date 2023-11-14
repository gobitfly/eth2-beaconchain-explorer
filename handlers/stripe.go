package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v72"
	portalsession "github.com/stripe/stripe-go/v72/billingportal/session"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/webhook"
)

func getCleanProductID(priceId string) string {
	if priceId == utils.Config.Frontend.Stripe.Whale {
		return "whale"
	}
	if priceId == utils.Config.Frontend.Stripe.Goldfish {
		return "goldfish"
	}
	if priceId == utils.Config.Frontend.Stripe.Plankton {
		return "plankton"
	}
	return ""
}

// StripeCreateCheckoutSession creates a session to checkout api pricing subscription
func StripeCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)

	// get the product that the user wants to subscribe to
	var req struct {
		Price string `json:"priceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Errorf("error decoding json.NewDecoder.Decode: %v", err)
		return
	}
	rq := "required"

	purchaseGroup := utils.GetPurchaseGroup(req.Price)

	if purchaseGroup == "" {
		http.Error(w, "Error invalid price item provided. Must be the price ID of Sapphire, Emerald or Diamond", http.StatusBadRequest)
		logger.Errorf("error invalid stripe price id provided: %v, expected one of [%v, %v, %v]", req.Price, utils.Config.Frontend.Stripe.Sapphire, utils.Config.Frontend.Stripe.Emerald, utils.Config.Frontend.Stripe.Diamond)
		return
	}

	// check if a subscription exists
	subscription, err := db.StripeGetUserSubscription(user.UserID, purchaseGroup)
	if err != nil {
		logger.Errorf("error retrieving user subscriptions %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// don't let the user checkout another subscription in the same group
	if subscription.Active != nil && *subscription.Active {
		logger.Errorf("error there is an active subscription cannot create another one %v", err)
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, struct {
			ErrorData string `json:"error"`
		}{
			ErrorData: "could not create a new stripe session",
		})
		return
	}

	// taxRates := utils.StripeDynamicRatesLive
	// if strings.HasPrefix(utils.Config.Frontend.Stripe.SecretKey, "sk_test") {
	// 	taxRates = utils.StripeDynamicRatesTest
	// }

	enabled := true
	auto := "auto"

	var successUrl = stripe.String("https://" + utils.Config.Frontend.SiteDomain + "/user/settings#api")
	var cancelUrl = stripe.String("https://" + utils.Config.Frontend.SiteDomain + "/pricing")
	if purchaseGroup == utils.GROUP_MOBILE {
		successUrl = stripe.String("https://" + utils.Config.Frontend.SiteDomain + "/user/settings#account")
		cancelUrl = stripe.String("https://" + utils.Config.Frontend.SiteDomain + "/premium")
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL: successUrl,
		CancelURL:  cancelUrl,
		// if the customer exists use the existing customer
		SubscriptionData:         &stripe.CheckoutSessionSubscriptionDataParams{},
		BillingAddressCollection: &rq,
		CustomerEmail:            &subscription.Email,
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		AllowPromotionCodes: &enabled,
		Mode:                stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.Price),
				Quantity: stripe.Int64(1),
				// DynamicTaxRates: taxRates,
			},
		},
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: &enabled,
		},
		TaxIDCollection: &stripe.CheckoutSessionTaxIDCollectionParams{
			Enabled: &enabled,
		},
	}
	if subscription.CustomerID != nil {
		params.CustomerEmail = nil
		params.Customer = subscription.CustomerID
		params.CustomerUpdate = &stripe.CheckoutSessionCustomerUpdateParams{
			Name:    &auto,
			Address: &auto,
		}
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
	user := getUser(r)

	var req struct {
		ReturnURL string `json:"returnURL"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.WithError(err).Error("json.NewDecoder.Decode")
		return
	}

	var customerID string
	err := db.FrontendWriterDB.Get(&customerID, `
	SELECT
		stripe_customer_id
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
	b, err := io.ReadAll(r.Body)
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
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if customer.Email == "" {
			utils.LogError(nil, fmt.Errorf("error no email provided when creating stripe customer %v", customer.ID), 0)
			http.Error(w, "error no email provided when creating stripe customer "+customer.ID, http.StatusBadRequest)
			return
		}
		err = db.StripeUpdateCustomerID(customer.Email, customer.ID)
		if err != nil {
			logger.WithError(err).Error("error could not update user with a stripe customerID ", customer.ID)
			http.Error(w, "error could not update user with a stripe customerID "+customer.ID+" err: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case "customer.deleted":
		var customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON", err)
			http.Error(w, "error parsing stripe webhook JSON", http.StatusInternalServerError)
			return
		}

		err = db.DisableAllSubscriptionsFromStripeUser(customer.ID)
		if err != nil {
			logger.WithError(err).Error("error could not disable stripe mobile subs: " + customer.ID + "err: ")
			// log & continue anyway
		}

		err = db.StripeRemoveCustomer(customer.ID)
		if err != nil {
			logger.WithError(err).Error("error could not delete user with customer ID: " + customer.ID + "err: ")
			http.Error(w, "error could not delete user with customer ID: "+customer.ID+"err: "+err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "error parsing stripe webhook JSON", http.StatusInternalServerError)
			return
		}

	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "error parsing stripe webhook JSON", http.StatusInternalServerError)
			return
		}

		if subscription.Items == nil {
			logger.WithError(err).Error("error creating subscription no items found", subscription)
			http.Error(w, "error creating subscription no items found", http.StatusBadRequest)
			return
		}

		if len(subscription.Items.Data) == 0 {
			logger.WithError(err).Error("error creating subscription no items found", subscription)
			http.Error(w, "error creating subscription no items found", http.StatusBadRequest)
			return
		}

		// to handle race condition errors where subscription.updated is executed before customer.subscription.created, do nothing since it's already processed
		_, err = db.StripeGetSubscription(subscription.ID)
		if err == sql.ErrNoRows {
			err = createNewStripeSubscription(subscription, event)
			if err != nil {
				logger.WithError(err).Error(err.Error(), event.Data.Object)
				http.Error(w, "error "+err.Error()+" customer: "+subscription.Customer.ID, http.StatusInternalServerError)
				return
			}
		}

	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "error parsing stripe webhook JSON", http.StatusInternalServerError)
			return
		}

		if subscription.Items == nil {
			utils.LogError(nil, fmt.Errorf("error updating subscription no items found %v", subscription), 0)
			http.Error(w, "error updating subscription no items found", http.StatusBadRequest)
			return
		}

		if len(subscription.Items.Data) == 0 {
			utils.LogError(nil, fmt.Errorf("error updating subscription no data found %v", subscription), 0)
			http.Error(w, "error updating subscription no items found", http.StatusBadRequest)
			return
		}
		priceID := subscription.Items.Data[0].Price.ID

		currSub, err := db.StripeGetSubscription(subscription.ID)

		if err != nil {
			logger.WithError(err).Error("error getting subscription from database with id ", subscription.ID)
			http.Error(w, "error updating subscription could not get current subscription err:"+err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := db.FrontendWriterDB.Begin()
		if err != nil {
			logger.WithError(err).Error("error creating transaction ", subscription.ID)
			http.Error(w, "error creating transaction :"+err.Error(), http.StatusInternalServerError)
		}
		defer tx.Rollback()

		err = db.StripeUpdateSubscription(tx, priceID, subscription.ID, event.Data.Raw)
		if err != nil {
			logger.WithError(err).Error("error updating user subscription", subscription.ID)
			http.Error(w, "error updating user subscription, customer: "+subscription.Customer.ID, http.StatusInternalServerError)
			return
		}

		if utils.GetPurchaseGroup(priceID) == utils.GROUP_MOBILE {
			err := db.ChangeProductIDFromStripe(tx, subscription.ID, getCleanProductID(priceID))
			if err != nil {
				logger.WithError(err).Error("error updating stripe mobile subscription", subscription.ID)
				http.Error(w, "error updating stripe mobile subscription customer: "+subscription.Customer.ID, http.StatusInternalServerError)
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			logger.WithError(err).Error("error committing transaction ", subscription.ID)
			http.Error(w, "error committing transaction :"+err.Error(), http.StatusInternalServerError)
		}

		if currSub.PriceID != nil && *currSub.PriceID != priceID && utils.GetPurchaseGroup(*currSub.PriceID) == utils.GetPurchaseGroup(priceID) {
			email, err := db.StripeGetCustomerEmail(subscription.Customer.ID)
			if err != nil {
				logger.WithError(err).Error("error retrieving customer email for subscription ", subscription.ID)
				http.Error(w, "error retrieving customer email for subscription err:"+err.Error(), http.StatusInternalServerError)
			}
			emailCustomerAboutPlanChange(email, priceID)
		}

	case "customer.subscription.deleted":
		// delete customer token
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "error parsing stripe webhook JSON", http.StatusInternalServerError)
			return
		}

		tx, err := db.FrontendWriterDB.Begin()
		if err != nil {
			logger.WithError(err).Error("error creating transaction ", subscription.ID)
			http.Error(w, "error creating transaction :"+err.Error(), http.StatusInternalServerError)
		}
		defer tx.Rollback()

		err = db.StripeUpdateSubscriptionStatus(tx, subscription.ID, false, &event.Data.Raw)
		if err != nil {
			logger.WithError(err).Error("error while deactivating subscription", event.Data.Object)
			http.Error(w, "error while deactivating subscription, customer:"+subscription.Customer.ID, http.StatusInternalServerError)
			return
		}

		if utils.GetPurchaseGroup(subscription.Items.Data[0].Price.ID) == utils.GROUP_MOBILE {
			appSubID, err := db.GetUserSubscriptionIDByStripe(subscription.ID)
			if err != nil {
				logger.WithError(err).Error("error updating stripe mobile subscription, no users_app_subs id found for subscription id", subscription.ID)
				http.Error(w, "error updating stripe mobile subscription, no users_app_subs id  found for subscription id, customer: "+subscription.Customer.ID, http.StatusInternalServerError)
				return
			}
			now := time.Now()
			nowTs := now.Unix()
			db.UpdateUserSubscription(tx, appSubID, false, nowTs, "user_canceled")
		}

		err = tx.Commit()
		if err != nil {
			logger.WithError(err).Error("error committing transaction ", subscription.ID)
			http.Error(w, "error committing transaction :"+err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if invoice.Lines == nil {
			logger.Warn("warning processing invoice and updating subscription no items found", invoice.ID)
			return
		}

		if len(invoice.Lines.Data) == 0 {
			logger.Warn("warning processing invoice and updating subscription no items found", invoice.ID)
			return
		}

		if len(invoice.Lines.Data[0].Subscription) == 0 {
			logger.Warn("error processing invoice and updating subscription no items found", invoice.ID)
			return
		}

		tx, err := db.FrontendWriterDB.Begin()
		if err != nil {
			logger.WithError(err).Error("error creating transaction ")
			http.Error(w, "error creating transaction :"+err.Error(), http.StatusInternalServerError)
		}
		defer tx.Rollback()

		err = db.StripeUpdateSubscriptionStatus(tx, invoice.Lines.Data[0].Subscription, true, nil)
		if err != nil {
			logger.WithError(err).Error("error processing invoice failed to activate subscription for customer", invoice.Customer.ID)
			http.Error(w, "error processing invoice failed to activate subscription for customer", http.StatusInternalServerError)
			return
		}

		if utils.GetPurchaseGroup(invoice.Lines.Data[0].Price.ID) == utils.GROUP_MOBILE {
			appSubID, err := db.GetUserSubscriptionIDByStripe(invoice.Lines.Data[0].Subscription)
			if err != nil {
				logger.WithError(err).Error("error updating stripe mobile subscription (paid), no users_app_subs id found for subscription id", invoice.Lines.Data[0].Subscription)
				http.Error(w, "error updating stripe mobile subscription, no users_app_subs id  found for subscription id, customer: "+invoice.Customer.ID, http.StatusInternalServerError)
				return
			}
			db.UpdateUserSubscription(tx, appSubID, true, 0, "")
		}

		err = tx.Commit()
		if err != nil {
			logger.WithError(err).Error("error committing transaction ")
			http.Error(w, "error committing transaction :"+err.Error(), http.StatusInternalServerError)
		}

	case "invoice.payment_failed":
		// The payment failed or the customer does not have a valid payment method.
		// The subscription becomes past_due. Notify your customer and send them to the
		// customer portal to update their payment information.
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			logger.WithError(err).Error("error parsing stripe webhook JSON")
			http.Error(w, "error parsing stripe webhook JSON", http.StatusInternalServerError)
			return
		}
		emailCustomerAboutFailedPayment(invoice.CustomerEmail)
	default:
		return
		// unhandled event type
	}
}

func createNewStripeSubscription(subscription stripe.Subscription, event stripe.Event) error {
	tx, err := db.FrontendWriterDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = db.StripeCreateSubscription(tx, subscription.Customer.ID, subscription.Items.Data[0].Price.ID, subscription.ID, event.Data.Raw)
	if err != nil {
		return err
	}

	if utils.GetPurchaseGroup(subscription.Items.Data[0].Price.ID) == utils.GROUP_MOBILE {
		userID, err := db.StripeGetCustomerUserId(subscription.Customer.ID)
		if err != nil {
			return err
		}
		details := types.MobileSubscription{
			ProductID:   getCleanProductID(subscription.Items.Data[0].Price.ID),
			PriceMicros: uint64(subscription.Items.Data[0].Price.UnitAmount),
			Currency:    string(subscription.Items.Data[0].Price.Currency),
			Transaction: types.MobileSubscriptionTransactionGeneric{
				Type:    "stripe",
				Receipt: subscription.ID,
				ID:      subscription.Items.Data[0].Price.ID,
			},
			Valid: false,
		}
		err = db.InsertMobileSubscription(tx, userID, details, details.Transaction.Type, details.Transaction.Receipt, 0, "", subscription.ID)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()

	return err
}

func emailCustomerAboutFailedPayment(email string) {
	msg := fmt.Sprintf("Payment processing failed. Could not activate your subscription. Please contact support at " + utils.Config.Frontend.Mail.Contact.SupportEmail + ". Manage Subscription: https://" + utils.Config.Frontend.SiteDomain + "/user/settings")
	// escape html
	msg = template.HTMLEscapeString(msg)
	err := mail.SendTextMail(email, "Failed Payment", msg, []types.EmailAttachment{})
	if err != nil {
		logger.Errorf("error sending failed payment mail: %v", err)
		return
	}
}

func emailCustomerAboutPlanChange(email, plan string) {
	p := "Sapphire"
	if plan == utils.Config.Frontend.Stripe.Emerald {
		p = "Emerald"
	} else if plan == utils.Config.Frontend.Stripe.Diamond {
		p = "Diamond"
	} else if plan == utils.Config.Frontend.Stripe.Plankton {
		p = "Plankton"
	} else if plan == utils.Config.Frontend.Stripe.Goldfish {
		p = "Goldfish"
	} else if plan == utils.Config.Frontend.Stripe.Whale {
		p = "Whale"
	}
	msg := fmt.Sprintf("You have successfully changed your payment plan to " + p + " to manage your subscription go to https://" + utils.Config.Frontend.SiteDomain + "/user/settings#api")
	// escape html
	msg = template.HTMLEscapeString(msg)
	err := mail.SendTextMail(email, "Payment Plan Change", msg, []types.EmailAttachment{})
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
