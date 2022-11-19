package order_book

import (
	"fmt"

	"github.com/albertsundjaja/order_book/config"
	"github.com/albertsundjaja/order_book/internal/message"
	mockDb "github.com/albertsundjaja/order_book/internal/mock/db"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrderBookManager", func() {
	var control *gomock.Controller
	var db *mockDb.MockIDbOrderBook
	config := &config.Config{}
	var orderBookManager *OrderBookManager

	BeforeEach(func() {
		control = gomock.NewController(GinkgoT())
		db = mockDb.NewMockIDbOrderBook(control)
		orderBookManager = NewOrderBookManager(config, make(chan bool), make(<-chan message.Message), make(chan<- string), db)
	})

	Describe("processMessage", func() {
		Context("valid raw added message", func() {
			It("should return the correct market depth string", func() {
				symbol := [3]byte{1, 2, 3}
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Symbol:  symbol,
					OrderId: orderId,
					Side:    [1]byte{message.SIDE_BUY},
					Price:   price,
					Size:    volume,
				}
				header := message.Header{
					Seq:  1,
					Size: 8,
				}
				rawMsg := message.Message{
					Symbol:    symbol,
					MsgType:   message.MSG_TYPE_ADDED,
					MsgHeader: header,
					MsgBody:   addMsg,
				}
				fakeDepth := "[(3, 1)], [(4, 2)]"
				db.EXPECT().AddOrder(addMsg).Return(true, nil)
				db.EXPECT().PrintDepth(symbol).Return(fakeDepth, nil)
				expectedDepth := fmt.Sprintf("%d, %s, %s\n", header.Seq, string(symbol[:]), fakeDepth)

				returnedDepth, err := orderBookManager.processMessage(rawMsg)
				Expect(err).To(BeNil())
				Expect(returnedDepth).To(Equal(expectedDepth))
			})
		})
	})
})
