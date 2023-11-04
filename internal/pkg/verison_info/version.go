package verison_info

import "time"

var (
	Version = ""
	Commit  = ""
	Date    = time.Now()
)

func SetVersion(version string) {
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
