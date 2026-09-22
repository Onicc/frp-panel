package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Onicc/frp-panel/conf"
)

const officialAPI = "https://api.github.com/repos/Onicc/frp-panel"

var commitPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

type Asset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type Release struct {
	ID              int64     `json:"id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	PublishedAt     time.Time `json:"published_at"`
	HTMLURL         string    `json:"html_url"`
	Assets          []Asset   `json:"assets"`
	Commit          string    `json:"-"`
}

type Snapshot struct {
	Channel        string    `json:"channel"`
	CurrentVersion string    `json:"currentVersion"`
	CurrentCommit  string    `json:"currentCommit"`
	LatestVersion  string    `json:"latestVersion,omitempty"`
	LatestCommit   string    `json:"latestCommit,omitempty"`
	ReleaseURL     string    `json:"releaseUrl,omitempty"`
	PublishedAt    time.Time `json:"publishedAt,omitempty"`
	Available      *bool     `json:"available"`
	CheckedAt      time.Time `json:"checkedAt"`
	Error          string    `json:"error,omitempty"`
}

type Checker struct {
	BaseURL string
	Client  *http.Client
	mu      sync.Mutex
	cache   map[string]cachedRelease
	compare map[string]cachedComparison
}

type cachedComparison struct {
	status string
	at     time.Time
}

type cachedRelease struct {
	release Release
	at      time.Time
}

func NewChecker() *Checker {
	return &Checker{BaseURL: officialAPI, Client: &http.Client{Timeout: 12 * time.Second}, cache: make(map[string]cachedRelease), compare: make(map[string]cachedComparison)}
}

var Default = NewChecker()

func Channel(version string) string {
	version = strings.TrimSpace(version)
	if version == "main" || version == "edge" || strings.HasPrefix(version, "edge-") {
		return "edge"
	}
	if strings.HasPrefix(version, "v") && validSemver(version) {
		return "stable"
	}
	return "unknown"
}

func (c *Checker) Check(ctx context.Context, current conf.VersionInfo, force bool) (Snapshot, Release) {
	s := Snapshot{Channel: Channel(current.GitVersion), CurrentVersion: current.GitVersion, CurrentCommit: current.GitCommit, CheckedAt: time.Now().UTC()}
	if s.Channel == "unknown" {
		s.Error = "development or unknown build; update channel cannot be inferred"
		return s, Release{}
	}
	rel, err := c.Latest(ctx, s.Channel, force)
	if err != nil {
		s.Error = err.Error()
		return s, Release{}
	}
	s.LatestVersion, s.LatestCommit, s.ReleaseURL, s.PublishedAt = rel.TagName, rel.Commit, rel.HTMLURL, rel.PublishedAt
	if s.Channel == "stable" {
		if !validSemver(current.GitVersion) || !validSemver(rel.TagName) {
			s.Error = "release version is not valid semver"
			return s, rel
		}
		available := compareSemver(current.GitVersion, rel.TagName) < 0
		s.Available = &available
		return s, rel
	}
	if !commitPattern.MatchString(current.GitCommit) || !commitPattern.MatchString(rel.Commit) {
		s.Error = "current or release commit is unavailable"
		return s, rel
	}
	if strings.EqualFold(current.GitCommit, rel.Commit) {
		available := false
		s.Available = &available
		return s, rel
	}
	comparison, err := c.compareCommits(ctx, current.GitCommit, rel.Commit, force)
	if err != nil {
		s.Error = "cannot determine commit ancestry: " + err.Error()
		return s, rel
	}
	if comparison != "ahead" && comparison != "identical" {
		s.Error = "release is not ahead of this build"
		return s, rel
	}
	available := comparison == "ahead"
	s.Available = &available
	return s, rel
}

func (c *Checker) compareCommits(ctx context.Context, current, latest string, force bool) (string, error) {
	key := strings.ToLower(current + "..." + latest)
	c.mu.Lock()
	if entry, ok := c.compare[key]; ok && !force && time.Since(entry.at) < 15*time.Minute {
		c.mu.Unlock()
		return entry.status, nil
	}
	c.mu.Unlock()
	var comparison struct {
		Status string `json:"status"`
	}
	compareURL := fmt.Sprintf("%s/compare/%s...%s", c.base(), url.PathEscape(current), url.PathEscape(latest))
	if err := c.getJSON(ctx, compareURL, &comparison); err != nil {
		return "", err
	}
	c.mu.Lock()
	if c.compare == nil {
		c.compare = make(map[string]cachedComparison)
	}
	c.compare[key] = cachedComparison{status: comparison.Status, at: time.Now()}
	c.mu.Unlock()
	return comparison.Status, nil
}

func (c *Checker) Latest(ctx context.Context, channel string, force bool) (Release, error) {
	if channel != "edge" && channel != "stable" {
		return Release{}, errors.New("unsupported update channel")
	}
	c.mu.Lock()
	if entry, ok := c.cache[channel]; ok && !force && time.Since(entry.at) < 15*time.Minute {
		c.mu.Unlock()
		return entry.release, nil
	}
	c.mu.Unlock()
	path := "/releases/latest"
	if channel == "edge" {
		path = "/releases/tags/edge"
	}
	var rel Release
	if err := c.getJSON(ctx, c.base()+path, &rel); err != nil {
		return Release{}, err
	}
	if rel.ID == 0 || rel.TagName == "" || len(rel.Assets) == 0 {
		return Release{}, errors.New("GitHub release has no usable assets")
	}
	commit := strings.TrimSpace(rel.TargetCommitish)
	if !commitPattern.MatchString(commit) {
		resolved, err := c.resolveTag(ctx, rel.TagName)
		if err != nil {
			return Release{}, err
		}
		commit = resolved
	}
	rel.Commit = strings.ToLower(commit)
	c.mu.Lock()
	if c.cache == nil {
		c.cache = make(map[string]cachedRelease)
	}
	c.cache[channel] = cachedRelease{release: rel, at: time.Now()}
	c.mu.Unlock()
	return rel, nil
}

func (c *Checker) resolveTag(ctx context.Context, tag string) (string, error) {
	var ref struct {
		Object struct{ SHA, Type string } `json:"object"`
	}
	if err := c.getJSON(ctx, c.base()+"/git/ref/tags/"+url.PathEscape(tag), &ref); err != nil {
		return "", err
	}
	for i := 0; i < 3 && ref.Object.Type == "tag"; i++ {
		var annotated struct {
			Object struct{ SHA, Type string } `json:"object"`
		}
		if err := c.getJSON(ctx, c.base()+"/git/tags/"+url.PathEscape(ref.Object.SHA), &annotated); err != nil {
			return "", err
		}
		ref.Object = annotated.Object
	}
	if ref.Object.Type != "commit" || !commitPattern.MatchString(ref.Object.SHA) {
		return "", errors.New("release tag does not resolve to a commit")
	}
	return ref.Object.SHA, nil
}

func (c *Checker) base() string {
	if c.BaseURL == "" {
		return officialAPI
	}
	return strings.TrimRight(c.BaseURL, "/")
}

func (c *Checker) getJSON(ctx context.Context, endpoint string, out any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "frp-panel-updater")
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub release API returned HTTP %d", response.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(out)
}

func validSemver(value string) bool {
	parts := strings.Split(strings.TrimPrefix(value, "v"), ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func compareSemver(a, b string) int {
	aa := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bb := strings.Split(strings.TrimPrefix(b, "v"), ".")
	for i := 0; i < 3; i++ {
		av, _ := strconv.ParseUint(aa[i], 10, 64)
		bv, _ := strconv.ParseUint(bb[i], 10, 64)
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}
