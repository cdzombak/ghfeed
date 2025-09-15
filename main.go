package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

var version = "<dev>"

// Commit represents a single commit with its metadata
type Commit struct {
	Hash    string
	Message string
	Link    string
}

// BranchActivity represents all commits to a specific repository/branch
type BranchActivity struct {
	Repo        string
	Branch      string
	Commits     []Commit
	LatestTime  *time.Time
	CompareLink string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if os.Args[1] == "-help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		os.Exit(0)
	}

	if os.Args[1] == "-version" || os.Args[1] == "--version" || os.Args[1] == "-v" {
		printVersion()
		os.Exit(0)
	}

	feedURL := os.Args[1]

	// Parse the feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing feed: %v\n", err)
		os.Exit(1)
	}

	// Process and consolidate the feed
	consolidatedFeed := consolidateCommits(feed)

	// Render as Atom
	err = consolidatedFeed.RenderAtom(os.Stdout, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering feed: %v\n", err)
		os.Exit(1)
	}
}

// consolidateCommits groups commit/push activities by repository/branch and returns a new feed
func consolidateCommits(feed *gofeed.Feed) *gofeed.Feed {
	// Create new feed with same metadata
	newFeed := &gofeed.Feed{
		Title:         feed.Title,
		Description:   feed.Description,
		Link:          feed.Link,
		FeedLink:      feed.FeedLink,
		Updated:       feed.Updated,
		UpdatedParsed: feed.UpdatedParsed,
		Language:      feed.Language,
		Copyright:     feed.Copyright,
		Generator:     feed.Generator,
		Categories:    feed.Categories,
		Authors:       feed.Authors,
		Image:         feed.Image,
		FeedType:      feed.FeedType,
		FeedVersion:   feed.FeedVersion,
		Items:         []*gofeed.Item{},
	}

	// Group items by repository/branch for commits/pushes
	branchGroups := make(map[string]*BranchActivity)
	nonCommitItems := []*gofeed.Item{}

	for _, item := range feed.Items {
		if isCommitOrPush(item.Title) {
			activity := extractBranchActivity(item)
			if activity != nil {
				key := fmt.Sprintf("%s/%s", activity.Repo, activity.Branch)
				if existing, exists := branchGroups[key]; exists {
					// Merge commits and update latest time
					existing.Commits = append(existing.Commits, activity.Commits...)
					if activity.LatestTime != nil && (existing.LatestTime == nil || activity.LatestTime.After(*existing.LatestTime)) {
						existing.LatestTime = activity.LatestTime
						existing.CompareLink = activity.CompareLink
					}
				} else {
					branchGroups[key] = activity
				}
			} else {
				// If we can't extract branch activity, keep as-is
				nonCommitItems = append(nonCommitItems, item)
			}
		} else {
			nonCommitItems = append(nonCommitItems, item)
		}
	}

	// Create consolidated items for each repository/branch
	for _, activity := range branchGroups {
		consolidatedItem := createConsolidatedBranchItem(activity)
		if consolidatedItem != nil {
			newFeed.Items = append(newFeed.Items, consolidatedItem)
		}
	}

	// Process and simplify non-commit items
	for _, item := range nonCommitItems {
		simplifiedItem := simplifyNonCommitItem(item)
		newFeed.Items = append(newFeed.Items, simplifiedItem)
	}

	// Sort items by published date (most recent first)
	sort.Slice(newFeed.Items, func(i, j int) bool {
		if newFeed.Items[i].PublishedParsed == nil {
			return false
		}
		if newFeed.Items[j].PublishedParsed == nil {
			return true
		}
		return newFeed.Items[i].PublishedParsed.After(*newFeed.Items[j].PublishedParsed)
	})

	return newFeed
}

// isCommitOrPush determines if an item represents a commit or push activity
func isCommitOrPush(title string) bool {
	commitPushPatterns := []string{
		"pushed",
		"created branch",
		"deleted branch",
		"created tag",
		"deleted tag",
	}

	titleLower := strings.ToLower(title)
	for _, pattern := range commitPushPatterns {
		if strings.Contains(titleLower, pattern) {
			return true
		}
	}
	return false
}

