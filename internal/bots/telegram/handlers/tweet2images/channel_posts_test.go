package tweet2images

import (
	"log"
	"os"
	"testing"

	"github.com/nekomeowww/perobot/internal/lib"
	"github.com/nekomeowww/perobot/internal/models/twitter"
	"github.com/nekomeowww/perobot/internal/thirdparty"
	"github.com/stretchr/testify/assert"
)

var h *Handler

func TestMain(m *testing.M) {
	logger := lib.NewLogger()()
	twitterPublic, err := thirdparty.NewTwitterPublic()()
	if err != nil {
		log.Fatal(err)
	}

	twitterModel := twitter.NewModel()(twitter.NewModelParam{
		Logger:        logger,
		TwitterPublic: twitterPublic,
	})

	h = NewHandler()(NewHandlerParam{
		Logger:       logger,
		TwitterModel: twitterModel,
	})

	os.Exit(m.Run())
}

func TestTweetIDFromText(t *testing.T) {
	assert := assert.New(t)

	tweetID := TweetIDFromText("https://twitter.com/test_10/status/1234?query=test")
	assert.Equal("1234", tweetID)

	tweetID = TweetIDFromText("https://twitter.com/testaccount/status/1234?query=test")
	assert.Equal("1234", tweetID)

	tweetID = TweetIDFromText("https://twitter.com/testaccount/status/1234")
	assert.Equal("1234", tweetID)
}
