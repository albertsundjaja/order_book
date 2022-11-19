package main

import (
	"flag"
	"log"

	"github.com/albertsundjaja/order_book/config"
	db "github.com/albertsundjaja/order_book/db/inmemory"
	"github.com/albertsundjaja/order_book/message"
	"github.com/albertsundjaja/order_book/order_book"
	"github.com/albertsundjaja/order_book/stream_handler"
)

const (
	// message header length
	HEADER_LENGTH = int64(8)
)

func main() {
	depthParam := flag.Int("depth", 3, "the depth that will be printed")
	flag.Parse()

	config := config.NewConfig()
	config.OrderBook.Depth = *depthParam
	// prepare components
	orderManagerChan := make(chan bool)
	streamHandlerChan := make(chan bool)
	commChan := make(chan message.Message)
	db := db.NewOrderBookDb(config)
	orderManager := order_book.NewOrderBookManager(config, orderManagerChan, commChan, db)
	streamHandler := stream_handler.NewStreamHandler(config, streamHandlerChan, commChan)

	// start component routines
	go streamHandler.Start()
	go orderManager.ProcessMessage()

	// wait until either finished or throw error
	select {
	case <-orderManagerChan:
		log.Println("OrderManager sends terminate signal")
		streamHandlerChan <- true
	case <-streamHandlerChan:
		log.Println("StreamHandler sends terminate signal")
		orderManagerChan <- true
	}
	log.Println("app shutting down")
}
