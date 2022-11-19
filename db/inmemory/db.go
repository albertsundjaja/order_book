// Package inmem_db is the IDbOrderBook implementation for in-memory data store
package inmem_db

import (
	"fmt"
	"log"

	"github.com/albertsundjaja/order_book/config"
	"github.com/albertsundjaja/order_book/message"
)

// OrderBook is the IDbOrderBook in-memory implementation
type OrderBookDb struct {
	config     *config.Config
	books      map[[3]byte]*orderBook // store the OrderBook of each symbols
	lastSymbol string                 // last symbol that was updated
}

// NewOrderBookDb return an instance of OrderBookDb
func NewOrderBookDb(config *config.Config) *OrderBookDb {
	return &OrderBookDb{
		config: config,
		books:  make(map[[3]byte]*orderBook),
	}
}

// AddSymbol add a new order book for the symbol to the manager
func (o *OrderBookDb) AddSymbol(symbol [3]byte, orderBook *orderBook) {
	o.books[symbol] = orderBook
}

// AddOrder add the order to the coressponding symbol order book
func (o *OrderBookDb) AddOrder(msg message.MessageAdded) (bool, error) {
	orderBook, ok := o.books[msg.Symbol]
	if !ok {
		orderBook = newOrderBook(o.config.OrderBook.Depth)
		o.AddSymbol(msg.Symbol, orderBook)
	}
	err := orderBook.addOrder(msg)
	if err != nil {
		log.Printf("Unable to add order. Error: %s \n", err.Error())
		return false, err
	}
	return orderBook.shouldPrint, nil
}

// UpdateOrder update the corresponding symbol OrderId
func (o *OrderBookDb) UpdateOrder(msg message.MessageUpdated) (bool, error) {
	orderBook, ok := o.books[msg.Symbol]
	if !ok {
		return false, fmt.Errorf("unable to update symbol %s. Symbol not found", msg.Symbol)
	}
	orderBook.shouldPrint = false
	err := orderBook.updateOrder(msg)
	if err != nil {
		log.Printf("Unable to update order. Error: %s \n", err.Error())
		return false, err
	}
	return orderBook.shouldPrint, nil
}

// DeleteOrder delete the corresponding symbol OrderId
func (o *OrderBookDb) DeleteOrder(msg message.MessageDeleted) (bool, error) {
	orderBook, ok := o.books[msg.Symbol]
	if !ok {
		return false, fmt.Errorf("unable to delete symbol %s. Symbol not found", msg.Symbol)
	}
	orderBook.shouldPrint = false
	err := orderBook.deleteOrder(msg)
	if err != nil {
		log.Printf("Unable to delete order. Error: %s \n", err.Error())
		return false, err
	}
	return orderBook.shouldPrint, nil
}

// ExecuteOrder execute the corresponding symbol OrderId
func (o *OrderBookDb) ExecuteOrder(msg message.MessageExecuted) (bool, error) {
	orderBook, ok := o.books[msg.Symbol]
	if !ok {
		return false, fmt.Errorf("unable to execute symbol %s. Symbol not found", msg.Symbol)
	}
	orderBook.shouldPrint = false
	err := orderBook.executeOrder(msg)
	if err != nil {
		log.Printf("Unable to execute order. Error: %s \n", err.Error())
		return false, err
	}
	return orderBook.shouldPrint, nil
}

// Print the depth for the last symbol action
func (o *OrderBookDb) PrintDepth(symbol [3]byte) (string, error) {
	orderBook, ok := o.books[symbol]
	if !ok {
		return "", fmt.Errorf("unexpected error occurred. last symbol was not found: %s", o.lastSymbol)
	}
	return orderBook.printDepth(), nil
}
