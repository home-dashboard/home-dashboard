package monitor_service

import (
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/pkg/database"
	"testing"
)

func TestFetchShortcutIconFromSimpleIcons(t *testing.T) {
	if _, err := fetchShortcutIconFromSimpleIcons(false); err != nil {
		t.Error(err)
	}
	if got, err := fetchShortcutIconFromSimpleIcons(false); err != nil {
		t.Error(err)
	} else if got != nil {
		t.Error("fetchShortcutIconFromSimpleIcons() should return nil")
	}
}

func TestFetchShortcutIconFromSimpleIconsForce(t *testing.T) {
	if got, err := fetchShortcutIconFromSimpleIcons(true); err != nil {
		t.Error(err)
	} else if got == nil {
		t.Error("fetchShortcutIconFromSimpleIconsForce() should not return nil")
	}
	if got, err := fetchShortcutIconFromSimpleIcons(true); err != nil {
		t.Error(err)
	} else if got == nil {
		t.Error("fetchShortcutIconFromSimpleIconsForce() should not return nil")
	}
}

func TestCreateOrUpdateShortcutIcons(t *testing.T) {
	// 使用默认的内存数据库
	db := database.GetDB()
	monitor_db.Initial(db)

	if err := RefreshShortcutIcons(); err != nil {
		t.Error(err)
	}
}
