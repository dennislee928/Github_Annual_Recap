\
package githubapi

import (
	"context"
	"time"

	"github.com/dennislee928/github-recap-2025/internal/model"
)

func (c *Client) FetchContributionsCollection(login string, from, to time.Time) (*model.ContributionsCollection, error) {
	const q = `
query($login:String!, $from:DateTime!, $to:DateTime!) {
  user(login:$login) {
    contributionsCollection(from:$from, to:$to) {
      totalCommitContributions
      totalPullRequestContributions
      totalIssueContributions
      totalPullRequestReviewContributions
      contributionCalendar {
        weeks {
          contributionDays {
            date
            contributionCount
          }
        }
      }
      commitContributionsByRepository(maxRepositories: 100) {
        repository { nameWithOwner isPrivate }
        contributions { totalCount }
      }
      pullRequestContributionsByRepository(maxRepositories: 100) {
        repository { nameWithOwner isPrivate }
        contributions { totalCount }
      }
      issueContributionsByRepository(maxRepositories: 100) {
        repository { nameWithOwner isPrivate }
        contributions { totalCount }
      }
      pullRequestReviewContributionsByRepository(maxRepositories: 100) {
        repository { nameWithOwner isPrivate }
        contributions { totalCount }
      }
    }
  }
}`

	type resp struct{
		User struct{
			ContributionsCollection struct{
				TotalCommitContributions int `json:"totalCommitContributions"`
				TotalPullRequestContributions int `json:"totalPullRequestContributions"`
				TotalIssueContributions int `json:"totalIssueContributions"`
				TotalPullRequestReviewContributions int `json:"totalPullRequestReviewContributions"`
				ContributionCalendar struct{
					Weeks []struct{
						ContributionDays []struct{
							Date string `json:"date"`
							ContributionCount int `json:"contributionCount"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`

				CommitByRepo []struct{
					Repository struct{
						NameWithOwner string `json:"nameWithOwner"`
						IsPrivate bool `json:"isPrivate"`
					} `json:"repository"`
					Contributions struct{
						TotalCount int `json:"totalCount"`
					} `json:"contributions"`
				} `json:"commitContributionsByRepository"`

				PRByRepo []struct{
					Repository struct{
						NameWithOwner string `json:"nameWithOwner"`
						IsPrivate bool `json:"isPrivate"`
					} `json:"repository"`
					Contributions struct{
						TotalCount int `json:"totalCount"`
					} `json:"contributions"`
				} `json:"pullRequestContributionsByRepository"`

				IssueByRepo []struct{
					Repository struct{
						NameWithOwner string `json:"nameWithOwner"`
						IsPrivate bool `json:"isPrivate"`
					} `json:"repository"`
					Contributions struct{
						TotalCount int `json:"totalCount"`
					} `json:"contributions"`
				} `json:"issueContributionsByRepository"`

				ReviewByRepo []struct{
					Repository struct{
						NameWithOwner string `json:"nameWithOwner"`
						IsPrivate bool `json:"isPrivate"`
					} `json:"repository"`
					Contributions struct{
						TotalCount int `json:"totalCount"`
					} `json:"contributions"`
				} `json:"pullRequestReviewContributionsByRepository"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	var out resp
	if err := c.doGraphQL(context.Background(), q, map[string]any{
		"login": login,
		"from": from.Format(time.RFC3339),
		"to": to.Format(time.RFC3339),
	}, &out); err != nil {
		return nil, err
	}

	cc := &model.ContributionsCollection{
		TotalCommits: out.User.ContributionsCollection.TotalCommitContributions,
		TotalPRs: out.User.ContributionsCollection.TotalPullRequestContributions,
		TotalIssues: out.User.ContributionsCollection.TotalIssueContributions,
		TotalReviews: out.User.ContributionsCollection.TotalPullRequestReviewContributions,
		Calendar: make([]model.ContributionDay, 0, 400),
		ByRepoCommits: map[string]model.RepoContribLite{},
		ByRepoPRs: map[string]model.RepoContribLite{},
		ByRepoIssues: map[string]model.RepoContribLite{},
		ByRepoReviews: map[string]model.RepoContribLite{},
	}

	for _, w := range out.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, d := range w.ContributionDays {
			cc.Calendar = append(cc.Calendar, model.ContributionDay{
				Date: d.Date,
				Count: d.ContributionCount,
			})
		}
	}

	for _, x := range out.User.ContributionsCollection.CommitByRepo {
		cc.ByRepoCommits[x.Repository.NameWithOwner] = model.RepoContribLite{
			Repo: x.Repository.NameWithOwner,
			Count: x.Contributions.TotalCount,
			IsPrivate: x.Repository.IsPrivate,
		}
	}
	for _, x := range out.User.ContributionsCollection.PRByRepo {
		cc.ByRepoPRs[x.Repository.NameWithOwner] = model.RepoContribLite{
			Repo: x.Repository.NameWithOwner,
			Count: x.Contributions.TotalCount,
			IsPrivate: x.Repository.IsPrivate,
		}
	}
	for _, x := range out.User.ContributionsCollection.IssueByRepo {
		cc.ByRepoIssues[x.Repository.NameWithOwner] = model.RepoContribLite{
			Repo: x.Repository.NameWithOwner,
			Count: x.Contributions.TotalCount,
			IsPrivate: x.Repository.IsPrivate,
		}
	}
	for _, x := range out.User.ContributionsCollection.ReviewByRepo {
		cc.ByRepoReviews[x.Repository.NameWithOwner] = model.RepoContribLite{
			Repo: x.Repository.NameWithOwner,
			Count: x.Contributions.TotalCount,
			IsPrivate: x.Repository.IsPrivate,
		}
	}

	return cc, nil
}
