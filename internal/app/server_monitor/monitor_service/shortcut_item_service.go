package monitor_service

import (
	"compress/flate"
	"compress/gzip"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/file_service"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"gorm.io/gorm/clause"
	"mime"
	url2 "net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var shortcutItemModel = monitor_model.ShortcutItem{}

func CreateOrUpdateShortcutItems(items []monitor_model.ShortcutItem) ([]monitor_model.ShortcutItem, error) {
	db := monitor_db.GetDB()

	affected := make([]monitor_model.ShortcutItem, len(items))
	for i, item := range items {
		model := db.Model(&shortcutItemModel)

		if item.ID != 0 {
			result := model.Where(monitor_model.ShortcutItem{Model: monitor_model.Model{ID: item.ID}}).Assign(item).FirstOrCreate(&item)
			if result.Error != nil {
				return nil, result.Error
			}
		}

		if result := db.Save(&item); result.Error != nil {
			return nil, result.Error
		}
		affected[i] = item
	}

	return affected, nil
}

// DeleteShortcutItems 删除 monitor_model.ShortcutItem.
// 如果快捷方式已被 monitor_model.ShortcutSection 引用, 则也会删除 monitor_model.ShortcutItem 与 monitor_model.ShortcutSection 的关联关系.
func DeleteShortcutItems(ids []uint) error {
	db := monitor_db.GetDB()

	items := make([]monitor_model.ShortcutItem, len(ids))
	for i, id := range ids {
		items[i] = monitor_model.ShortcutItem{Model: monitor_model.Model{ID: id}}
	}
	if result := db.Select("Sections").Delete(&items); result.Error != nil {
		return result.Error
	}

	return nil
}

func ListShortcutItemsByQuery(query monitor_model.ShortcutItem, preload []string) (*[]monitor_model.ShortcutItem, error) {
	return listShortcutItemsWithPreload(query, preload, []string{})
}

func ListShortcutItemsByFuzzyQuery(query monitor_model.ShortcutItem, likes []string, preload []string) (*[]monitor_model.ShortcutItem, error) {
	return listShortcutItemsWithPreload(query, preload, likes)
}

func CountShortcutItem(query monitor_model.ShortcutItem) (int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&shortcutItemModel).Where(query).Count(&count)

	return count, result.Error
}

func listShortcutItemsWithPreload(query monitor_model.ShortcutItem, preload []string, likes []string) (*[]monitor_model.ShortcutItem, error) {
	db := monitor_db.GetDB()

	items := make([]monitor_model.ShortcutItem, 0)

	model := db.Model(&shortcutItemModel)
	for _, p := range preload {
		model = model.Preload(p)
	}

	// 如果有 like 条件, 则将 like 条件从 query 中移除, 并将 like 条件添加到 likeClauses 中.
	queryValue := reflect.ValueOf(&query).Elem()
	likeClauses := make([]clause.Expression, len(likes))
	for _, l := range likes {
		field := queryValue.FieldByName(l)
		likeClauses = append(likeClauses, clause.Like{Column: l, Value: strings.Join([]string{"%", field.String(), "%"}, "")})
		field.Set(reflect.Zero(field.Type()))
	}
	model = model.Clauses(likeClauses...)

	result := model.Where(&query).Find(&items)

	return &items, result.Error
}

func RefreshCachedShortcutItemImageIcon(items *[]monitor_model.ShortcutItem) error {
	for _, item := range *items {
		if cachedUrl := GetCachedShortcutItemImageIconUrl(item); len(cachedUrl) <= 0 {
			continue
		} else {
			item.IconCachedUrl = cachedUrl
		}

		if _, err := CreateOrUpdateShortcutItems([]monitor_model.ShortcutItem{item}); err != nil {
			return err
		}
	}

	return nil
}

// GetCachedShortcutItemImageIconUrl 获取快捷方式图标的缓存 url. 如果快捷方式图标不是 monitor_model.ShortcutItemIconTypeUrl 类型,
// 或者快捷方式的 iconUrl 为空, 则返回空字符串.
func GetCachedShortcutItemImageIconUrl(item monitor_model.ShortcutItem) string {
	if item.IconType != monitor_model.ShortcutItemIconTypeUrl || item.IconUrl == "" {
		logger.Error("item %s(%d) icon type is not url or icon url is empty", item.Title, item.ID)
		return ""
	}

	if url, err := url2.Parse(item.IconUrl); err != nil {
		return ""
	} else if iconCachedUrl, err := cacheImageIconFromUrl(url); err != nil {
		logger.Error("cache %s image icon from %s failed: %w", item.Title, item.IconUrl, err)
		return ""
	} else {
		return iconCachedUrl
	}
}

// cacheImageIconFromUrl 从 url 拉取图标, 并将图标缓存到文件系统中. 返回图标在文件系统中的相对路径.
func cacheImageIconFromUrl(url *url2.URL) (string, error) {
	req, err := httpClient.Get(url.String())
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Host", url.Host)
	ua, err := RandomUserAgent()
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua.UserAgent)
	res, err := httpClient.Send(req)
	if err != nil {
		return "", err
	}

	decompressedReader := res.Body
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		decompressedReader, err = gzip.NewReader(res.Body)
		if err != nil {
			return "", err
		}
	case "deflate":
		decompressedReader = flate.NewReader(res.Body)
	}

	ext := ""
	if extensions, err := mime.ExtensionsByType(res.Header.Get("Content-Type")); len(extensions) <= 0 || err != nil {
		ext = filepath.Ext(filepath.FromSlash(url.Path))
	} else {
		ext = extensions[0]
	}

	fs, err := file_service.Get()
	if err != nil {
		return "", err
	}

	fileSystemPath := filepath.Join("shortcut", "icon", strconv.FormatInt(time.Now().UnixNano(), 10)+ext)
	if err := fs.Save(fileSystemPath, decompressedReader); err != nil {
		return "", err
	}
	iconCachedUrl, err := url2.JoinPath("", strings.Split(fileSystemPath, string(filepath.Separator))...)
	if err != nil {
		return "", err
	}

	return iconCachedUrl, nil
}
