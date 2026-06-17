package update

import (
	"strconv"
	"strings"
)

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tag
	}
	if !strings.HasPrefix(tag, "v") {
		return "v" + tag
	}
	return tag
}

func isReleaseVersion(v string) bool {
	_, ok := parseSemver(normalizeTag(v))
	return ok
}

// compareVersion 比较两个 semver 标签；a>b 返回 1，相等返回 0，a<b 返回 -1。
// 无法解析的版本视为低于任何合法 release 版本。
func compareVersion(a, b string) int {
	av, aok := parseSemver(normalizeTag(a))
	bv, bok := parseSemver(normalizeTag(b))
	if !aok && !bok {
		return 0
	}
	if !aok {
		return -1
	}
	if !bok {
		return 1
	}
	if av.major != bv.major {
		return cmpInt(av.major, bv.major)
	}
	if av.minor != bv.minor {
		return cmpInt(av.minor, bv.minor)
	}
	return cmpInt(av.patch, bv.patch)
}

func cmpInt(a, b int) int {
	switch {
	case a > b:
		return 1
	case a < b:
		return -1
	default:
		return 0
	}
}

type semverParts struct {
	major, minor, patch int
}

func parseSemver(tag string) (semverParts, bool) {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	parts := strings.Split(tag, ".")
	if len(parts) < 3 {
		return semverParts{}, false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semverParts{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semverParts{}, false
	}
	patchStr := parts[2]
	if i := strings.IndexAny(patchStr, "-+"); i >= 0 {
		patchStr = patchStr[:i]
	}
	patch, err := strconv.Atoi(patchStr)
	if err != nil {
		return semverParts{}, false
	}
	return semverParts{major: major, minor: minor, patch: patch}, true
}
