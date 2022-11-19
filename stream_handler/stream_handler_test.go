package stream_handler

import (
	"bytes"
	"encoding/binary"

	"github.com/albertsundjaja/order_book/config"
	"github.com/albertsundjaja/order_book/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StreamHandler", func() {
	var (
		streamHandler *StreamHandler
	)
	managerChan := make(chan bool)
	orderBookChan := make(chan message.Message)
	config := &config.Config{
		Stream: struct {
			HeaderLength int64 `mapstructure:"headerLength"`
		}{
			HeaderLength: 8,
		},
	}

	BeforeEach(func() {
		streamHandler = NewStreamHandler(config, managerChan, orderBookChan)
	})

	Describe("eat", func() {
		Context("eating with count less than buffer", func() {
			It("should return the byte correctly", func() {
				streamHandler.buffer = []byte{1, 2, 3, 4}
				result, err := streamHandler.eat(1)
				Expect(err).To(BeNil())
				Expect(result).To(Equal([]byte{1}))
			})
		})
		Context("eating with count more than buffer", func() {
			It("should return an error", func() {
				streamHandler.buffer = []byte{1, 2, 3, 4}
				_, err := streamHandler.eat(100)
				Expect(err).To(Not(BeNil()))
			})
		})
	})

	Describe("ParseMsg", func() {
		Context("with raw msg as MSG_TYPE_ADDED", func() {
			It("should parse the message correctly", func() {
				addMsg := message.MessageAdded{
					Symbol:      [3]byte{1, 2, 3},
					OrderId:     uint64(123),
					Side:        [1]byte{1},
					ReservedOne: [3]byte{1, 2, 3},
					Size:        uint64(123),
					Price:       int32(123),
					ReservedTwo: [4]byte{1, 2, 3, 4},
				}
				var msg bytes.Buffer
				binary.Write(&msg, binary.LittleEndian, addMsg)

				parsedMsg, err := ParseMsg(message.MSG_TYPE_ADDED, msg.Bytes())
				_, ok := parsedMsg.MsgBody.(message.MessageAdded)
				Expect(err).To(BeNil())
				Expect(ok).To(BeTrue())
			})
		})
		Context("with raw msg as MSG_TYPE_UPDATED", func() {
			It("should parse the message correctly", func() {
				updateMsg := message.MessageUpdated{
					Symbol:      [3]byte{1, 2, 3},
					OrderId:     uint64(123),
					Side:        [1]byte{1},
					ReservedOne: [3]byte{1, 2, 3},
					Size:        uint64(123),
					Price:       int32(123),
					ReservedTwo: [4]byte{1, 2, 3, 4},
				}
				var msg bytes.Buffer
				binary.Write(&msg, binary.LittleEndian, updateMsg)

				parsedMsg, err := ParseMsg(message.MSG_TYPE_UPDATED, msg.Bytes())
				_, ok := parsedMsg.MsgBody.(message.MessageUpdated)
				Expect(err).To(BeNil())
				Expect(ok).To(BeTrue())
			})
		})
		Context("with raw msg as MSG_TYPE_DELETED", func() {
			It("should parse the message correctly", func() {
				delMsg := message.MessageDeleted{
					Symbol:  [3]byte{1, 2, 3},
					OrderId: uint64(123),
					Side:    [1]byte{1},
				}
				var msg bytes.Buffer
				binary.Write(&msg, binary.LittleEndian, delMsg)

				parsedMsg, err := ParseMsg(message.MSG_TYPE_DELETED, msg.Bytes())
				_, ok := parsedMsg.MsgBody.(message.MessageDeleted)
				Expect(err).To(BeNil())
				Expect(ok).To(BeTrue())
			})
		})
		Context("with raw msg as MSG_TYPE_EXECUTED", func() {
			It("should parse the message correctly", func() {
				exMsg := message.MessageExecuted{
					Symbol:    [3]byte{1, 2, 3},
					OrderId:   uint64(123),
					Side:      [1]byte{1},
					Reserved:  [3]byte{1, 2, 3},
					TradedQty: uint64(123),
				}
				var msg bytes.Buffer
				binary.Write(&msg, binary.LittleEndian, exMsg)

				parsedMsg, err := ParseMsg(message.MSG_TYPE_EXECUTED, msg.Bytes())
				_, ok := parsedMsg.MsgBody.(message.MessageExecuted)
				Expect(err).To(BeNil())
				Expect(ok).To(BeTrue())
			})
		})
	})

	Describe("Read", func() {

	})
})
