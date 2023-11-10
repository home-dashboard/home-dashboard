package monitor_controller

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_http_client"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
	"time"
)

var httpClient *comfy_http_client.Client

func init() {
	httpClient, _ = comfy_http_client.New("", http.Header{}, time.Minute*5)
}

// ExtractShortcutItemInfoFromURL 从 URL 中提取网站信息, 并以 monitor_model.ShortcutItem 的形式返回.
// @Summary ExtractShortcutItemInfoFromURL
// @Description ExtractShortcutItemInfoFromURL
// @Tags ExtractShortcutItemInfoFromURL
// @Produce json
// @Param url query string true "url"
// @Success 200 {object} monitor_model.ShortcutItem
// @Router shortcut/item/extract-from-url [get]
func ExtractShortcutItemInfoFromURL(c *gin.Context) {
	var query struct {
		URL string `form:"url" json:"url" binding:"required"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	websiteUrl, err := url2.Parse(query.URL)
	if err != nil {
		respondUnknownError(c, err.Error())
		return
	}
	if len(strings.Split(websiteUrl.Host, ".")) <= 2 {
		websiteUrl.Host = strings.Join([]string{"www", websiteUrl.Host}, ".")
	}

	item, err := extractWebsiteInfoFromUrl(websiteUrl, c.Request.Header)
	if err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	if alternatives, err := monitor_service.ListShortcutItemsByFuzzyQuery(monitor_model.ShortcutItem{Title: item.Title, URL: websiteUrl.Hostname()}, []string{"Title", "URL"}, []string{"Icon"}); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"alternatives": alternatives, "item": item})
	}
}

func extractWebsiteInfoFromUrl(url *url2.URL, header http.Header) (*monitor_model.ShortcutItem, error) {
	req, err := httpClient.Get(url.String())
	if err != nil {
		return nil, err
	}
	req.Header = header
	// 不支持 brotli 压缩, 因为 Golang 没有原生的支持.
	//req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	res, err := httpClient.Send(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	decompressedReader := res.Body
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		decompressedReader, err = gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
	case "deflate":
		decompressedReader = flate.NewReader(res.Body)
	}

	htmlBytes, _ := io.ReadAll(decompressedReader)
	_, charsetValue, _ := charset.DetermineEncoding(htmlBytes, res.Header.Get("Content-Type"))
	decodeReader, err := charset.NewReader(bytes.NewReader(htmlBytes), charsetValue)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(decodeReader)
	if err != nil {
		return nil, err
	}

	var bestIconUrl string
	doc.Find("link[rel~='icon'][type='image/svg+xml']").Each(func(i int, s *goquery.Selection) {
		if iconUrl, ok := s.Attr("href"); ok {
			bestIconUrl = iconUrl
		}
	})
	if len(bestIconUrl) <= 0 {
		doc.Find("link[rel~='icon']").Each(func(i int, s *goquery.Selection) {
			if iconUrl, ok := s.Attr("href"); ok {
				bestIconUrl = iconUrl
			}
		})
	}
	if len(bestIconUrl) <= 0 {
		bestIconUrl = "/favicon.ico"
	}
	if iconUrl, err := url2.Parse(bestIconUrl); err != nil {
		return nil, err
	} else {
		bestIconUrl = url.ResolveReference(iconUrl).String()
	}

	return &monitor_model.ShortcutItem{
		Title:           doc.Find("title").Text(),
		Description:     doc.Find("meta[name*='description']").AttrOr("content", ""),
		URL:             url.String(),
		IconType:        monitor_model.ShortcutItemIconTypeUrl,
		IconUrl:         bestIconUrl,
		Tags:            "",
		Target:          monitor_model.ShortcutItemTargetTypeNewTab,
		StatusCheck:     true,
		StatusCheckUrl:  url.String(),
		BackgroundColor: "",
	}, nil
}

// CreateShortcutItem 创建 shortcut item.
// @Summary CreateShortcutItem
// @Description CreateShortcutItem
// @Tags CreateShortcutItem
// @Accept json
// @Produce json
// @Param shortcutItem body monitor_model.ShortcutItem true "body"
// @Success 200 {object} monitor_model.ShortcutItem
// @Router shortcut/item/create [post]
func CreateShortcutItem(c *gin.Context) {
	var body monitor_model.ShortcutItem

	if err := c.ShouldBindJSON(&body); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else if body.ID != 0 {
		respondEntityValidationError(c, "shortcut item should not have id")
		return
	}

	if count, err := monitor_service.CountShortcutItem(monitor_model.ShortcutItem{Title: body.Title}); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else if count > 0 {
		respondEntityAlreadyExistError(c, "shortcut item with title %s already exists", body.Title)
		return
	}

	if created, err := monitor_service.CreateOrUpdateShortcutItems([]monitor_model.ShortcutItem{body}); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, created[0])
	}
}

// ListShortcutItems 根据 sectionId 返回对应的 shortcut items.
// @Summary ListShortcutItems
// @Description ListShortcutItems
// @Tags ListShortcutItems
// @Produce json
// @Param sectionId query number false "section id"
// @Success 200 {array} monitor_model.ShortcutItem
// @Router shortcut/item/list [get]
func ListShortcutItems(c *gin.Context) {
	var body struct {
		SectionId uint `form:"sectionId" json:"sectionId" binding:"required"`
	}

	if err := c.ShouldBindQuery(&body); err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	sections, err := monitor_service.ListShortcutSectionsByQuery(0, monitor_model.ShortcutSection{Model: monitor_model.Model{ID: body.SectionId}}, []string{"Items"})
	if err != nil {
		respondUnknownError(c, err.Error())
		return
	} else if len(*sections) <= 0 {
		respondEntityNotFoundError(c, "section %d not found", body.SectionId)
		return
	}
	items := (*sections)[0].Items

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

// UpdateShortcutItem 更新 shortcut item.
// @Summary UpdateShortcutItem
// @Description UpdateShortcutItem
// @Tags UpdateShortcutItem
// @Accept json
// @Produce json
// @Param id path number true "id"
// @Param shortcutItem body monitor_model.ShortcutItem true "body"
// @Success 200
// @Router shortcut/item//update/{id} [put]
func UpdateShortcutItem(c *gin.Context) {
	var body monitor_model.ShortcutItem

	if err := c.ShouldBindJSON(&body); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else if ID, err := strconv.ParseUint(c.Param("id"), 10, 0); err != nil {
		respondEntityValidationError(c, "id should be number")
		return
	} else {
		body.ID = uint(ID)
	}

	if updated, err := monitor_service.CreateOrUpdateShortcutItems([]monitor_model.ShortcutItem{body}); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, updated[0])
	}
}

// DeleteShortcutItem 删除 shortcut item.
// @Summary DeleteShortcutItem
// @Description DeleteShortcutItem
// @Tags DeleteShortcutItem
// @Produce json
// @Success 200
// @Router shortcut/item/delete [delete]
func DeleteShortcutItem(c *gin.Context) {
	ids := make([]uint, 0)
	for _, idStr := range c.QueryArray("ids") {
		id, err := strconv.ParseUint(idStr, 10, 0)
		if err != nil {
			respondUnknownError(c, err.Error())
			return
		}

		ids = append(ids, uint(id))
	}

	if err := monitor_service.DeleteShortcutItems(ids); err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
