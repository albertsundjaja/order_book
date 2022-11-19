package order_book

import (
	"fmt"
	"log"
	"sort"

	"github.com/albertsundjaja/order_book/message"
)

const (
	SORT_ORDER_BUY  = false // sort order descending
	SORT_ORDER_SELL = true  // sort order ascending
	SIDE_BUY        = 66    // Buy side. "B" in uint8
	SIDE_SELL       = 83    // Sell side. "S" in uint8
)

// OrderBook contains all the buys and sells data for a symbol
type OrderBook struct {
	depth       int               // a parameter to indicate how deep we should keep the aggregate
	Buy         map[uint64]*Order // store map of all the buy orders with OrderId as key
	Sell        map[uint64]*Order // store map of all the sell orders with OrderId as key
	AggBuy      map[int32]*Order  // store aggregated buy data with price as key
	AggSell     map[int32]*Order  // store aggregated sell data with price as key
	BuyDepth    []int32           // store all prices in AggBuy that is used for buy depth, sorted descending
	SellDepth   []int32           // store all prices in AggSell that is used for sell depth, sorted ascending
	shouldPrint bool              // flag indicating whether an update to OrderBook should print new depth
}

// NewOrderBook init an empty OrderBook
func NewOrderBook(depth int) *OrderBook {
	return &OrderBook{
		Buy:     make(map[uint64]*Order),
		Sell:    make(map[uint64]*Order),
		AggBuy:  make(map[int32]*Order),
		AggSell: make(map[int32]*Order),
		depth:   depth,
	}
}

// PrintDepth print the depth to the console
func (o *OrderBook) PrintDepth(sequence uint32, symbol string) {
	buyDepth := ""
	lenBuyDepth := min(o.depth, len(o.BuyDepth))
	for idx, val := range o.BuyDepth[:lenBuyDepth] {
		buyDepth += fmt.Sprintf("(%d, %d)", o.AggBuy[val].Price, o.AggBuy[val].Volume)
		if idx < lenBuyDepth-1 {
			buyDepth += ","
		}
	}
	sellDepth := ""
	lenSellDepth := min(o.depth, len(o.SellDepth))
	for idx, val := range o.SellDepth[:lenSellDepth] {
		sellDepth += fmt.Sprintf("(%d, %d)", o.AggSell[val].Price, o.AggSell[val].Volume)
		if idx < lenSellDepth-1 {
			sellDepth += ","
		}
	}
	fmt.Printf("%d, %s, [%s], [%s] \n", sequence, symbol, buyDepth, sellDepth)
}

