package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v68/github"
	"github.com/gorilla/mux"
)

// ReleaseInfo represents the release version and its asset size in bytes.
type ReleaseInfo struct {
	Version string
	Size    int64
}

// Deltas represents the response from querying a subset of releases and comparing
// the difference in size between releases
type Deltas struct {
	Current  string
	Previous string
	Delta    int64
}

// fetchReleases fetches all releases from the GitHub repository and filters them by valid semantic version tags.
// It returns a slice of ReleaseInfo with the relevant release information.
func fetchReleases(ctx context.Context, client *github.Client, owner, repo string) ([]ReleaseInfo, error) {
	releases, _, err := client.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}

	semverRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	var releaseInfo []ReleaseInfo

	for _, release := range releases {
		if release.TagName == nil {
			continue
		}
		tag := *release.TagName

		// Filter out invalid tags.
		if !semverRegex.MatchString(tag) {
			continue
		}

		var assetSize int64
		for _, asset := range release.Assets {
			// TODO_TECHDEBT: This conditional constraints to fetch only newer versions of Airflow as is.
			// To make it work with different releases, the user should supply the name prefix it wants to
			// retrieve, otherwise results may be not as expected.
			if asset.Name != nil && strings.Contains(*asset.Name, fmt.Sprintf("apache_airflow-%s.tar.gz", tag)) && asset.Size != nil {
				assetSize = int64(*asset.Size)
				break
			}
		}
		if assetSize > 0 {
			releaseInfo = append(releaseInfo, ReleaseInfo{Version: tag, Size: assetSize})
		}
	}

	return releaseInfo, nil
}

// queryReleaseDeltas prints the size deltas between releases for the specified initial and final versions.
// TODO_TECHDEBT: Supplied values are not validated in any form. Non-existent release values will return
// non-expected results both ath initial and final versions.
// In the future this might implement some library that validates semantic versioning.
func queryReleaseDeltas(releases []ReleaseInfo, initialVersion, finalVersion string) ([]Deltas, error) {
	var begin, end int
	var deltas []Deltas
	for index, release := range releases {
		if release.Version == finalVersion {
			end = index
		}

		if release.Version == initialVersion {
			begin = index
		}
	}

	// Given the order in which the API response from GitHub comes, end or finalVersion
	// has to be always in a smaller index than begin or initialVersion
	if begin < end {
		return []Deltas{}, fmt.Errorf("invalid range: initial version '%s' must be older than final version '%s'", initialVersion, finalVersion)
	}

	for i := end + 1; i <= begin; i++ {
		curr := releases[i-1]
		prev := releases[i]
		delta := curr.Size - prev.Size

		deltas = append(deltas, Deltas{
			Current:  curr.Version,
			Previous: prev.Version,
			Delta:    delta,
		})
	}

	return deltas, nil
}

func handleDeltas(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owner := vars["owner"]
	repo := vars["repo"]

	initialVersion := r.URL.Query().Get("initial")
	finalVersion := r.URL.Query().Get("final")

	if initialVersion == "" || finalVersion == "" {
		http.Error(w, "Missing query parameters 'initial' or 'final'", http.StatusBadRequest)
		return
	}

	// Fetch releases from the GitHub API.
	ctx := context.Background()
	client := github.NewClient(nil)
	releases, err := fetchReleases(ctx, client, owner, repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch releases: %v", err), http.StatusInternalServerError)
		return
	}

	// Query versions and calculate deltas.
	deltas, err := queryReleaseDeltas(releases, initialVersion, finalVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Write response to handler
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deltas); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}
