package order_book

// Order is the data for individual order
type Order struct {
	Index  int
	Volume uint64
	Price  int32
}

// NewOrder create new Order
func NewOrder(volume uint64, price int32) *Order {
	return &Order{
		Volume: volume,
		Price:  price,
	}
}
