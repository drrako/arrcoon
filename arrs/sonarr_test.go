package arrs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/h2non/gock"
)

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
