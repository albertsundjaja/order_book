package db

import "github.com/albertsundjaja/order_book/message"

// IDbOrderBook is an interface to store order book for easy DB replacement
type IDbOrderBook interface {
	AddOrder(message.MessageAdded) error        // add order to db
	UpdateOrder(message.MessageUpdated) error   // update order
	DeleteOrder(message.MessageDeleted) error   // delete order
	ExecuteOrder(message.MessageExecuted) error // execute order
	ShouldPrint() bool                          // check whether we should print the depth after any of the above operation
	PrintDepth() (string, error)                // return string that prints the latest depth
}
