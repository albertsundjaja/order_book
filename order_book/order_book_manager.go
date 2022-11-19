package order_book

import (
	"fmt"

	"log"

	"github.com/albertsundjaja/order_book/config"
	"github.com/albertsundjaja/order_book/db"
	"github.com/albertsundjaja/order_book/message"
)

// OrderBookManager contains the books of all the symbols
type OrderBookManager struct {
	config      *config.Config         // store app config
	db          db.IDbOrderBook        // store all our order data
	streamChan  <-chan message.Message // channel for receiving message from StreamHandler
	managerChan chan bool              // for communicating with the main routine
}

// NewOrderBook manager init the OrderBookManager
func NewOrderBookManager(config *config.Config, managerChan chan bool, streamChan <-chan message.Message, db db.IDbOrderBook) *OrderBookManager {
	return &OrderBookManager{
		config:      config,
		streamChan:  streamChan,
		managerChan: managerChan,
		db:          db,
	}
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

func (o *OrderBookManager) processMessage(msg message.Message) error {
	switch msg.MsgType {
	case message.MSG_TYPE_ADDED:
		addedMsg := msg.MsgBody.(message.MessageAdded)
		err := o.db.AddOrder(addedMsg)
		if err != nil {
			log.Printf("Unable to add order. Error: %s \n", err.Error())
			return err
		}
	case message.MSG_TYPE_UPDATED:
		updatedMsg := msg.MsgBody.(message.MessageUpdated)
		err := o.db.UpdateOrder(updatedMsg)
		if err != nil {
			log.Printf("Unable to update order. Error: %s \n", err.Error())
			return err
		}
	case message.MSG_TYPE_DELETED:
		delMsg := msg.MsgBody.(message.MessageDeleted)
		err := o.db.DeleteOrder(delMsg)
		if err != nil {
			log.Printf("Unable to delete order. Error: %s \n", err.Error())
			return err
		}
	case message.MSG_TYPE_EXECUTED:
		exMsg := msg.MsgBody.(message.MessageExecuted)
		err := o.db.ExecuteOrder(exMsg)
		if err != nil {
			log.Printf("Unable to execute order. Error: %s \n", err.Error())
			return err
		}
	}
	if o.db.ShouldPrint() {
		marketDepth, err := o.db.PrintDepth()
		if err != nil {
			log.Printf("Unable to get market depth: %s", err.Error())
			return err
		}
		fmt.Printf("%d, %s \n", msg.MsgHeader.Seq, marketDepth)
	}
	return nil
}

// processMessage will extract the message and pass the message to corresponding symbol order book
// func (s *OrderBookManager) processMessage(msg message.Message) error {
// 	var orderBook *OrderBook
// 	var ok bool
// 	var symbol string
// 	switch msg.MsgType {
// 	case message.MSG_TYPE_ADDED:
// 		addedMsg := msg.MsgBody.(message.MessageAdded)
// 		symbol = string(addedMsg.Symbol[:])
// 		orderBook, ok = s.books[symbol]
// 		if !ok {
// 			orderBook = NewOrderBook(s.config.OrderBook.Depth)
// 			s.AddSymbol(symbol, orderBook)
// 		}
// 		orderBook.shouldPrint = false
// 		err := orderBook.AddOrder(addedMsg)
// 		if err != nil {
// 			log.Printf("Unable to add order. Error: %s \n", err.Error())
// 			return err
// 		}
// 	case message.MSG_TYPE_UPDATED:
// 		updatedMsg := msg.MsgBody.(message.MessageUpdated)
// 		symbol = string(updatedMsg.Symbol[:])
// 		orderBook, ok = s.books[symbol]
// 		if !ok {
// 			return fmt.Errorf("unable to update symbol %s. Symbol not found", symbol)
// 		}
// 		orderBook.shouldPrint = false
// 		err := orderBook.UpdateOrder(updatedMsg)
// 		if err != nil {
// 			log.Printf("Unable to update order. Error: %s \n", err.Error())
// 			return err
// 		}
// 	case message.MSG_TYPE_DELETED:
// 		delMsg := msg.MsgBody.(message.MessageDeleted)
// 		symbol = string(delMsg.Symbol[:])
// 		orderBook, ok = s.books[symbol]
// 		if !ok {
// 			return fmt.Errorf("unable to delete symbol %s. Symbol not found", symbol)
// 		}
// 		orderBook.shouldPrint = false
// 		err := orderBook.DeleteOrder(delMsg)
// 		if err != nil {
// 			log.Printf("Unable to delete order. Error: %s \n", err.Error())
// 			return err
// 		}
// 	case message.MSG_TYPE_EXECUTED:
// 		exMsg := msg.MsgBody.(message.MessageExecuted)
// 		symbol = string(exMsg.Symbol[:])
// 		orderBook, ok = s.books[symbol]
// 		if !ok {
// 			return fmt.Errorf("unable to execute symbol %s. Symbol not found", symbol)
// 		}
// 		orderBook.shouldPrint = false
// 		err := orderBook.ExecuteOrder(exMsg)
// 		if err != nil {
// 			log.Printf("Unable to execute order. Error: %s \n", err.Error())
// 			return err
// 		}
// 	}
// 	if orderBook.shouldPrint {
// 		orderBook.PrintDepth(msg.MsgHeader.Seq, symbol)
// 	}
// 	return nil
// }
