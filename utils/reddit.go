package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
)

type Subreddit struct {
	Name string
}

type RedditResponse []RedditResponseElement

type RedditResponseElement struct {
	Kind string         `json:"kind"`
	Data RedditPostData `json:"data"`
}

type RedditPostData struct {
	After     interface{} `json:"after"`
	Dist      *int64      `json:"dist"`
	Modhash   string      `json:"modhash"`
	GeoFilter string      `json:"geo_filter"`
	Children  []Child     `json:"children"`
	Before    interface{} `json:"before"`
}

type Child struct {
	Kind string     `json:"kind"`
	Data RedditPost `json:"data"`
}

type RedditPost struct {
	ApprovedAtUTC              interface{}   `json:"approved_at_utc"`
	Subreddit                  string        `json:"subreddit"`
	Selftext                   string        `json:"selftext"`
	UserReports                []interface{} `json:"user_reports"`
	Saved                      bool          `json:"saved"`
	ModReasonTitle             interface{}   `json:"mod_reason_title"`
	Gilded                     int64         `json:"gilded"`
	Clicked                    bool          `json:"clicked"`
	Title                      string        `json:"title"`
	LinkFlairRichtext          []interface{} `json:"link_flair_richtext"`
	SubredditNamePrefixed      string        `json:"subreddit_name_prefixed"`
	Hidden                     bool          `json:"hidden"`
	Pwls                       int64         `json:"pwls"`
	LinkFlairCSSClass          interface{}   `json:"link_flair_css_class"`
	Downs                      int64         `json:"downs"`
	ThumbnailHeight            int64         `json:"thumbnail_height"`
	TopAwardedType             interface{}   `json:"top_awarded_type"`
	ParentWhitelistStatus      string        `json:"parent_whitelist_status"`
	HideScore                  bool          `json:"hide_score"`
	Name                       string        `json:"name"`
	Quarantine                 bool          `json:"quarantine"`
	LinkFlairTextColor         string        `json:"link_flair_text_color"`
	UpvoteRatio                int64         `json:"upvote_ratio"`
	AuthorFlairBackgroundColor interface{}   `json:"author_flair_background_color"`
	SubredditType              string        `json:"subreddit_type"`
	UPS                        int64         `json:"ups"`
	TotalAwardsReceived        int64         `json:"total_awards_received"`
	MediaEmbed                 Gildings      `json:"media_embed"`
	ThumbnailWidth             int64         `json:"thumbnail_width"`
	AuthorFlairTemplateID      interface{}   `json:"author_flair_template_id"`
	IsOriginalContent          bool          `json:"is_original_content"`
	AuthorFullname             string        `json:"author_fullname"`
	SecureMedia                interface{}   `json:"secure_media"`
	IsRedditMediaDomain        bool          `json:"is_reddit_media_domain"`
	IsMeta                     bool          `json:"is_meta"`
	Category                   interface{}   `json:"category"`
	SecureMediaEmbed           Gildings      `json:"secure_media_embed"`
	LinkFlairText              interface{}   `json:"link_flair_text"`
	CanModPost                 bool          `json:"can_mod_post"`
	Score                      int64         `json:"score"`
	ApprovedBy                 interface{}   `json:"approved_by"`
	IsCreatedFromAdsUI         bool          `json:"is_created_from_ads_ui"`
	AuthorPremium              bool          `json:"author_premium"`
	Thumbnail                  string        `json:"thumbnail"`
	Edited                     bool          `json:"edited"`
	AuthorFlairCSSClass        interface{}   `json:"author_flair_css_class"`
	AuthorFlairRichtext        []interface{} `json:"author_flair_richtext"`
	Gildings                   Gildings      `json:"gildings"`
	PostHint                   string        `json:"post_hint"`
	ContentCategories          interface{}   `json:"content_categories"`
	IsSelf                     bool          `json:"is_self"`
	ModNote                    interface{}   `json:"mod_note"`
	Created                    int64         `json:"created"`
	LinkFlairType              string        `json:"link_flair_type"`
	Wls                        int64         `json:"wls"`
	RemovedByCategory          interface{}   `json:"removed_by_category"`
	BannedBy                   interface{}   `json:"banned_by"`
	AuthorFlairType            string        `json:"author_flair_type"`
	Domain                     string        `json:"domain"`
	AllowLiveComments          bool          `json:"allow_live_comments"`
	SelftextHTML               interface{}   `json:"selftext_html"`
	Likes                      interface{}   `json:"likes"`
	SuggestedSort              interface{}   `json:"suggested_sort"`
	BannedAtUTC                interface{}   `json:"banned_at_utc"`
	URLOverriddenByDest        string        `json:"url_overridden_by_dest"`
	ViewCount                  interface{}   `json:"view_count"`
	Archived                   bool          `json:"archived"`
	NoFollow                   bool          `json:"no_follow"`
	IsCrosspostable            bool          `json:"is_crosspostable"`
	Pinned                     bool          `json:"pinned"`
	Over18                     bool          `json:"over_18"`
	Preview                    Preview       `json:"preview"`
	AllAwardings               []interface{} `json:"all_awardings"`
	Awarders                   []interface{} `json:"awarders"`
	MediaOnly                  bool          `json:"media_only"`
	CanGild                    bool          `json:"can_gild"`
	Spoiler                    bool          `json:"spoiler"`
	Locked                     bool          `json:"locked"`
	AuthorFlairText            interface{}   `json:"author_flair_text"`
	TreatmentTags              []interface{} `json:"treatment_tags"`
	Visited                    bool          `json:"visited"`
	RemovedBy                  interface{}   `json:"removed_by"`
	NumReports                 interface{}   `json:"num_reports"`
	Distinguished              interface{}   `json:"distinguished"`
	SubredditID                string        `json:"subreddit_id"`
	AuthorIsBlocked            bool          `json:"author_is_blocked"`
	ModReasonBy                interface{}   `json:"mod_reason_by"`
	RemovalReason              interface{}   `json:"removal_reason"`
	LinkFlairBackgroundColor   string        `json:"link_flair_background_color"`
	ID                         string        `json:"id"`
	IsRobotIndexable           bool          `json:"is_robot_indexable"`
	NumDuplicates              int64         `json:"num_duplicates"`
	ReportReasons              interface{}   `json:"report_reasons"`
	Author                     string        `json:"author"`
	DiscussionType             interface{}   `json:"discussion_type"`
	NumComments                int64         `json:"num_comments"`
	SendReplies                bool          `json:"send_replies"`
	Media                      interface{}   `json:"media"`
	ContestMode                bool          `json:"contest_mode"`
	AuthorPatreonFlair         bool          `json:"author_patreon_flair"`
	AuthorFlairTextColor       interface{}   `json:"author_flair_text_color"`
	Permalink                  string        `json:"permalink"`
	WhitelistStatus            string        `json:"whitelist_status"`
	Stickied                   bool          `json:"stickied"`
	URL                        string        `json:"url"`
	SubredditSubscribers       int64         `json:"subreddit_subscribers"`
	CreatedUTC                 int64         `json:"created_utc"`
	NumCrossposts              int64         `json:"num_crossposts"`
	ModReports                 []interface{} `json:"mod_reports"`
	IsVideo                    bool          `json:"is_video"`
}

type Gildings struct {
}

type Preview struct {
	Images  []Image `json:"images"`
	Enabled bool    `json:"enabled"`
}

type Image struct {
	Source      Source   `json:"source"`
	Resolutions []Source `json:"resolutions"`
	Variants    Gildings `json:"variants"`
	ID          string   `json:"id"`
}

type Source struct {
	URL    string `json:"url"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

func (s *Subreddit) RandomHot() (*RedditPost, error) {
	client := resty.New().SetRedirectPolicy(
		resty.FlexibleRedirectPolicy(3),
	)

	resp, err := client.R().SetHeader("Host", "reddit.com").SetHeader("User-Agent", "go:io.astralapp.bots:v1.0.0 (by /u/AmusedGrap)").Get(fmt.Sprintf("https://www.reddit.com/r/%s/random.json", s.Name))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New("Error fetching random post")
	}

	var posts RedditResponseElement

	err = json.Unmarshal(resp.Body(), &posts)

	if err != nil {
		return nil, err
	}

	if len(posts.Data.Children) == 0 {
		return nil, errors.New("No posts found")
	}

	rand.Seed(time.Now().Unix())
	random := rand.Intn(len(posts.Data.Children))

	return &posts.Data.Children[random].Data, nil
}
