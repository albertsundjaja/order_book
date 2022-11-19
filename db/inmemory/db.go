// Package inmem_db is the IDbOrderBook implementation for in-memory data store
package inmem_db

import (
	"fmt"
	"log"

	"github.com/albertsundjaja/order_book/config"
	"github.com/albertsundjaja/order_book/message"
)

const (
	SORT_ORDER_BUY  = false // sort order descending
	SORT_ORDER_SELL = true  // sort order ascending
	SIDE_BUY        = 66    // Buy side. "B" in uint8
	SIDE_SELL       = 83    // Sell side. "S" in uint8
)

// OrderBook is the IDbOrderBook in-memory implementation
type OrderBookDb struct {
	config      *config.Config
	books       map[string]*orderBook // store the OrderBook of each symbols
	shouldPrint bool                  // check whether last update should print
	lastSymbol  string                // last symbol that was updated
}

// NewOrderBookDb return an instance of OrderBookDb
func NewOrderBookDb(config *config.Config) *OrderBookDb {
	return &OrderBookDb{
		config: config,
		books:  make(map[string]*orderBook),
	}
}

// AddSymbol add a new order book for the symbol to the manager
func (o *OrderBookDb) AddSymbol(symbol string, orderBook *orderBook) {
	o.books[symbol] = orderBook
}

// AddOrder add the order to the coressponding symbol order book
func (o *OrderBookDb) AddOrder(msg message.MessageAdded) error {
	symbol := string(msg.Symbol[:])
	orderBook, ok := o.books[symbol]
	if !ok {
		orderBook = newOrderBook(o.config.OrderBook.Depth)
		o.AddSymbol(symbol, orderBook)
	}
	o.shouldPrint = false
	err := orderBook.addOrder(msg)
	if err != nil {
		log.Printf("Unable to add order. Error: %s \n", err.Error())
		return err
	}
	o.shouldPrint = orderBook.shouldPrint
	o.lastSymbol = symbol
	return nil
}

// UpdateOrder update the corresponding symbol OrderId
func (o *OrderBookDb) UpdateOrder(msg message.MessageUpdated) error {
	symbol := string(msg.Symbol[:])
	orderBook, ok := o.books[symbol]
	if !ok {
		return fmt.Errorf("unable to update symbol %s. Symbol not found", symbol)
	}
	orderBook.shouldPrint = false
	err := orderBook.updateOrder(msg)
	if err != nil {
		log.Printf("Unable to update order. Error: %s \n", err.Error())
		return err
	}
	o.shouldPrint = orderBook.shouldPrint
	o.lastSymbol = symbol
	return nil
}

// DeleteOrder delete the corresponding symbol OrderId
func (o *OrderBookDb) DeleteOrder(msg message.MessageDeleted) error {
	symbol := string(msg.Symbol[:])
	orderBook, ok := o.books[symbol]
	if !ok {
		return fmt.Errorf("unable to delete symbol %s. Symbol not found", symbol)
	}
	orderBook.shouldPrint = false
	err := orderBook.deleteOrder(msg)
	if err != nil {
		log.Printf("Unable to delete order. Error: %s \n", err.Error())
		return err
	}
	o.shouldPrint = orderBook.shouldPrint
	o.lastSymbol = symbol
	return nil
}

// ExecuteOrder execute the corresponding symbol OrderId
func (o *OrderBookDb) ExecuteOrder(msg message.MessageExecuted) error {
	symbol := string(msg.Symbol[:])
	orderBook, ok := o.books[symbol]
	if !ok {
		return fmt.Errorf("unable to execute symbol %s. Symbol not found", symbol)
	}
	orderBook.shouldPrint = false
	err := orderBook.executeOrder(msg)
	if err != nil {
		log.Printf("Unable to execute order. Error: %s \n", err.Error())
		return err
	}
	o.shouldPrint = orderBook.shouldPrint
	o.lastSymbol = symbol
	return nil
}

// ShouldPrint returns the flag whether the last message triggered a reprint
func (o *OrderBookDb) ShouldPrint() bool {
	return o.shouldPrint
}

// Print the depth for the last symbol action
func (o *OrderBookDb) PrintDepth() (string, error) {
	orderBook, ok := o.books[o.lastSymbol]
	if !ok {
		return "", fmt.Errorf("unexpected error occurred. last symbol was not found: %s", o.lastSymbol)
	}
	o.shouldPrint = false
	return fmt.Sprintf("%s, %s", o.lastSymbol, orderBook.printDepth()), nil
}
