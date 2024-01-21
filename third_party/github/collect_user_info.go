package github

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/shurcooL/githubv4"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_errors"
	"net/http"
	"time"
)

var latestUserInfo *map[string]any

// GetUserInfo 获取用户信息.
func GetUserInfo(c *gin.Context) {
	if latestUserInfo == nil {
		if err := collectUserInfo(); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, comfy_errors.NewResponseError(comfy_errors.UnknownError, "get user info failed, %w", err))
		}
	}

	c.JSON(http.StatusOK, latestUserInfo)
}

type repositoryInfo struct {
	TotalDiskUsage int
	StargazerCount int
	WatcherCount   int
	ForkCount      int
}

// startFetchUserInfoLoop 启动定时获取用户信息的循环.
func startFetchUserInfoLoop(context context.Context) {
	// 一小时获取一次用户统计信息
	ticker := time.NewTicker(time.Hour)

	go func() {
		for {
			select {
			case <-context.Done():
				ticker.Stop()
				logger.Info("stop fetch user info loop\n")
				return
			case <-ticker.C:
				if err := collectUserInfo(); err != nil {
					logger.Error("get user info failed, %w\n", err)
					continue
				} else {
					logger.Info("update user info success\n")
				}
			}
		}
	}()
}

// collectUserInfo 收集用户统计信息.
func collectUserInfo() error {
	// from https://medium.com/@yuichkun/how-to-retrieve-contribution-graph-data-from-the-github-api-dc3a151b4af
	type userInfoQuery struct {
		Viewer struct {
			Name      githubv4.String
			Followers struct {
				TotalCount githubv4.Int
			}
			Following struct {
				TotalCount githubv4.Int
			}
			Email     githubv4.String
			AvatarUrl githubv4.String
			Url       githubv4.String

			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions githubv4.Int
					Weeks              []struct {
						ContributionDays []struct {
							ContributionCount githubv4.Int    `json:"contributionCount"`
							Date              githubv4.String `json:"date"`
							Color             githubv4.String `json:"color"`
							ContributionLevel githubv4.String `json:"contributionLevel"`
						} `json:"contributionDays"`
					}
				}
			}
		}
	}

	type repositoryQuery struct {
		Viewer struct {
			Repositories struct {
				Edges []struct {
					Cursor githubv4.String
					Node   struct {
						Name           githubv4.String
						NameWithOwner  githubv4.String
						Description    githubv4.String
						StargazerCount githubv4.Int
						Watchers       struct {
							TotalCount githubv4.Int
						}
						ForkCount githubv4.Int
					}
				}

				PageInfo struct {
					EndCursor       githubv4.String
					StartCursor     githubv4.String
					HasNextPage     githubv4.Boolean
					HasPreviousPage githubv4.Boolean
				}

				TotalCount     githubv4.Int
				TotalDiskUsage githubv4.Int
			} `graphql:"repositories(isFork: $isFork, privacy: $privacy, first: $first, after: $after)"`
		}
	}

	// 获取用户信息
	var userInfoQueryResult userInfoQuery
	if err := graphqlClient.Query(context.Background(), &userInfoQueryResult, nil); err != nil {
		return errors.New(err)
	}

	// 仓库的统计信息
	var repositoryInfoResult repositoryInfo
	// 循环获取仓库列表信息
	var repositoryQueryResult repositoryQuery
	for {
		queryVariables := map[string]interface{}{
			"isFork":  githubv4.Boolean(false),
			"privacy": githubv4.RepositoryPrivacyPublic,
			"first":   githubv4.Int(50),
		}
		if len(repositoryQueryResult.Viewer.Repositories.PageInfo.EndCursor) > 0 {
			queryVariables["after"] = repositoryQueryResult.Viewer.Repositories.PageInfo.EndCursor
		} else {
			queryVariables["after"] = (*githubv4.String)(nil)
		}

		if err := graphqlClient.Query(context.Background(), &repositoryQueryResult, queryVariables); err != nil {
			return errors.New(err)
		}

		repositoryInfoResult.TotalDiskUsage = int(repositoryQueryResult.Viewer.Repositories.TotalDiskUsage)
		for _, edge := range repositoryQueryResult.Viewer.Repositories.Edges {
			repositoryInfoResult.StargazerCount += int(edge.Node.StargazerCount)
			repositoryInfoResult.WatcherCount += int(edge.Node.Watchers.TotalCount)
			repositoryInfoResult.ForkCount += int(edge.Node.ForkCount)
		}

		if !repositoryQueryResult.Viewer.Repositories.PageInfo.HasNextPage {
			break
		}
	}

	userInfoResult := userInfoQueryResult.Viewer

	latestUserInfo = &map[string]any{
		"name":           userInfoResult.Name,
		"followerCount":  userInfoResult.Followers.TotalCount,
		"followingCount": userInfoResult.Following.TotalCount,
		"email":          userInfoResult.Email,
		"avatarUrl":      userInfoResult.AvatarUrl,
		"url":            userInfoResult.Url,
		"contributionCalendar": map[string]any{
			"totalContributions": userInfoResult.ContributionsCollection.ContributionCalendar.TotalContributions,
			"weeks":              userInfoResult.ContributionsCollection.ContributionCalendar.Weeks,
		},
		"repositoryInfo": map[string]any{
			"totalDiskUsage": repositoryInfoResult.TotalDiskUsage,
			"stargazerCount": repositoryInfoResult.StargazerCount,
			"watcherCount":   repositoryInfoResult.WatcherCount,
			"forkCount":      repositoryInfoResult.ForkCount,
		},
	}

	return nil
}
