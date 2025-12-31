package analyze

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/dennislee928/github-recap-2025/internal/model"
)

func BuildRecap(user string, year int, cc *model.ContributionsCollection, prs []model.PRItem, issuesOpened, issuesClosed []model.IssueItem, langs model.LanguageBytes, growth *model.GrowthMetrics) *model.Recap {
	var recap model.Recap
	recap.Meta.User = user
	recap.Meta.Year = year
	recap.Meta.GeneratedAt = time.Now().UTC()

	// Totals
	recap.Totals.Commits = cc.TotalCommits
	recap.Totals.PullRequests = cc.TotalPRs
	recap.Totals.Issues = cc.TotalIssues
	recap.Totals.Reviews = cc.TotalReviews
	recap.Totals.Overall = recap.Totals.Commits + recap.Totals.PullRequests + recap.Totals.Issues + recap.Totals.Reviews

	// Calendar + streak + most productive day/week
	recap.Calendar.Days = cc.Calendar
	recap.Calendar.LongestStreak = longestStreak(cc.Calendar)
	recap.Calendar.MostProductiveDay = mostProductiveDay(cc.Calendar)
	week, cnt := mostProductiveISOWeek(cc.Calendar)
	recap.Calendar.MostProductiveISOWeek.ISOWeek = week
	recap.Calendar.MostProductiveISOWeek.Count = cnt

	// Top repos by activity
	repoMap := map[string]*model.RepoContrib{}
	mergeLite := func(m map[string]model.RepoContribLite, field string) {
		for repo, v := range m {
			r, ok := repoMap[repo]
			if !ok {
				r = &model.RepoContrib{Repo: repo, IsPrivate: v.IsPrivate}
				repoMap[repo] = r
			}
			switch field {
			case "commit":
				r.CommitCount = v.Count
			case "pr":
				r.PRCount = v.Count
			case "issue":
				r.IssueCount = v.Count
			case "review":
				r.ReviewCount = v.Count
			}
		}
	}
	mergeLite(cc.ByRepoCommits, "commit")
	mergeLite(cc.ByRepoPRs, "pr")
	mergeLite(cc.ByRepoIssues, "issue")
	mergeLite(cc.ByRepoReviews, "review")

	top := make([]model.RepoContrib, 0, len(repoMap))
	for _, r := range repoMap {
		r.TotalActivity = r.CommitCount + r.PRCount + r.IssueCount + r.ReviewCount
		top = append(top, *r)
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].TotalActivity == top[j].TotalActivity {
			return top[i].Repo < top[j].Repo
		}
		return top[i].TotalActivity > top[j].TotalActivity
	})
	if len(top) > 12 {
		top = top[:12]
	}
	recap.TopRepos = top

	// PR stats
	recap.PRStats.Opened = len(prs)
	merged := 0
	var totalMergeHours float64
	var mergedCount int
	var biggest *model.PRItem
	for i := range prs {
		p := prs[i]
		if p.Merged && p.MergedAt != nil {
			merged++
			mergedCount++
			totalMergeHours += p.MergedAt.Sub(p.CreatedAt).Hours()
		}
		// biggest PR by churn
		score := p.Additions + p.Deletions
		if biggest == nil || score > (biggest.Additions+biggest.Deletions) {
			cp := p
			biggest = &cp
		}
	}
	recap.PRStats.Merged = merged
	if recap.PRStats.Opened > 0 {
		recap.PRStats.MergeRate = float64(merged) / float64(recap.PRStats.Opened)
	}
	if mergedCount > 0 {
		recap.PRStats.AvgTimeToMergeHours = totalMergeHours / float64(mergedCount)
	}
	recap.PRStats.BiggestPR = biggest
	recap.PRStats.TimeOfDayHistogram = hourHistogramPR(prs)

	// Issue stats
	recap.IssueStats.Opened = len(issuesOpened)
	recap.IssueStats.Closed = len(issuesClosed)
	recap.IssueStats.TimeOfDayHistogram = hourHistogramIssues(issuesOpened)

	// Reviews
	recap.Reviews.Total = cc.TotalReviews
	recap.Reviews.ByRepo = topReviewsByRepo(cc.ByRepoReviews)

	// Languages
	recap.Languages.WeightedBytes = langs
	recap.Languages.Note = "Language bytes are aggregated from the current repo language breakdown (not time-series). Weighted by bytes across repos you contributed to in 2025."
	recap.Languages.Top = topLanguages(langs, 10)

	recap.Growth = growth

	return &recap
}

func longestStreak(days []model.ContributionDay) int {
	cur := 0
	best := 0
	for _, d := range days {
		if d.Count > 0 {
			cur++
			if cur > best { best = cur }
		} else {
			cur = 0
		}
	}
	return best
}

func mostProductiveDay(days []model.ContributionDay) model.ContributionDay {
	best := model.ContributionDay{Date:"", Count:-1}
	for _, d := range days {
		if d.Count > best.Count {
			best = d
		}
	}
	if best.Count < 0 { best.Count = 0 }
	return best
}

func mostProductiveISOWeek(days []model.ContributionDay) (string, int) {
	weekSum := map[string]int{}
	for _, d := range days {
		t, err := time.Parse("2006-01-02", d.Date)
		if err != nil { continue }
		y, w := t.ISOWeek()
		key := fmtISOWeek(y, w)
		weekSum[key] += d.Count
	}
	bestKey := ""
	best := -1
	for k, v := range weekSum {
		if v > best {
			best = v
			bestKey = k
		}
	}
	if best < 0 { best = 0 }
	return bestKey, best
}

func fmtISOWeek(y, w int) string {
	return fmt.Sprintf("%04d-W%02d", y, w)
}

func hourHistogramPR(prs []model.PRItem) map[string]int {
	h := make(map[string]int, 24)
	for i := 0; i < 24; i++ {
		h[fmt2(i)] = 0
	}
	for _, p := range prs {
		h[fmt2(p.CreatedAt.UTC().Hour())]++
	}
	return h
}

func hourHistogramIssues(issues []model.IssueItem) map[string]int {
	h := make(map[string]int, 24)
	for i := 0; i < 24; i++ {
		h[fmt2(i)] = 0
	}
	for _, it := range issues {
		h[fmt2(it.CreatedAt.UTC().Hour())]++
	}
	return h
}

func fmt2(n int) string {
	if n < 10 { return "0" + strconv.Itoa(n) }
	return strconv.Itoa(n)
}

func topReviewsByRepo(m map[string]model.RepoContribLite) []model.RepoContribLite {
	out := make([]model.RepoContribLite, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Repo < out[j].Repo
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > 8 { out = out[:8] }
	return out
}

func topLanguages(langs model.LanguageBytes, n int) []struct{
	Language string `json:"language"`
	Bytes int64 `json:"bytes"`
	Share float64 `json:"share"`
} {
	type kv struct{ k string; v int64 }
	var items []kv
	var total int64
	for k, v := range langs {
		items = append(items, kv{k, v})
		total += v
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].v == items[j].v { return items[i].k < items[j].k }
		return items[i].v > items[j].v
	})
	if n > len(items) { n = len(items) }
	out := make([]struct{
		Language string `json:"language"`
		Bytes int64 `json:"bytes"`
		Share float64 `json:"share"`
	}, 0, n)
	for i := 0; i < n; i++ {
		share := 0.0
		if total > 0 {
			share = float64(items[i].v) / float64(total)
		}
		out = append(out, struct{
			Language string `json:"language"`
			Bytes int64 `json:"bytes"`
			Share float64 `json:"share"`
		}{Language: items[i].k, Bytes: items[i].v, Share: share})
	}
	return out
}
