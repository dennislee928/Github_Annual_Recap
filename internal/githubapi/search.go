\
package githubapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dennislee928/github-recap-2025/internal/model"
)

func (c *Client) SearchPullRequests(login string, from, to time.Time, maxResults int) ([]model.PRItem, error) {
	qstr := fmt.Sprintf("author:%s is:pr created:%s..%s", login, from.Format("2006-01-02"), to.Format("2006-01-02"))
	return c.searchPR(qstr, maxResults)
}

func (c *Client) SearchIssuesOpenedClosed(login string, from, to time.Time, maxResults int) (opened []model.IssueItem, closed []model.IssueItem, err error) {
	qOpened := fmt.Sprintf("author:%s is:issue created:%s..%s", login, from.Format("2006-01-02"), to.Format("2006-01-02"))
	qClosed := fmt.Sprintf("author:%s is:issue closed:%s..%s", login, from.Format("2006-01-02"), to.Format("2006-01-02"))

	opened, err = c.searchIssues(qOpened, maxResults, true)
	if err != nil { return nil, nil, err }
	closed, err = c.searchIssues(qClosed, maxResults, false)
	if err != nil { return nil, nil, err }
	return opened, closed, nil
}

func (c *Client) searchPR(query string, maxResults int) ([]model.PRItem, error) {
	const q = `
query($q:String!, $after:String) {
  search(query:$q, type:ISSUE, first:100, after:$after) {
    issueCount
    pageInfo { hasNextPage endCursor }
    nodes {
      ... on PullRequest {
        number
        title
        url
        createdAt
        merged
        mergedAt
        additions
        deletions
        repository { nameWithOwner isPrivate }
      }
    }
  }
}`

	var all []model.PRItem
	var after *string
	for len(all) < maxResults {
		type resp struct{
			Search struct{
				IssueCount int `json:"issueCount"`
				PageInfo struct{
					HasNextPage bool `json:"hasNextPage"`
					EndCursor *string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []struct{
					Number int `json:"number"`
					Title string `json:"title"`
					URL string `json:"url"`
					CreatedAt time.Time `json:"createdAt"`
					Merged bool `json:"merged"`
					MergedAt *time.Time `json:"mergedAt"`
					Additions int `json:"additions"`
					Deletions int `json:"deletions"`
					Repository struct{
						NameWithOwner string `json:"nameWithOwner"`
						IsPrivate bool `json:"isPrivate"`
					} `json:"repository"`
				} `json:"nodes"`
			} `json:"search"`
		}
		var out resp
		vars := map[string]any{"q": query}
		if after != nil { vars["after"] = *after } else { vars["after"] = nil }
		if err := c.doGraphQL(context.Background(), q, vars, &out); err != nil {
			return nil, err
		}
		for _, n := range out.Search.Nodes {
			all = append(all, model.PRItem{
				Repo: n.Repository.NameWithOwner,
				Number: n.Number,
				Title: n.Title,
				URL: n.URL,
				CreatedAt: n.CreatedAt,
				Merged: n.Merged,
				MergedAt: n.MergedAt,
				Additions: n.Additions,
				Deletions: n.Deletions,
			})
			if len(all) >= maxResults { break }
		}
		if !out.Search.PageInfo.HasNextPage || out.Search.PageInfo.EndCursor == nil {
			break
		}
		after = out.Search.PageInfo.EndCursor
	}
	return all, nil
}

func (c *Client) searchIssues(query string, maxResults int, wantCreated bool) ([]model.IssueItem, error) {
	const q = `
query($q:String!, $after:String) {
  search(query:$q, type:ISSUE, first:100, after:$after) {
    issueCount
    pageInfo { hasNextPage endCursor }
    nodes {
      ... on Issue {
        number
        title
        url
        createdAt
        closedAt
        repository { nameWithOwner isPrivate }
      }
    }
  }
}`

	var all []model.IssueItem
	var after *string
	for len(all) < maxResults {
		type resp struct{
			Search struct{
				IssueCount int `json:"issueCount"`
				PageInfo struct{
					HasNextPage bool `json:"hasNextPage"`
					EndCursor *string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []struct{
					Number int `json:"number"`
					Title string `json:"title"`
					URL string `json:"url"`
					CreatedAt time.Time `json:"createdAt"`
					ClosedAt *time.Time `json:"closedAt"`
					Repository struct{
						NameWithOwner string `json:"nameWithOwner"`
						IsPrivate bool `json:"isPrivate"`
					} `json:"repository"`
				} `json:"nodes"`
			} `json:"search"`
		}
		var out resp
		vars := map[string]any{"q": query}
		if after != nil { vars["after"] = *after } else { vars["after"] = nil }
		if err := c.doGraphQL(context.Background(), q, vars, &out); err != nil {
			return nil, err
		}
		for _, n := range out.Search.Nodes {
			all = append(all, model.IssueItem{
				Repo: n.Repository.NameWithOwner,
				Number: n.Number,
				Title: n.Title,
				URL: n.URL,
				CreatedAt: n.CreatedAt,
				ClosedAt: n.ClosedAt,
			})
			if len(all) >= maxResults { break }
		}
		if !out.Search.PageInfo.HasNextPage || out.Search.PageInfo.EndCursor == nil {
			break
		}
		after = out.Search.PageInfo.EndCursor
	}
	return all, nil
}
