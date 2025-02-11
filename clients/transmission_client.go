package clients

import (
	"context"
	"net/url"

	"github.com/hekmon/transmissionrpc/v3"
	log "github.com/sirupsen/logrus"
)

type TransmissionClient struct {
	transmissionClient TransmissionRPCInterface
}

type TransmissionRPCInterface interface {
	RPCVersion(ctx context.Context) (ok bool, serverVersion int64, serverMinimumVersion int64, err error)
	TorrentGetAllForHashes(ctx context.Context, hashes []string) (torrents []transmissionrpc.Torrent, err error)
	TorrentRemove(ctx context.Context, payload transmissionrpc.TorrentRemovePayload) (err error)
}

type TransmissionRPC struct {
	transmissionClient *transmissionrpc.Client
}

func (trpcw *TransmissionRPC) RPCVersion(ctx context.Context) (ok bool, serverVersion int64, serverMinimumVersion int64, err error) {
	return trpcw.transmissionClient.RPCVersion(ctx)
}

func (trpcw *TransmissionRPC) TorrentGetAllForHashes(ctx context.Context, hashes []string) (torrents []transmissionrpc.Torrent, err error) {
	return trpcw.transmissionClient.TorrentGetAllForHashes(ctx, hashes)
}

func (trpcw *TransmissionRPC) TorrentRemove(ctx context.Context, payload transmissionrpc.TorrentRemovePayload) (err error) {
	return trpcw.transmissionClient.TorrentRemove(ctx, payload)
}

func NewTransmissionClient(config ClientConfig) TorrentClient {
	endpoint, err := url.Parse(config["host"].(string))
	if err != nil {
		panic(err)
	}
	tbt, err := transmissionrpc.New(endpoint, nil)
	if err != nil {
		panic(err)
	}
	return &TransmissionClient{transmissionClient: &TransmissionRPC{tbt}}
}

func (tc TransmissionClient) Test() bool {
	ctx := context.Background()
	ok, serverVersion, serverMinimumVersion, err := tc.transmissionClient.RPCVersion(ctx)
	if err != nil {
		log.WithError(err).Error("Couldn't connect to transmission")
		return false
	}
	if !ok {
		log.WithFields(log.Fields{
			"Server version":  serverVersion,
			"Minimum version": serverMinimumVersion,
			"Library version": transmissionrpc.RPCVersion,
		}).Error("Remote transmission RPC version is incompatible with the transmission library")
		return false
	}
	log.WithFields(log.Fields{
		"Server version":         serverVersion,
		"Server minimum version": serverMinimumVersion,
	}).Info("Succesfully connected to transmission")
	return true
}

func (tc TransmissionClient) RemoveTorrents(hashes []string) {
	ctx := context.Background()
	if len(hashes) == 0 {
		return
	}
	torrents, err := tc.transmissionClient.TorrentGetAllForHashes(ctx, hashes)

	log.WithFields(log.Fields{
		"Torrents": torrents,
	}).Trace("Requested by hash torrents")

	if err != nil {
		log.WithError(err).Error("Could torrents hash data")
		return
	}

	for _, torrent := range torrents {
		payload := transmissionrpc.TorrentRemovePayload{
			IDs:             []int64{*torrent.ID},
			DeleteLocalData: true,
		}
		err := tc.transmissionClient.TorrentRemove(ctx, payload)
		if err != nil {
			log.
				WithFields(log.Fields{
					"Torrent hash": torrent.HashString,
					"Torrent ID":   torrent.ID,
				}).
				WithError(err).
				Error("Couldn't remove torrent")
		}
		log.WithFields(log.Fields{
			"Torrent hash": torrent.HashString,
			"Torrent ID":   torrent.ID,
		}).Info("Torrent has been removed")
	}
}
