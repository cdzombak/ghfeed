package main

import (
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

// Test data with actual GitHub feed HTML content from your feed
const (
	// Actual push HTML with commits from your dotfiles repo
	pushHTML = `<div class="git-push js-feed-item-view"><div class="body">
<!-- push -->
<div class="d-flex flex-items-baseline py-4">
  <div class="d-flex flex-column width-full">
    <div class="d-flex flex-items-baseline">
      <div class="color-fg-muted">
        <span class="mr-2"><a class="d-inline-block" href="/cdzombak" rel="noreferrer"><img class="avatar avatar-user" src="https://avatars.githubusercontent.com/u/102904?s=64&amp;v=4" width="32" height="32" alt="@cdzombak"></a></span>
        <a class="Link--primary no-underline wb-break-all" href="/cdzombak" rel="noreferrer">cdzombak</a>

        pushed to
        <a class="branch-name" href="/cdzombak/dotfiles/tree/master" rel="noreferrer">master</a>
        in
        <a class="Link--primary no-underline wb-break-all" href="/cdzombak/dotfiles" rel="noreferrer">cdzombak/dotfiles</a>

        <span>
          · <relative-time tense="past" datetime="2025-09-15T01:28:02Z" data-view-component="true">September 15, 2025 01:28</relative-time>
        </span>
      </div>
    </div>

    <div class="commits pusher-is-only-committer">
      <ul class="list-style-none">
          <li class="d-flex flex-items-baseline">
            <span>
              <a class="d-inline-block" href="/cdzombak" rel="noreferrer"><img class="mr-1 avatar-user" src="https://avatars.githubusercontent.com/u/102904?s=32&amp;v=4" width="16" height="16" alt="@cdzombak"></a>
            </span>
            <code><a class="mr-1" href="/cdzombak/dotfiles/commit/8e9b024bede1064de870417f7e3f7aa876fa3b47" rel="noreferrer">8e9b024</a></code>
            <div class="dashboard-break-word lh-condensed">
              <blockquote>
                remove Instapaper Save app
              </blockquote>
            </div>
          </li>
          <li class="d-flex flex-items-baseline">
            <span>
              <a class="d-inline-block" href="/cdzombak" rel="noreferrer"><img class="mr-1 avatar-user" src="https://avatars.githubusercontent.com/u/102904?s=32&amp;v=4" width="16" height="16" alt="@cdzombak"></a>
            </span>
            <code><a class="mr-1" href="/cdzombak/dotfiles/commit/b19a1b604e77908604438ab33529c6a6a9d7f9d1" rel="noreferrer">b19a1b6</a></code>
            <div class="dashboard-break-word lh-condensed">
              <blockquote>
                fix Red Eye install
              </blockquote>
            </div>
          </li>
      </ul>
    </div>
  </div>
</div>
</div></div>`

	pullRequestHTML = `<div class="git-pull-request js-feed-item-view"><div class="body">
<div class="d-flex flex-items-baseline py-4">
  <div class="d-flex flex-column width-full">
    <div>
      <div class="d-flex flex-items-baseline">
        <div class="color-fg-muted">
          <span class="mr-2"><a class="d-inline-block" href="/cdzombak" rel="noreferrer"><img class="avatar avatar-user" src="https://avatars.githubusercontent.com/u/102904?s=64&amp;v=4" width="32" height="32" alt="@cdzombak"></a></span>
          <a class="Link--primary no-underline wb-break-all" href="/cdzombak" rel="noreferrer">cdzombak</a>
          opened
          <a class="Link--primary no-underline wb-break-all" aria-label="mmcdole/gofeed#264" href="https://github.com/mmcdole/gofeed/pull/264" rel="noreferrer">mmcdole/gofeed#264</a>
          <span>
            · <relative-time tense="past" datetime="2025-09-14T15:58:34-07:00" data-view-component="true">September 14, 2025 15:58</relative-time>
          </span>
        </div>
      </div>
    </div>

    <div class="Box p-3 my-2 color-shadow-medium color-bg-overlay">
      <div class="ml-4">
        <div>
          <span class="f4 lh-condensed text-bold color-fg-default"><a class="color-fg-default text-bold" aria-label="Allow outputting RSS, Atom, and JSON feeds" href="https://github.com/mmcdole/gofeed/pull/264" rel="noreferrer">Allow outputting RSS, Atom, and JSON feeds</a></span>
          <span class="f4 color-fg-muted ml-1">#264</span>
            <div class="lh-condensed mb-2 mt-1">
              <p dir="auto">This PR allows gofeed to convert the universal <code class="notranslate">Feed</code> representation back into RSS, Atom, or JSON feed structures...</p>
            </div>
        </div>

          <div class="diffstat d-inline-block mt-1">
            <span class="color-fg-success">+3,415</span>
            <span class="color-fg-danger">-146</span>
          </div>

      </div>
    </div>
  </div>
</div>
</div></div>`

	forkHTML = `<div class="fork js-feed-item-view"><div class="body">
<div class="d-flex flex-items-baseline py-4">
  <div class="d-flex flex-column width-full">
      <div class="d-flex flex-items-baseline">
        <div class="color-fg-muted">
          <span class="mr-2"><a class="d-inline-block" href="/cdzombak" rel="noreferrer"><img class="avatar avatar-user" src="https://avatars.githubusercontent.com/u/102904?s=64&amp;v=4" width="32" height="32" alt="@cdzombak"></a></span>
          <a class="Link--primary no-underline wb-break-all" href="/cdzombak" rel="noreferrer">cdzombak</a>
          forked
          <a class="Link--primary no-underline wb-break-all" title="mmcdole/gofeed" href="/mmcdole/gofeed" rel="noreferrer">mmcdole/gofeed</a>
          from
          <a class="Link--primary no-underline wb-break-all" href="/cdzombak/gofeed" rel="noreferrer">cdzombak/gofeed</a>
          <span>
            · <relative-time tense="past" datetime="2025-09-14T09:25:35-07:00" data-view-component="true">September 14, 2025 09:25</relative-time>
          </span>
        </div>
      </div>
  </div>
</div>
</div></div>`

	branchCreateHTML = `<div class="git-branch js-feed-item-view"><div class="body">
<div class="d-flex flex-items-baseline py-4">
  <div class="d-flex flex-column width-full">
        <div class="d-flex flex-items-baseline">
          <div class="color-fg-muted">
                <span class="mr-2"><a class="d-inline-block" href="/cdzombak" rel="noreferrer"><img class="avatar avatar-user" src="https://avatars.githubusercontent.com/u/102904?s=64&amp;v=4" width="32" height="32" alt="@cdzombak"></a></span>
              <a class="Link--primary no-underline wb-break-all" href="/cdzombak" rel="noreferrer">cdzombak</a>
              created a
              branch
              in
              <a class="Link--primary no-underline wb-break-all" href="/cdzombak/gofeed" rel="noreferrer">cdzombak/gofeed</a>

            <span>
              · <relative-time tense="past" datetime="2025-09-14T22:52:36Z" data-view-component="true">September 14, 2025 22:52</relative-time>
            </span>
          </div>
        </div>

    <div class="Box p-3 mt-2 color-shadow-medium color-bg-overlay">
      <div>
        <div class="f4 lh-condensed text-bold color-fg-default">
          <div class="d-inline-block">
            <div class="d-flex">
              <div class="mr-2">
                <a class="d-flex" href="/cdzombak" rel="noreferrer"><img src="https://avatars.githubusercontent.com/u/102904?s=40&amp;v=4" alt="@cdzombak" size="20" height="20" width="20" data-view-component="true" class="avatar avatar-small circle"></a>
              </div>
                  <div class="d-inline-block">
                    <a class="css-truncate css-truncate-target branch-name v-align-middle" title="refs/heads/cdz/feed-creation" href="https://github.com/cdzombak/gofeed/tree/refs/heads/cdz/feed-creation" rel="noreferrer">refs/heads/cdz/feed-creation</a> in <a class="Link--primary text-bold no-underline wb-break-all d-inline-block" href="/cdzombak/gofeed" rel="noreferrer">cdzombak/gofeed</a>
                  </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
</div></div>`

	tagDeleteHTML = `<div class="git-branch js-feed-item-view"><div class="body">
<div class="d-flex flex-items-baseline py-4">
  <div class="d-flex flex-column width-full">
    <div class="color-fg-muted">
      <span class="mr-2"><a class="d-inline-block" href="/cdzombak" rel="noreferrer"><img class="avatar avatar-user" src="https://avatars.githubusercontent.com/u/102904?s=64&amp;v=4" width="32" height="32" alt="@cdzombak"></a></span>
      <a class="Link--primary no-underline wb-break-all" href="/cdzombak" rel="noreferrer">cdzombak</a>
      deleted
      tag
      <span class="branch-name">refs/tags/v0.0.6</span>
      in
      <a class="Link--primary no-underline wb-break-all" href="/cdzombak/homebrew-gomod" rel="noreferrer">cdzombak/homebrew-gomod</a>
      <span>
        · <relative-time tense="past" datetime="2025-09-13T21:13:56Z" data-view-component="true">September 13, 2025 21:13</relative-time>
      </span>
    </div>
  </div>
</div>
</div></div>`
)

func TestExtractCommitsFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Commit
	}{
		{
			name:    "Extract commits from push HTML",
			content: pushHTML,
			expected: []Commit{
				{
					Hash:    "b19a1b6",
					Message: "fix Red Eye install",
					Link:    "https://github.com/cdzombak/dotfiles/commit/b19a1b604e77908604438ab33529c6a6a9d7f9d1",
				},
				{
					Hash:    "8e9b024",
					Message: "remove Instapaper Save app",
					Link:    "https://github.com/cdzombak/dotfiles/commit/8e9b024bede1064de870417f7e3f7aa876fa3b47",
				},
			},
		},
		{
			name:     "No commits in PR HTML",
			content:  pullRequestHTML,
			expected: []Commit{},
		},
		{
			name:     "Empty content",
			content:  "",
			expected: []Commit{},
		},
		{
			name: "Single commit",
			content: `<code><a href="/cdzombak/test/commit/abc123def" rel="noreferrer">abc123d</a></code>
			<div class="dashboard-break-word lh-condensed">
				<blockquote>fix bug</blockquote>
			</div>`,
			expected: []Commit{
				{
					Hash:    "abc123d",
					Message: "fix bug",
					Link:    "https://github.com/cdzombak/test/commit/abc123def",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCommitsFromContent(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("extractCommitsFromContent() = %d commits, want %d", len(result), len(tt.expected))
				return
			}

			for i, commit := range result {
				expected := tt.expected[i]
				if commit.Hash != expected.Hash {
					t.Errorf("extractCommitsFromContent()[%d].Hash = %v, want %v", i, commit.Hash, expected.Hash)
				}
				if commit.Message != expected.Message {
					t.Errorf("extractCommitsFromContent()[%d].Message = %v, want %v", i, commit.Message, expected.Message)
				}
				if commit.Link != expected.Link {
					t.Errorf("extractCommitsFromContent()[%d].Link = %v, want %v", i, commit.Link, expected.Link)
				}
			}
		})
	}
}

