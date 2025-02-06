package clients

import (
	"net"
	"net/http"
	"time"

	"github.com/kolo/xmlrpc"
	log "github.com/sirupsen/logrus"
)

type RtorrentClient struct {
	xmlrpcClient *xmlrpc.Client
}

func NewRtorrentClient(config Clients) TorrentClient {
	defaultTransport := http.DefaultTransport.(*http.Transport)
	transport := &http.Transport{
		Proxy: defaultTransport.Proxy,
		DialContext: (&net.Dialer{
			Timeout:   4 * time.Second,
			KeepAlive: 4 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     defaultTransport.ForceAttemptHTTP2,
		MaxIdleConns:          defaultTransport.MaxIdleConns,
		IdleConnTimeout:       defaultTransport.IdleConnTimeout,
		TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
		ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
	}
	xmlrpcClient, err := xmlrpc.NewClient(config.Rtorrent.Host, transport)
	if err != nil {
		log.Error("Coulnd't initialize rtorrent client")
		return nil
	}
	return &RtorrentClient{xmlrpcClient: xmlrpcClient}
}

func (rc RtorrentClient) Test() bool {
	log.Info("Testing Rtorrent accessibility")
	var listedMethods []string
	err := rc.xmlrpcClient.Call("system.listMethods", []interface{}{}, &listedMethods)
	if err != nil {
		log.WithError(err).Error("Couldn't connect to rTorrent")
		return false
	}
	for _, m := range listedMethods {
		if m == "d.erase" {
			log.Info("Succesfully connected to rTorrent client")
			return true
		}
	}
	log.Error("Unsupported rTorrent instance")
	return false
}

func (rc RtorrentClient) RemoveTorrents(hashes []string) {
	if len(hashes) == 0 {
		return
	}
	log.WithFields(log.Fields{
		"Hashes": hashes,
	}).Info("Requesting torrent files removal")
	for _, hash := range hashes {
		var response interface{}
		deleteParams := []map[string]interface{}{
			{
				"methodName": "d.custom5.set",
				"params":     []interface{}{hash, "1"},
			},
			{
				"methodName": "d.delete_tied",
				"params":     []interface{}{hash},
			},
			{
				"methodName": "d.erase",
				"params":     []interface{}{hash},
			},
		}
		err := rc.xmlrpcClient.Call("system.multicall", deleteParams, &response)

		if err != nil {
			log.WithError(err).Error("Couldn't run erase call")
		}

		responseSlice, ok := response.([]interface{})
		if !ok {
			log.WithFields(log.Fields{
				"Error responses": response,
				"Hash":            hash,
			}).Error("Unknown response type")
			continue
		}

		var successResponses []interface{}
		var errorResponses []interface{}

		for _, responseMap := range responseSlice {
			if responseMapSlice, ok := responseMap.([]interface{}); ok {
				successResponse, ok := responseMapSlice[0].(string)
				if ok {
					successResponses = append(successResponses, successResponse)
				}
			} else if responseMapMap, ok := responseMap.(map[string]interface{}); ok {
				errorResponses = append(errorResponses, responseMapMap)
			}
		}

		if len(errorResponses) > 0 {
			// Debug level because it is expected that the same hash code could be passed to rTorrent multiple times
			log.WithFields(log.Fields{
				"Error responses": errorResponses,
				"Hash":            hash,
			}).Debug("Couldn't remove torrents")
			continue
		}

		if len(successResponses) > 0 {
			log.WithFields(log.Fields{
				"Hash": hash,
			}).Info("Torrent has been removed")
		}
	}
}
