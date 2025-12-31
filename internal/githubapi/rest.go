\
package githubapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dennislee928/github-recap-2025/internal/model"
	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

func (c *Client) rest() *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.cfg.Token})
	tc := oauth2.NewClient(context.Background(), ts)
	// allow overriding API base? go-github supports Enterprise URLs, but keep default for now.
	return github.NewClient(tc)
}

func (c *Client) FetchLanguagesForContributedRepos(cc *model.ContributionsCollection) (model.LanguageBytes, error) {
	repos := map[string]bool{}
	for k := range cc.ByRepoCommits { repos[k] = true }
	for k := range cc.ByRepoPRs { repos[k] = true }
	for k := range cc.ByRepoIssues { repos[k] = true }
	for k := range cc.ByRepoReviews { repos[k] = true }

	client := c.rest()

	type job struct{ owner, repo string }
	jobs := make([]job, 0, len(repos))
	for full := range repos {
		parts := strings.Split(full, "/")
		if len(parts) != 2 { continue }
		jobs = append(jobs, job{owner: parts[0], repo: parts[1]})
	}

	agg := model.LanguageBytes{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// modest concurrency to avoid rate-limit bursts
	sem := make(chan struct{}, 6)

	var firstErr error
	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func(){ <-sem }()

			langs, _, err := client.Repositories.ListLanguages(context.Background(), j.owner, j.repo)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("list languages %s/%s: %w", j.owner, j.repo, err)
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			for lang, bytes := range langs {
				agg[lang] += int64(bytes)
			}
			mu.Unlock()
		}(j)
	}
	wg.Wait()
	return agg, firstErr
}

func (c *Client) CalcStarsForksGainedOwnedRepos(login string, from, to time.Time) (*model.GrowthMetrics, error) {
	client := c.rest()

	// list owned repos
	var allRepos []*github.Repository
	opt := &github.RepositoryListByUserOptions{
		Type: "owner",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := client.Repositories.ListByUser(context.Background(), login, opt)
		if err != nil { return nil, err }
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 { break }
		opt.Page = resp.NextPage
	}

	metrics := &model.GrowthMetrics{
		Year: from.Year(),
		Repos: make([]model.GrowthRepo, 0, len(allRepos)),
		Note: "Stars gained derived from stargazer timestamps when available; forks gained derived from fork creation timestamps. Private repo stars/forks are usually not meaningful. Use --skip-growth to disable.",
	}

	type result struct{
		gr model.GrowthRepo
		err error
	}
	sem := make(chan struct{}, 4)
	outCh := make(chan result, len(allRepos))
	var wg sync.WaitGroup

	for _, r := range allRepos {
		r := r
		if r.GetName() == "" || r.GetOwner() == nil { continue }
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func(){ <-sem }()
			owner := r.GetOwner().GetLogin()
			repo := r.GetName()
			full := fmt.Sprintf("%s/%s", owner, repo)

			gr := model.GrowthRepo{
				Repo: full,
				StarsNow: r.GetStargazersCount(),
				ForksNow: r.GetForksCount(),
				IsPrivate: r.GetPrivate(),
			}

			// stars gained in year: /stargazers with starred_at
			starsGained, err := c.countStarsInRange(client, owner, repo, from, to)
			if err != nil {
				// keep going, but report error
				outCh <- result{gr: gr, err: fmt.Errorf("stars %s: %w", full, err)}
				return
			}
			gr.StarsGainedInYear = starsGained

			// forks gained in year: /forks list includes CreatedAt
			forksGained, err := c.countForksInRange(client, owner, repo, from, to)
			if err != nil {
				outCh <- result{gr: gr, err: fmt.Errorf("forks %s: %w", full, err)}
				return
			}
			gr.ForksGainedInYear = forksGained

			outCh <- result{gr: gr, err: nil}
		}()
	}

	wg.Wait()
	close(outCh)

	var firstErr error
	for r := range outCh {
		metrics.Repos = append(metrics.Repos, r.gr)
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		metrics.TotalStarsGained += r.gr.StarsGainedInYear
		metrics.TotalForksGained += r.gr.ForksGainedInYear
		metrics.TotalStarsNow += r.gr.StarsNow
		metrics.TotalForksNow += r.gr.ForksNow
	}

	// sort repos by stars gained desc
	sort.Slice(metrics.Repos, func(i, j int) bool {
		if metrics.Repos[i].StarsGainedInYear == metrics.Repos[j].StarsGainedInYear {
			return metrics.Repos[i].StarsNow > metrics.Repos[j].StarsNow
		}
		return metrics.Repos[i].StarsGainedInYear > metrics.Repos[j].StarsGainedInYear
	})

	return metrics, firstErr
}

func (c *Client) countStarsInRange(client *github.Client, owner, repo string, from, to time.Time) (int, error) {
	// GitHub returns starred_at only with a special accept header.
	// go-github supports this via a custom request; easiest is to call Repositories.ListStargazers with StarListOptions.
	// In go-github, StarListOptions returns []*github.Stargazer with StarredAt.
	opt := &github.ListOptions{PerPage: 100}
	count := 0
	for {
		sg, resp, err := client.Activity.ListStargazers(context.Background(), owner, repo, opt)
		if err != nil {
			return count, err
		}
		for _, s := range sg {
			if s.StarredAt == nil { continue }
			t := s.StarredAt.Time
			if !t.Before(from) && !t.After(to) {
				count++
			}
		}
		if resp.NextPage == 0 { break }
		opt.Page = resp.NextPage
	}
	return count, nil
}

func (c *Client) countForksInRange(client *github.Client, owner, repo string, from, to time.Time) (int, error) {
	opt := &github.RepositoryListForksOptions{
		Sort: "newest",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	count := 0
	for {
		forks, resp, err := client.Repositories.ListForks(context.Background(), owner, repo, opt)
		if err != nil {
			return count, err
		}
		for _, f := range forks {
			t := f.GetCreatedAt().Time
			if !t.Before(from) && !t.After(to) {
				count++
			}
			// Since sorted newest, if we are past range (older than from), we can early-stop
			if t.Before(from) {
				// We can stop scanning this repo
				return count, nil
			}
		}
		if resp.NextPage == 0 { break }
		opt.Page = resp.NextPage
	}
	return count, nil
}

// NOTE: go-github's Activity.ListStargazers uses the correct Accept header internally for starred_at.
func init() {
	_ = http.MethodGet
}
