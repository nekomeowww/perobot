package twitter_public_types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayText(t *testing.T) {
	result := &TweetResultsResult{
		Legacy: &TweetResultsResultLegacy{
			DisplayTextRange: []int{0, 5},
			FullText:         "你好，世界 https://t.co/SHORTURL",
		},
	}

	displayText := result.DisplayText()
	assert.Equal(t, "你好，世界", displayText)
}

func TestDisplayTextWithURLsMapped(t *testing.T) {
	result := &TweetResultsResult{
		Legacy: &TweetResultsResultLegacy{
			DisplayTextRange: []int{0, 59},
			Entities: &TweetResultsResultLegacyEntities{
				URLs: []TweetResultsResultLegacyEntitiesURL{
					{
						DisplayURL:  "pixiv.net/users/1234",
						ExpandedURL: "https://pixiv.net/users/1234",
						URL:         "https://t.co/ABCD",
						Indices: []int{
							71,
							94,
						},
					},
				},
			},
			FullText: "你好，世界（some）word word2 #hashtag1 #hashtag2 https://t.co/ABCD https://t.co/SHORTURL",
		},
	}

	displayText := result.DisplayTextWithURLsMapped()
	assert.Equal(t, "你好，世界（some）word word2 #hashtag1 #hashtag2 https://pixiv.net/users/1234", displayText)
}

func TestDisplayTextWithURLsMappedEmbeddedInMarkdownURL(t *testing.T) {
	result := &TweetResultsResult{
		Legacy: &TweetResultsResultLegacy{
			DisplayTextRange: []int{0, 59},
			Entities: &TweetResultsResultLegacyEntities{
				URLs: []TweetResultsResultLegacyEntitiesURL{
					{
						DisplayURL:  "pixiv.net/users/1234",
						ExpandedURL: "https://pixiv.net/users/1234",
						URL:         "https://t.co/ABCD",
						Indices: []int{
							71,
							94,
						},
					},
				},
			},
			FullText: "你好，世界（some）word word2 #hashtag1 #hashtag2 https://t.co/ABCD https://t.co/SHORTURL",
		},
	}

	displayText := result.DisplayTextWithURLsMappedEmbeddedInMarkdownURL()
	assert.Equal(t, "你好，世界（some）word word2 #hashtag1 #hashtag2 [pixiv.net/users/1234](https://pixiv.net/users/1234)", displayText)
}
