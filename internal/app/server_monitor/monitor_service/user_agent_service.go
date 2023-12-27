package monitor_service

import (
	"encoding/xml"
	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"io"
	"net/http"
)

var userAgentModel = monitor_model.UserAgent{}

func CountUserAgent(query monitor_model.UserAgent) (int64, error) {
	db := monitor_db.GetDB()

	count := int64(0)
	result := db.Model(&userAgentModel).Where(query).Count(&count)

	return count, result.Error
}

// RefreshUserAgent 从 https://techpatterns.com/downloads/firefox/useragentswitcher.xml 拉取 userAgent 列表, 并更新到数据库.
func RefreshUserAgent() error {
	db := monitor_db.GetDB()

	storedUserAgents, err := listUserAgents(0, monitor_model.UserAgent{})
	storedUserAgentMap := make(map[string]*monitor_model.UserAgent, len(*storedUserAgents))
	for _, userAgent := range *storedUserAgents {
		storedUserAgentMap[userAgent.UserAgent] = &userAgent
	}

	userAgents, err := fetchUserAgent()
	if err != nil {
		return err
	} else if userAgents == nil {
		return nil
	}

	// 需要被创建的 userAgent
	createdUserAgents := make([]monitor_model.UserAgent, 0)
	// 需要被更新的 userAgent
	updatedUserAgents := make([]monitor_model.UserAgent, 0)
	for _, userAgent := range *userAgents {
		if storedUserAgent, ok := storedUserAgentMap[userAgent.UserAgent]; !ok {
			createdUserAgents = append(createdUserAgents, userAgent)
		} else {
			// 如果 userAgent 的描述和 UserAgent 都相同, 则不需要更新.
			if storedUserAgent.Description == userAgent.Description && storedUserAgent.UserAgent == userAgent.UserAgent {
				continue
			}

			userAgent.ID = storedUserAgent.ID
			updatedUserAgents = append(updatedUserAgents, userAgent)
		}
	}

	if result := db.Model(&userAgentModel).Create(&createdUserAgents); result.Error != nil {
		return result.Error
	} else {
		logger.Info("Created %d user agents.", len(createdUserAgents))
	}

	for _, userAgent := range updatedUserAgents {
		if result := db.Model(&userAgentModel).Save(&userAgent); result.Error != nil {
			return result.Error
		}
	}
	logger.Info("Updated %d user agents.", len(updatedUserAgents))

	return nil
}

func RandomUserAgent() (*monitor_model.UserAgent, error) {
	db := monitor_db.GetDB()

	userAgent := monitor_model.UserAgent{}
	result := db.Model(&userAgentModel).Order("RANDOM()").First(&userAgent)

	return &userAgent, result.Error
}

func listUserAgents(max int64, query monitor_model.UserAgent) (*[]monitor_model.UserAgent, error) {
	db := monitor_db.GetDB()

	count, _ := CountUserAgent(query)

	// 如果 max 大于记录总数或者 max 小于等于 0, 则返回所有记录.
	if max >= count || max <= 0 {
		max = count
	}

	icons := make([]monitor_model.UserAgent, max)

	result := db.Where(&query).Limit(int(max)).Model(&userAgentModel).Find(&icons)

	return &icons, result.Error
}

// fetchUserAgent 从 https://techpatterns.com/downloads/firefox/useragentswitcher.xml 拉取 userAgent 列表
func fetchUserAgent() (*[]monitor_model.UserAgent, error) {
	req, err := httpClient.Get("https://techpatterns.com/downloads/firefox/useragentswitcher.xml")
	if err != nil {
		return nil, err
	}

	res, err := httpClient.Send(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, err
	}

	type _UserAgent struct {
		XMLName     xml.Name `xml:"useragent"`
		UserAgent   string   `xml:"useragent,attr"`
		Description string   `xml:"description,attr"`
		AppCodeName string   `xml:"appcodename,attr"`
		AppName     string   `xml:"appname,attr"`
		AppVersion  string   `xml:"appversion,attr"`
		Platform    string   `xml:"platform,attr"`
		Vendor      string   `xml:"vendor,attr"`
		VendorSub   string   `xml:"vendorsub,attr"`
	}
	type _UserAgentFolder struct {
		Description string       `xml:"description,attr"`
		UserAgents  []_UserAgent `xml:"useragent"`
	}
	type _UserAgentNestedFolder struct {
		_UserAgentFolder
		Folders []_UserAgentFolder `xml:"folder"`
	}
	type _UserAgentSwitcher struct {
		XMLName xml.Name                 `xml:"useragentswitcher"`
		Folders []_UserAgentNestedFolder `xml:"folder"`
	}

	xmlBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	userAgentSwitcher := _UserAgentSwitcher{}
	err = xml.Unmarshal(xmlBytes, &userAgentSwitcher)
	if err != nil {
		return nil, err
	}

	for _, folder := range userAgentSwitcher.Folders {
		for _, userAgent := range folder.UserAgents {
			_ = userAgent
		}

		for _, folder := range folder.Folders {
			for _, userAgent := range folder.UserAgents {
				_ = userAgent
			}
		}
	}

	xmlUserAgents := lo.FlatMap(userAgentSwitcher.Folders, func(folder _UserAgentNestedFolder, index int) []_UserAgent {
		list := lo.Map(folder.Folders, func(folder _UserAgentFolder, index int) []_UserAgent {
			return folder.UserAgents
		})
		list = append(list, folder.UserAgents)
		return lo.Flatten(list)
	})

	xmlUserAgents = lo.Filter(xmlUserAgents, func(item _UserAgent, index int) bool {
		return len(item.UserAgent) > 0
	})

	userAgents := make([]monitor_model.UserAgent, len(xmlUserAgents), len(xmlUserAgents))
	for i, _userAgent := range xmlUserAgents {
		userAgents[i] = monitor_model.UserAgent{
			UserAgent:   _userAgent.UserAgent,
			Description: _userAgent.Description,
			AppCodeName: _userAgent.AppCodeName,
			AppName:     _userAgent.AppName,
			AppVersion:  _userAgent.AppVersion,
			Platform:    _userAgent.Platform,
			Vendor:      _userAgent.Vendor,
			VendorSub:   _userAgent.VendorSub,
		}
	}

	return &userAgents, nil
}
