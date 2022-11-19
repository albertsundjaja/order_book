package inmem_db

import (
	"fmt"
	"log"

	"github.com/albertsundjaja/order_book/message"
)

const (
	SORT_ORDER_BUY  = false // sort order descending
	SORT_ORDER_SELL = true  // sort order ascending
)

// orderBook is the item that stores all the orderId for a given symbol
type orderBook struct {
	depth       int               // a parameter to indicate how deep we should keep the aggregate
	Buy         map[uint64]*order // store map of all the buy orders with OrderId as key
	Sell        map[uint64]*order // store map of all the sell orders with OrderId as key
	AggBuy      map[int32]*order  // store aggregated buy data with price as key
	AggSell     map[int32]*order  // store aggregated sell data with price as key
	BuyDepth    []int32           // store all prices in AggBuy that is used for buy depth, sorted descending
	SellDepth   []int32           // store all prices in AggSell that is used for sell depth, sorted ascending
	shouldPrint bool              // flag indicating whether an update to orderBook should print new depth
}

// order is the data for individual order
type order struct {
	Index  int
	Volume uint64
	Price  int32
}

// newOrder create new Order
func newOrder(volume uint64, price int32) *order {
	return &order{
		Volume: volume,
		Price:  price,
	}
}

// newOrderBook init an empty orderBook
func newOrderBook(depth int) *orderBook {
	return &orderBook{
		Buy:     make(map[uint64]*order),
		Sell:    make(map[uint64]*order),
		AggBuy:  make(map[int32]*order),
		AggSell: make(map[int32]*order),
		depth:   depth,
	}
}

// PrintDepth print the depth to the console
func (o *orderBook) printDepth() string {
	buyDepth := ""
	lenBuyDepth := min(o.depth, len(o.BuyDepth))
	for idx, val := range o.BuyDepth[:lenBuyDepth] {
		buyDepth += fmt.Sprintf("(%d, %d)", o.AggBuy[val].Price, o.AggBuy[val].Volume)
		if idx < lenBuyDepth-1 {
			buyDepth += ", "
		}
	}
	sellDepth := ""
	lenSellDepth := min(o.depth, len(o.SellDepth))
	for idx, val := range o.SellDepth[:lenSellDepth] {
		sellDepth += fmt.Sprintf("(%d, %d)", o.AggSell[val].Price, o.AggSell[val].Volume)
		if idx < lenSellDepth-1 {
			sellDepth += ", "
		}
	}
	return fmt.Sprintf("[%s], [%s]", buyDepth, sellDepth)
}

// ShouldPrint return the flag whether we should print after the prev update
func (o *orderBook) ShouldPrint() bool {
	return o.shouldPrint
}

// AddOrder add the buy/sell order from the symbol into the symbol order book map
func (o *orderBook) addOrder(addMsg message.MessageAdded) error {
	order := newOrder(addMsg.Size, addMsg.Price)
	switch addMsg.Side[0] {
	case message.SIDE_BUY:
		if _, ok := o.Buy[addMsg.OrderId]; ok {
			return fmt.Errorf("unable to add order for OrderId %d. OrderId already exists", addMsg.OrderId)
		}
		o.Buy[addMsg.OrderId] = order
		o.addAggBuy(order.Price, order.Volume)
	case message.SIDE_SELL:
		if _, ok := o.Sell[addMsg.OrderId]; ok {
			return fmt.Errorf("unable to add order for OrderId %d. OrderId already exists", addMsg.OrderId)
		}
		o.Sell[addMsg.OrderId] = order
		o.addAggSell(order.Price, order.Volume)
	default:
		return fmt.Errorf("unrecognized side for Add Msg. OrderId: %d. Received side: %s", addMsg.OrderId, string(addMsg.Side[:]))
	}
	return nil
}

// UpdateOrder update the specified order with new volume and price
func (o *orderBook) updateOrder(updateMsg message.MessageUpdated) error {
	var order *order
	var ok bool
	switch updateMsg.Side[0] {
	case message.SIDE_BUY:
		order, ok = o.Buy[updateMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to update order, orderId %d does not exist", updateMsg.OrderId)
		}
		o.decAggBuy(order.Price, order.Volume)
		o.addAggBuy(updateMsg.Price, updateMsg.Size)
	case message.SIDE_SELL:
		order, ok = o.Sell[updateMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to update order, orderId %d does not exist", updateMsg.OrderId)
		}
		o.decAggSell(order.Price, order.Volume)
		o.addAggSell(updateMsg.Price, updateMsg.Size)
	default:
		return fmt.Errorf("unrecognized side for Update Msg. OrderId: %d. Received side: %s", updateMsg.OrderId, string(updateMsg.Side[:]))
	}

	order.Price = updateMsg.Price
	order.Volume = updateMsg.Size
	return nil
}

// DeleteOrder delete the order for the given order
func (o *orderBook) deleteOrder(delMsg message.MessageDeleted) error {
	switch delMsg.Side[0] {
	case message.SIDE_BUY:
		order, ok := o.Buy[delMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to delete orderId %d. It does not exist", delMsg.OrderId)
		}
		o.decAggBuy(order.Price, order.Volume)
		delete(o.Buy, delMsg.OrderId)
	case message.SIDE_SELL:
		order, ok := o.Sell[delMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to delete orderId %d. It does not exist", delMsg.OrderId)
		}
		o.decAggSell(order.Price, order.Volume)
		delete(o.Sell, delMsg.OrderId)
	default:
		return fmt.Errorf("unrecognized side for Delete Msg. OrderId: %d. Received side: %s", delMsg.OrderId, string(delMsg.Side[:]))
	}
	return nil
}

