package twitter_public_types

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// GetTweetDetailParamVariables TweetDetail GraphQL 参数，variables 参数部分
//
// https://github.com/fa0311/TwitterInternalAPIDocument/blob/master/docs/markdown/GraphQL.md#variables-63
type GetTweetDetailParamVariables struct {
	FocalTweetID                           string `json:"focalTweetId"`
	WithRuxInjections                      bool   `json:"with_rux_injections"`
	IncludePromotedContent                 bool   `json:"includePromotedContent"`
	WithCommunity                          bool   `json:"withCommunity"`
	WithQuickPromoteEligibilityTweetFields bool   `json:"withQuickPromoteEligibilityTweetFields"`
	WithBirdwatchNotes                     bool   `json:"withBirdwatchNotes"`
	WithSuperFollowsUserFields             bool   `json:"withSuperFollowsUserFields"`
	WithDownvotePerspective                bool   `json:"withDownvotePerspective"`
	WithReactionsMetadata                  bool   `json:"withReactionsMetadata"`
	WithReactionsPerspective               bool   `json:"withReactionsPerspective"`
	WithSuperFollowsTweetFields            bool   `json:"withSuperFollowsTweetFields"`
	WithVoice                              bool   `json:"withVoice"`
	WithV2Timeline                         bool   `json:"withV2Timeline"`
}

// GetTweetDetailParamFeatures TweetDetail GraphQL 参数，features 参数部分
//
// https://github.com/fa0311/TwitterInternalAPIDocument/blob/master/docs/markdown/GraphQL.md#features-63
type GetTweetDetailParamFeatures struct {
	ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled                 bool `json:"responsive_web_twitter_blue_verified_badge_is_enabled"`
	VerifiedPhoneLabelEnabled                                      bool `json:"verified_phone_label_enabled"`
	ResponsiveWebGraphqlTimelineNavigationEnabled                  bool `json:"responsive_web_graphql_timeline_navigation_enabled"`
	ViewCountsPublicVisibilityEnabled                              bool `json:"view_counts_public_visibility_enabled"`
	ViewCountsEverywhereAPIEnabled                                 bool `json:"view_counts_everywhere_api_enabled"`
	LongformNotetweetsConsumptionEnabled                           bool `json:"longform_notetweets_consumption_enabled"`
	TweetypieUnmentionOptimizationEnabled                          bool `json:"tweetypie_unmention_optimization_enabled"`
	ResponsiveWebUcGqlEnabled                                      bool `json:"responsive_web_uc_gql_enabled"`
	VibeAPIEnabled                                                 bool `json:"vibe_api_enabled"`
	ResponsiveWebEditTweetAPIEnabled                               bool `json:"responsive_web_edit_tweet_api_enabled"`
	GraphqlIsTranslatableRwebTweetIsTranslatableEnabled            bool `json:"graphql_is_translatable_rweb_tweet_is_translatable_enabled"`
	StandardizedNudgesMisinfo                                      bool `json:"standardized_nudges_misinfo"`
	TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled bool `json:"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled"`
	InteractiveTextEnabled                                         bool `json:"interactive_text_enabled"`
	ResponsiveWebTextConversationsEnabled                          bool `json:"responsive_web_text_conversations_enabled"`
	ResponsiveWebEnhanceCardsEnabled                               bool `json:"responsive_web_enhance_cards_enabled"`
}

var (
	// DefaultGetTweetDetailParamVariables 默认的 TweetDetail GET 请求参数，Variables 部分
	DefaultGetTweetDetailParamVariables = GetTweetDetailParamVariables{
		WithRuxInjections:                      false,
		IncludePromotedContent:                 true,
		WithCommunity:                          true,
		WithQuickPromoteEligibilityTweetFields: true,
		WithBirdwatchNotes:                     true,
		WithSuperFollowsUserFields:             true,
		WithDownvotePerspective:                false,
		WithReactionsMetadata:                  false,
		WithReactionsPerspective:               false,
		WithSuperFollowsTweetFields:            true,
		WithVoice:                              true,
		WithV2Timeline:                         true,
	}

	// DefaultGetTweetDetailParamFeatures 默认的 TweetDetail GET 请求参数，Features 部分
	DefaultGetTweetDetailParamFeatures = GetTweetDetailParamFeatures{
		ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled:                 true,
		VerifiedPhoneLabelEnabled:                                      false,
		ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
		ViewCountsPublicVisibilityEnabled:                              true,
		ViewCountsEverywhereAPIEnabled:                                 true,
		LongformNotetweetsConsumptionEnabled:                           false,
		TweetypieUnmentionOptimizationEnabled:                          true,
		ResponsiveWebUcGqlEnabled:                                      true,
		VibeAPIEnabled:                                                 true,
		ResponsiveWebEditTweetAPIEnabled:                               true,
		GraphqlIsTranslatableRwebTweetIsTranslatableEnabled:            true,
		StandardizedNudgesMisinfo:                                      true,
		TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
		InteractiveTextEnabled:                                         true,
		ResponsiveWebTextConversationsEnabled:                          false,
		ResponsiveWebEnhanceCardsEnabled:                               false,
	}
)

