package arrs

import (
	testutils "arrcoon/testing"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/mock"
)

type MockTorrentClient struct {
	mock.Mock
}

func (m *MockTorrentClient) RemoveTorrents(hashes []string) {
	m.Called(hashes)
}

func (m *MockTorrentClient) Test() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestSonarrTestEvent(t *testing.T) {
	defer gock.Off()

	testUrl := "http://localhost"

	sonarr := NewSonarr("testdir", testUrl, "testtoken", nil)
	gock.InterceptClient(sonarr.restClient.GetClient())

	gock.New(testUrl).
		Get("/api").
		Reply(200).
		BodyString(`{"current": "v3"}`)

	assert.True(t, sonarr.testApi())
	assert.True(t, gock.IsDone())
}

func TestPartiallyRemovedSeason(t *testing.T) {
	defer gock.Off()

	mockTorrentClient := &MockTorrentClient{}
	// Assert that nothing is removed
	mockTorrentClient.On("RemoveTorrents", []string(nil)).Return(nil)

	testUrl := "http://localhost"

	sonarr := NewSonarr("testdir", testUrl, "testtoken", mockTorrentClient)
	gock.InterceptClient(sonarr.restClient.GetClient())

	gock.New(testUrl).
		Get("/api/v3/history/series").
		MatchParams(map[string]string{
			"includeEpisode": "false",
			"includeSeries":  "false",
			"seriesId":       "85",
		}).
		Reply(200).
		JSON(testutils.LoadJson("history_season_partially_removed"))

	sonarr.removeOutdatedTorrents(85, nil)

	assert.True(t, gock.IsDone())
	mock.AssertExpectationsForObjects(t, mockTorrentClient)
}

func TestEntirelyRemovedSeason(t *testing.T) {
	defer gock.Off()

	mockTorrentClient := &MockTorrentClient{}
	// Assert that season with all removed episodes is removed from the torrent client
	mockTorrentClient.On("RemoveTorrents", []string{"BBBBB4F4132C4AC7031F5692F36AC77A2ECBCCBB"}).Return(nil)

	testUrl := "http://localhost"

	sonarr := NewSonarr("testdir", testUrl, "testtoken", mockTorrentClient)
	gock.InterceptClient(sonarr.restClient.GetClient())

	gock.New(testUrl).
		Get("/api/v3/history/series").
		MatchParams(map[string]string{
			"includeEpisode": "false",
			"includeSeries":  "false",
			"seriesId":       "85",
		}).
		Reply(200).
		JSON(testutils.LoadJson("history_season_removed"))

	os.Setenv("sonarr_series_id", "85")
	os.Setenv("sonarr_episodefile_id", "1512")
	os.Setenv("sonarr_episodefile_episodeids", "3752")

	sonarr.HandleEvent("EpisodeFileDelete")

	assert.True(t, gock.IsDone())
	mock.AssertExpectationsForObjects(t, mockTorrentClient)
}
