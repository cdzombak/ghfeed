# ghfeed

A GitHub activity feed consolidator that transforms verbose GitHub Atom feeds into clean, readable summaries.

## Features

- **Commit Consolidation**: Groups multiple commits by repository and branch into single entries
- **Activity Simplification**: Converts verbose GitHub HTML into clean, minimal descriptions
- **Data Preservation**: Maintains all commit messages, links, and essential metadata
- **HTML Output**: Generates properly formatted entries with clickable links

### What It Does

**Before**: 20+ individual "pushed" entries cluttering your feed
**After**: Clean consolidated entries like "cdzombak pushed 15 commits to dotfiles/master"

**Before**: Verbose GitHub HTML with excessive markup
**After**: Simple entries like "cdzombak opened PR #264 in mmcdole/gofeed"

The program processes these GitHub activities:
- Push/commit consolidation by repository and branch
- Pull request creation and merging
- Repository forks
- Branch and tag management
- Other GitHub activities

### Output

Generates a clean Atom feed with:
- Consolidated commit entries showing individual commit messages and links
- Simplified non-commit activities with essential information preserved
- Proper chronological ordering
- Valid Atom XML format

Perfect for making GitHub activity feeds more readable in feed readers.

> [!NOTE]  
> Real-world output feed: TK

## Usage

```bash
ghfeed <feed-url> > /path/to/output.atom
```

### Docker

TK

## Installation

## Debian via apt repository

TK

## Homebrew

TK

## Manual from build artifacts

TK

## License

GNU GPL v3; see [LICENSE](LICENSE) in this repo for details.

## Author

Chris Dzombak
- [dzombak.com](https://www.dzombak.com)
- [GitHub @cdzombak](https://github.com/cdzombak)
