package metrics

import (
	"regexp"
	"strings"

	"github.com/mattermost/platform/model"
)

// rawSubstitutions holds the raw strings that will be compiled at run time
var rawSubstitutions = map[string]string{
	//Channel Routes
	"/channels/cid/":       "/channels/[a-z0-9]{26}/",       //Channel ID
	"/channels/name/cname": "/channels/name/[A-Za-z0-9_-]+", //Get Channel By Name

	//Team Routes
	"/teams/tid/": "/teams/[a-z0-9]{26}/", //Team ID
}

// MetricNamesByCommonPath allows lookup of metric name by path
var MetricNamesByCommonPath = map[string]string{
	"/general/ping":                        MetricAPIGeneralPing,
	"/users/login":                         MetricAPIGeneralLogin,
	"/teams/tid/channels/name/cname":       MetricAPIChannelGetByName,
	"/teams/tid/channels/cid/join":         MetricAPIChannelJoin,
	"/teams/tid/channels/cid/posts/create": MetricAPIPostCreate,
}

// Subtitutions holds the compiled regex
var Subtitutions = map[string]*regexp.Regexp{}

func init() {
	for k, v := range rawSubstitutions {
		Subtitutions[k] = regexp.MustCompile(v)
	}
}

// CollatePaths organize and clean up path names based off known formats in Subtitutions
func CollatePaths(path string) string {
	result := strings.TrimPrefix(path, model.API_URL_SUFFIX)
	for sub, reg := range Subtitutions {
		result = reg.ReplaceAllString(result, sub)
	}
	return result
}

// LookupMetricNameByPath will search for metricname by path
func LookupMetricNameByPath(path string) (string, bool) {
	cp := CollatePaths(path)
	name, ok := MetricNamesByCommonPath[cp]
	return name, ok
}
