package order_book

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("SortedContainsInt32", func() {
		Context("with a slice that contains the number in ascending order", func() {
			It("should return the idx of the number", func() {
				expectedNum := int32(555)
				slice := []int32{1, 2, 3, 4, expectedNum, 1000, 2000}
				idx := SortedContainsInt32(true, slice, expectedNum)
				Expect(slice[idx]).To(Equal(expectedNum))
			})
		})
		Context("with a slice that does not contain the number in ascending order", func() {
			It("should return -1", func() {
				expectedNum := int32(555)
				slice := []int32{1, 2, 3, 4, 1000, 2000}
				idx := SortedContainsInt32(true, slice, expectedNum)
				Expect(idx).To(Equal(-1))
			})
		})
		Context("with a slice that contains the number in descending order", func() {
			It("should return the idx of the number", func() {
				expectedNum := int32(555)
				slice := []int32{777, 666, expectedNum}
				idx := SortedContainsInt32(false, slice, expectedNum)
				Expect(slice[idx]).To(Equal(expectedNum))
			})
		})
		Context("with a slice that does not contain the number in descending order", func() {
			It("should return -1", func() {
				expectedNum := int32(555)
				slice := []int32{777, 666, 444}
				idx := SortedContainsInt32(true, slice, expectedNum)
				Expect(idx).To(Equal(-1))
			})
		})
	})
})
