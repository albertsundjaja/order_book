package test

import (
	"bufio"
	"os"

	"github.com/albertsundjaja/order_book/config"
	db "github.com/albertsundjaja/order_book/internal/db/inmemory"
	"github.com/albertsundjaja/order_book/internal/message"
	"github.com/albertsundjaja/order_book/internal/order_book"
	"github.com/albertsundjaja/order_book/internal/stream_handler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2e test", func() {
	os.Setenv("ENV", "test")

	Describe("testing with input1.stream", func() {
		It("should print out the same output as output1.log", func() {
			f, _ := os.Open("input1.stream")
			reader := bufio.NewReader(f)
			config := config.NewConfig()
			config.OrderBook.Depth = 3

			orderManagerChan := make(chan bool)
			streamHandlerChan := make(chan bool)
			printChan := make(chan string)
			commChan := make(chan message.Message)
			db := db.NewOrderBookDb(config)
			orderManager := order_book.NewOrderBookManager(config, orderManagerChan, commChan, printChan, db)
			streamHandler := stream_handler.NewStreamHandler(config, reader, streamHandlerChan, commChan)

			// start component routines
			go streamHandler.Start()
			go orderManager.ProcessMessage()

			var result string

			// expect 10 message
			counter := 0
			for {
				if counter == 9 {
					break
				}
				msg := <-printChan
				result += msg
				counter += 1
			}
			// cleanup
			orderManagerChan <- true

			expectedResult, _ := os.ReadFile("output1.log")
			Expect(string(expectedResult)).To(Equal(result))
		})
	})
})