// ExecuteOrder execute the given order and deleting from the order book if it exhaust all the volume
func (o *orderBook) executeOrder(exMsg message.MessageExecuted) error {
	switch exMsg.Side[0] {
	case message.SIDE_BUY:
		order, ok := o.Buy[exMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to execute orderId %d. It does not exist", exMsg.OrderId)
		}
		order.Volume -= exMsg.TradedQty
		o.decAggBuy(order.Price, exMsg.TradedQty)
		if order.Volume <= 0 {
			delete(o.Buy, exMsg.OrderId)
		}
	case message.SIDE_SELL:
		order, ok := o.Sell[exMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to execute orderId %d. It does not exist", exMsg.OrderId)
		}
		order.Volume -= exMsg.TradedQty
		o.decAggSell(order.Price, exMsg.TradedQty)
		if order.Volume <= 0 {
			delete(o.Sell, exMsg.OrderId)
		}
	default:
		return fmt.Errorf("unrecognized side for Execute Msg. OrderId: %d. Received side: %s", exMsg.OrderId, string(exMsg.Side[:]))
	}
	return nil
}

// add to AggBuy
func (o *orderBook) addAggBuy(price int32, size uint64) {
	order, ok := o.AggBuy[price]
	if !ok {
		order = newOrder(0, price)
		o.AggBuy[price] = order
	}
	order.Volume += size
	o.addBuyDepth(price)
	if i := SortedContainsInt32(SORT_ORDER_BUY, o.BuyDepth[0:min(len(o.BuyDepth), o.depth)], price); i != -1 {
		o.shouldPrint = true
	}
}

// dec AggBuy
func (o *orderBook) decAggBuy(price int32, size uint64) {
	order, ok := o.AggBuy[price]
	if !ok {
		log.Fatalf("price (%d) is not found when decreasing aggBuy! this is not supposed to happen.", price)
	}
	order.Volume -= size
	if i := SortedContainsInt32(SORT_ORDER_BUY, o.BuyDepth[0:min(len(o.BuyDepth), o.depth)], price); i != -1 {
		o.shouldPrint = true
	}
	if order.Volume == 0 {
		o.removeBuyDepth(price)
		delete(o.AggBuy, price)
	}
}

// add to AggSell
func (o *orderBook) addAggSell(price int32, size uint64) {
	order, ok := o.AggSell[price]
	if !ok {
		order = newOrder(0, price)
		o.AggSell[price] = order
	}
	order.Volume += size
	o.addSellDepth(price)
	if i := SortedContainsInt32(SORT_ORDER_SELL, o.SellDepth[0:min(len(o.SellDepth), o.depth)], price); i != -1 {
		o.shouldPrint = true
	}
}

// dec from AgSell
func (o *orderBook) decAggSell(price int32, size uint64) {
	order, ok := o.AggSell[price]
	if !ok {
		log.Fatalf("price (%d) is not found when decreasing aggSell! this is not supposed to happen.", price)
	}
	order.Volume -= size
	if i := SortedContainsInt32(SORT_ORDER_SELL, o.SellDepth[0:min(len(o.SellDepth), o.depth)], price); i != -1 {
		o.shouldPrint = true
	}
	if order.Volume == 0 {
		o.removeSellDepth(price)
		delete(o.AggSell, price)
	}
}

// add price into BuyDepth, ignoring it if it's already there
func (o *orderBook) addBuyDepth(price int32) {
	// search if price is already in BuyDepth
	var i int
	if i = SortedContainsInt32(SORT_ORDER_BUY, o.BuyDepth, price); i != -1 {
		return
	}
	o.BuyDepth = append(o.BuyDepth, price)
	// sort descending
	// insertion sort is used as BuyDepth is originally sorted, it is roughly O(n) for almost sorted array
	InsertiontSortInt32(o.BuyDepth, SORT_ORDER_BUY)
}

// remove price from BuyDepth, ignoring it if it's not present
func (o *orderBook) removeBuyDepth(price int32) {
	var i int
	if i = SortedContainsInt32(SORT_ORDER_BUY, o.BuyDepth, price); i == -1 {
		return
	}
	if len(o.BuyDepth) <= 1 {
		o.BuyDepth = nil
		return
	}
	o.BuyDepth = append(o.BuyDepth[:i], o.BuyDepth[i+1:]...)
}

// add price to SellDepth, ignoring it if it's present
func (o *orderBook) addSellDepth(price int32) {
	// search if price is already in SellDepth
	var i int
	if i = SortedContainsInt32(SORT_ORDER_SELL, o.SellDepth, price); i != -1 {
		return
	}
	o.SellDepth = append(o.SellDepth, price)
	// sort ascending
	// insertion sort is used as SellDepth is originally sorted, it is roughly O(n) for almost sorted array
	InsertiontSortInt32(o.SellDepth, SORT_ORDER_SELL)
}

// remove price from SellDepth, ignoring it if it's not present
func (o *orderBook) removeSellDepth(price int32) {
	var i int
	if i = SortedContainsInt32(SORT_ORDER_SELL, o.SellDepth, price); i == -1 {
		return
	}

	if len(o.SellDepth) <= 1 {
		o.SellDepth = nil
		return
	}
	o.SellDepth = append(o.SellDepth[:i], o.SellDepth[i+1:]...)
}
