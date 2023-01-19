package twitter_public_types

type TimelineThreadedConversationWithInjectionsV2 struct {
	Instructions []*TimelineInstruction `json:"instructions"`
}

func (t *TimelineThreadedConversationWithInjectionsV2) FindOneTweetEntry() *TimelineInstructionEntryItem {
	for _, instruction := range t.Instructions {
		if instruction.Type == TimelineInstructionTypeTimelineAddEntries {
			return instruction.FindOneTweetEntry()
		}
	}

	return nil
}

type TimelineInstructionType string

const (
	TimelineInstructionTypeTimelineAddEntries        TimelineInstructionType = "TimelineAddEntries"
	TimelineInstructionTypeTimelineTerminateTimeline TimelineInstructionType = "TimelineTerminateTimeline"
)

type TimelineInstruction struct {
	Type      TimelineInstructionType         `json:"type"`
	Direction string                          `json:"direction,omitempty"`
	Entries   []*TimelineInstructionEntryItem `json:"entries,omitempty"`
}

func (i *TimelineInstruction) FindOneTweetEntry() *TimelineInstructionEntryItem {
	return i.FindOneEntryByContentEntryType(TimelineInstructionEntryItemEntryTypeTimelineTimelineItem)
}

func (i *TimelineInstruction) FindOneEntryByContentEntryType(entryType TimelineInstructionEntryItemContentEntryType) *TimelineInstructionEntryItem {
	for _, entry := range i.Entries {
		if entry.Content != nil && entry.Content.EntryType == entryType {
			return entry
		}
	}

	return nil
}

type TimelineInstructionEntryItem struct {
	EntryID   string                               `json:"entryId"`
	SortIndex string                               `json:"sortIndex"`
	Content   *TimelineInstructionEntryItemContent `json:"content"`
}

func (e *TimelineInstructionEntryItem) TweetResults() *TweetResultsResult {
	if e.Content == nil {
		return nil
	}
	if e.Content.ItemContent == nil {
		return nil
	}
	if e.Content.ItemContent.TweetResults == nil {
		return nil
	}

	return e.Content.ItemContent.TweetResults.Result
}

type TimelineInstructionEntryItemContentEntryType string

const (
	TimelineInstructionEntryItemEntryTypeTimelineTimelineItem   TimelineInstructionEntryItemContentEntryType = "TimelineTimelineItem"
	TimelineInstructionEntryItemEntryTypeTimelineTimelineModule TimelineInstructionEntryItemContentEntryType = "TimelineTimelineModule"
)

type TimelineInstructionEntryItemContent struct {
	EntryType   TimelineInstructionEntryItemContentEntryType    `json:"entryType"`
	Typename    string                                          `json:"__typename"`
	ItemContent *TimelineInstructionEntryItemContentItemContent `json:"itemContent"`
}

type TimelineInstructionEntryItemContentItemContentItemType string

const (
	TimelineInstructionEntryItemContentItemContentTimelineTweet TimelineInstructionEntryItemContentItemContentItemType = "TimelineTweet"
)

type TimelineInstructionEntryItemContentItemContent struct {
	ItemType            TimelineInstructionEntryItemContentItemContentItemType `json:"itemType"`
	Typename            string                                                 `json:"__typename"`
	TweetResults        *TweetResults                                          `json:"tweet_results"`
	TweetDisplayType    string                                                 `json:"tweetDisplayType"`
	HasModeratedReplies bool                                                   `json:"hasModeratedReplies"`
}
