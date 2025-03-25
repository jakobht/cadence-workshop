package shippingcomplete

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	WorkflowName = "OrderProcessingWorkflowComplete"
)

// RegisterWorkflow registers the OrderProcessingWorkflow.
func RegisterWorkflow(w worker.Worker) {
	w.RegisterWorkflowWithOptions(PackageProcessingWorkflow, workflow.RegisterOptions{Name: WorkflowName})

	// Register your activities here
	w.RegisterActivityWithOptions(validatePayment, activity.RegisterOptions{Name: "validatePaymentComplete"})
	w.RegisterActivityWithOptions(shipProduct, activity.RegisterOptions{Name: "shipProductComplete"})
	w.RegisterActivityWithOptions(estimatedDeliveryTime, activity.RegisterOptions{Name: "estimatedDeliveryTime"})
}

// Order represents an order with basic details like the ID, customer name, and order amount.
type Order struct {
	ID       string  `json:"id"`
	Customer string  `json:"customer"`
	Amount   float64 `json:"amount"`
	Address  string  `json:"address"`
	SendFrom string  `json:"sendFrom"`
}

type ScanSignalValue struct {
	Location string `json:"location"`
}

type QueryResult struct {
	Delivered       bool     `json:"delivered"`
	LocationHistory []string `json:"locationHistory"`
}

// PackageProcessingWorkflow processes an order through several steps:
// 1. It first validates the payment for the order.
// 2. Then, it proceeds to ship the package.
// 3. Finally, it returns a result indicating success or failure based on the payment and shipping status.
func PackageProcessingWorkflow(ctx workflow.Context, order Order) (string, error) {
	locations := []string{order.SendFrom}
	packageDelivered := false

	err := workflow.SetQueryHandler(ctx, "current_status", func() (QueryResult, error) {
		return QueryResult{
			Delivered:       packageDelivered,
			LocationHistory: locations,
		}, nil
	})
	if err != nil {
		return "", fmt.Errorf("set query handler: %v", err)
	}

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Important: We need to use the Cadence supplied logger.
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting PackageProcessingWorkflow", zap.String("orderID", order.ID), zap.String("customer", order.Customer))

	// Step 1: Validate the payment.
	// The payment validation step checks if the payment for the order is valid.
	// In this example, we simulate the payment validation by calling the `validatePayment` activity.
	// If validation fails, the workflow stops early and returns an appropriate error.

	// Retry policy configuration: exponential backoff with a maximum of 3 retries.
	var paymentRetryPolicy = &cadence.RetryPolicy{
		InitialInterval:    1 * time.Second,  // Start with 1 second.
		BackoffCoefficient: 2.0,              // Exponential backoff.
		MaximumInterval:    10 * time.Second, // Max retry interval.
		MaximumAttempts:    3,                // Retry up to 3 times.
	}

	// Configure activity options with retry policy.
	var activityOptions = workflow.ActivityOptions{
		RetryPolicy:            paymentRetryPolicy, // Attach retry policy.
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}

	// Add the activity options to the context.
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

	var paymentValidationResult string
	err = workflow.ExecuteActivity(activityCtx, validatePayment, order).Get(ctx, &paymentValidationResult)
	if err != nil {
		return "", fmt.Errorf("validate payment for order: %v", err)
	}

	// Step 2: Ship the package
	// Once the payment is validated, we proceed to ship the package.
	// The ship the package activity is called to simulate the shipping process.
	// If shipping fails or encounters an error, the workflow returns an error.
	var shipProductResult string
	err = workflow.ExecuteActivity(ctx, shipProduct, order).Get(ctx, &shipProductResult)
	if err != nil {
		return "", fmt.Errorf("ship product for order: %v", err)
	}

	// Step 3: Scan the package
	// Every time the package is scanned, we will get a signal
	// We will use the signal to update the package location
	signalChan := workflow.GetSignalChannel(ctx, "ScanSignal")
	s := workflow.NewSelector(ctx)
	s.AddReceive(signalChan, func(c workflow.Channel, more bool) {
		var signalVal ScanSignalValue
		c.Receive(ctx, &signalVal)
		workflow.GetLogger(ctx).Info("Received signal!", zap.String("signal", "ScanSignal"), zap.Any("value", signalVal))
		locations = append(locations, signalVal.Location)
	})

	deliveredChan := workflow.GetSignalChannel(ctx, "DeliveredSignal")
	s.AddReceive(deliveredChan, func(c workflow.Channel, more bool) {
		packageDelivered = true
	})

	var cancelFunc workflow.CancelFunc
	justCanceledTimer := false
	notificationSent := false
	for !packageDelivered {
		if !notificationSent && !justCanceledTimer {
			cancelFunc, err = setupNotificationTimer(ctx, order, locations, s, cancelFunc, &justCanceledTimer, &notificationSent)
			if err != nil {
				logger.Error("Error setting up notification timer", zap.Error(err))
			}
		}

		justCanceledTimer = false
		s.Select(ctx)
	}

	// Step 3: Return a success message
	// If both payment validation and shipping were successful, we return a success message indicating the order was processed.
	return fmt.Sprintf("Order %s processed successfully for customer %s.", order.ID, order.Customer), nil
}

