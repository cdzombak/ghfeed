package main

import (
	"encoding/json"
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

	// Parse command line arguments
	var feedURL string
	var customTitle string
	var format = "atom"          // default format
	var consolidatePushes = true // default to true for backward compatibility

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-retitle" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: -retitle flag requires a title argument\n")
				os.Exit(1)
			}
			customTitle = args[i+1]
			i++ // Skip the next argument since we consumed it
		} else if arg == "-format" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: -format flag requires a format argument (atom, rss, or json)\n")
				os.Exit(1)
			}
			format = args[i+1]
			if format != "atom" && format != "rss" && format != "json" {
				fmt.Fprintf(os.Stderr, "Error: format must be 'atom', 'rss', or 'json'\n")
				os.Exit(1)
			}
			i++ // Skip the next argument since we consumed it
		} else if arg == "-consolidate-pushes" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: -consolidate-pushes flag requires a boolean argument (true or false)\n")
				os.Exit(1)
			}
			switch args[i+1] {
			case "true":
				consolidatePushes = true
			case "false":
				consolidatePushes = false
			default:
				fmt.Fprintf(os.Stderr, "Error: -consolidate-pushes must be 'true' or 'false'\n")
				os.Exit(1)
			}
			i++ // Skip the next argument since we consumed it
		} else if feedURL == "" {
			feedURL = arg
		} else {
			fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n", arg)
			os.Exit(1)
		}
	}

	if feedURL == "" {
		printUsage()
		os.Exit(1)
	}

	// Parse the feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing feed: %v\n", err)
		os.Exit(1)
	}

	// Process and consolidate the feed
	consolidatedFeed := consolidateCommits(feed, customTitle, consolidatePushes)

	// Render in the specified format
	err = renderFeed(consolidatedFeed, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering feed: %v\n", err)
		os.Exit(1)
	}
}