// extractBranchActivity extracts repository, branch, and commit data from a push item
func extractBranchActivity(item *gofeed.Item) *BranchActivity {
	// Extract repo name from link
	repoName := ""
	if item.Link != "" {
		repoLinkRegex := regexp.MustCompile(`github\.com/cdzombak/([\w-]+)`)
		matches := repoLinkRegex.FindStringSubmatch(item.Link)
		if len(matches) > 1 {
			repoName = matches[1]
		}
	}

	if repoName == "" {
		return nil
	}

	// Extract branch name from content
	branchName := "master" // default
	if item.Content != "" {
		branchRegex := regexp.MustCompile(`<a class="branch-name"[^>]*href="[^"]*/tree/([^"]*)"[^>]*>([^<]+)</a>`)
		matches := branchRegex.FindStringSubmatch(item.Content)
		if len(matches) > 2 {
			branchName = matches[2]
		}
	}

	// Extract commits from content
	commits := extractCommitsFromContent(item.Content)

	activity := &BranchActivity{
		Repo:        repoName,
		Branch:      branchName,
		Commits:     commits,
		LatestTime:  item.PublishedParsed,
		CompareLink: item.Link,
	}

	return activity
}

// extractCommitsFromContent parses commit information from HTML content
func extractCommitsFromContent(content string) []Commit {
	var commits []Commit

	// More flexible regex to match commit entries in the HTML
	// This accounts for the div wrapper around blockquote
	commitRegex := regexp.MustCompile(`<code[^>]*><a[^>]*href="([^"]*commit/([a-f0-9]+))"[^>]*>([a-f0-9]+)</a></code>\s*<div[^>]*>\s*<blockquote[^>]*>\s*([^<]*?)\s*</blockquote>`)
	matches := commitRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 5 {
			commit := Commit{
				Hash:    match[3], // short hash
				Message: strings.TrimSpace(match[4]),
				Link:    "https://github.com" + match[1], // full commit URL
			}
			commits = append(commits, commit)
		}
	}

	// If no commits found with the above regex, try without the div wrapper
	if len(commits) == 0 {
		simpleBlockquoteRegex := regexp.MustCompile(`<code[^>]*><a[^>]*href="([^"]*commit/([a-f0-9]+))"[^>]*>([a-f0-9]+)</a></code>.*?<blockquote[^>]*>\s*([^<]*?)\s*</blockquote>`)
		matches = simpleBlockquoteRegex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) >= 5 {
				commit := Commit{
					Hash:    match[3], // short hash
					Message: strings.TrimSpace(match[4]),
					Link:    "https://github.com" + match[1], // full commit URL
				}
				commits = append(commits, commit)
			}
		}
	}

	// If still no commits found, try to extract just from links to commits
	if len(commits) == 0 {
		linkRegex := regexp.MustCompile(`href="([^"]*commit/([a-f0-9]+))"[^>]*>([a-f0-9]+)</a>`)
		linkMatches := linkRegex.FindAllStringSubmatch(content, -1)

		for _, match := range linkMatches {
			if len(match) >= 4 {
				commit := Commit{
					Hash:    match[3],                        // short hash
					Message: "Commit " + match[3],            // fallback message
					Link:    "https://github.com" + match[1], // full commit URL
				}
				commits = append(commits, commit)
			}
		}
	}

	return commits
}

// createConsolidatedBranchItem creates a single item representing all commits to a repository/branch
func createConsolidatedBranchItem(activity *BranchActivity) *gofeed.Item {
	if len(activity.Commits) == 0 {
		return nil
	}

	// Count commits for title
	commitCount := len(activity.Commits)
	title := fmt.Sprintf("cdzombak pushed %d commits to %s/%s", commitCount, activity.Repo, activity.Branch)

	// Create HTML description with commit details
	var htmlParts []string
	htmlParts = append(htmlParts, "<div>")

	for _, commit := range activity.Commits {
		commitHTML := fmt.Sprintf(
			"<div style='margin-bottom: 12px;'>"+
				"<code><a href='%s'>%s</a></code>: %s"+
				"</div>",
			commit.Link,
			commit.Hash,
			commit.Message,
		)
		htmlParts = append(htmlParts, commitHTML)
	}

	// Add compare link if available
	if activity.CompareLink != "" {
		htmlParts = append(htmlParts, fmt.Sprintf(
			"<div style='margin-top: 16px; border-top: 1px solid #eee; padding-top: 8px;'>"+
				"<a href='%s'>View all changes</a>"+
				"</div>",
			activity.CompareLink,
		))
	}

	htmlParts = append(htmlParts, "</div>")
	htmlContent := strings.Join(htmlParts, "")

	// Create consolidated item
	consolidatedItem := &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            activity.CompareLink,
		Published:       activity.LatestTime.Format(time.RFC3339),
		PublishedParsed: activity.LatestTime,
		Updated:         activity.LatestTime.Format(time.RFC3339),
		UpdatedParsed:   activity.LatestTime,
		GUID:            fmt.Sprintf("consolidated-%s-%s-%d", activity.Repo, activity.Branch, activity.LatestTime.Unix()),
	}

	return consolidatedItem
}

