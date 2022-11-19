// Package message contains all the message format that are expected from the input
package message

const (
	MSG_TYPE_ADDED    = "A"
	MSG_TYPE_UPDATED  = "U"
	MSG_TYPE_DELETED  = "D"
	MSG_TYPE_EXECUTED = "E"
)

type Message struct {
	MsgType   string      // store the message type
	MsgHeader Header      // header of the message
	MsgBody   interface{} // body of the message can be MessageAdded, MessageDeleted, MessageUpdated, MessageExecuted
}

type MessageAdded struct {
	Symbol      [3]byte
	OrderId     uint64
	Side        [1]byte
	ReservedOne [3]byte
	Size        uint64
	Price       int32
	ReservedTwo [4]byte
}

type MessageUpdated struct {
	Symbol      [3]byte
	OrderId     uint64
	Side        [1]byte
	ReservedOne [3]byte
	Size        uint64
	Price       int32
	ReservedTwo [4]byte
}

type MessageDeleted struct {
	Symbol  [3]byte
	OrderId uint64
	Side    [1]byte
}

type MessageExecuted struct {
	Symbol    [3]byte
	OrderId   uint64
	Side      [1]byte
	Reserved  [3]byte
	TradedQty uint64
}

type Header struct {
	Seq  uint32
	Size uint32
}
