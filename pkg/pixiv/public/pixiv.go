package pixiv_public

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"

	"github.com/nekomeowww/perobot/pkg/options"
	pixiv_public_types "github.com/nekomeowww/perobot/pkg/pixiv/public/types"
)

type ClientOptions struct {
	Logger *logrus.Entry
}

func WithLogger(logger *logrus.Entry) options.CallOptions[ClientOptions] {
	return options.NewCallOptions(func(o *ClientOptions) {
		o.Logger = logger
	})
}

type Client struct {
	reqClient *req.Client
	logger    *logrus.Entry
}

func NewClient(phpSESSID string, callOpts ...options.CallOptions[ClientOptions]) (*Client, error) {
	if phpSESSID == "" {
		return nil, fmt.Errorf("must supply a valid Pixiv Token PHPSESSID in configs or environment variable")
	}

	opts := options.ApplyCallOptions(callOpts, ClientOptions{
		Logger: logrus.NewEntry(logrus.New()),
	})

	c := req.
		C().
		SetCommonHeader("authority", "www.pixiv.net").
		SetCommonHeader("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6").
		SetCommonHeader("cache-control", "no-cache").
		SetCommonHeader("Referer", "https://www.pixiv.net/").
		SetCommonCookies(&http.Cookie{
			Name:  "PHPSESSID",
			Value: phpSESSID,
		}).
		SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.52")

	client := &Client{
		reqClient: c,
		logger:    opts.Logger,
	}

	return client, nil
}

type ArtworkPage struct {
	Global       *pixiv_public_types.Global           `json:"global"`
	Preload      *pixiv_public_types.Preload          `json:"preload"`
	IllustDetail *pixiv_public_types.IllustDetailResp `json:"illustDetail"`
}

func (c *Client) ArtworkPage(illustID string) (*ArtworkPage, error) {
	resp, err := c.reqClient.R().
		SetHeader("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9").
		Get(fmt.Sprintf("https://www.pixiv.net/artworks/%s", illustID))
	if err != nil {
		c.logger.Errorf("failed to get pixiv artwork page, err: %v", err)
		return nil, err
	}
	if !resp.IsSuccess() {
		c.logger.Errorf("failed to pixiv artwork page, status code: %d, full request: %s", resp.StatusCode, resp.Dump())
		return nil, fmt.Errorf("request to %s failed: status code: %d", resp.Request.URL, resp.StatusCode)
	}

	dom, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.logger.Errorf("failed to parse pixiv artwork page from response body, err: %v", err)
		return nil, err
	}

	globalDataContent, ok := dom.Find("#meta-global-data").Attr("content")
	if !ok {
		c.logger.Error("failed to find global data content in pixiv artwork page body")
		return nil, fmt.Errorf("failed to find global data content in pixiv artwork page body")
	}

	var global pixiv_public_types.Global
	err = json.Unmarshal([]byte(globalDataContent), &global)
	if err != nil {
		c.logger.Errorf("failed to unmarshal global data content, err: %v", err)
		return nil, err
	}

	preloadDataContent, ok := dom.Find("#meta-preload-data").Attr("content")
	if !ok {
		c.logger.Error("failed to find preload data content")
		return nil, fmt.Errorf("failed to find preload data content")
	}

	var preload pixiv_public_types.Preload
	err = json.Unmarshal([]byte(preloadDataContent), &preload)
	if err != nil {
		c.logger.Errorf("failed to unmarshal preload data content, err: %v", err)
		return nil, err
	}

	illustDetail, err := c.IllustDetail(illustID)
	if err != nil {
		return nil, err
	}

	return &ArtworkPage{
		Global:       &global,
		Preload:      &preload,
		IllustDetail: illustDetail,
	}, nil
}

// IllustDetail 获取 Pixiv 画作详情
//
// https://natescarlet.github.io/pixiv/artwork.html
//
// https://pkg.go.dev/github.com/NateScarlet/pixiv@v0.7.0/pkg/artwork#Artwork
func (c *Client) IllustDetail(illustID string) (*pixiv_public_types.IllustDetailResp, error) {
	var illustDetail pixiv_public_types.IllustDetailResp

	resp, err := c.reqClient.R().
		SetResult(&illustDetail).
		Get(fmt.Sprintf("https://www.pixiv.net/ajax/illust/%s", illustID))
	if err != nil {
		c.logger.Errorf("failed to get illust detail: %v, full request: %s", err, resp.Dump())
		return nil, err
	}
	if !resp.IsSuccess() {
		c.logger.Errorf("failed to get illust detail, status code: %d, full request: %s", resp.StatusCode, resp.Dump())
		return nil, fmt.Errorf("request to %s failed: status code: %d", resp.Request.URL, resp.StatusCode)
	}
	if illustDetail.Error {
		c.logger.Errorf("failed to get illust detail, error: %s, full request: %s", illustDetail.Message, resp.Dump())
		return nil, fmt.Errorf("request to %s failed: error: %s", resp.Request.URL, illustDetail.Message)
	}

	return &illustDetail, nil
}

// IllustDetailPages 获取 Pixiv 画作分页详情
//
// https://natescarlet.github.io/pixiv/artwork.html#id2
//
// https://pkg.go.dev/github.com/NateScarlet/pixiv@v0.7.0/pkg/artwork#Artwork.FetchPages
func (c *Client) IllustDetailPages(illustID string) (*pixiv_public_types.IllustDetailPagesResp, error) {
	var illustDetailPages pixiv_public_types.IllustDetailPagesResp

	resp, err := c.reqClient.R().
		SetResult(&illustDetailPages).
		Get(fmt.Sprintf("https://www.pixiv.net/ajax/illust/%s/pages", illustID))
	if err != nil {
		c.logger.Errorf("failed to get illust detail pages: %v, full request: %s", err, resp.Dump())
		return nil, err
	}
	if !resp.IsSuccess() {
		c.logger.Errorf("failed to get illust detail pages, status code: %d, full request: %s", resp.StatusCode, resp.Dump())
		return nil, fmt.Errorf("request to %s failed: status code: %d", resp.Request.URL, resp.StatusCode)
	}
	if illustDetailPages.Error {
		c.logger.Errorf("failed to get illust detail pages, error: %s, full request: %s", illustDetailPages.Message, resp.Dump())
		return nil, fmt.Errorf("request to %s failed: error: %s", resp.Request.URL, illustDetailPages.Message)
	}

	return &illustDetailPages, nil
}

func (c *Client) GetImage(link string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)

	resp, err := c.reqClient.R().
		SetOutput(buffer).
		Get(link)
	if err != nil {
		c.logger.Errorf("failed to fetch pixiv image, err: %v, full request: %s", err, resp.Dump())
		return nil, err
	}
	if !resp.IsSuccess() {
		c.logger.Errorf("failed to fetch pixiv image, status code: %d, full request: %s", resp.StatusCode, resp.Dump())
		return nil, fmt.Errorf("failed to fetch pixiv image, status code: %d", resp.StatusCode)
	}

	return buffer, nil
}
