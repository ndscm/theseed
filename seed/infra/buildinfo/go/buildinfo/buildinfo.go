package buildinfo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

//go:embed stable-status.txt
var bazelStableStatus string

//go:embed volatile-status.txt
var bazelVolatileStatus string

type BuildInfo struct {
	// bazel status
	BuildHost string `json:"buildHost"`
	BuildUser string `json:"buildUser"`

	// git status
	BuildBranch    string `json:"buildBranch"`
	BuildBrief     string `json:"buildBrief"`
	BuildDirty     string `json:"buildDirty"`
	BuildGitCommit string `json:"buildGitCommit"`
	BuildTag       string `json:"buildTag"`

	// volatile status
	BuildTime time.Time `json:"buildTime"`
}

type SingletonBuildInfo struct {
	mutex     sync.RWMutex
	buildInfo *BuildInfo
}

var singletonBuildInfo = SingletonBuildInfo{}

func parseBazelStatus(statusText string) map[string]string {
	lines := strings.Split(statusText, "\n")
	status := map[string]string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			status[key] = value
		} else {
			key := parts[0]
			status[key] = ""
		}
	}
	return status
}

func GetBuildInfo() BuildInfo {
	if singletonBuildInfo.buildInfo == nil {
		singletonBuildInfo.mutex.Lock()
		defer singletonBuildInfo.mutex.Unlock()
		if singletonBuildInfo.buildInfo == nil {
			buildInfo := &BuildInfo{}

			stableStatusMap := parseBazelStatus(bazelStableStatus)
			buildInfo.BuildHost = stableStatusMap["BUILD_HOST"]
			buildInfo.BuildUser = stableStatusMap["BUILD_USER"]
			buildInfo.BuildBranch = stableStatusMap["STABLE_BUILD_BRANCH"]
			buildInfo.BuildBrief = stableStatusMap["STABLE_BUILD_BRIEF"]
			buildInfo.BuildDirty = stableStatusMap["STABLE_BUILD_DIRTY"]
			buildInfo.BuildGitCommit = stableStatusMap["STABLE_GIT_COMMIT"]
			buildInfo.BuildTag = stableStatusMap["STABLE_BUILD_TAG"]

			volatileStatusMap := parseBazelStatus(bazelVolatileStatus)
			buildTimestamp, ok := volatileStatusMap["BUILD_TIMESTAMP"]
			if ok {
				buildTimeInt, err := strconv.ParseInt(buildTimestamp, 10, 64)
				if err == nil {
					buildInfo.BuildTime = time.Unix(buildTimeInt, 0)
				}
			}

			singletonBuildInfo.buildInfo = buildInfo
		}
		return *singletonBuildInfo.buildInfo
	}
	singletonBuildInfo.mutex.RLock()
	defer singletonBuildInfo.mutex.RUnlock()
	return *singletonBuildInfo.buildInfo
}

var webappHeadInjectionTemplate = `
<script>
window.__SEED_BUILD_INFO__ =%s;
</script>
`

func GenerateWebappHeadInjection() (string, error) {
	buildInfo := GetBuildInfo()
	jsonBuildInfo, err := json.Marshal(buildInfo)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	headInjection := fmt.Sprintf(webappHeadInjectionTemplate, jsonBuildInfo)
	return headInjection, nil
}
