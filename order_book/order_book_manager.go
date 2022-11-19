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
	managerChan chan bool              // for communicating with the main routine for termination
	printChan   chan<- string          // for sending out the result of the market depth
}

// NewOrderBook manager init the OrderBookManager
func NewOrderBookManager(config *config.Config, managerChan chan bool, streamChan <-chan message.Message, printChan chan<- string, db db.IDbOrderBook) *OrderBookManager {
	return &OrderBookManager{
		config:      config,
		streamChan:  streamChan,
		managerChan: managerChan,
		printChan:   printChan,
		db:          db,
	}
}

// ProcessMessage process the message received from the stream
func (o *OrderBookManager) ProcessMessage() {
mainLoop:
	for {
		select {
		case msg := <-o.streamChan:
			marketDepth, err := o.processMessage(msg)
			if err != nil {
				log.Printf("error occurred in ProcessMessage: %s \n", err.Error())
				o.managerChan <- true
				break mainLoop
			}
			if marketDepth != "" {
				o.printChan <- marketDepth
			}
		case <-o.managerChan:
			break mainLoop
		}
	}
}

// processMessage parse the raw msg and send it to DB
// returns empty string if the msg does not update the top N depth otherwise, it returns the complete string for the market depth
func (o *OrderBookManager) processMessage(msg message.Message) (string, error) {
	var shouldPrint bool
	var err error
	switch msg.MsgType {
	case message.MSG_TYPE_ADDED:
		addedMsg := msg.MsgBody.(message.MessageAdded)
		shouldPrint, err = o.db.AddOrder(addedMsg)
		if err != nil {
			log.Printf("Unable to add order. Error: %s \n", err.Error())
			return "", err
		}
	case message.MSG_TYPE_UPDATED:
		updatedMsg := msg.MsgBody.(message.MessageUpdated)
		shouldPrint, err = o.db.UpdateOrder(updatedMsg)
		if err != nil {
			log.Printf("Unable to update order. Error: %s \n", err.Error())
			return "", err
		}
	case message.MSG_TYPE_DELETED:
		delMsg := msg.MsgBody.(message.MessageDeleted)
		shouldPrint, err = o.db.DeleteOrder(delMsg)
		if err != nil {
			log.Printf("Unable to delete order. Error: %s \n", err.Error())
			return "", err
		}
	case message.MSG_TYPE_EXECUTED:
		exMsg := msg.MsgBody.(message.MessageExecuted)
		shouldPrint, err = o.db.ExecuteOrder(exMsg)
		if err != nil {
			log.Printf("Unable to execute order. Error: %s \n", err.Error())
			return "", err
		}
	default:
		return "", fmt.Errorf("unrecognized message type %s", msg.MsgType)
	}
	if shouldPrint {
		marketDepth, err := o.db.PrintDepth(msg.Symbol)
		if err != nil {
			log.Printf("Unable to get market depth: %s", err.Error())
			return "", err
		}
		// e.g. 4, VC0, [(318800, 4709), (315000, 2986)], [(318900, 360)]
		return fmt.Sprintf("%d, %s, %s\n", msg.MsgHeader.Seq, string(msg.Symbol[:]), marketDepth), nil
	}
	return "", nil
}
