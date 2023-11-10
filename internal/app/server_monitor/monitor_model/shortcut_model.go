package monitor_model

type ShortcutSection struct {
	Model
	Name    string         `json:"name"`
	Icon    string         `json:"icon"`
	Default bool           `json:"default"`
	Items   []ShortcutItem `json:"items" gorm:"many2many:shortcut_section_link_shortcut_item;"`
}

type ShortcutItemTargetType int

var (
	ShortcutItemTargetTypeSelfTab ShortcutItemTargetType = 0
	ShortcutItemTargetTypeNewTab  ShortcutItemTargetType = 1
	ShortcutItemTargetTypeEmbed   ShortcutItemTargetType = 2
)

type ShortcutItemIconType int

var (
	ShortcutItemIconTypeIcon ShortcutItemIconType = 0
	ShortcutItemIconTypeUrl  ShortcutItemIconType = 1
	ShortcutItemIconTypeText ShortcutItemIconType = 2
)

type ShortcutItem struct {
	Model
	Title           string                     `json:"title"`
	Description     string                     `json:"description"`
	URL             string                     `json:"url"`
	IconType        ShortcutItemIconType       `json:"iconType"`
	IconUrl         string                     `json:"iconUrl"`
	IconText        string                     `json:"iconText"`
	IconID          uint                       `json:"IconId"`
	Icon            ShortcutIcon               `json:"icon" gorm:"foreignKey:IconID"`
	Tags            string                     `json:"tags"`
	Target          ShortcutItemTargetType     `json:"target"`
	StatusCheck     bool                       `json:"statusCheck"`
	StatusCheckUrl  string                     `json:"statusCheckUrl"`
	BackgroundColor string                     `json:"backgroundColor"`
	Sections        []ShortcutSection          `json:"sections" gorm:"many2many:shortcut_section_link_shortcut_item;"`
	Usages          []ShortcutSectionItemUsage `json:"usages" gorm:"many2many:shortcut_section_item_link_shortcut_usage"`
}

type ShortcutIcon struct {
	Model
	Brand string `json:"brand"`
	Slug  string `json:"slug" gorm:"unique"`
	Color string `json:"color"`
}

// ShortcutSectionItemUsage 记录分组下的快捷方式使用情况.
type ShortcutSectionItemUsage struct {
	SectionId uint `json:"sectionId" gorm:"primaryKey;autoIncrement:false"`
	ItemId    uint `json:"itemId" gorm:"primaryKey;autoIncrement:false"`
	// 点击次数.
	ClickCount int `json:"clickCount"`
}
