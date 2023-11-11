package verison_info

import (
	"golang.org/x/mod/semver"
	"strings"
	"time"
)

var (
	Version = ""
	Commit  = ""
	Date    = time.Now()
)

// SetVersion 设置版本号. 会自动添加 "v" 前缀.
func SetVersion(version string) {
	if !strings.HasPrefix(strings.ToLower(version), "v") {
		v := strings.Join([]string{"v", version}, "")
		if semver.IsValid(v) {
			version = v
		}
	}

	Version = version
}

func SetCommit(commit string) {
	Commit = commit
}

func SetDate(date time.Time) {
	Date = date
}

func Set(version, commit string, date time.Time) {
	SetVersion(version)
	SetCommit(commit)
	SetDate(date)
}