// ActivityType represents different types of GitHub activities
type ActivityType int

const (
	ActivityPullRequest ActivityType = iota
	ActivityFork
	ActivityBranchCreate
	ActivityBranchDelete
	ActivityTagDelete
	ActivityOther
)

// detectActivityType determines what type of GitHub activity an item represents
func detectActivityType(item *gofeed.Item) ActivityType {
	title := strings.ToLower(item.Title)

	if strings.Contains(title, "opened a pull request") || strings.Contains(title, "contributed to") || (strings.Contains(title, "opened") && strings.Contains(item.Content, "pull_request")) {
		return ActivityPullRequest
	}
	if strings.Contains(title, "forked") {
		return ActivityFork
	}
	if strings.Contains(title, "created a branch") || strings.Contains(title, "created branch") {
		return ActivityBranchCreate
	}
	if strings.Contains(title, "deleted branch") {
		return ActivityBranchDelete
	}
	if strings.Contains(title, "deleted") && (strings.Contains(title, "tag") || strings.Contains(item.Content, "tag")) {
		return ActivityTagDelete
	}

	return ActivityOther
}

// simplifyNonCommitItem creates a simplified version of non-commit GitHub activities
func simplifyNonCommitItem(item *gofeed.Item) *gofeed.Item {
	activityType := detectActivityType(item)

	switch activityType {
	case ActivityPullRequest:
		return simplifyPullRequest(item)
	case ActivityFork:
		return simplifyFork(item)
	case ActivityBranchCreate:
		return simplifyBranchCreate(item)
	case ActivityBranchDelete:
		return simplifyBranchDelete(item)
	case ActivityTagDelete:
		return simplifyTagDelete(item)
	default:
		// For other activities, create a basic simplified version
		return simplifyOtherActivity(item)
	}
}

// simplifyPullRequest creates a clean, simple pull request entry
func simplifyPullRequest(item *gofeed.Item) *gofeed.Item {
	// Extract PR number and repository from link
	prNumber := ""
	targetRepo := ""

	if item.Link != "" {
		// Extract PR number: /pull/264
		prRegex := regexp.MustCompile(`/pull/(\d+)`)
		matches := prRegex.FindStringSubmatch(item.Link)
		if len(matches) > 1 {
			prNumber = matches[1]
		}

		// Extract repository: github.com/user/repo
		repoRegex := regexp.MustCompile(`github\.com/([^/]+/[^/]+)`)
		matches = repoRegex.FindStringSubmatch(item.Link)
		if len(matches) > 1 {
			targetRepo = matches[1]
		}
	}

	// Extract PR title from content
	prTitle := ""
	if item.Content != "" {
		titleRegex := regexp.MustCompile(`<span[^>]*class="[^"]*text-bold[^"]*"[^>]*><a[^>]*>([^<]+)</a></span>`)
		matches := titleRegex.FindStringSubmatch(item.Content)
		if len(matches) > 1 {
			prTitle = strings.TrimSpace(matches[1])
		}
	}

	// Extract diff stats
	diffStats := ""
	if item.Content != "" {
		diffRegex := regexp.MustCompile(`<span class="color-fg-success">([^<]+)</span>\s*<span class="color-fg-danger">([^<]+)</span>`)
		matches := diffRegex.FindStringSubmatch(item.Content)
		if len(matches) > 2 {
			diffStats = matches[1] + " " + matches[2]
		}
	}

	// Create simplified title
	title := fmt.Sprintf("cdzombak opened PR #%s in %s", prNumber, targetRepo)
	if prTitle != "" {
		title += ": " + prTitle
	}

	// Create simplified HTML content
	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`<a href='%s'>View PR #%s</a>`, item.Link, prNumber)
	if prTitle != "" {
		htmlContent += fmt.Sprintf(`<div style='margin-top: 8px; font-weight: bold;'>%s</div>`, prTitle)
	}
	if diffStats != "" {
		htmlContent += fmt.Sprintf(`<div style='margin-top: 8px; font-family: monospace; color: #666;'>%s</div>`, diffStats)
	}
	htmlContent += `</div>`

	return &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            item.Link,
		Published:       item.Published,
		PublishedParsed: item.PublishedParsed,
		Updated:         item.Updated,
		UpdatedParsed:   item.UpdatedParsed,
		Authors:         item.Authors,
		GUID:            item.GUID,
	}
}

