package monitor_controller

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_service"
	"net/http"
	"strconv"
)

// CreateShortcutSection 创建快捷方式分组.
// @Summary CreateShortcutSection
// @Description CreateShortcutSection
// @Tags CreateShortcutSection
// @Produce json
// @Param shortcutSection body monitor_model.ShortcutSection true "body"
// @Success 200
// @Router shortcut/section/create [post]
func CreateShortcutSection(c *gin.Context) {
	var body monitor_model.ShortcutSection

	if err := c.ShouldBindJSON(&body); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else if body.ID != 0 {
		respondEntityValidationError(c, "ID must be 0")
		return
	}

	if count, err := monitor_service.CountShortcutSection(monitor_model.ShortcutSection{Name: body.Name}); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else if count > 0 {
		respondEntityAlreadyExistError(c, "shortcut section with name %s already exists", body.Name)
		return
	}

	if created, err := monitor_service.CreateOrUpdateShortcutSections([]monitor_model.ShortcutSection{body}); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, created[0])
	}
}

// ListShortcutSections 获取快捷方式分组.
// @Summary ListShortcutSections
// @Description ListShortcutSections
// @Tags ListShortcutSections
// @Produce json
// @Success 200 {array} monitor_model.ShortcutSection
// @Router shortcut/section/list [get]
func ListShortcutSections(c *gin.Context) {
	var body monitor_model.ShortcutSection

	if err := c.ShouldBindQuery(&body); err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	sections, err := monitor_service.ListShortcutSectionsByQuery(0, body, []string{"Items", "Items.Icon", "Items.Usages"})
	if err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	for i, section := range *sections {
		for j, item := range section.Items {
			(*sections)[i].Items[j].Usages = lo.Filter[monitor_model.ShortcutSectionItemUsage](item.Usages, func(usage monitor_model.ShortcutSectionItemUsage, _ int) bool {
				return usage.SectionId == section.ID
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sections": sections,
	})
}

// UpdateShortcutSection 更新快捷方式分组.
// @Summary UpdateShortcutSection
// @Description UpdateShortcutSection
// @Tags UpdateShortcutSection
// @Accept json
// @Produce json
// @Param shortcutSection body monitor_model.ShortcutSection true "body"
// @Success 200
// @Router shortcut/section/update/{id} [post]
func UpdateShortcutSection(c *gin.Context) {
	var body monitor_model.ShortcutSection

	if err := c.ShouldBindJSON(&body); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else if ID, err := strconv.ParseUint(c.Param("id"), 10, 0); err != nil {
		respondEntityValidationError(c, "id should be number")
		return
	} else {
		body.ID = uint(ID)
	}

	if affected, err := monitor_service.CreateOrUpdateShortcutSections([]monitor_model.ShortcutSection{body}); err != nil {
		respondUnknownError(c, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, affected[0])
	}
}

// DeleteShortcutSection 删除快捷方式分组.
// @Summary DeleteShortcutSection
// @Description DeleteShortcutSection
// @Tags DeleteShortcutSection
// @Produce json
// @Router shortcut/section/delete/{id} [delete]
func DeleteShortcutSection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	if err := monitor_service.DeleteShortcutSections([]uint{uint(id)}); err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// DeleteShortcutSectionItems 删除快捷方式分组中的项目.
// @Summary DeleteShortcutSectionItems
// @Description DeleteShortcutSectionItems
// @Tags DeleteShortcutSectionItems
// @Accept query
// @Produce json
// @Router shortcut/section/delete/{id}/items [delete]
func DeleteShortcutSectionItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 0)
	if err != nil {
		respondUnknownError(c, err.Error())
		return
	}

	itemIds := make([]uint, 0)
	for _, idStr := range c.QueryArray("itemIds") {
		id, err := strconv.ParseUint(idStr, 10, 0)
		if err != nil {
			respondUnknownError(c, err.Error())
			return
		}

		itemIds = append(itemIds, uint(id))
	}

	if err := monitor_service.DeleteShortcutSectionItems(uint(id), itemIds); err != nil {
		respondUnknownError(c, err.Error())
	}
	c.JSON(http.StatusOK, gin.H{})
}
