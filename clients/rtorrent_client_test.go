package clients

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockXmlrpcClient struct {
	mock.Mock
}

func (m *MockXmlrpcClient) Call(method string, args any, reply any) error {
	callArgs := m.Called(method, args, reply)
	return callArgs.Error(0)
}

func TestRtorrentRemove_EmptyHashes(t *testing.T) {
	mockXmlrpc := new(MockXmlrpcClient)
	client := RtorrentClient{xmlrpcClient: mockXmlrpc}

	client.RemoveTorrents(nil)
	mockXmlrpc.AssertNotCalled(t, "Call")
}

func TestRtorrentRemove_OnlyEmptyHash(t *testing.T) {
	mockXmlrpc := new(MockXmlrpcClient)
	client := RtorrentClient{xmlrpcClient: mockXmlrpc}

	client.RemoveTorrents([]string{""})
	mockXmlrpc.AssertNotCalled(t, "Call")
}

func TestRtorrentRemove_Success(t *testing.T) {
	mockXmlrpc := new(MockXmlrpcClient)
	client := RtorrentClient{xmlrpcClient: mockXmlrpc}

	hash := "AAA65110BA16EF7839C27604B41AB083C832D83C"

	mockXmlrpc.On("Call", "system.multicall", mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			reply := args.Get(2).(*any)
			*reply = []any{
				[]any{"1"},
			}
		})

	client.RemoveTorrents([]string{hash})

	mock.AssertExpectationsForObjects(t, mockXmlrpc)
}

func TestRtorrentRemove_RetryOnError(t *testing.T) {
	mockXmlrpc := new(MockXmlrpcClient)
	client := RtorrentClient{xmlrpcClient: mockXmlrpc}

	hash := "AAA65110BA16EF7839C27604B41AB083C832D83C"

	mockXmlrpc.On("Call", "system.multicall", mock.Anything, mock.Anything).
		Return(errors.New("Fault(-503): could not find expected element string")).
		Times(3)

	client.RemoveTorrents([]string{hash})

	mock.AssertExpectationsForObjects(t, mockXmlrpc)
}

func TestRtorrentRemove_MixedEmptyAndValidHashes(t *testing.T) {
	mockXmlrpc := new(MockXmlrpcClient)
	client := RtorrentClient{xmlrpcClient: mockXmlrpc}

	hash := "BBB65110BA16EF7839C27604B41AB083C832D83C"

	mockXmlrpc.On("Call", "system.multicall", mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			reply := args.Get(2).(*any)
			*reply = []any{
				[]any{"1"},
			}
		})

	client.RemoveTorrents([]string{"", hash, ""})

	mock.AssertExpectationsForObjects(t, mockXmlrpc)
}
