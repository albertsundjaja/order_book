package order_book

import (
	"fmt"

	"log"

	"github.com/albertsundjaja/order_book/config"
	"github.com/albertsundjaja/order_book/stream_handler"
)

// OrderBookManager contains the books of all the symbols
type OrderBookManager struct {
	config      *config.Config                // store app config
	books       map[string]*OrderBook         // store the OrderBook of each symbols
	streamChan  <-chan stream_handler.Message // channel for receiving message from StreamHandler
	managerChan chan bool                     // for communicating with the main routine
}

// NewOrderBook manager init the OrderBookManager
func NewOrderBookManager(config *config.Config, managerChan chan bool, streamChan <-chan stream_handler.Message) *OrderBookManager {
	return &OrderBookManager{
		config:      config,
		books:       make(map[string]*OrderBook),
		streamChan:  streamChan,
		managerChan: managerChan,
	}
}

// AddSymbol add a new order book for the symbol to the manager
func (s *OrderBookManager) AddSymbol(symbol string, orderBook *OrderBook) {
	s.books[symbol] = orderBook
}

// ProcessMessage process the message received from the stream
func (s *OrderBookManager) ProcessMessage() {
	for {
		select {
		case msg := <-s.streamChan:
			err := s.processMessage(msg)
			if err != nil {
				log.Printf("error occurred in ProcessMessage: %s \n", err.Error())
				s.managerChan <- true
			}
		case <-s.managerChan:
			return
		}
	}
}

// processMessage will extract the message and pass the message to corresponding symbol order book
func (s *OrderBookManager) processMessage(msg stream_handler.Message) error {
	var orderBook *OrderBook
	var ok bool
	var symbol string
	switch msg.MsgType {
	case stream_handler.MSG_TYPE_ADDED:
		addedMsg := msg.MsgBody.(stream_handler.MessageAdded)
		symbol = string(addedMsg.Symbol[:])
		orderBook, ok = s.books[symbol]
		if !ok {
			orderBook = NewOrderBook(s.config.OrderBook.Depth)
			s.AddSymbol(symbol, orderBook)
		}
		orderBook.shouldPrint = false
		err := orderBook.AddOrder(addedMsg)
		if err != nil {
			log.Printf("Unable to add order. Error: %s \n", err.Error())
			return err
		}
	case stream_handler.MSG_TYPE_UPDATED:
		updatedMsg := msg.MsgBody.(stream_handler.MessageUpdated)
		symbol = string(updatedMsg.Symbol[:])
		orderBook, ok = s.books[symbol]
		if !ok {
			return fmt.Errorf("unable to update symbol %s. Symbol not found", symbol)
		}
		orderBook.shouldPrint = false
		err := orderBook.UpdateOrder(updatedMsg)
		if err != nil {
			log.Printf("Unable to update order. Error: %s \n", err.Error())
			return err
		}
	case stream_handler.MSG_TYPE_DELETED:
		delMsg := msg.MsgBody.(stream_handler.MessageDeleted)
		symbol = string(delMsg.Symbol[:])
		orderBook, ok = s.books[symbol]
		if !ok {
			return fmt.Errorf("unable to delete symbol %s. Symbol not found", symbol)
		}
		orderBook.shouldPrint = false
		err := orderBook.DeleteOrder(delMsg)
		if err != nil {
			log.Printf("Unable to delete order. Error: %s \n", err.Error())
			return err
		}
	case stream_handler.MSG_TYPE_EXECUTED:
		exMsg := msg.MsgBody.(stream_handler.MessageExecuted)
		symbol = string(exMsg.Symbol[:])
		orderBook, ok = s.books[symbol]
		if !ok {
			return fmt.Errorf("unable to execute symbol %s. Symbol not found", symbol)
		}
		orderBook.shouldPrint = false
		err := orderBook.ExecuteOrder(exMsg)
		if err != nil {
			log.Printf("Unable to execute order. Error: %s \n", err.Error())
			return err
		}
	}
	if orderBook.shouldPrint {
		orderBook.PrintDepth(msg.MsgHeader.Seq, symbol)
	}
	return nil
}
