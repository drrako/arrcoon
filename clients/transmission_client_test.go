package clients

import (
	"context"
	"testing"

	"github.com/hekmon/transmissionrpc/v3"
	"github.com/stretchr/testify/mock"
)

type MockTransmissionRPC struct {
	mock.Mock
}

func (m *MockTransmissionRPC) RPCVersion(ctx context.Context) (ok bool, serverVersion int64, serverMinimumVersion int64, err error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Get(1).(int64), args.Get(2).(int64), args.Error(3)
}

func (m *MockTransmissionRPC) TorrentRemove(ctx context.Context, payload transmissionrpc.TorrentRemovePayload) (err error) {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockTransmissionRPC) TorrentGetAllForHashes(ctx context.Context, hashes []string) (torrents []transmissionrpc.Torrent, err error) {
	args := m.Called(ctx, hashes)
	return args.Get(0).([]transmissionrpc.Torrent), args.Error(1)
}

func TestTransmissionRemove(t *testing.T) {
	mockTransmissionClient := new(MockTransmissionRPC)
	id2 := int64(22)
	hashString2 := "BBB65110BA16EF7839C27604B41AB083C832D83C"
	removeHashes := []string{"AAA65110BA16EF7839C27604B41AB083C832D83C", "BBB65110BA16EF7839C27604B41AB083C832D83C"}
	mockTransmissionClient.On("TorrentGetAllForHashes", mock.Anything, removeHashes).Return([]transmissionrpc.Torrent{
		{ID: &id2, HashString: &hashString2},
	}, nil)
	mockTransmissionClient.On("TorrentRemove", mock.Anything, transmissionrpc.TorrentRemovePayload{IDs: []int64{id2}, DeleteLocalData: true}).Return(nil)

	client := TransmissionClient{transmissionClient: mockTransmissionClient}

	client.RemoveTorrents([]string{"AAA65110BA16EF7839C27604B41AB083C832D83C", "BBB65110BA16EF7839C27604B41AB083C832D83C"})

	mock.AssertExpectationsForObjects(t, mockTransmissionClient)
}
