package example

import (
	"errors"
	"fmt"
	"github.com/watertreestar/canary/state"
	"strings"
)

const (
	CreatingOrder     state.StateType = "CreatingOrder"
	OrderFailed       state.StateType = "OrderFailed"
	OrderPlaced       state.StateType = "OrderPlaced"
	ChargingCard      state.StateType = "ChargingCard"
	TransactionFailed state.StateType = "TransactionFailed"
	OrderShipped      state.StateType = "OrderShipped"

	CreateOrder     state.EventType = "CreateOrder"
	FailOrder       state.EventType = "FailOrder"
	PlaceOrder      state.EventType = "PlaceOrder"
	ChargeCard      state.EventType = "ChargeCard"
	FailTransaction state.EventType = "FailTransaction"
	ShipOrder       state.EventType = "ShipOrder"
)

// state context definition

type OrderCreationContext struct {
	items []string
	err   error
}

func (c *OrderCreationContext) String() string {
	return fmt.Sprintf("OrderCreationContext [ items: %s, err: %v ]",
		strings.Join(c.items, ","), c.err)
}

type OrderShipmentContext struct {
	cardNumber string
	address    string
	err        error
}

func (c *OrderShipmentContext) String() string {
	return fmt.Sprintf("OrderShipmentContext [ cardNumber: %s, address: %s, err: %v ]",
		c.cardNumber, c.address, c.err)
}

// state action definition

type CreatingOrderAction struct{}

func (a *CreatingOrderAction) Execute(eventCtx state.EventContext) state.EventType {
	order := eventCtx.(*OrderCreationContext)
	fmt.Println("Validating, order:", order)
	if len(order.items) == 0 {
		order.err = errors.New("insufficient number of items in order")
		return FailOrder
	}
	return PlaceOrder
}

type OrderFailedAction struct{}

func (a *OrderFailedAction) Execute(eventCtx state.EventContext) state.EventType {
	order := eventCtx.(*OrderCreationContext)
	fmt.Println("Order failed, err:", order.err)
	return state.NoOp
}

type OrderPlacedAction struct{}

func (a *OrderPlacedAction) Execute(eventCtx state.EventContext) state.EventType {
	order := eventCtx.(*OrderCreationContext)
	fmt.Println("Order placed, items:", order.items)
	return state.NoOp
}

type ChargingCardAction struct{}

func (a *ChargingCardAction) Execute(eventCtx state.EventContext) state.EventType {
	shipment := eventCtx.(*OrderShipmentContext)
	fmt.Println("Validating card, shipment:", shipment)
	if shipment.cardNumber == "" {
		shipment.err = errors.New("Card number is invalid")
		return FailTransaction
	}
	return ShipOrder
}

type TransactionFailedAction struct{}

func (a *TransactionFailedAction) Execute(eventCtx state.EventContext) state.EventType {
	shipment := eventCtx.(*OrderShipmentContext)
	fmt.Println("Transaction failed, err:", shipment.err)
	return state.NoOp
}

type OrderShippedAction struct{}

func (a *OrderShippedAction) Execute(eventCtx state.EventContext) state.EventType {
	shipment := eventCtx.(*OrderShipmentContext)
	fmt.Println("Order shipped, address:", shipment.address)
	return state.NoOp
}

// state machine definition

func newOrderFSM() *state.StateMachine {
	return &state.StateMachine{
		States: state.States{
			state.Default: state.State{
				Events: state.Events{
					CreateOrder: CreatingOrder,
				},
			},
			CreatingOrder: state.State{
				Action: &CreatingOrderAction{},
				Events: state.Events{
					FailOrder:  OrderFailed,
					PlaceOrder: OrderPlaced,
				},
			},
			OrderFailed: state.State{
				Action: &OrderFailedAction{},
				Events: state.Events{
					CreateOrder: CreatingOrder,
				},
			},
			OrderPlaced: state.State{
				Action: &OrderPlacedAction{},
				Events: state.Events{
					ChargeCard: ChargingCard,
				},
			},
			ChargingCard: state.State{
				Action: &ChargingCardAction{},
				Events: state.Events{
					FailTransaction: TransactionFailed,
					ShipOrder:       OrderShipped,
				},
			},
			TransactionFailed: state.State{
				Action: &TransactionFailedAction{},
				Events: state.Events{
					ChargeCard: ChargingCard,
				},
			},
			OrderShipped: state.State{
				Action: &OrderShippedAction{},
			},
		},
	}
}