// TweetDetailResp TweetDetail GraphQL 返回数据
//
// https://github.com/fa0311/TwitterInternalAPIDocument/blob/master/docs/markdown/GraphQL.md#features-63
type TweetDetailResp struct {
	Data *TweetDetail `json:"data"`
}

type TweetDetail struct {
	ThreadedConversationWithInjectionsV2 *TimelineThreadedConversationWithInjectionsV2 `json:"threaded_conversation_with_injections_v2"`
}

type TweetResults struct {
	Result *TweetResultsResult `json:"result"`
}

type TweetResultsResult struct {
	Typename                string                                     `json:"__typename"`
	RestID                  string                                     `json:"rest_id"`
	HasBirdwatchNotes       bool                                       `json:"has_birdwatch_notes"`
	Core                    *TweetResultsResultCore                    `json:"core"`
	UnmentionData           any                                        `json:"unmention_data"`
	EditControl             *TweetResultsResultEditControl             `json:"edit_control"`
	EditPerspective         *TweetResultsResultEditPerspective         `json:"edit_perspective"`
	IsTranslatable          bool                                       `json:"is_translatable"`
	Legacy                  *TweetResultsResultLegacy                  `json:"legacy"`
	QuickPromoteEligibility *TweetResultsResultQuickPromoteEligibility `json:"quick_promote_eligibility"`
	Views                   *TweetResultsResultViews                   `json:"views"`
}

func (r *TweetResultsResult) FullText() string {
	if r.Legacy == nil {
		return ""
	}

	return r.Legacy.FullText
}

func (r *TweetResultsResult) DisplayText() string {
	if r.Legacy == nil {
		return ""
	}

	return applyDisplayTextRangeOnFullText(r.Legacy.FullText, r.Legacy.DisplayTextRange)
}

func (r *TweetResultsResult) DisplayTextWithURLsMapped() string {
	if r.Legacy == nil {
		return ""
	}

	displayText := r.DisplayText()
	if len(r.Legacy.Entities.URLs) == 0 {
		return displayText
	}

	for _, url := range r.Legacy.Entities.URLs {
		displayText = strings.ReplaceAll(displayText, url.URL, url.ExpandedURL)
	}

	return displayText
}

func (r *TweetResultsResult) DisplayTextWithURLsMappedEmbeddedInMarkdownURL() string {
	if r.Legacy == nil {
		return ""
	}

	displayText := r.DisplayText()
	if len(r.Legacy.Entities.URLs) == 0 {
		return displayText
	}

	for _, url := range r.Legacy.Entities.URLs {
		displayText = strings.ReplaceAll(displayText, url.URL, fmt.Sprintf("[%s](%s)", url.DisplayURL, url.ExpandedURL))
	}

	return displayText
}

func (r *TweetResultsResult) DisplayTextWithURLsMappedEmbeddedInHTML() string {
	if r.Legacy == nil {
		return ""
	}

	displayText := r.DisplayText()
	if len(r.Legacy.Entities.URLs) == 0 {
		return displayText
	}

	for _, url := range r.Legacy.Entities.URLs {
		displayText = strings.ReplaceAll(displayText, url.URL, fmt.Sprintf(`<a href="%s">%s</a>`, url.ExpandedURL, url.DisplayURL))
	}

	return displayText
}

func (r *TweetResultsResult) User() *UserResultsResultLegacy {
	if r.Core == nil {
		return nil
	}
	if r.Core.UserResults == nil {
		return nil
	}
	if r.Core.UserResults.Result == nil {
		return nil
	}

	return r.Core.UserResults.Result.Legacy
}

func applyDisplayTextRangeOnFullText(fullText string, displayTextRange []int) string {
	var substringStart int
	var substringEnd int
	if len(displayTextRange) == 2 {
		substringStart = displayTextRange[0]
		substringEnd = displayTextRange[1]
	} else {
		return fullText
	}

	return string([]rune(fullText)[substringStart:substringEnd])
}

func (r *TweetResultsResult) PhotoURLs() []string {
	if r.Legacy == nil {
		fmt.Println("Unrecognizable tweet, no legacy field found")
		return make([]string, 0)
	}
	if r.Legacy.ExtendedEntities == nil {
		fmt.Println("Unrecognizable tweet, no extended_entities field found")
		return make([]string, 0)
	}
	if len(r.Legacy.ExtendedEntities.Media) == 0 {
		fmt.Println("Tweet has no medias")
		return make([]string, 0)
	}

	photos := lo.Filter(r.Legacy.Entities.Media, func(media *EntityMedia, _ int) bool {
		return media.Type == TweetLegacyExtendedEntityMediaTypePhoto && media.MediaURLHTTPS != ""
	})

	return lo.Map(photos, func(photo *EntityMedia, _ int) string {
		return photo.MediaURLHTTPS
	})
}