// AddOrder add the buy/sell order from the symbol into the symbol order book map
func (o *OrderBook) AddOrder(addMsg message.MessageAdded) error {
	order := NewOrder(addMsg.Size, addMsg.Price)
	switch addMsg.Side[0] {
	case SIDE_BUY:
		if _, ok := o.Buy[addMsg.OrderId]; ok {
			return fmt.Errorf("unable to add order for OrderId %d. OrderId already exists", addMsg.OrderId)
		}
		o.Buy[addMsg.OrderId] = order
		o.addAggBuy(order.Price, order.Volume)
	case SIDE_SELL:
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
func (o *OrderBook) UpdateOrder(updateMsg message.MessageUpdated) error {
	var order *Order
	var ok bool
	switch updateMsg.Side[0] {
	case SIDE_BUY:
		order, ok = o.Buy[updateMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to update order, orderId %d does not exist", updateMsg.OrderId)
		}
		o.decAggBuy(order.Price, order.Volume)
		o.addAggBuy(updateMsg.Price, updateMsg.Size)
	case SIDE_SELL:
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
func (o *OrderBook) DeleteOrder(delMsg message.MessageDeleted) error {
	switch delMsg.Side[0] {
	case SIDE_BUY:
		order, ok := o.Buy[delMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to delete orderId %d. It does not exist", delMsg.OrderId)
		}
		o.decAggBuy(order.Price, order.Volume)
		delete(o.Buy, delMsg.OrderId)
	case SIDE_SELL:
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
func (o *OrderBook) ExecuteOrder(exMsg message.MessageExecuted) error {
	switch exMsg.Side[0] {
	case SIDE_BUY:
		order, ok := o.Buy[exMsg.OrderId]
		if !ok {
			return fmt.Errorf("unable to execute orderId %d. It does not exist", exMsg.OrderId)
		}
		order.Volume -= exMsg.TradedQty
		o.decAggBuy(order.Price, exMsg.TradedQty)
		if order.Volume <= 0 {
			delete(o.Buy, exMsg.OrderId)
		}
	case SIDE_SELL:
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
func (o *OrderBook) addAggBuy(price int32, size uint64) {
	order, ok := o.AggBuy[price]
	if !ok {
		order = NewOrder(0, price)
		o.AggBuy[price] = order
	}
	order.Volume += size
	o.addBuyDepth(price)
	if i := SortedContainsInt32(SORT_ORDER_BUY, o.BuyDepth[0:min(len(o.BuyDepth), o.depth)], price); i != -1 {
		o.shouldPrint = true
	}
}

// dec AggBuy
func (o *OrderBook) decAggBuy(price int32, size uint64) {
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
func (o *OrderBook) addAggSell(price int32, size uint64) {
	order, ok := o.AggSell[price]
	if !ok {
		order = NewOrder(0, price)
		o.AggSell[price] = order
	}
	order.Volume += size
	o.addSellDepth(price)
	if i := SortedContainsInt32(SORT_ORDER_SELL, o.SellDepth[0:min(len(o.SellDepth), o.depth)], price); i != -1 {
		o.shouldPrint = true
	}
}

// dec from AgSell
func (o *OrderBook) decAggSell(price int32, size uint64) {
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
func (o *OrderBook) addBuyDepth(price int32) {
	// search if price is already in BuyDepth
	var i int
	if i = SortedContainsInt32(SORT_ORDER_BUY, o.BuyDepth, price); i != -1 {
		return
	}
	o.BuyDepth = append(o.BuyDepth, price)
	// sort descending
	sort.Slice(o.BuyDepth, func(i int, j int) bool { return o.BuyDepth[i] > o.BuyDepth[j] })
}

// remove price from BuyDepth, ignoring it if it's not present
func (o *OrderBook) removeBuyDepth(price int32) {
	var i int
	if i = SortedContainsInt32(SORT_ORDER_BUY, o.BuyDepth, price); i == -1 {
		return
	}
	if len(o.BuyDepth) <= 1 {
		o.BuyDepth = nil
		return
	}
	// replacing the removed index with last, slice the last and sort is faster (O(log n)) than moving all the elements (O(n))
	o.BuyDepth[i] = o.BuyDepth[len(o.BuyDepth)-1]
	o.BuyDepth = o.BuyDepth[:len(o.BuyDepth)-1]
	sort.Slice(o.BuyDepth, func(i int, j int) bool { return o.BuyDepth[i] > o.BuyDepth[j] })
}

// add price to SellDepth, ignoring it if it's present
func (o *OrderBook) addSellDepth(price int32) {
	// search if price is already in SellDepth
	var i int
	if i = SortedContainsInt32(SORT_ORDER_SELL, o.SellDepth, price); i != -1 {
		return
	}
	o.SellDepth = append(o.SellDepth, price)
	// sort ascending
	sort.Slice(o.SellDepth, func(i int, j int) bool { return o.SellDepth[i] < o.SellDepth[j] })
}

// remove price from SellDepth, ignoring it if it's not present
func (o *OrderBook) removeSellDepth(price int32) {
	var i int
	if i = SortedContainsInt32(SORT_ORDER_SELL, o.SellDepth, price); i == -1 {
		return
	}

	if len(o.SellDepth) <= 1 {
		o.SellDepth = nil
		return
	}
	// replacing the removed index with last, slice the last and sort is faster (O(log n)) than moving all the elements (O(n))
	o.SellDepth[i] = o.SellDepth[len(o.SellDepth)-1]
	o.SellDepth = o.SellDepth[:len(o.SellDepth)-1]
	sort.Slice(o.SellDepth, func(i int, j int) bool { return o.SellDepth[i] < o.SellDepth[j] })
}
