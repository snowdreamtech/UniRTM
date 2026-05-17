package env

import (
	"crypto/rand"
	"os"
	"strings"
)

// Get returns the value of the environment variable with the given key,
// searching with prefixes in order: UNIRTM_, MISE_, and then the raw key.
func Get(key string) string {
	// 1. UNIRTM_ prefix
	if v := os.Getenv("UNIRTM_" + key); v != "" {
		return v
	}
	// 2. MISE_ prefix
	if v := os.Getenv("MISE_" + key); v != "" {
		return v
	}
	// 3. Raw key (Native)
	return os.Getenv(key)
}

// ProxyBypassDomains contains a list of common domestic mirror domains
// that should bypass any configured HTTP proxies to prevent connection drops.
var ProxyBypassDomains = []string{
	"aliyun.com",
	"npmmirror.com",
	"tencent.com",
	"huaweicloud.com",
	"163.com",
	"ustc.edu.cn",
	"tsinghua.edu.cn",
	"sjtu.edu.cn",
	"bfsu.edu.cn",
	"lzu.edu.cn",
	"nju.edu.cn",
	"cqu.edu.cn",
	"hit.edu.cn",
	"zju.edu.cn",
	"douban.com",
	"rsproxy.cn",
	"r.cnpmjs.org",
	"goproxy.cn",
	"goproxy.io",
	"gems.ruby-china.com",
	"sn0wdr1am.com",
}

// ShouldBypassProxy returns true if the given host should bypass the proxy.
func ShouldBypassProxy(host string) bool {
	if strings.Contains(host, "mirror") || strings.HasSuffix(host, ".cn") {
		return true
	}
	for _, domain := range ProxyBypassDomains {
		if strings.Contains(host, domain) {
			return true
		}
	}
	return false
}

var (
	//ProjectName Project Name
	ProjectName string = "unirtm"

	//Author Author
	Author string = "Snowdream Tech <snowdreamtech@qq.com>"

	//BuildTime Build Time
	BuildTime string = "N/A"

	//GitTag Git Tag
	GitTag string = "N/A"

	//CommitHash Commit Hash
	CommitHash string = "N/A"

	//CommitHashFull Commit Hash
	CommitHashFull string = "N/A"

	//COPYRIGHT COPYRIGHT
	COPYRIGHT string = "Copyright (c) 2023-present SnowdreamTech Inc."

	//LICENSE LICENSE
	LICENSE string = "MIT <https://github.com/snowdreamtech/unirtm/blob/main/LICENSE>"

	//Config Config File Path
	Config string = "unirtm.toml"

	// Debug indicates whether the application should run in debug mode.
	Debug bool

	// Trace indicates whether the application should run in trace mode.
	Trace bool

	// Quiet indicates whether the application should run in quiet mode.
	Quiet bool

	// Cwd specifies the current working directory for the application.
	Cwd string

	// EnvName specifies the environment name for loading environment-specific configs.
	EnvName string

	// Jobs specifies the number of parallel jobs to run.
	Jobs int

	// Yes indicates whether to automatically answer yes to all confirmation prompts.
	Yes bool

	// Locked indicates whether to require lockfile URLs to be present during installation.
	Locked bool

	// Silent indicates whether to suppress all output and non-error messages.
	Silent bool
)

// RandomString returns a random string of the specified length.
func RandomString(n int) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