// simplifyFork creates a clean fork entry
func simplifyFork(item *gofeed.Item) *gofeed.Item {
	// Extract source and target repos from title or content
	sourceRepo := ""
	targetRepo := ""

	if item.Title != "" {
		// Example: "cdzombak forked cdzombak/gofeed from mmcdole/gofeed"
		forkRegex := regexp.MustCompile(`forked ([^/]+/[^/\s]+) from ([^/]+/[^/\s]+)`)
		matches := forkRegex.FindStringSubmatch(item.Title)
		if len(matches) > 2 {
			targetRepo = matches[1]
			sourceRepo = matches[2]
		}
	}

	title := fmt.Sprintf("cdzombak forked %s", sourceRepo)

	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`<a href='%s'>View fork: %s</a>`, item.Link, targetRepo)
	htmlContent += `</div>`

	return &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            item.Link,
		Published:       item.Published,
		PublishedParsed: item.PublishedParsed,
		Updated:         item.Updated,
		UpdatedParsed:   item.UpdatedParsed,
		Authors:         item.Authors,
		GUID:            item.GUID,
	}
}

// simplifyBranchCreate creates a clean branch creation entry
func simplifyBranchCreate(item *gofeed.Item) *gofeed.Item {
	// Extract branch name and repository
	branchName := ""
	repoName := ""

	if item.Content != "" {
		branchRegex := regexp.MustCompile(`title="([^"]*)"[^>]*>([^<]+)</a> in <a[^>]*>([^/]+/[^<]+)</a>`)
		matches := branchRegex.FindStringSubmatch(item.Content)
		if len(matches) > 3 {
			branchName = strings.TrimPrefix(matches[1], "refs/heads/")
			repoName = matches[3]
		}
	}

	if branchName == "" && item.Link != "" {
		// Fallback: extract from link
		if strings.Contains(item.Link, "/tree/") {
			parts := strings.Split(item.Link, "/tree/")
			if len(parts) > 1 {
				branchName = parts[1]
			}
		}
	}

	title := fmt.Sprintf("cdzombak created branch %s", branchName)
	if repoName != "" {
		title += fmt.Sprintf(" in %s", repoName)
	}

	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`<a href='%s'>View branch: %s</a>`, item.Link, branchName)
	htmlContent += `</div>`

	return &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            item.Link,
		Published:       item.Published,
		PublishedParsed: item.PublishedParsed,
		Updated:         item.Updated,
		UpdatedParsed:   item.UpdatedParsed,
		Authors:         item.Authors,
		GUID:            item.GUID,
	}
}

// simplifyBranchDelete creates a clean branch deletion entry
func simplifyBranchDelete(item *gofeed.Item) *gofeed.Item {
	title := "cdzombak deleted a branch"

	htmlContent := `<div style='margin-bottom: 12px;'>Branch deleted</div>`

	return &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            item.Link,
		Published:       item.Published,
		PublishedParsed: item.PublishedParsed,
		Updated:         item.Updated,
		UpdatedParsed:   item.UpdatedParsed,
		Authors:         item.Authors,
		GUID:            item.GUID,
	}
}

