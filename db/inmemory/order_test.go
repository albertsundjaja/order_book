package inmem_db

import (
	"github.com/albertsundjaja/order_book/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OrderBook", func() {

	var (
		orderBook *orderBook
	)

	BeforeEach(func() {
		orderBook = newOrderBook(5)
	})
	Describe("AddOrder", func() {
		Context("adding order to buy side with a new OrderId", func() {
			It("add to the Buy, AggBuy and BuyDepth correctly", func() {
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_BUY},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				Expect(orderBook.Buy[orderId].Price).To(Equal(price))
				Expect(orderBook.Buy[orderId].Volume).To(Equal(volume))
				Expect(orderBook.AggBuy[price].Volume).To(Equal(volume))
				Expect(SortedContainsInt32(SORT_ORDER_BUY, orderBook.BuyDepth, price)).To(Equal(0))
			})
		})

		Context("adding order to sell side with a new OrderId", func() {
			It("add to the Sell, AggSell, SellDepth correctly", func() {
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_SELL},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				Expect(orderBook.Sell[orderId].Price).To(Equal(price))
				Expect(orderBook.Sell[orderId].Volume).To(Equal(volume))
				Expect(orderBook.AggSell[price].Volume).To(Equal(volume))
				Expect(SortedContainsInt32(SORT_ORDER_SELL, orderBook.SellDepth, price)).To(Equal(0))
			})
		})
	})

	Describe("UpdateOrder", func() {
		Context("updating existing Buy OrderId", func() {
			It("should update correctly", func() {
				// add order first
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_BUY},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				updatedPrice := int32(20)
				updatedVolume := uint64(20)
				updateMsg := message.MessageUpdated{
					Side:    [1]byte{SIDE_BUY},
					OrderId: orderId,
					Price:   updatedPrice,
					Size:    updatedVolume,
				}
				err = orderBook.updateOrder(updateMsg)
				Expect(err).To(BeNil())

				Expect(orderBook.Buy[orderId].Price).To(Equal(updatedPrice))
				Expect(orderBook.Buy[orderId].Volume).To(Equal(updatedVolume))
				Expect(orderBook.AggBuy[updatedPrice].Volume).To(Equal(updatedVolume))
				Expect(SortedContainsInt32(SORT_ORDER_BUY, orderBook.BuyDepth, updatedPrice)).To(Equal(0))

				// check old price is correctly handled
				_, ok := orderBook.AggBuy[price]
				Expect(ok).To(Equal(false))
				Expect(SortedContainsInt32(SORT_ORDER_BUY, orderBook.BuyDepth, price)).To(Equal(-1))
			})
		})

		Context("updating existing Sell OrderId", func() {
			It("should update correctly", func() {
				// add order first
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_SELL},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				updatedPrice := int32(20)
				updatedVolume := uint64(20)
				updateMsg := message.MessageUpdated{
					Side:    [1]byte{SIDE_SELL},
					OrderId: orderId,
					Price:   updatedPrice,
					Size:    updatedVolume,
				}
				err = orderBook.updateOrder(updateMsg)
				Expect(err).To(BeNil())

				Expect(orderBook.Sell[orderId].Price).To(Equal(updatedPrice))
				Expect(orderBook.Sell[orderId].Volume).To(Equal(updatedVolume))
				Expect(orderBook.AggSell[updatedPrice].Volume).To(Equal(updatedVolume))
				Expect(SortedContainsInt32(SORT_ORDER_SELL, orderBook.SellDepth, updatedPrice)).To(Equal(0))

				// check old price is correctly handled
				_, ok := orderBook.AggSell[price]
				Expect(ok).To(Equal(false))
				Expect(SortedContainsInt32(SORT_ORDER_SELL, orderBook.SellDepth, price)).To(Equal(-1))
			})
		})
	})

	Describe("DeleteOrder", func() {
		Context("deleting existing Buy OrderId", func() {
			It("should delete correctly", func() {
				// add order first
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_BUY},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				delMsg := message.MessageDeleted{
					Side:    [1]byte{SIDE_BUY},
					OrderId: orderId,
				}
				err = orderBook.deleteOrder(delMsg)
				Expect(err).To(BeNil())

				_, ok := orderBook.Buy[orderId]
				Expect(ok).To(BeFalse())
				_, ok = orderBook.AggBuy[price]
				Expect(ok).To(BeFalse())
				Expect(SortedContainsInt32(SORT_ORDER_BUY, orderBook.BuyDepth, price)).To(Equal(-1))
			})
		})

		Context("deleting existing Sell OrderId", func() {
			It("should delete correctly", func() {
				// add order first
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_SELL},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				delMsg := message.MessageDeleted{
					Side:    [1]byte{SIDE_SELL},
					OrderId: orderId,
				}
				err = orderBook.deleteOrder(delMsg)
				Expect(err).To(BeNil())

				_, ok := orderBook.Sell[orderId]
				Expect(ok).To(BeFalse())
				_, ok = orderBook.AggSell[price]
				Expect(ok).To(BeFalse())
				Expect(SortedContainsInt32(SORT_ORDER_SELL, orderBook.SellDepth, price)).To(Equal(-1))
			})
		})
	})

	Describe("Execute Order", func() {
		Context("executing existing Buy OrderId", func() {
			It("should execute correctly", func() {
				// add order first
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_BUY},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				exMsg := message.MessageExecuted{
					Side:      [1]byte{SIDE_BUY},
					OrderId:   orderId,
					TradedQty: volume,
				}
				err = orderBook.executeOrder(exMsg)
				Expect(err).To(BeNil())

				_, ok := orderBook.Buy[orderId]
				Expect(ok).To(BeFalse())
				_, ok = orderBook.AggBuy[price]
				Expect(ok).To(BeFalse())
				Expect(SortedContainsInt32(SORT_ORDER_BUY, orderBook.BuyDepth, price)).To(Equal(-1))
			})

		})

		Context("executing existing Sell OrderId", func() {
			It("should execute correctly", func() {
				// add order first
				orderId := uint64(123)
				price := int32(1)
				volume := uint64(1)
				addMsg := message.MessageAdded{
					Side:    [1]byte{SIDE_SELL},
					OrderId: orderId,
					Price:   price,
					Size:    volume,
				}
				err := orderBook.addOrder(addMsg)
				Expect(err).To(BeNil())

				exMsg := message.MessageExecuted{
					Side:      [1]byte{SIDE_SELL},
					OrderId:   orderId,
					TradedQty: volume,
				}
				err = orderBook.executeOrder(exMsg)
				Expect(err).To(BeNil())

				_, ok := orderBook.Sell[orderId]
				Expect(ok).To(BeFalse())
				_, ok = orderBook.AggSell[price]
				Expect(ok).To(BeFalse())
				Expect(SortedContainsInt32(SORT_ORDER_SELL, orderBook.SellDepth, price)).To(Equal(-1))
			})

		})
	})
})
