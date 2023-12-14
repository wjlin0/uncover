package update

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/logrusorgru/aurora"
	"github.com/tj/go-update"
	githubUpdateStore "github.com/tj/go-update/stores/github"
	"github.com/wjlin0/pathScan/pkg/util"
	"strings"
)

// Aurora instance
var Aurora aurora.Aurora = aurora.NewAurora(true)

func CheckVersion(Owner, Repo, ordVersion string) (string, error) {

	version := strings.Replace(ordVersion, "v", "", 1)
	m := &update.Manager{
		Store: &githubUpdateStore.Store{
			Owner:   Owner,
			Repo:    Repo,
			Version: version,
		},
	}
	releases, err := m.LatestReleases()
	if err != nil {
		return ordVersion, err
	}
	if len(releases) == 0 {
		return ordVersion, err
	}
	if util.CheckVersion(version, releases[0].Version) {
		return releases[0].Version, nil
	}

	return ordVersion, err

}

// GetVersionDescription returns tags like (latest) or (outdated) or (dev)
func GetVersionDescription(current string, latest string) string {
	if strings.HasSuffix(current, "-dev") {
		if IsDevReleaseOutdated(current, latest) {
			return fmt.Sprintf("(%v)", Aurora.BrightRed("outdated"))
		} else {
			return fmt.Sprintf("(%v)", Aurora.Blue("development"))
		}
	}
	if IsOutdated(current, latest) {
		return fmt.Sprintf("(%v)", Aurora.BrightRed("outdated"))
	} else {
		return fmt.Sprintf("(%v)", Aurora.BrightGreen("latest"))
	}
}

// IsOutdated returns true if current version is outdated
func IsOutdated(current, latest string) bool {
	if strings.HasSuffix(current, "-dev") {
		return IsDevReleaseOutdated(current, latest)
	}
	currentVer, _ := semver.NewVersion(current)
	latestVer, _ := semver.NewVersion(latest)
	if currentVer == nil || latestVer == nil {
		// fallback to naive comparison
		return current != latest
	}
	return latestVer.GreaterThan(currentVer)
}

// IsDevReleaseOutdated returns true if installed tool (dev version) is outdated
// ex: if installed tools is v2.9.1-dev and latest release is v2.9.1 then it is outdated
// since v2.9.1-dev is released and merged into main/master branch
func IsDevReleaseOutdated(current string, latest string) bool {
	// remove -dev suffix
	current = strings.TrimSuffix(current, "-dev")
	currentVer, _ := semver.NewVersion(current)
	latestVer, _ := semver.NewVersion(latest)
	if currentVer == nil || latestVer == nil {
		if current == latest {
			return true
		} else {
			// can't compare, so consider it latest
			return false
		}
	}
	if latestVer.GreaterThan(currentVer) || latestVer.Equal(currentVer) {
		return true
	}
	return false
}
