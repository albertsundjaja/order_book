package stream_handler

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/albertsundjaja/order_book/config"
	"github.com/albertsundjaja/order_book/internal/message"
)

// StreamHandler is the handler for reading stdin
type StreamHandler struct {
	config        *config.Config         // store app config
	buffer        []byte                 // store the buffer of the input stream
	lastHeader    *message.Header        // store last fully constructed header
	lastMsgType   string                 // store last read msg type
	orderBookChan chan<- message.Message // channel for sending message to OrderBook
	managerChan   chan bool              // for communicating with main routine
	input         io.Reader              // where to get the input from
}

func NewStreamHandler(config *config.Config, input io.Reader, managerChan chan bool, orderBookChan chan<- message.Message) *StreamHandler {
	return &StreamHandler{
		config:        config,
		lastHeader:    nil,
		orderBookChan: orderBookChan,
		managerChan:   managerChan,
		input:         input,
	}
}

// eat returns the slice from 0:count from the buffer
// it will then consume it after returning
// return an error if not enough bytes in the buffer
func (s *StreamHandler) eat(count int64) ([]byte, error) {
	if int64(len(s.buffer)) < count {
		return nil, fmt.Errorf("not enough buffer present to eat")
	}
	retBuf := s.buffer[0:count]
	s.buffer = s.buffer[count:]
	return retBuf, nil
}

// Start is the main process that read from stdin and parse the chunks
func (s *StreamHandler) Start() {
	//read the stdin in chunks
	reader := bufio.NewReader(s.input)
	part := make([]byte, 4096)
	var err error
	var count int
	for {
		// read until EOF
		if count, err = reader.Read(part); err != nil {
			break
		}
		// pass the data into our stream handler
		s.Read(part[:count])
	}
	if err == io.EOF {
		// extra time to allow OrderBook to finish (not required, but here so that the print statements are nicely ordered)
		time.Sleep(500 * time.Microsecond)
		log.Println("stream finished")
	} else {
		log.Printf("error while reading: %s \n", err.Error())
	}
	s.managerChan <- true
}

// Read read the raw message buffered from stdin
func (s *StreamHandler) Read(rawMsg []byte) {
	s.buffer = append(s.buffer, rawMsg...)
	for {
		if s.lastHeader == nil {
			rawHeader, err := s.eat(s.config.Stream.HeaderLength)
			if err != nil {
				break
			}

			var header message.Header
			err = binary.Read(bytes.NewReader(rawHeader), binary.LittleEndian, &header)
			if err != nil {
				panic(err)
			}
			s.lastHeader = &header
		}
		if s.lastMsgType == "" {
			rawType, err := s.eat(1)
			if err != nil {
				break
			}

			var msgType [1]byte
			err = binary.Read(bytes.NewReader(rawType), binary.LittleEndian, &msgType)
			if err != nil {
				panic(err)
			}
			s.lastMsgType = string(msgType[:])
		}
		if s.lastHeader != nil {
			// remove 1 as we extracted the msg type
			body, err := s.eat(int64(s.lastHeader.Size - 1))
			if err != nil {
				break
			}
			msg, err := ParseMsg(s.lastMsgType, body)
			if err != nil {
				panic(err)
			}
			msg.MsgHeader = *s.lastHeader
			s.lastHeader = nil
			s.lastMsgType = ""
			s.orderBookChan <- msg
		}
	}
}

// ParseMsg unmarshall the raw body received into a complete Message
func ParseMsg(msgType string, msg []byte) (message.Message, error) {
	var decodedMsg message.Message
	decodedMsg.MsgType = msgType
	switch msgType {
	case message.MSG_TYPE_ADDED:
		var msgAdded message.MessageAdded
		err := binary.Read(bytes.NewBuffer(msg), binary.LittleEndian, &msgAdded)
		if err != nil {
			return message.Message{}, err
		}
		decodedMsg.Symbol = msgAdded.Symbol
		decodedMsg.MsgBody = msgAdded
	case message.MSG_TYPE_UPDATED:
		var msgUpdated message.MessageUpdated
		err := binary.Read(bytes.NewBuffer(msg), binary.LittleEndian, &msgUpdated)
		if err != nil {
			return message.Message{}, err
		}
		decodedMsg.Symbol = msgUpdated.Symbol
		decodedMsg.MsgBody = msgUpdated
	case message.MSG_TYPE_DELETED:
		var msgDeleted message.MessageDeleted
		err := binary.Read(bytes.NewBuffer(msg), binary.LittleEndian, &msgDeleted)
		if err != nil {
			return message.Message{}, err
		}
		decodedMsg.Symbol = msgDeleted.Symbol
		decodedMsg.MsgBody = msgDeleted
	case message.MSG_TYPE_EXECUTED:
		var msgExecuted message.MessageExecuted
		err := binary.Read(bytes.NewBuffer(msg), binary.LittleEndian, &msgExecuted)
		if err != nil {
			return message.Message{}, err
		}
		decodedMsg.Symbol = msgExecuted.Symbol
		decodedMsg.MsgBody = msgExecuted
	default:
		return message.Message{}, fmt.Errorf("unrecognized message type")
	}

	return decodedMsg, nil
}
