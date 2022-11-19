package order_book_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOrderBook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OrderBook Suite")
}
