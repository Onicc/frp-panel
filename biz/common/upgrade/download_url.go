package upgrade

import (
	"fmt"
	"regexp"
	"strings"
)

var stableReleaseTag = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

func buildDownloadURL(opt Options) (string, error) {
	if u := strings.TrimSpace(opt.DownloadURL); len(u) > 0 {
		return u, nil
	}
	version := strings.TrimSpace(opt.Version)
	if len(version) == 0 {
		version = "latest"
	}
	if version != "latest" && !stableReleaseTag.MatchString(version) {
		return "", fmt.Errorf("version must be latest or a stable vX.X.X release")
	}

	asset, err := detectAssetName()
	if err != nil {
		return "", err
	}
	var url string
	if version == "latest" {
		url = fmt.Sprintf("https://github.com/Onicc/frp-panel/releases/latest/download/%s", asset)
	} else {
		url = fmt.Sprintf("https://github.com/Onicc/frp-panel/releases/download/%s/%s", version, asset)
	}
	if opt.UseGithubProxy && len(strings.TrimSpace(opt.GithubProxy)) > 0 {
		url = fmt.Sprintf("%s/%s", strings.TrimRight(strings.TrimSpace(opt.GithubProxy), "/"), url)
	}
	return url, nil
}

func buildChecksumURL(downloadURL string) string {
	index := strings.LastIndex(downloadURL, "/")
	if index < 0 {
		return ""
	}
	return downloadURL[:index+1] + "checksums.txt"
}
