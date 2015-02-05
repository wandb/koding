package main

import (
	"encoding/json"
	"fmt"
	"socialapi/workers/payment/paymentemail"
	"socialapi/workers/payment/paymentwebhook/webhookmodels"
	"socialapi/workers/payment/stripe"
)

func stripePaymentSucceeded(raw []byte, c *Controller) error {
	var invoice *webhookmodels.StripeInvoice

	err := json.Unmarshal(raw, &invoice)
	if err != nil {
		return err
	}

	err = stripe.InvoiceCreatedWebhook(invoice)
	if err != nil {
		return err
	}

	return stripePaymentSucceededEmail(invoice, c)
}

func stripePaymentSucceededEmail(req *webhookmodels.StripeInvoice, c *Controller) error {
	emailAddress, err := getEmailForCustomer(req.CustomerId)
	if err != nil {
		return err
	}

	if req.Lines.Data == nil {
		return fmt.Errorf(
			"Invoice: %s for %s has nil line items", req.ID, req.CustomerId,
		)
	}

	if len(req.Lines.Data) < 0 {
		return fmt.Errorf(
			"Invoice: %s for %s has 0 line items", req.ID, req.CustomerId,
		)
	}

	planName := req.Lines.Data[0].Plan.Name
	opts := map[string]string{
		"planName": planName,
		"price":    formatStripeAmount(req.Currency, req.AmountDue),
	}

	Log.Info("Stripe: Sent payment email to: %s with plan: %s", emailAddress, planName)

	return paymentemail.Send(
		c.Email, paymentemail.PaymentCreated, emailAddress, opts,
	)
}
