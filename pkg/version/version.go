package version

import "fmt"

var (
	Version = "dev"
	Commit  = "none"
	BuiltBy = "unknown"
)

// GetVersionInfo 返回版本信息.
func GetVersionInfo() string {
	switch BuiltBy {
	case "goreleaser":
		return fmt.Sprintf("%s-%s", Version, Commit)
	default:
		return Version
	}
}