func TestExtractUsername(t *testing.T) {
	tests := []struct {
		name     string
		feed     *gofeed.Feed
		expected string
	}{
		{
			name: "Extract from feed link with .atom",
			feed: &gofeed.Feed{
				Link: "https://github.com/testuser.atom",
			},
			expected: "testuser",
		},
		{
			name: "Extract from feed link without .atom",
			feed: &gofeed.Feed{
				Link: "https://github.com/anotheruser",
			},
			expected: "anotheruser",
		},
		{
			name: "Extract from item link when feed link missing",
			feed: &gofeed.Feed{
				Link: "",
				Items: []*gofeed.Item{
					{Link: "https://github.com/itemuser/repo/commit/abc123"},
				},
			},
			expected: "itemuser",
		},
		{
			name: "Extract from second item when first is invalid",
			feed: &gofeed.Feed{
				Items: []*gofeed.Item{
					{Link: "https://example.com/invalid"},
					{Link: "https://github.com/validuser/repo/pull/1"},
				},
			},
			expected: "validuser",
		},
		{
			name: "Fallback when no valid links",
			feed: &gofeed.Feed{
				Link: "https://example.com/invalid",
				Items: []*gofeed.Item{
					{Link: "https://example.com/also-invalid"},
				},
			},
			expected: "user",
		},
		{
			name: "Extract cdzombak from feed",
			feed: &gofeed.Feed{
				Link: "https://github.com/cdzombak.atom",
			},
			expected: "cdzombak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractUsername(tt.feed)
			if result != tt.expected {
				t.Errorf("extractUsername() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectActivityType(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected ActivityType
	}{
		{
			name: "Pull request (opened)",
			item: &gofeed.Item{
				Title:   "cdzombak opened a pull request in gofeed",
				Content: pullRequestHTML,
			},
			expected: ActivityPullRequest,
		},
		{
			name: "Pull request (contributed)",
			item: &gofeed.Item{
				Title:   "cdzombak contributed to cdzombak/things2md",
				Content: pullRequestHTML,
			},
			expected: ActivityPullRequest,
		},
		{
			name: "Fork",
			item: &gofeed.Item{
				Title:   "cdzombak forked cdzombak/gofeed from mmcdole/gofeed",
				Content: forkHTML,
			},
			expected: ActivityFork,
		},
		{
			name: "Branch creation",
			item: &gofeed.Item{
				Title:   "cdzombak created a branch",
				Content: branchCreateHTML,
			},
			expected: ActivityBranchCreate,
		},
		{
			name: "Branch deletion",
			item: &gofeed.Item{
				Title:   "cdzombak deleted branch feature-test",
				Content: "",
			},
			expected: ActivityBranchDelete,
		},
		{
			name: "Tag deletion",
			item: &gofeed.Item{
				Title:   "cdzombak deleted",
				Content: tagDeleteHTML,
			},
			expected: ActivityTagDelete,
		},
		{
			name: "Push (should not be detected as these are handled separately)",
			item: &gofeed.Item{
				Title:   "cdzombak pushed dotfiles",
				Content: pushHTML,
			},
			expected: ActivityOther,
		},
		{
			name: "Unknown activity",
			item: &gofeed.Item{
				Title:   "cdzombak starred a repository",
				Content: "",
			},
			expected: ActivityOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectActivityType(tt.item)
			if result != tt.expected {
				t.Errorf("detectActivityType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSimplifyPullRequest(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected struct {
			title        string
			containsURL  string
			containsText string
		}
	}{
		{
			name: "Full PR with title and stats",
			item: &gofeed.Item{
				Title:   "cdzombak opened a pull request in gofeed",
				Content: pullRequestHTML,
				Link:    "https://github.com/mmcdole/gofeed/pull/264",
			},
			expected: struct {
				title        string
				containsURL  string
				containsText string
			}{
				title:        "cdzombak opened PR #264 in mmcdole/gofeed: Allow outputting RSS, Atom, and JSON feeds",
				containsURL:  "https://github.com/mmcdole/gofeed/pull/264",
				containsText: "+3,415 -146",
			},
		},
		{
			name: "PR without diff stats",
			item: &gofeed.Item{
				Title:   "cdzombak opened a pull request in test",
				Content: `<span class="f4 lh-condensed text-bold color-fg-default"><a class="color-fg-default text-bold" href="/test/pull/1">Test PR</a></span>`,
				Link:    "https://github.com/test/repo/pull/1",
			},
			expected: struct {
				title        string
				containsURL  string
				containsText string
			}{
				title:        "cdzombak opened PR #1 in test/repo: Test PR",
				containsURL:  "https://github.com/test/repo/pull/1",
				containsText: "Test PR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simplifyPullRequest(tt.item, "cdzombak")

			if result.Title != tt.expected.title {
				t.Errorf("simplifyPullRequest().Title = %v, want %v", result.Title, tt.expected.title)
			}

			if !strings.Contains(result.Content, tt.expected.containsURL) {
				t.Errorf("simplifyPullRequest().Content should contain URL %v, got %v", tt.expected.containsURL, result.Content)
			}

			if tt.expected.containsText != "" && !strings.Contains(result.Content, tt.expected.containsText) {
				t.Errorf("simplifyPullRequest().Content should contain text %v, got %v", tt.expected.containsText, result.Content)
			}

			if result.Link != tt.item.Link {
				t.Errorf("simplifyPullRequest().Link = %v, want %v", result.Link, tt.item.Link)
			}
		})
	}
}

func TestSimplifyFork(t *testing.T) {
	item := &gofeed.Item{
		Title:   "cdzombak forked cdzombak/gofeed from mmcdole/gofeed",
		Content: forkHTML,
		Link:    "https://github.com/cdzombak/gofeed",
	}

	result := simplifyFork(item, "cdzombak")

	expectedTitle := "cdzombak forked mmcdole/gofeed"
	if result.Title != expectedTitle {
		t.Errorf("simplifyFork().Title = %v, want %v", result.Title, expectedTitle)
	}

	if !strings.Contains(result.Content, "cdzombak/gofeed") {
		t.Errorf("simplifyFork().Content should contain fork target, got %v", result.Content)
	}

	if result.Link != item.Link {
		t.Errorf("simplifyFork().Link = %v, want %v", result.Link, item.Link)
	}
}

func TestSimplifyBranchCreate(t *testing.T) {
	item := &gofeed.Item{
		Title:   "cdzombak created a branch",
		Content: branchCreateHTML,
		Link:    "https://github.com/cdzombak/gofeed/tree/refs/heads/cdz/feed-creation",
	}

	result := simplifyBranchCreate(item, "cdzombak")

	expectedTitle := "cdzombak created branch cdz/feed-creation in cdzombak/gofeed"
	if result.Title != expectedTitle {
		t.Errorf("simplifyBranchCreate().Title = %v, want %v", result.Title, expectedTitle)
	}

	if !strings.Contains(result.Content, "cdz/feed-creation") {
		t.Errorf("simplifyBranchCreate().Content should contain branch name, got %v", result.Content)
	}
}

func TestSimplifyTagDelete(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected struct {
			title   string
			content string
		}
	}{
		{
			name: "Tag deletion with refs/tags/ prefix",
			item: &gofeed.Item{
				Title:   "cdzombak deleted",
				Content: tagDeleteHTML,
				Link:    "https://github.com/cdzombak/homebrew-gomod/compare/2b377a2203...0000000000",
			},
			expected: struct {
				title   string
				content string
			}{
				title:   "cdzombak deleted tag v0.0.6 in homebrew-gomod",
				content: "v0.0.6",
			},
		},
		{
			name: "Tag deletion from title (fallback case)",
			item: &gofeed.Item{
				Title:   "cdzombak deleted refs/tags/v1.2.3",
				Content: "",
				Link:    "https://github.com/cdzombak/test/compare/abc...def",
			},
			expected: struct {
				title   string
				content string
			}{
				title:   "cdzombak deleted tag refs/tags/v1.2.3 in test",
				content: "refs/tags/v1.2.3",
			},
		},
		{
			name: "Tag deletion with clean tag name",
			item: &gofeed.Item{
				Title:   "cdzombak deleted",
				Content: `<span class="branch-name">refs/tags/v2.1.0</span> in <a href="/cdzombak/project">cdzombak/project</a>`,
				Link:    "https://github.com/cdzombak/project/compare/abc...def",
			},
			expected: struct {
				title   string
				content string
			}{
				title:   "cdzombak deleted tag v2.1.0 in project",
				content: "v2.1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simplifyTagDelete(tt.item, "cdzombak")

			if result.Title != tt.expected.title {
				t.Errorf("simplifyTagDelete().Title = %v, want %v", result.Title, tt.expected.title)
			}

			if !strings.Contains(result.Content, tt.expected.content) {
				t.Errorf("simplifyTagDelete().Content should contain %v, got %v", tt.expected.content, result.Content)
			}
		})
	}
}

func TestExtractBranchActivity(t *testing.T) {
	publishedTime, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")

	tests := []struct {
		name     string
		item     *gofeed.Item
		expected *BranchActivity
	}{
		{
			name: "Push with commits",
			item: &gofeed.Item{
				Title:           "cdzombak pushed dotfiles",
				Content:         pushHTML,
				Link:            "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
				PublishedParsed: &publishedTime,
			},
			expected: &BranchActivity{
				Repo:   "dotfiles",
				Branch: "master",
				Commits: []Commit{
					{
						Hash:    "b19a1b6",
						Message: "fix Red Eye install",
						Link:    "https://github.com/cdzombak/dotfiles/commit/b19a1b604e77908604438ab33529c6a6a9d7f9d1",
					},
					{
						Hash:    "8e9b024",
						Message: "remove Instapaper Save app",
						Link:    "https://github.com/cdzombak/dotfiles/commit/8e9b024bede1064de870417f7e3f7aa876fa3b47",
					},
				},
				LatestTime:  &publishedTime,
				CompareLink: "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
			},
		},
		{
			name: "Invalid repo link",
			item: &gofeed.Item{
				Title:   "cdzombak pushed something",
				Content: pushHTML,
				Link:    "https://example.com/invalid",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBranchActivity(tt.item, "cdzombak")

			if tt.expected == nil {
				if result != nil {
					t.Errorf("extractBranchActivity() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("extractBranchActivity() = nil, want non-nil")
				return
			}

			if result.Repo != tt.expected.Repo {
				t.Errorf("extractBranchActivity().Repo = %v, want %v", result.Repo, tt.expected.Repo)
			}

			if result.Branch != tt.expected.Branch {
				t.Errorf("extractBranchActivity().Branch = %v, want %v", result.Branch, tt.expected.Branch)
			}

			if len(result.Commits) != len(tt.expected.Commits) {
				t.Errorf("extractBranchActivity().Commits length = %d, want %d", len(result.Commits), len(tt.expected.Commits))
				return
			}

			for i, commit := range result.Commits {
				expected := tt.expected.Commits[i]
				if commit.Hash != expected.Hash {
					t.Errorf("extractBranchActivity().Commits[%d].Hash = %v, want %v", i, commit.Hash, expected.Hash)
				}
				if commit.Message != expected.Message {
					t.Errorf("extractBranchActivity().Commits[%d].Message = %v, want %v", i, commit.Message, expected.Message)
				}
				if commit.Link != expected.Link {
					t.Errorf("extractBranchActivity().Commits[%d].Link = %v, want %v", i, commit.Link, expected.Link)
				}
			}
		})
	}
}

func TestCreateConsolidatedBranchItem(t *testing.T) {
	publishedTime, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")

	activity := &BranchActivity{
		Repo:   "dotfiles",
		Branch: "master",
		Commits: []Commit{
			{
				Hash:    "8e9b024",
				Message: "remove Instapaper Save app",
				Link:    "https://github.com/cdzombak/dotfiles/commit/8e9b024bede1064de870417f7e3f7aa876fa3b47",
			},
			{
				Hash:    "b19a1b6",
				Message: "fix Red Eye install",
				Link:    "https://github.com/cdzombak/dotfiles/commit/b19a1b604e77908604438ab33529c6a6a9d7f9d1",
			},
		},
		LatestTime:  &publishedTime,
		CompareLink: "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
	}

	result := createConsolidatedBranchItem(activity, "cdzombak")

	expectedTitle := "cdzombak pushed 2 commits to dotfiles/master"
	if result.Title != expectedTitle {
		t.Errorf("createConsolidatedBranchItem().Title = %v, want %v", result.Title, expectedTitle)
	}

	// Check that both commits are in the content
	if !strings.Contains(result.Content, "8e9b024") || !strings.Contains(result.Content, "remove Instapaper Save app") {
		t.Errorf("createConsolidatedBranchItem().Content missing first commit")
	}

	if !strings.Contains(result.Content, "b19a1b6") || !strings.Contains(result.Content, "fix Red Eye install") {
		t.Errorf("createConsolidatedBranchItem().Content missing second commit")
	}

	// Check for compare link
	if !strings.Contains(result.Content, "View all changes") {
		t.Errorf("createConsolidatedBranchItem().Content missing compare link")
	}

	if result.Link != activity.CompareLink {
		t.Errorf("createConsolidatedBranchItem().Link = %v, want %v", result.Link, activity.CompareLink)
	}

	if result.PublishedParsed != activity.LatestTime {
		t.Errorf("createConsolidatedBranchItem().PublishedParsed = %v, want %v", result.PublishedParsed, activity.LatestTime)
	}
}

func TestCreateConsolidatedBranchItemSingleCommit(t *testing.T) {
	publishedTime, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")

	activity := &BranchActivity{
		Repo:   "dotfiles",
		Branch: "master",
		Commits: []Commit{
			{
				Hash:    "8e9b024",
				Message: "remove Instapaper Save app",
				Link:    "https://github.com/cdzombak/dotfiles/commit/8e9b024bede1064de870417f7e3f7aa876fa3b47",
			},
		},
		LatestTime:  &publishedTime,
		CompareLink: "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
	}

	result := createConsolidatedBranchItem(activity, "cdzombak")

	expectedTitle := "cdzombak pushed 1 commit to dotfiles/master"
	if result.Title != expectedTitle {
		t.Errorf("createConsolidatedBranchItem().Title = %v, want %v", result.Title, expectedTitle)
	}

	// Check that the commit is in the content
	if !strings.Contains(result.Content, "8e9b024") || !strings.Contains(result.Content, "remove Instapaper Save app") {
		t.Errorf("createConsolidatedBranchItem().Content missing commit")
	}
}

func TestCreateConsolidatedBranchItemNoCommits(t *testing.T) {
	activity := &BranchActivity{
		Repo:    "test",
		Branch:  "main",
		Commits: []Commit{},
	}

	result := createConsolidatedBranchItem(activity, "cdzombak")

	if result != nil {
		t.Errorf("createConsolidatedBranchItem() with no commits = %v, want nil", result)
	}
}

func TestIsCommitOrPush(t *testing.T) {
	tests := []struct {
		title    string
		expected bool
	}{
		{"cdzombak pushed dotfiles", true},
		{"cdzombak created branch feature-test", true},
		{"cdzombak deleted branch old-feature", true},
		{"cdzombak created tag v1.0.0", true},
		{"cdzombak deleted tag v0.9.0", true},
		{"cdzombak opened a pull request", false},
		{"cdzombak forked repository", false},
		{"cdzombak starred repository", false},
		{"cdzombak contributed to project", false},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := isCommitOrPush(tt.title)
			if result != tt.expected {
				t.Errorf("isCommitOrPush(%v) = %v, want %v", tt.title, result, tt.expected)
			}
		})
	}
}

// Integration test with realistic feed data
func TestConsolidateCommitsIntegration(t *testing.T) {
	publishedTime1, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")
	publishedTime2, _ := time.Parse(time.RFC3339, "2025-09-14T23:03:04Z")
	publishedTime3, _ := time.Parse(time.RFC3339, "2025-09-14T15:58:34-07:00")

	// Create a mock feed with various activity types from your actual feed
	inputFeed := &gofeed.Feed{
		Title:       "GitHub Public Timeline Feed",
		Description: "GitHub activities for cdzombak",
		Items: []*gofeed.Item{
			// Push to dotfiles/master
			{
				Title:           "cdzombak pushed dotfiles",
				Content:         pushHTML,
				Link:            "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
				PublishedParsed: &publishedTime1,
				GUID:            "push-dotfiles-1",
			},
			// Push to gofeed/cdz/feed-creation
			{
				Title: "cdzombak pushed gofeed",
				Content: `<div class="git-push js-feed-item-view"><div class="body">
				<a class="branch-name" href="/cdzombak/gofeed/tree/cdz/feed-creation" rel="noreferrer">cdz/feed-creation</a>
				<code><a href="/cdzombak/gofeed/commit/21507eb3406370a5fe2ccdb809456058e3d0dd9f" rel="noreferrer">21507eb</a></code>
				<div class="dashboard-break-word lh-condensed">
					<blockquote>add tests for feed.Render* methods</blockquote>
				</div>
				</div></div>`,
				Link:            "https://github.com/cdzombak/gofeed/compare/6958a905fb...21507eb340",
				PublishedParsed: &publishedTime2,
				GUID:            "push-gofeed-1",
			},
			// Pull request
			{
				Title:           "cdzombak opened a pull request in gofeed",
				Content:         pullRequestHTML,
				Link:            "https://github.com/mmcdole/gofeed/pull/264",
				PublishedParsed: &publishedTime3,
				GUID:            "pr-gofeed-264",
			},
			// Fork
			{
				Title:           "cdzombak forked cdzombak/gofeed from mmcdole/gofeed",
				Content:         forkHTML,
				Link:            "https://github.com/cdzombak/gofeed",
				PublishedParsed: &publishedTime3,
				GUID:            "fork-gofeed-1",
			},
		},
	}

	result := consolidateCommits(inputFeed, "", true)

	// Verify feed metadata is preserved
	if result.Title != inputFeed.Title {
		t.Errorf("consolidateCommits() title = %v, want %v", result.Title, inputFeed.Title)
	}

	if result.Description != inputFeed.Description {
		t.Errorf("consolidateCommits() description = %v, want %v", result.Description, inputFeed.Description)
	}

	// Should have 4 items: 2 consolidated pushes + 1 PR + 1 fork
	if len(result.Items) != 4 {
		t.Errorf("consolidateCommits() items count = %d, want 4", len(result.Items))
		for i, item := range result.Items {
			t.Logf("Item %d: %s", i, item.Title)
		}
		return
	}

	// Find consolidated dotfiles entry
	var dotfilesItem *gofeed.Item
	for _, item := range result.Items {
		if strings.Contains(item.Title, "dotfiles/master") {
			dotfilesItem = item
			break
		}
	}

	if dotfilesItem == nil {
		t.Fatal("consolidateCommits() missing consolidated dotfiles entry")
	}

	// Verify consolidated entry structure
	expectedTitle := "cdzombak pushed 2 commits to dotfiles/master"
	if dotfilesItem.Title != expectedTitle {
		t.Errorf("consolidated dotfiles title = %v, want %v", dotfilesItem.Title, expectedTitle)
	}

	// Verify commits are included in content
	if !strings.Contains(dotfilesItem.Content, "8e9b024") ||
		!strings.Contains(dotfilesItem.Content, "remove Instapaper Save app") {
		t.Error("consolidated dotfiles missing first commit")
	}

	if !strings.Contains(dotfilesItem.Content, "b19a1b6") ||
		!strings.Contains(dotfilesItem.Content, "fix Red Eye install") {
		t.Error("consolidated dotfiles missing second commit")
	}

	// Find simplified PR entry
	var prItem *gofeed.Item
	for _, item := range result.Items {
		if strings.Contains(item.Title, "opened PR #264") {
			prItem = item
			break
		}
	}

	if prItem == nil {
		t.Fatal("consolidateCommits() missing simplified PR entry")
	}

	expectedPRTitle := "cdzombak opened PR #264 in mmcdole/gofeed: Allow outputting RSS, Atom, and JSON feeds"
	if prItem.Title != expectedPRTitle {
		t.Errorf("simplified PR title = %v, want %v", prItem.Title, expectedPRTitle)
	}

	// Verify PR content is simplified and contains link
	if !strings.Contains(prItem.Content, "View PR <tt>#264</tt>") {
		t.Error("simplified PR missing view link")
	}

	// Find simplified fork entry
	var forkItem *gofeed.Item
	for _, item := range result.Items {
		if strings.Contains(item.Title, "forked mmcdole/gofeed") {
			forkItem = item
			break
		}
	}

	if forkItem == nil {
		t.Fatal("consolidateCommits() missing simplified fork entry")
	}

	expectedForkTitle := "cdzombak forked mmcdole/gofeed"
	if forkItem.Title != expectedForkTitle {
		t.Errorf("simplified fork title = %v, want %v", forkItem.Title, expectedForkTitle)
	}
}

// Test branch activity merging (multiple pushes to same branch)
func TestBranchActivityMerging(t *testing.T) {
	publishedTime1, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")
	publishedTime2, _ := time.Parse(time.RFC3339, "2025-09-15T01:30:00Z")

	inputFeed := &gofeed.Feed{
		Title: "Test Feed",
		Items: []*gofeed.Item{
			// First push to dotfiles/master
			{
				Title: "cdzombak pushed dotfiles",
				Content: `<div class="git-push js-feed-item-view">
					<a class="branch-name" href="/cdzombak/dotfiles/tree/master" rel="noreferrer">master</a>
					<code><a href="/cdzombak/dotfiles/commit/abc123" rel="noreferrer">abc123</a></code>
					<div class="dashboard-break-word lh-condensed">
						<blockquote>first commit</blockquote>
					</div>
					</div>`,
				Link:            "https://github.com/cdzombak/dotfiles/compare/xyz...abc123",
				PublishedParsed: &publishedTime1,
				GUID:            "push-dotfiles-1",
			},
			// Second push to same branch
			{
				Title: "cdzombak pushed dotfiles",
				Content: `<div class="git-push js-feed-item-view">
					<a class="branch-name" href="/cdzombak/dotfiles/tree/master" rel="noreferrer">master</a>
					<code><a href="/cdzombak/dotfiles/commit/def456" rel="noreferrer">def456</a></code>
					<div class="dashboard-break-word lh-condensed">
						<blockquote>second commit</blockquote>
					</div>
					</div>`,
				Link:            "https://github.com/cdzombak/dotfiles/compare/abc123...def456",
				PublishedParsed: &publishedTime2,
				GUID:            "push-dotfiles-2",
			},
		},
	}

	result := consolidateCommits(inputFeed, "", true)

	// Should have 1 consolidated item (both pushes merged)
	if len(result.Items) != 1 {
		t.Errorf("consolidateCommits() items count = %d, want 1", len(result.Items))
		return
	}

	item := result.Items[0]

	// Should show total commit count
	expectedTitle := "cdzombak pushed 2 commits to dotfiles/master"
	if item.Title != expectedTitle {
		t.Errorf("merged branch title = %v, want %v", item.Title, expectedTitle)
	}

	// Should contain both commits
	if !strings.Contains(item.Content, "abc123") || !strings.Contains(item.Content, "first commit") {
		t.Error("merged branch missing first commit")
	}

	if !strings.Contains(item.Content, "def456") || !strings.Contains(item.Content, "second commit") {
		t.Error("merged branch missing second commit")
	}

	// Should use the latest timestamp
	if item.PublishedParsed.Before(publishedTime2) {
		t.Error("merged branch should use latest timestamp")
	}
}

// Test that the code works for different GitHub usernames
func TestDifferentUsernames(t *testing.T) {
	testUsers := []string{"alice", "bob123", "cool-user", "testuser"}

	for _, username := range testUsers {
		t.Run("username_"+username, func(t *testing.T) {
			// Test with custom HTML content for this user
			pushContent := strings.ReplaceAll(pushHTML, "cdzombak", username)
			prContent := strings.ReplaceAll(pullRequestHTML, "cdzombak", username)

			publishedTime, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")

			// Create feed for this user
			inputFeed := &gofeed.Feed{
				Title: "GitHub Activities for " + username,
				Link:  "https://github.com/" + username + ".atom",
				Items: []*gofeed.Item{
					{
						Title:           username + " pushed dotfiles",
						Content:         pushContent,
						Link:            "https://github.com/" + username + "/dotfiles/compare/b19a1b604e...8e9b024bed",
						PublishedParsed: &publishedTime,
						GUID:            "push-1",
					},
					{
						Title:           username + " opened a pull request in gofeed",
						Content:         prContent,
						Link:            "https://github.com/mmcdole/gofeed/pull/264",
						PublishedParsed: &publishedTime,
						GUID:            "pr-1",
					},
				},
			}

			result := consolidateCommits(inputFeed, "", true)

			// Verify username extraction worked
			extractedUsername := extractUsername(inputFeed)
			if extractedUsername != username {
				t.Errorf("extractUsername() = %v, want %v", extractedUsername, username)
			}

			// Should have 2 items: 1 consolidated push + 1 simplified PR
			if len(result.Items) != 2 {
				t.Errorf("consolidateCommits() items count = %d, want 2", len(result.Items))
				return
			}

			// Find consolidated push entry
			var pushItem *gofeed.Item
			for _, item := range result.Items {
				if strings.Contains(item.Title, "pushed") && strings.Contains(item.Title, "dotfiles") {
					pushItem = item
					break
				}
			}

			if pushItem == nil {
				t.Fatal("missing consolidated push entry for " + username)
			}

			// Verify title uses correct username
			expectedPushTitle := username + " pushed 2 commits to dotfiles/master"
			if pushItem.Title != expectedPushTitle {
				t.Errorf("push title = %v, want %v", pushItem.Title, expectedPushTitle)
			}

			// Find simplified PR entry
			var prItem *gofeed.Item
			for _, item := range result.Items {
				if strings.Contains(item.Title, "opened PR") {
					prItem = item
					break
				}
			}

			if prItem == nil {
				t.Fatal("missing simplified PR entry for " + username)
			}

			// Verify PR title uses correct username
			expectedPRTitle := username + " opened PR #264 in mmcdole/gofeed: Allow outputting RSS, Atom, and JSON feeds"
			if prItem.Title != expectedPRTitle {
				t.Errorf("PR title = %v, want %v", prItem.Title, expectedPRTitle)
			}
		})
	}
}

// Test individual simplify functions with different usernames
func TestSimplifyFunctionsWithDifferentUsers(t *testing.T) {
	testUsers := []string{"alice", "bob123", "test-user"}

	for _, username := range testUsers {
		t.Run("user_"+username, func(t *testing.T) {
			// Test pull request simplification
			prItem := &gofeed.Item{
				Title:   username + " opened a pull request in test/repo",
				Content: `<span class="f4 lh-condensed text-bold color-fg-default"><a class="color-fg-default text-bold" href="/test/repo/pull/123">Test PR Title</a></span>`,
				Link:    "https://github.com/test/repo/pull/123",
			}

			result := simplifyPullRequest(prItem, username)
			expectedTitle := username + " opened PR #123 in test/repo: Test PR Title"
			if result.Title != expectedTitle {
				t.Errorf("simplifyPullRequest title = %v, want %v", result.Title, expectedTitle)
			}

			// Test fork simplification
			forkItem := &gofeed.Item{
				Title: username + " forked " + username + "/test from original/test",
				Link:  "https://github.com/" + username + "/test",
			}

			forkResult := simplifyFork(forkItem, username)
			expectedForkTitle := username + " forked original/test"
			if forkResult.Title != expectedForkTitle {
				t.Errorf("simplifyFork title = %v, want %v", forkResult.Title, expectedForkTitle)
			}

			// Test branch creation
			branchItem := &gofeed.Item{
				Title:   username + " created a branch",
				Content: `<a class="css-truncate css-truncate-target branch-name" title="refs/heads/feature" href="https://github.com/` + username + `/repo/tree/refs/heads/feature">refs/heads/feature</a> in <a href="/` + username + `/repo">` + username + `/repo</a>`,
				Link:    "https://github.com/" + username + "/repo/tree/refs/heads/feature",
			}

			branchResult := simplifyBranchCreate(branchItem, username)
			expectedBranchTitle := username + " created branch feature in " + username + "/repo"
			if branchResult.Title != expectedBranchTitle {
				t.Errorf("simplifyBranchCreate title = %v, want %v", branchResult.Title, expectedBranchTitle)
			}

			// Test tag deletion
			tagItem := &gofeed.Item{
				Title:   username + " deleted tag refs/tags/v1.0.0",
				Content: `<span class="branch-name">refs/tags/v1.0.0</span> in <a href="/` + username + `/repo">` + username + `/repo</a>`,
				Link:    "https://github.com/" + username + "/repo/compare/abc123...000000",
			}

			tagResult := simplifyTagDelete(tagItem, username)
			expectedTagTitle := username + " deleted tag refs/tags/v1.0.0 in repo"
			if tagResult.Title != expectedTagTitle {
				t.Errorf("simplifyTagDelete title = %v, want %v", tagResult.Title, expectedTagTitle)
			}
		})
	}
}

// Test edge cases for HTML parsing
func TestHTMLParsingEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int // expected number of commits
	}{
		{
			name: "commits with extra whitespace",
			content: `
			<code><a href="/test/commit/abc123" rel="noreferrer">abc123</a></code>

			<div class="dashboard-break-word lh-condensed">
				<blockquote>

					commit with extra whitespace

				</blockquote>
			</div>`,
			expected: 1,
		},
		{
			name: "malformed HTML",
			content: `<code><a href="/test/commit/abc123">abc123</a></code>
			<blockquote>no closing div</blockquote>`,
			expected: 1, // Fallback parsing should still extract the commit link
		},
		{
			name: "commit with special characters in message",
			content: `<code><a href="/test/commit/abc123" rel="noreferrer">abc123</a></code>
			<div class="dashboard-break-word lh-condensed">
				<blockquote>fix: handle "quotes" & ampersands</blockquote>
			</div>`,
			expected: 1,
		},
		{
			name:     "completely empty content",
			content:  "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits := extractCommitsFromContent(tt.content)
			if len(commits) != tt.expected {
				t.Errorf("extractCommitsFromContent() = %d commits, want %d", len(commits), tt.expected)
			}

			// If we expect commits, verify they have proper content
			if tt.expected > 0 && len(commits) > 0 {
				commit := commits[0]
				if commit.Hash == "" {
					t.Error("commit hash should not be empty")
				}
				if commit.Link == "" {
					t.Error("commit link should not be empty")
				}
				if commit.Message == "" {
					t.Error("commit message should not be empty")
				}
			}
		})
	}
}

func TestGenerateComparisonLink(t *testing.T) {
	tests := []struct {
		name     string
		activity *BranchActivity
		username string
		expected string
	}{
		{
			name: "single commit links directly to commit",
			activity: &BranchActivity{
				Repo:   "test-repo",
				Branch: "main",
				Commits: []Commit{
					{
						Hash:    "abc123def456", // matches URL hash
						Message: "Test commit",
						Link:    "https://github.com/testuser/test-repo/commit/abc123def456",
					},
				},
				CompareLink: "https://github.com/testuser/test-repo/compare/abc123def456",
			},
			username: "testuser",
			expected: "https://github.com/testuser/test-repo/commit/abc123def456",
		},
		{
			name: "multiple commits creates comparison link",
			activity: &BranchActivity{
				Repo:   "test-repo",
				Branch: "main",
				Commits: []Commit{
					{
						Hash:    "def456abc789", // newest first - matches URL hash (all hex chars)
						Message: "Second commit",
						Link:    "https://github.com/testuser/test-repo/commit/def456abc789",
					},
					{
						Hash:    "abc123def456", // oldest last - matches URL hash
						Message: "First commit",
						Link:    "https://github.com/testuser/test-repo/commit/abc123def456",
					},
				},
				CompareLink: "https://github.com/testuser/test-repo/compare/abc123def456...def456abc789",
			},
			username: "testuser",
			expected: "https://github.com/testuser/test-repo/compare/abc123def456^...def456abc789",
		},
		{
			name: "empty commits falls back to original",
			activity: &BranchActivity{
				Repo:        "test-repo",
				Branch:      "main",
				Commits:     []Commit{},
				CompareLink: "https://github.com/testuser/test-repo/compare/original",
			},
			username: "testuser",
			expected: "https://github.com/testuser/test-repo/compare/original",
		},
		{
			name: "invalid commit links falls back to newest",
			activity: &BranchActivity{
				Repo:   "test-repo",
				Branch: "main",
				Commits: []Commit{
					{
						Hash:    "def456abc789", // matches URL hash (all hex chars)
						Message: "Valid commit",
						Link:    "https://github.com/testuser/test-repo/commit/def456abc789",
					},
					{
						Hash:    "abc123",
						Message: "Invalid link",
						Link:    "invalid-url",
					},
				},
				CompareLink: "https://github.com/testuser/test-repo/compare/original",
			},
			username: "testuser",
			expected: "https://github.com/testuser/test-repo/commit/def456abc789",
		},
		{
			name: "same commit hash for multiple commits",
			activity: &BranchActivity{
				Repo:   "test-repo",
				Branch: "main",
				Commits: []Commit{
					{
						Hash:    "abc123def456", // matches URL hash
						Message: "Same commit",
						Link:    "https://github.com/testuser/test-repo/commit/abc123def456",
					},
					{
						Hash:    "abc123def456", // matches URL hash
						Message: "Same commit duplicate",
						Link:    "https://github.com/testuser/test-repo/commit/abc123def456",
					},
				},
				CompareLink: "https://github.com/testuser/test-repo/compare/original",
			},
			username: "testuser",
			expected: "https://github.com/testuser/test-repo/commit/abc123def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateComparisonLink(tt.activity, tt.username)
			if result != tt.expected {
				t.Errorf("generateComparisonLink() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConsolidateCommitsWithConsolidationDisabled(t *testing.T) {
	publishedTime1, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")
	publishedTime2, _ := time.Parse(time.RFC3339, "2025-09-14T23:03:04Z")

	// Create a mock feed with multiple pushes to the same repo/branch
	inputFeed := &gofeed.Feed{
		Title:       "GitHub Public Timeline Feed",
		Description: "GitHub activities for cdzombak",
		Items: []*gofeed.Item{
			// First push to dotfiles/master
			{
				Title:           "cdzombak pushed dotfiles",
				Content:         pushHTML,
				Link:            "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
				PublishedParsed: &publishedTime1,
				GUID:            "push-dotfiles-1",
			},
			// Second push to dotfiles/master (would normally be consolidated)
			{
				Title: "cdzombak pushed dotfiles",
				Content: `<div class="git-push js-feed-item-view"><div class="body">
				<a class="branch-name" href="/cdzombak/dotfiles/tree/master" rel="noreferrer">master</a>
				<code><a href="/cdzombak/dotfiles/commit/abc123def456" rel="noreferrer">abc123d</a></code>
				<div class="dashboard-break-word lh-condensed">
					<blockquote>another commit message</blockquote>
				</div>
				</div></div>`,
				Link:            "https://github.com/cdzombak/dotfiles/compare/abc123def456...def456abc123",
				PublishedParsed: &publishedTime2,
				GUID:            "push-dotfiles-2",
			},
		},
	}

	// Test with consolidation disabled
	result := consolidateCommits(inputFeed, "", false)

	// Should have 2 separate push items (not consolidated)
	if len(result.Items) != 2 {
		t.Errorf("consolidateCommits(consolidate=false) items count = %d, want 2", len(result.Items))
		for i, item := range result.Items {
			t.Logf("Item %d: %s", i, item.Title)
		}
		return
	}

	for i, item := range result.Items {
		if !strings.Contains(item.Title, "pushed") || !strings.Contains(item.Title, "dotfiles/master") {
			t.Errorf("Item %d title = %v, should be individual push entry", i, item.Title)
		}

		// Should have individual GUID format
		if !strings.HasPrefix(item.GUID, "individual-") {
			t.Errorf("Item %d GUID = %v, should start with 'individual-'", i, item.GUID)
		}
	}
}

func TestConsolidateCommitsWithConsolidationEnabled(t *testing.T) {
	publishedTime1, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")
	publishedTime2, _ := time.Parse(time.RFC3339, "2025-09-14T23:03:04Z")

	inputFeed := &gofeed.Feed{
		Title:       "GitHub Public Timeline Feed",
		Description: "GitHub activities for cdzombak",
		Items: []*gofeed.Item{
			{
				Title:           "cdzombak pushed dotfiles",
				Content:         pushHTML,
				Link:            "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
				PublishedParsed: &publishedTime1,
				GUID:            "push-dotfiles-1",
			},
			{
				Title: "cdzombak pushed dotfiles",
				Content: `<div class="git-push js-feed-item-view"><div class="body">
				<a class="branch-name" href="/cdzombak/dotfiles/tree/master" rel="noreferrer">master</a>
				<code><a href="/cdzombak/dotfiles/commit/abc123def456" rel="noreferrer">abc123d</a></code>
				<div class="dashboard-break-word lh-condensed">
					<blockquote>another commit message</blockquote>
				</div>
				</div></div>`,
				Link:            "https://github.com/cdzombak/dotfiles/compare/abc123def456...def456abc123",
				PublishedParsed: &publishedTime2,
				GUID:            "push-dotfiles-2",
			},
		},
	}

	// Test with consolidation enabled (default behavior)
	result := consolidateCommits(inputFeed, "", true)

	// Should have 1 consolidated item
	if len(result.Items) != 1 {
		t.Errorf("consolidateCommits(consolidate=true) items count = %d, want 1", len(result.Items))
		return
	}

	// Should be consolidated entry
	item := result.Items[0]
	if !strings.Contains(item.Title, "pushed") || !strings.Contains(item.Title, "commits to dotfiles/master") {
		t.Errorf("Item title = %v, should be consolidated push entry", item.Title)
	}

	// Should have consolidated GUID format
	if !strings.HasPrefix(item.GUID, "consolidated-") {
		t.Errorf("Item GUID = %v, should start with 'consolidated-'", item.GUID)
	}
}

func TestCommitOrderingWithConsolidationDisabled(t *testing.T) {
	publishedTime1, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")
	publishedTime2, _ := time.Parse(time.RFC3339, "2025-09-15T01:30:00Z")

	// Test that commits within individual push items are ordered newest-first
	// when consolidation is disabled (each push becomes its own feed item)
	items := []*gofeed.Item{
		{
			Title:           "cdzombak pushed dotfiles",
			Content:         pushHTML,
			Link:            "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed",
			PublishedParsed: &publishedTime1,
		},
		{
			Title:           "cdzombak pushed another-repo",
			Content:         `<code><a href="/cdzombak/another-repo/commit/abc123def456" rel="noreferrer">abc123d</a></code>
			<div class="dashboard-break-word lh-condensed">
				<blockquote>older commit message</blockquote>
			</div>
			<code><a href="/cdzombak/another-repo/commit/def456abc789" rel="noreferrer">def456a</a></code>
			<div class="dashboard-break-word lh-condensed">
				<blockquote>newer commit message</blockquote>
			</div>`,
			Link:            "https://github.com/cdzombak/another-repo/compare/abc123def456...def456abc789",
			PublishedParsed: &publishedTime2,
		},
	}

	username := "cdzombak"

	var result []*gofeed.Item
	var activities []*BranchActivity

	for _, item := range items {
		if isCommitOrPush(item.Title) {
			activity := extractBranchActivity(item, username)
			activities = append(activities, activity)
			result = append(result, createIndividualPushItem(activity, username))
		}
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}

	// Verify first push item has commits in newest-first order
	activity1 := activities[0]
	if len(activity1.Commits) != 2 {
		t.Fatalf("Expected 2 commits in first item, got %d", len(activity1.Commits))
	}
	// Based on our test data, b19a1b6 should be first (newest) and 8e9b024 should be second (oldest)
	if activity1.Commits[0].Hash != "b19a1b6" {
		t.Errorf("Expected first commit to be b19a1b6 (newest), got %s", activity1.Commits[0].Hash)
	}
	if activity1.Commits[1].Hash != "8e9b024" {
		t.Errorf("Expected second commit to be 8e9b024 (oldest), got %s", activity1.Commits[1].Hash)
	}

	// Verify second push item has commits in newest-first order
	activity2 := activities[1]
	if len(activity2.Commits) != 2 {
		t.Fatalf("Expected 2 commits in second item, got %d", len(activity2.Commits))
	}
	// def456a should be first (newest) and abc123d should be second (oldest)
	if activity2.Commits[0].Hash != "def456a" {
		t.Errorf("Expected first commit to be def456a (newest), got %s", activity2.Commits[0].Hash)
	}
	if activity2.Commits[1].Hash != "abc123d" {
		t.Errorf("Expected second commit to be abc123d (oldest), got %s", activity2.Commits[1].Hash)
	}

	// Verify that commits are displayed in newest-first order in the HTML content
	// First item should show b19a1b6 (newer) before 8e9b024 (older)
	firstContent := result[0].Content
	b19Pos := strings.Index(firstContent, ">b19a1b6<")
	e9bPos := strings.Index(firstContent, ">8e9b024<")
	if b19Pos == -1 || e9bPos == -1 {
		t.Errorf("Could not find both commit hashes in first item content")
	} else if b19Pos > e9bPos {
		t.Errorf("Commits not in newest-first order: b19a1b6 at pos %d, 8e9b024 at pos %d", b19Pos, e9bPos)
	}

	// Second item should show def456a (newer) before abc123d (older)
	secondContent := result[1].Content
	defPos := strings.Index(secondContent, ">def456a<")
	abcPos := strings.Index(secondContent, ">abc123d<")
	if defPos == -1 || abcPos == -1 {
		t.Errorf("Could not find both commit hashes in second item content")
	} else if defPos > abcPos {
		t.Errorf("Commits not in newest-first order: def456a at pos %d, abc123d at pos %d", defPos, abcPos)
	}

	// Verify the comparison links use the original format from the feed item
	expectedLink1 := "https://github.com/cdzombak/dotfiles/compare/b19a1b604e...8e9b024bed"
	if !strings.Contains(result[0].Content, expectedLink1) {
		t.Errorf("Expected comparison link %s not found in first item content", expectedLink1)
	}

	expectedLink2 := "https://github.com/cdzombak/another-repo/compare/abc123def456...def456abc789"
	if !strings.Contains(result[1].Content, expectedLink2) {
		t.Errorf("Expected comparison link %s not found in second item content", expectedLink2)
	}
}

func TestCreateIndividualPushItem(t *testing.T) {
	publishedTime, _ := time.Parse(time.RFC3339, "2025-09-15T01:28:02Z")

	activity := &BranchActivity{
		Repo:   "dotfiles",
		Branch: "master",
		Commits: []Commit{
			{Hash: "8e9b024", Message: "remove Instapaper Save app", Link: "https://github.com/cdzombak/dotfiles/commit/8e9b024"},
			{Hash: "b19a1b6", Message: "fix Red Eye install", Link: "https://github.com/cdzombak/dotfiles/commit/b19a1b6"},
		},
		LatestTime:  &publishedTime,
		CompareLink: "https://github.com/cdzombak/dotfiles/compare/b19a1b6^...8e9b024",
	}

	result := createIndividualPushItem(activity, "cdzombak")

	if result == nil {
		t.Fatal("createIndividualPushItem() returned nil")
	}

	expectedTitle := "cdzombak pushed 2 commits to dotfiles/master"
	if result.Title != expectedTitle {
		t.Errorf("createIndividualPushItem() title = %v, want %v", result.Title, expectedTitle)
	}

	// Should contain both commits in content
	if !strings.Contains(result.Content, "8e9b024") || !strings.Contains(result.Content, "remove Instapaper Save app") {
		t.Error("createIndividualPushItem() missing first commit in content")
	}

	if !strings.Contains(result.Content, "b19a1b6") || !strings.Contains(result.Content, "fix Red Eye install") {
		t.Error("createIndividualPushItem() missing second commit in content")
	}

	// Should have compare link
	if result.Link != activity.CompareLink {
		t.Errorf("createIndividualPushItem() link = %v, want %v", result.Link, activity.CompareLink)
	}

	// Should have individual GUID format
	expectedGUIDPrefix := "individual-dotfiles-master-"
	if !strings.HasPrefix(result.GUID, expectedGUIDPrefix) {
		t.Errorf("createIndividualPushItem() GUID = %v, should start with %v", result.GUID, expectedGUIDPrefix)
	}
}
