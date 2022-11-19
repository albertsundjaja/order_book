package db

import "github.com/albertsundjaja/order_book/message"

// IDbOrderBook is an interface to store order book for easy DB replacement
// all data manipulation return bool that indicates whether that transaction changes the top N depth
type IDbOrderBook interface {
	AddOrder(message.MessageAdded) (bool, error)        // add order to db
	UpdateOrder(message.MessageUpdated) (bool, error)   // update order
	DeleteOrder(message.MessageDeleted) (bool, error)   // delete order
	ExecuteOrder(message.MessageExecuted) (bool, error) // execute order
	PrintDepth(symbol [3]byte) (string, error)          // return string that gives the symbol depth e.g. [(2, 1)], [(5, 1), (6, 1)]
}
