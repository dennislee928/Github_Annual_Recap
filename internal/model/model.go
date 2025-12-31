package model

import "time"

type ContributionDay struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"`
}

type RepoContrib struct {
	Repo           string `json:"repo"` // owner/name
	CommitCount    int    `json:"commit_count"`
	PRCount        int    `json:"pr_count"`
	IssueCount     int    `json:"issue_count"`
	ReviewCount    int    `json:"review_count"`
	IsPrivate      bool   `json:"is_private"`
	TotalActivity  int    `json:"total_activity"`
}

type PRItem struct {
	Repo       string    `json:"repo"`
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"created_at"`
	Merged     bool      `json:"merged"`
	MergedAt   *time.Time `json:"merged_at,omitempty"`
	Additions  int       `json:"additions"`
	Deletions  int       `json:"deletions"`
}

type IssueItem struct {
	Repo      string     `json:"repo"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	URL       string     `json:"url"`
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
}

type LanguageBytes map[string]int64

type GrowthRepo struct {
	Repo              string `json:"repo"`
	StarsGainedInYear int    `json:"stars_gained_in_year"`
	ForksGainedInYear int    `json:"forks_gained_in_year"`
	StarsNow          int    `json:"stars_now"`
	ForksNow          int    `json:"forks_now"`
	IsPrivate         bool   `json:"is_private"`
}

type GrowthMetrics struct {
	Year int `json:"year"`
	Repos []GrowthRepo `json:"repos"`
	TotalStarsGained int `json:"total_stars_gained"`
	TotalForksGained int `json:"total_forks_gained"`
	TotalStarsNow int `json:"total_stars_now"`
	TotalForksNow int `json:"total_forks_now"`
	Note string `json:"note"`
}

type ContributionsCollection struct {
	TotalCommits int `json:"total_commits"`
	TotalPRs int `json:"total_prs"`
	TotalIssues int `json:"total_issues"`
	TotalReviews int `json:"total_reviews"`

	Calendar []ContributionDay `json:"calendar_days"` // flattened
	ByRepoCommits map[string]RepoContribLite `json:"by_repo_commits"`
	ByRepoPRs map[string]RepoContribLite `json:"by_repo_prs"`
	ByRepoIssues map[string]RepoContribLite `json:"by_repo_issues"`
	ByRepoReviews map[string]RepoContribLite `json:"by_repo_reviews"`
}

type RepoContribLite struct {
	Repo string `json:"repo"`
	Count int `json:"count"`
	IsPrivate bool `json:"is_private"`
}

type Recap struct {
	Meta struct {
		User string `json:"user"`
		Year int `json:"year"`
		GeneratedAt time.Time `json:"generated_at"`
	} `json:"meta"`

	Totals struct {
		Commits int `json:"commits"`
		PullRequests int `json:"pull_requests"`
		Issues int `json:"issues"`
		Reviews int `json:"reviews"`
		Overall int `json:"overall"`
	} `json:"totals"`

	Calendar struct {
		Days []ContributionDay `json:"days"`
		LongestStreak int `json:"longest_streak"`
		MostProductiveDay ContributionDay `json:"most_productive_day"`
		MostProductiveISOWeek struct{
			ISOWeek string `json:"iso_week"`
			Count int `json:"count"`
		} `json:"most_productive_iso_week"`
	} `json:"calendar"`

	TopRepos []RepoContrib `json:"top_repos"`

	PRStats struct {
		Opened int `json:"opened"`
		Merged int `json:"merged"`
		MergeRate float64 `json:"merge_rate"`
		AvgTimeToMergeHours float64 `json:"avg_time_to_merge_hours"`
		BiggestPR *PRItem `json:"biggest_pr,omitempty"`
		TimeOfDayHistogram map[string]int `json:"time_of_day_histogram"` // hour "00".."23"
	} `json:"pr_stats"`

	IssueStats struct {
		Opened int `json:"opened"` // issues authored, created in year
		Closed int `json:"closed"` // issues authored, closed in year
		TimeOfDayHistogram map[string]int `json:"time_of_day_histogram"`
	} `json:"issue_stats"`

	Reviews struct {
		Total int `json:"total"`
		ByRepo []RepoContribLite `json:"by_repo"`
	} `json:"reviews"`

	Languages struct {
		WeightedBytes LanguageBytes `json:"weighted_bytes"`
		Top []struct{
			Language string `json:"language"`
			Bytes int64 `json:"bytes"`
			Share float64 `json:"share"`
		} `json:"top"`
		Note string `json:"note"`
	} `json:"languages"`

	Growth *GrowthMetrics `json:"growth,omitempty"`
}
