package example

import (
	"github.com/watertreestar/canary/state"
	"testing"
)

func TestOrderFSM(t *testing.T) {
	orderFsm := newOrderFSM()

	// Define the context for order creation.
	creationCtx := &OrderCreationContext{
		items: []string{},
	}

	// Try to create an order with invalid set of items.
	err := orderFsm.SendEvent(CreateOrder, creationCtx)
	if err != nil {
		t.Errorf("Failed to send create order event, err: %v", err)
	}

	// The state machine should enter the OrderFailed state.
	if orderFsm.Current != OrderFailed {
		t.Errorf("Expected the FSM to be in the OrderFailed state, actual: %s",
			orderFsm.Current)
	}

	// Let's fix the order creation context.
	creationCtx = &OrderCreationContext{
		items: []string{"foo", "bar"},
	}

	// Let's now retry the same order with a valid set of items.
	err = orderFsm.SendEvent(CreateOrder, creationCtx)
	if err != nil {
		t.Errorf("Failed to send create order event, err: %v", err)
	}

	// The state machine should enter the OrderPlaced state.
	if orderFsm.Current != OrderPlaced {
		t.Errorf("Expected the FSM to be in the OrderPlaced state, actual: %s",
			orderFsm.Current)
	}

	// Let's now define the context for shipping the order.
	shipmentCtx := &OrderShipmentContext{
		cardNumber: "",
		address:    "123 Foo Street, Bar Baz, QU 45678, USA",
	}

	// Try to charge the card using an invalid card number.
	err = orderFsm.SendEvent(ChargeCard, shipmentCtx)
	if err != nil {
		t.Errorf("Failed to send charge card event, err: %v", err)
	}

	// The state machine should enter the TransactionFailed state.
	if orderFsm.Current != TransactionFailed {
		t.Errorf("Expected the FSM to be in the TransactionFailed state, actual: %s",
			orderFsm.Current)
	}

	// Let's fix the shipment context.
	shipmentCtx = &OrderShipmentContext{
		cardNumber: "0000-0000-0000-0000",
		address:    "123 Foo Street, Bar Baz, QU 45678, USA",
	}

	// Let's now retry the transaction with a valid card number.
	err = orderFsm.SendEvent(ChargeCard, shipmentCtx)
	if err != nil {
		t.Errorf("Failed to send charge card event, err: %v", err)
	}

	// The state machine should enter the OrderShipped state.
	if orderFsm.Current != OrderShipped {
		t.Errorf("Expected the FSM to be in the OrderShipped state, actual: %s",
			orderFsm.Current)
	}

	// Let's try charging the card again for the same order. We should see the
	// event getting rejected as the card has been already charged once and the
	// state machine is in the OrderShipped state.
	err = orderFsm.SendEvent(ChargeCard, shipmentCtx)
	if err != state.ErrEventRejected {
		t.Errorf("Expected the FSM to return a rejected event error")
	}
}
