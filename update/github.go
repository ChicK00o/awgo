// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"sort"

	aw "github.com/deanishe/awgo"
)

const ghBaseURL = "https://api.github.com/repos/"

var rxWorkflowFile = regexp.MustCompile(`\.alfred(\d+)?workflow$`)

// GitHub is a Workflow Option. It sets a Workflow Updater for the specified GitHub repo.
// Repo name should be of the form "username/repo", e.g. "deanishe/alfred-ssh".
func GitHub(repo string) aw.Option {
	return func(wf *aw.Workflow) aw.Option {
		u, _ := NewUpdater(
			&githubSource{Repo: repo, fetch: getURL},
			wf.Version(),
			filepath.Join(wf.CacheDir(), "_aw/update"),
		)
		return aw.Update(u)(wf)
	}
}

type githubSource struct {
	Repo  string
	dls   []Download
	fetch func(URL string) ([]byte, error)
}

// Downloads implements Source.
func (src *githubSource) Downloads() ([]Download, error) {
	if src.dls == nil {
		src.dls = []Download{}
		// rels := []*Release{}
		js, err := src.fetch(src.url())
		if err != nil {
			// log.Printf("error: fetch GitHub releases: %s", err)
			return nil, err
		}
		// log.Printf("%d bytes of JSON", len(js))
		if src.dls, err = parseGitHubReleases(js); err != nil {
			// log.Printf("error: parse GitHub releases: %s", err)
			return nil, err
		}
	}
	log.Printf("%d download(s) in repo %s", len(src.dls), src.Repo)
	return src.dls, nil
}

// url returns URL of releases list.
func (src *githubSource) url() string { return fmt.Sprintf("%s%s/releases", ghBaseURL, src.Repo) }

// ghRelease is the data model for GitHub releases JSON.
type ghRelease struct {
	Name       string     `json:"name"`
	Prerelease bool       `json:"prerelease"`
	Assets     []*ghAsset `json:"assets"`
	Tag        string     `json:"tag_name"`
}

// ghAsset is the data model for GitHub releases JSON.
type ghAsset struct {
	Name             string `json:"name"`
	URL              string `json:"browser_download_url"`
	MinAlfredVersion SemVer `json:"-"`
}

// parseGitHubReleases parses GitHub releases JSON.
func parseGitHubReleases(js []byte) ([]Download, error) {
	var (
		dls  = []Download{}
		rels = []*ghRelease{}
	)
	if err := json.Unmarshal(js, &rels); err != nil {
		return nil, err
	}
	for _, r := range rels {
		if len(r.Assets) == 0 {
			continue
		}
		v, err := NewSemVer(r.Tag)
		if err != nil {
			log.Printf("ignored release %s: not semantic: %v", r.Tag, err)
			continue
		}
		var all []Download
		for _, a := range r.Assets {
			m := rxWorkflowFile.FindStringSubmatch(a.Name)
			if len(m) != 2 {
				log.Printf("ignored release %s: no workflow files", r.Tag)
				continue
			}
			w := Download{
				URL:        a.URL,
				Filename:   a.Name,
				Version:    v,
				Prerelease: r.Prerelease,
			}
			all = append(all, w)
		}
		if err := validRelease(all); err != nil {
			log.Printf("ignored release %s: %v", r.Tag, err)
			continue
		}
		dls = append(dls, all...)
	}
	sort.Sort(sort.Reverse(byVersion(dls)))
	return dls, nil
}

// Reject releases that contain multiple files with the same extension.
func validRelease(dls []Download) error {
	if len(dls) == 0 {
		return errors.New("empty slice")
	}
	dupes := map[string]int{}
	for _, dl := range dls {
		x := filepath.Ext(dl.Filename)
		dupes[x] = dupes[x] + 1
	}
	for x, n := range dupes {
		if n > 1 {
			return fmt.Errorf("multiple files with extension %q", x)
		}
	}
	return nil
}