func (r *TweetResultsResult) ExtendedPhotoURLs() []string {
	if r.Legacy == nil {
		fmt.Println("Unrecognizable tweet, no legacy field found")
		return make([]string, 0)
	}
	if r.Legacy.ExtendedEntities == nil {
		fmt.Println("Unrecognizable tweet, no extended_entities field found")
		return make([]string, 0)
	}
	if len(r.Legacy.ExtendedEntities.Media) == 0 {
		fmt.Println("Tweet has no medias")
		return make([]string, 0)
	}

	photos := lo.Filter(r.Legacy.ExtendedEntities.Media, func(media *ExtendedEntityMedia, _ int) bool {
		return media.Type == TweetLegacyExtendedEntityMediaTypePhoto && media.MediaURLHTTPS != ""
	})

	return lo.Map(photos, func(photo *ExtendedEntityMedia, _ int) string {
		return photo.MediaURLHTTPS
	})
}

func (r *TweetResultsResult) ExtendedGIFURLs() []string {
	if r.Legacy == nil {
		fmt.Println("Unrecognizable tweet, no legacy field found")
		return make([]string, 0)
	}
	if r.Legacy.ExtendedEntities == nil {
		fmt.Println("Unrecognizable tweet, no extended_entities field found")
		return make([]string, 0)
	}
	if len(r.Legacy.ExtendedEntities.Media) == 0 {
		fmt.Println("Tweet has no medias")
		return make([]string, 0)
	}

	gifs := lo.Filter(r.Legacy.ExtendedEntities.Media, func(media *ExtendedEntityMedia, _ int) bool {
		return media.Type == TweetLegacyExtendedEntityMediaTypeAnimatedGIF && media.VideoInfo != nil
	})

	return lo.Map(gifs, func(gif *ExtendedEntityMedia, _ int) string {
		return gif.VideoInfo.Variants[0].URL
	})
}

type TweetResultsResultCore struct {
	UserResults *UserResults `json:"user_results"`
}

type TweetResultsResultEditControl struct {
	EditTweetIds       []string `json:"edit_tweet_ids"`
	EditableUntilMsecs string   `json:"editable_until_msecs"`
	IsEditEligible     bool     `json:"is_edit_eligible"`
	EditsRemaining     string   `json:"edits_remaining"`
}

type TweetResultsResultEditPerspective struct {
	Favorited bool `json:"favorited"`
	Retweeted bool `json:"retweeted"`
}

type TweetResultsResultLegacy struct {
	CreatedAt                 string                                    `json:"created_at"`
	ConversationIDStr         string                                    `json:"conversation_id_str"`
	DisplayTextRange          []int                                     `json:"display_text_range"`
	Entities                  *TweetResultsResultLegacyEntities         `json:"entities"`
	ExtendedEntities          *TweetResultsResultLegacyExtendedEntities `json:"extended_entities"`
	FavoriteCount             int                                       `json:"favorite_count"`
	Favorited                 bool                                      `json:"favorited"`
	FullText                  string                                    `json:"full_text"`
	IsQuoteStatus             bool                                      `json:"is_quote_status"`
	Lang                      string                                    `json:"lang"`
	PossiblySensitive         bool                                      `json:"possibly_sensitive"`
	PossiblySensitiveEditable bool                                      `json:"possibly_sensitive_editable"`
	QuoteCount                int                                       `json:"quote_count"`
	ReplyCount                int                                       `json:"reply_count"`
	RetweetCount              int                                       `json:"retweet_count"`
	Retweeted                 bool                                      `json:"retweeted"`
	UserIDStr                 string                                    `json:"user_id_str"`
	IDStr                     string                                    `json:"id_str"`
}

type TweetResultsResultLegacyEntities struct {
	Media        []*EntityMedia                        `json:"media"`
	UserMentions []interface{}                         `json:"user_mentions"`
	URLs         []TweetResultsResultLegacyEntitiesURL `json:"urls"`
	Hashtags     []interface{}                         `json:"hashtags"`
	Symbols      []interface{}                         `json:"symbols"`
}

type TweetResultsResultLegacyEntitiesURL struct {
	DisplayURL  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
	URL         string `json:"url"`
	Indices     []int  `json:"indices"`
}

type TweetResultsResultLegacyExtendedEntities struct {
	Media []*ExtendedEntityMedia `json:"media"`
}

type TweetResultsResultQuickPromoteEligibility struct {
	Eligibility string `json:"eligibility"`
}

type TweetResultsResultViews struct {
	Count string `json:"count"`
	State string `json:"state"`
}
