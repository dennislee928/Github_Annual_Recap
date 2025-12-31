package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dennislee928/github-recap-2025/internal/analyze"
	"github.com/dennislee928/github-recap-2025/internal/config"
	"github.com/dennislee928/github-recap-2025/internal/githubapi"
	"github.com/dennislee928/github-recap-2025/internal/model"
)

func main() {
	var (
		user      = flag.String("user", "", "GitHub username (login)")
		year      = flag.Int("year", 2025, "Year for recap (e.g., 2025)")
		out       = flag.String("out", "./web/recap_2025.json", "Output JSON path")
		skipGrowth = flag.Bool("skip-growth", false, "Skip stars/forks gained calculation (rate-limit heavy)")
		maxSearch = flag.Int("max-search", 1000, "Max results to pull from GraphQL search queries")
	)
	flag.Parse()

	if *user == "" {
		log.Fatal("--user is required")
	}

	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	from := time.Date(*year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(*year, 12, 31, 23, 59, 59, 0, time.UTC)

	client := githubapi.New(cfg)

	log.Printf("Fetching contributionsCollection for %s %d...", *user, *year)
	cc, err := client.FetchContributionsCollection(*user, from, to)
	if err != nil {
		log.Fatalf("FetchContributionsCollection: %v", err)
	}

	log.Printf("Fetching PR details via GraphQL search (max=%d)...", *maxSearch)
	prs, err := client.SearchPullRequests(*user, from, to, *maxSearch)
	if err != nil {
		log.Fatalf("SearchPullRequests: %v", err)
	}

	log.Printf("Fetching Issue details via GraphQL search (opened/closed)...")
	issuesOpened, issuesClosed, err := client.SearchIssuesOpenedClosed(*user, from, to, *maxSearch)
	if err != nil {
		log.Fatalf("SearchIssuesOpenedClosed: %v", err)
	}

	log.Printf("Fetching repo language stats via REST (weighted bytes)...")
	langAgg, err := client.FetchLanguagesForContributedRepos(cc)
	if err != nil {
		log.Printf("WARN: FetchLanguagesForContributedRepos: %v", err)
	}

	var growth *model.GrowthMetrics
	if !*skipGrowth {
		log.Printf("Calculating stars/forks gained in %d for owned repos (may be slow)...", *year)
		growth, err = client.CalcStarsForksGainedOwnedRepos(*user, from, to)
		if err != nil {
			log.Printf("WARN: CalcStarsForksGainedOwnedRepos: %v", err)
		}
	}

	recap := analyze.BuildRecap(*user, *year, cc, prs, issuesOpened, issuesClosed, langAgg, growth)

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatalf("mkdir out dir: %v", err)
	}

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("create out: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(recap); err != nil {
		log.Fatalf("write json: %v", err)
	}

	fmt.Printf("OK: wrote %s\n", *out)
}
