package monitor_service

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_http_client"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var shortcutIconModel = monitor_model.ShortcutIcon{}

func ListShortcutIconsByQuery(max int64, query monitor_model.ShortcutIcon) (*[]monitor_model.ShortcutIcon, error) {
	return listShortcutIcons(max, query)
}

func CountShortcutIcon() (int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&shortcutIconModel).Count(&count)

	return count, result.Error
}

func listShortcutIcons(max int64, query monitor_model.ShortcutIcon) (*[]monitor_model.ShortcutIcon, error) {
	db := monitor_db.GetDB()

	count, _ := CountShortcutIcon()

	// 如果 max 大于记录总数或者 max 小于等于 0, 则返回所有记录.
	if max >= count || max <= 0 {
		max = count
	}

	icons := make([]monitor_model.ShortcutIcon, max)

	result := db.Where(&query).Limit(int(max)).Model(&shortcutIconModel).Find(&icons)

	return &icons, result.Error
}

// RefreshShortcutIcons 从 https://github.com/simple-icons/simple-icons 拉取图标列表, 并更新到数据库.
func RefreshShortcutIcons() error {
	db := monitor_db.GetDB()

	storedIcons, err := listShortcutIcons(0, monitor_model.ShortcutIcon{})
	storedIconSlugMap := make(map[string]*monitor_model.ShortcutIcon, len(*storedIcons))
	for _, icon := range *storedIcons {
		storedIconSlugMap[icon.Slug] = &icon
	}

	icons, err := fetchShortcutIconFromSimpleIcons(false)
	if err != nil {
		return err
	} else if icons == nil {
		return nil
	}

	// 需要被创建的图标
	createdIcons := make([]monitor_model.ShortcutIcon, 0)
	// 需要被更新的图标
	updatedIcons := make([]monitor_model.ShortcutIcon, 0)
	for _, icon := range *icons {
		if storedIcon, ok := storedIconSlugMap[icon.Slug]; !ok {
			createdIcons = append(createdIcons, icon)
		} else {
			// 如果图标的品牌, 颜色和 slug 都相同, 则不需要更新.
			if storedIcon.Brand == icon.Brand && storedIcon.Color == icon.Color && storedIcon.Slug == icon.Slug {
				continue
			}

			icon.ID = storedIcon.ID
			updatedIcons = append(updatedIcons, icon)
		}
	}

	if result := db.Model(&shortcutIconModel).Create(&createdIcons); result.Error != nil {
		return result.Error
	} else {
		logger.Info("Created %d shortcut icons.", len(createdIcons))
	}

	for _, icon := range updatedIcons {
		if result := db.Model(&shortcutIconModel).Save(&icon); result.Error != nil {
			return result.Error
		}
	}
	logger.Info("Updated %d shortcut icons.", len(updatedIcons))

	return nil
}

var httpClient *comfy_http_client.Client
var eTag string

func init() {
	httpClient, _ = comfy_http_client.New("https://simpleicons.org", http.Header{}, time.Minute*5)
}

// fetchShortcutIconFromSimpleIcons 从 https://github.com/simple-icons/simple-icons 获取图标列表.
// 如果 force 为 true, 则强制从远程获取. 否则仅在 ETag 不匹配时拉取图标列表.
func fetchShortcutIconFromSimpleIcons(force bool) (*[]monitor_model.ShortcutIcon, error) {
	req, err := httpClient.Get("")
	if err != nil {
		return nil, err
	}

	if !force {
		req.Header.Set("If-None-Match", eTag)
	}

	res, err := httpClient.Send(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode == http.StatusNotModified {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, err
	}

	eTag = res.Header.Get("ETag")

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	icons := make([]monitor_model.ShortcutIcon, 0)
	doc.Find("li.grid-item").Each(func(i int, s *goquery.Selection) {
		// 忽略已废弃的图标
		if s.Find(".deprecated").Size() > 0 {
			return
		}

		iconUrl := s.Find("img.icon-preview").AttrOr("d-src", ".svg")
		slug := strings.Replace(filepath.Base(iconUrl), ".svg", "", 1)
		if len(slug) <= 0 {
			return
		}

		icon := monitor_model.ShortcutIcon{
			Brand: s.Find(".grid-item__title").Text(),
			Slug:  slug,
			Color: s.Find(".grid-item__color").Text(),
		}

		icons = append(icons, icon)
	})

	return &icons, nil
}