func setupNotificationTimer(ctx workflow.Context, order Order, locations []string, s workflow.Selector, cancelFunc workflow.CancelFunc, justCanceledTimer *bool, notificationSent *bool) (workflow.CancelFunc, error) {
	// Cancel the previous timer if it exists
	if cancelFunc != nil {
		cancelFunc()
	}

	// Get the notification time based on the order and the last location
	last_location := locations[len(locations)-1]
	var deliveryTimeInDays int
	err := workflow.ExecuteActivity(ctx, estimatedDeliveryTime, order, last_location).Get(ctx, &deliveryTimeInDays)
	if err != nil {
		return nil, fmt.Errorf("getting delivery time", zap.Error(err))
	}

	// Cancelable timer context
	timerCtx, cancel := workflow.WithCancel(ctx)

	// Create the timer and add it to the selector (we only wait minutes, not days)
	timer := workflow.NewTimer(timerCtx, time.Duration(deliveryTimeInDays)*time.Minute)
	s.AddFuture(timer, func(f workflow.Future) {
		// Check if the timer was canceled
		err := f.Get(ctx, nil)
		var canceledErr *cadence.CanceledError
		if errors.As(err, &canceledErr) {
			workflow.GetLogger(ctx).Info("This timer was canceled")
			*justCanceledTimer = true
			return
		}
		// We should call an activity that sends a notification to the customer,
		// but for now we will just log the event, and remember that we sent the notification
		*notificationSent = true
		workflow.GetLogger(ctx).Info("Notification timer fired!")
	})

	return cancel, nil
}

// Add an activity here that validates a payment.
// The validation fails if the order amount is greater than 25 (for example, due to payment policy restrictions).
func validatePayment(ctx context.Context, order Order) (string, error) {
	// Simulate a failure if this is the 0th or 1st attempt
	info := activity.GetInfo(ctx)
	if info.Attempt < 2 {
		activity.GetLogger(ctx).Info("Temporary failure in payment processing")
		return "", fmt.Errorf("temporary issue, please retry")
	}

	if order.Amount > 25 {
		return "", fmt.Errorf("payment validation failed for order: %v", order.ID)
	}

	activity.GetLogger(ctx).Info("Payment processed successfully")
	return "Payment successful", nil
}

func shipProduct(ctx context.Context, order Order) (string, error) {
	if order.Customer == "" {
		return "", fmt.Errorf("customer is required")
	}
	activity.GetLogger(ctx).Info("Shipping product", zap.String("orderID", order.ID))
	return "Product shipped successfully", nil
}

func estimatedDeliveryTime(ctx context.Context, order Order, currentLocation string) (int, error) {
	return rand.IntN(3) + 1, nil
}