// renderFeed outputs the feed in the specified format
func renderFeed(feed *gofeed.Feed, format string) error {
	switch format {
	case "atom":
		return feed.RenderAtom(os.Stdout, nil)
	case "rss":
		return feed.RenderRSS(os.Stdout, nil)
	case "json":
		return renderJSON(feed)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// renderJSON outputs the feed as JSON
func renderJSON(feed *gofeed.Feed) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(feed)
}

// consolidateCommits groups commit/push activities by repository/branch and returns a new feed
func consolidateCommits(feed *gofeed.Feed, customTitle string, consolidatePushes bool) *gofeed.Feed {
	// Extract username from feed link or items
	username := extractUsername(feed)

	// Create new feed with same metadata
	title := feed.Title
	if customTitle != "" {
		title = customTitle
	}
	newFeed := &gofeed.Feed{
		Title:         title,
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

	// Group items by repository/branch for commits/pushes (if consolidating)
	if consolidatePushes {
		branchGroups := make(map[string]*BranchActivity)
		nonCommitItems := []*gofeed.Item{}

		for _, item := range feed.Items {
			if isCommitOrPush(item.Title) {
				activity := extractBranchActivity(item, username)
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
			// Generate proper comparison link that encompasses all commits
			activity.CompareLink = generateComparisonLink(activity, username)
			consolidatedItem := createConsolidatedBranchItem(activity, username)
			if consolidatedItem != nil {
				newFeed.Items = append(newFeed.Items, consolidatedItem)
			}
		}

		// Process and simplify non-commit items
		for _, item := range nonCommitItems {
			simplifiedItem := simplifyNonCommitItem(item, username)
			newFeed.Items = append(newFeed.Items, simplifiedItem)
		}
	} else {
		// Process each item individually without consolidation
		for _, item := range feed.Items {
			if isCommitOrPush(item.Title) {
				activity := extractBranchActivity(item, username)
				if activity != nil {
					// Generate proper comparison link that encompasses all commits in this push
					activity.CompareLink = generateComparisonLink(activity, username)
					individualItem := createIndividualPushItem(activity, username)
					if individualItem != nil {
						newFeed.Items = append(newFeed.Items, individualItem)
					}
				} else {
					// If we can't extract branch activity, keep as-is
					simplifiedItem := simplifyNonCommitItem(item, username)
					newFeed.Items = append(newFeed.Items, simplifiedItem)
				}
			} else {
				simplifiedItem := simplifyNonCommitItem(item, username)
				newFeed.Items = append(newFeed.Items, simplifiedItem)
			}
		}
	}

	// Sort items by updated date (most recent first), falling back to published date
	sort.Slice(newFeed.Items, func(i, j int) bool {
		// Get effective date for item i (updated or published)
		dateI := newFeed.Items[i].UpdatedParsed
		if dateI == nil {
			dateI = newFeed.Items[i].PublishedParsed
		}
		if dateI == nil {
			return false
		}

		// Get effective date for item j (updated or published)
		dateJ := newFeed.Items[j].UpdatedParsed
		if dateJ == nil {
			dateJ = newFeed.Items[j].PublishedParsed
		}
		if dateJ == nil {
			return true
		}

		return dateI.After(*dateJ)
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

// extractUsername extracts the GitHub username from the feed
func extractUsername(feed *gofeed.Feed) string {
	// Try to extract from feed link first (e.g., https://github.com/username.atom)
	if feed.Link != "" {
		userRegex := regexp.MustCompile(`github\.com/([^/\.]+)(?:\.atom)?`)
		matches := userRegex.FindStringSubmatch(feed.Link)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Try to extract from feed items
	for _, item := range feed.Items {
		if item.Link != "" {
			userRegex := regexp.MustCompile(`github\.com/([^/]+)/`)
			matches := userRegex.FindStringSubmatch(item.Link)
			if len(matches) > 1 {
				return matches[1]
			}
		}
	}

	// Fallback to "user" if we can't extract username
	return "user"
}

// extractBranchActivity extracts repository, branch, and commit data from a push item
func extractBranchActivity(item *gofeed.Item, username string) *BranchActivity {
	// Extract repo name from link
	repoName := ""
	if item.Link != "" {
		repoLinkRegex := regexp.MustCompile(`github\.com/` + regexp.QuoteMeta(username) + `/([\w-]+)`)
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

// generateComparisonLink creates a GitHub comparison link that encompasses all commits in the activity
func generateComparisonLink(activity *BranchActivity, username string) string {
	if len(activity.Commits) == 0 {
		return activity.CompareLink // fallback to original
	}

	// If only one commit, link directly to it
	if len(activity.Commits) == 1 {
		return activity.Commits[0].Link
	}

	// For multiple commits, create comparison link from oldest to newest
	// GitHub compares show oldest..newest
	oldestCommit := activity.Commits[len(activity.Commits)-1] // commits are typically in newest-first order
	newestCommit := activity.Commits[0]

	// Extract commit hashes from their links
	oldestHash := extractCommitHashFromLink(oldestCommit.Link)
	newestHash := extractCommitHashFromLink(newestCommit.Link)

	if oldestHash != "" && newestHash != "" && oldestHash != newestHash {
		return fmt.Sprintf("https://github.com/%s/%s/compare/%s^...%s", username, activity.Repo, oldestHash, newestHash)
	}

	// Fallback to newest commit if we can't create comparison
	return newestCommit.Link
}

// extractCommitHashFromLink extracts the commit hash from a GitHub commit URL
func extractCommitHashFromLink(link string) string {
	commitRegex := regexp.MustCompile(`/commit/([a-f0-9]+)`)
	matches := commitRegex.FindStringSubmatch(link)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// createConsolidatedBranchItem creates a single item representing all commits to a repository/branch
func createConsolidatedBranchItem(activity *BranchActivity, username string) *gofeed.Item {
	if len(activity.Commits) == 0 {
		return nil
	}

	// Count commits for title
	commitCount := len(activity.Commits)
	commitWord := "commits"
	if commitCount == 1 {
		commitWord = "commit"
	}
	title := fmt.Sprintf("%s pushed %d %s to %s/%s", username, commitCount, commitWord, activity.Repo, activity.Branch)

	// Create HTML description with commit details
	var htmlParts []string
	htmlParts = append(htmlParts, "<div>")

	for _, commit := range activity.Commits {
		commitHTML := fmt.Sprintf(
			"<div style='margin-bottom: 12px;'>"+
				"<tt><a href='%s'>%s</a></tt>: %s"+
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

// createIndividualPushItem creates a single item representing one push to a repository/branch
func createIndividualPushItem(activity *BranchActivity, username string) *gofeed.Item {
	if len(activity.Commits) == 0 {
		return nil
	}

	// Count commits for title
	commitCount := len(activity.Commits)
	commitWord := "commits"
	if commitCount == 1 {
		commitWord = "commit"
	}
	title := fmt.Sprintf("%s pushed %d %s to %s/%s", username, commitCount, commitWord, activity.Repo, activity.Branch)

	// Create HTML description with commit details (same format as consolidated)
	var htmlParts []string
	htmlParts = append(htmlParts, "<div>")

	for _, commit := range activity.Commits {
		commitHTML := fmt.Sprintf(
			"<div style='margin-bottom: 12px;'>"+
				"<tt><a href='%s'>%s</a></tt>: %s"+
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

	// Create individual push item
	individualItem := &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            activity.CompareLink,
		Published:       activity.LatestTime.Format(time.RFC3339),
		PublishedParsed: activity.LatestTime,
		Updated:         activity.LatestTime.Format(time.RFC3339),
		UpdatedParsed:   activity.LatestTime,
		GUID:            fmt.Sprintf("individual-%s-%s-%d", activity.Repo, activity.Branch, activity.LatestTime.Unix()),
	}

	return individualItem
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
func simplifyNonCommitItem(item *gofeed.Item, username string) *gofeed.Item {
	activityType := detectActivityType(item)

	switch activityType {
	case ActivityPullRequest:
		return simplifyPullRequest(item, username)
	case ActivityFork:
		return simplifyFork(item, username)
	case ActivityBranchCreate:
		return simplifyBranchCreate(item, username)
	case ActivityBranchDelete:
		return simplifyBranchDelete(item, username)
	case ActivityTagDelete:
		return simplifyTagDelete(item, username)
	default:
		// For other activities, create a basic simplified version
		return simplifyOtherActivity(item, username)
	}
}

// simplifyPullRequest creates a clean, simple pull request entry
func simplifyPullRequest(item *gofeed.Item, username string) *gofeed.Item {
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
	title := fmt.Sprintf("%s opened PR #%s in %s", username, prNumber, targetRepo)
	if prTitle != "" {
		title += ": " + prTitle
	}

	// Create simplified HTML content
	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`<a href='%s'>View PR <tt>#%s</tt></a>`, item.Link, prNumber)
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
func simplifyFork(item *gofeed.Item, username string) *gofeed.Item {
	// Extract source and target repos from title or content
	sourceRepo := ""
	targetRepo := ""

	if item.Title != "" {
		// Example: "username forked username/gofeed from mmcdole/gofeed"
		forkRegex := regexp.MustCompile(`forked ([^/]+/[^/\s]+) from ([^/]+/[^/\s]+)`)
		matches := forkRegex.FindStringSubmatch(item.Title)
		if len(matches) > 2 {
			targetRepo = matches[1]
			sourceRepo = matches[2]
		}
	}

	title := fmt.Sprintf("%s forked %s", username, sourceRepo)

	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`<a href='%s'>View fork: <tt>%s</tt></a>`, item.Link, targetRepo)
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
func simplifyBranchCreate(item *gofeed.Item, username string) *gofeed.Item {
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

	title := fmt.Sprintf("%s created branch %s", username, branchName)
	if repoName != "" {
		title += fmt.Sprintf(" in %s", repoName)
	}

	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`<a href='%s'>View branch: <tt>%s</tt></a>`, item.Link, branchName)
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
func simplifyBranchDelete(item *gofeed.Item, username string) *gofeed.Item {
	title := fmt.Sprintf("%s deleted a branch", username)

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
func simplifyTagDelete(item *gofeed.Item, username string) *gofeed.Item {
	tagName := ""
	repoName := ""

	// Try to extract from title first
	if item.Title != "" {
		titleRegex := regexp.MustCompile(regexp.QuoteMeta(username) + ` deleted (?:tag )?(.*?)(?:\s|$)`)
		matches := titleRegex.FindStringSubmatch(item.Title)
		if len(matches) > 1 {
			tagName = strings.TrimSpace(matches[1])
		}
	}

	// Try to extract from content
	if item.Content != "" {
		// First, try to extract tag name from branch-name span
		if tagName == "" {
			tagRegex := regexp.MustCompile(`<span class="branch-name">([^<]+)</span>`)
			matches := tagRegex.FindStringSubmatch(item.Content)
			if len(matches) > 1 {
				rawTagName := matches[1]
				// Clean up refs/tags/ prefix if present
				if strings.HasPrefix(rawTagName, "refs/tags/") {
					tagName = strings.TrimPrefix(rawTagName, "refs/tags/")
				} else {
					tagName = rawTagName
				}
			}
		}

		// Then extract repo name from links
		repoRegex := regexp.MustCompile(`<a[^>]*>([^/]+/([^<]+))</a>`)
		matches := repoRegex.FindAllStringSubmatch(item.Content, -1)
		for _, match := range matches {
			if len(match) > 2 && !strings.Contains(match[1], "@") { // Skip user links
				repoName = match[2] // Just the repo name without username
				break
			}
		}
	}

	// Extract repo from link if not found
	if repoName == "" && item.Link != "" {
		repoRegex := regexp.MustCompile(`github\.com/` + regexp.QuoteMeta(username) + `/([^/]+)`)
		matches := repoRegex.FindStringSubmatch(item.Link)
		if len(matches) > 1 {
			repoName = matches[1]
		}
	}

	if tagName == "" {
		tagName = "tag"
	}

	title := fmt.Sprintf("%s deleted tag %s", username, tagName)
	if repoName != "" {
		title += fmt.Sprintf(" in %s", repoName)
	}

	htmlContent := `<div style='margin-bottom: 12px;'>`
	htmlContent += fmt.Sprintf(`Deleted tag: <tt>%s</tt>`, tagName)
	if repoName != "" {
		htmlContent += fmt.Sprintf(` in <tt>%s</tt>`, repoName)
	}
	htmlContent += `</div>`

	// For deleted tags, link to repo homepage instead of the original link
	link := item.Link
	if repoName != "" {
		link = fmt.Sprintf("https://github.com/%s/%s", username, repoName)
	}

	return &gofeed.Item{
		Title:           title,
		Description:     htmlContent,
		Content:         htmlContent,
		Link:            link,
		Published:       item.Published,
		PublishedParsed: item.PublishedParsed,
		Updated:         item.Updated,
		UpdatedParsed:   item.UpdatedParsed,
		Authors:         item.Authors,
		GUID:            item.GUID,
	}
}

// simplifyOtherActivity creates a basic simplified version for unrecognized activities
func simplifyOtherActivity(item *gofeed.Item, username string) *gofeed.Item {
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
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <feed-url>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       %s -help\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  -retitle <title>    Set custom title for the output feed\n")
	fmt.Fprintf(os.Stderr, "  -format <format>    Output format: atom, rss, or json (default: atom)\n")
	fmt.Fprintf(os.Stderr, "  -consolidate-pushes <bool>  Consolidate pushes into single entries (default: true)\n")
}

func printVersion() {
	fmt.Printf("ghfeed version %s\n", version)
}

// printHelp prints detailed help information
func printHelp() {
	fmt.Printf("ghfeed - GitHub Activity Feed Consolidator\n\n")
	fmt.Printf("version %s\n\n", version)

	fmt.Printf("USAGE:\n")
	fmt.Printf("  %s [options] <feed-url>\n\n", os.Args[0])

	fmt.Printf("OPTIONS:\n")
	fmt.Printf("  -retitle <title>    Set custom title for the output feed\n")
	fmt.Printf("  -format <format>    Output format: atom, rss, or json (default: atom)\n")
	fmt.Printf("  -consolidate-pushes <bool>  Consolidate pushes into single entries (default: true)\n\n")

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

	fmt.Printf("EXAMPLES:\n")
	fmt.Printf("  %s https://github.com/username.atom\n", os.Args[0])
	fmt.Printf("  %s -retitle \"My Custom Feed\" https://github.com/username.atom\n", os.Args[0])
	fmt.Printf("  %s -format rss https://github.com/username.atom\n", os.Args[0])
	fmt.Printf("  %s -format json -retitle \"JSON Feed\" https://github.com/username.atom\n\n", os.Args[0])

	fmt.Printf("AUTHOR:\n")
	fmt.Printf("  Chris Dzombak: https://dzombak.com, https://github.com/cdzombak\n\n")

	fmt.Printf("GITHUB:\n")
	fmt.Printf("  https://github.com/cdzombak/ghfeed\n\n")
}
