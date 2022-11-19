package stream_handler_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStreamHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StreamHandler Suite")
}