// simplifyTagDelete creates a clean tag deletion entry
func simplifyTagDelete(item *gofeed.Item) *gofeed.Item {
	tagName := ""
	repoName := ""

	// Try to extract from title first
	if item.Title != "" {
		titleRegex := regexp.MustCompile(`cdzombak deleted (?:tag )?(.*?)(?:\s|$)`)
		matches := titleRegex.FindStringSubmatch(item.Title)
		if len(matches) > 1 {
			tagName = strings.TrimSpace(matches[1])
		}
	}

	// Try to extract from content
	if item.Content != "" {
		tagRegex := regexp.MustCompile(`<span class="branch-name">([^<]+)</span>.*?<a[^>]*>([^/]+/[^<]+)</a>`)
		matches := tagRegex.FindStringSubmatch(item.Content)
		if len(matches) > 2 {
			if tagName == "" {
				tagName = matches[1]
			}
			repoName = matches[2]
		}
	}

	// Extract repo from link if not found
	if repoName == "" && item.Link != "" {
		repoRegex := regexp.MustCompile(`github\.com/cdzombak/([^/]+)`)
		matches := repoRegex.FindStringSubmatch(item.Link)
		if len(matches) > 1 {
			repoName = matches[1]
		}
	}

	if tagName == "" {
		tagName = "tag"
	}

	title := fmt.Sprintf("cdzombak deleted tag %s", tagName)
	if repoName != "" {
		title += fmt.Sprintf(" in %s", repoName)
	}

	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`Deleted tag: %s`, tagName)
	if repoName != "" {
		htmlContent += fmt.Sprintf(` in %s`, repoName)
	}
	htmlContent += `</div>`

	return &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            item.Link,
		Published:       item.Published,
		PublishedParsed: item.PublishedParsed,
		Updated:         item.Updated,
		UpdatedParsed:   item.UpdatedParsed,
		Authors:         item.Authors,
		GUID:            item.GUID,
	}
}

// simplifyOtherActivity creates a basic simplified version for unrecognized activities
func simplifyOtherActivity(item *gofeed.Item) *gofeed.Item {
	// Keep the original title but create simpler content
	htmlContent := `<div style='margin-bottom: 12px;'>`
	if item.Link != "" {
		htmlContent += fmt.Sprintf(`<a href='%s'>View activity</a>`, item.Link)
	} else {
		htmlContent += `GitHub activity`
	}
	htmlContent += `</div>`

	return &gofeed.Item{
		Title:           item.Title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            item.Link,
		Published:       item.Published,
		PublishedParsed: item.PublishedParsed,
		Updated:         item.Updated,
		UpdatedParsed:   item.UpdatedParsed,
		Authors:         item.Authors,
		GUID:            item.GUID,
	}
}

// printUsage prints basic usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <feed-url>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s -help\n", os.Args[0])
}

func printVersion() {
	fmt.Printf("ghfeed version %s\n", version)
}

// printHelp prints detailed help information
func printHelp() {
	fmt.Printf("ghfeed - GitHub Activity Feed Consolidator\n\n")
	fmt.Printf("version %s\n\n", version)

	fmt.Printf("USAGE:\n")
	fmt.Printf("  %s <feed-url>\n\n", os.Args[0])

	fmt.Printf("DESCRIPTION:\n")
	fmt.Printf("  Transforms verbose GitHub Atom feeds into clean, readable summaries.\n")
	fmt.Printf("  Consolidates multiple commits by repository/branch and simplifies\n")
	fmt.Printf("  other GitHub activities while preserving all essential information.\n\n")

	fmt.Printf("FEATURES:\n")
	fmt.Printf("  • Consolidates commits: \"cdzombak pushed 15 commits to dotfiles/master\"\n")
	fmt.Printf("  • Simplifies activities: \"cdzombak opened PR #264 in mmcdole/gofeed\"\n")
	fmt.Printf("  • Preserves commit messages, links, and metadata\n")
	fmt.Printf("  • Generates clean HTML with clickable links\n")
	fmt.Printf("  • Maintains chronological ordering\n\n")

	fmt.Printf("EXAMPLE:\n")
	fmt.Printf("  %s https://github.com/username.atom\n\n", os.Args[0])

	fmt.Printf("AUTHOR:\n")
	fmt.Printf("  Chris Dzombak: https://dzombak.com, https://github.com/cdzombak\n\n")

	fmt.Printf("GITHUB:\n")
	fmt.Printf("  https://github.com/cdzombak/ghfeed\n\n")
}
