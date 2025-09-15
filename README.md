# ghfeed

A GitHub activity feed consolidator that transforms verbose GitHub Atom feeds into clean, readable summaries.

## Features

- **Commit Consolidation**: Groups multiple commits by repository and branch into single entries
- **Activity Simplification**: Converts verbose GitHub HTML into clean, minimal descriptions
- **Data Preservation**: Maintains all commit messages, links, and essential metadata
- **HTML Output**: Generates properly formatted entries with clickable links

### What It Does

| Before | After |
|--------|-------|
| 20+ individual "pushed" entries cluttering your feed | Clean consolidated entries like "cdzombak pushed 15 commits to dotfiles/master" |
| Verbose GitHub HTML with excessive markup | Simple entries like "cdzombak opened PR #264 in mmcdole/gofeed" |

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
> Real-world output feed example: [dzombak.com/feeds/github-cdzombak.atom](https://www.dzombak.com/feeds/github-cdzombak.atom)

## Usage

```bash
ghfeed [options] <feed-url>  > /path/to/output.atom
```

### Options

- `-format rss|json|atom`: Set the format of the output feed
- `-retitle "new title"`: Set the title of the output feed

### Docker

```shell
docker run --rm cdzombak/ghfeed:1 <feed-url>  > /path/to/output.atom
```

## Installation

## Debian via apt repository

Set up my `oss` apt repository:

```shell
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://dist.cdzombak.net/keys/dist-cdzombak-net.gpg -o /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo chmod 644 /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo mkdir -p /etc/apt/sources.list.d
sudo curl -fsSL https://dist.cdzombak.net/cdzombak-oss.sources -o /etc/apt/sources.list.d/cdzombak-oss.sources
sudo chmod 644 /etc/apt/sources.list.d/cdzombak-oss.sources
sudo apt update
```

Then install `ghfeed` via `apt-get`:

```shell
sudo apt-get install ghfeed
```

## Homebrew

```shell
brew install cdzombak/oss/ghfeed
```

## Manual from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/ghfeed/releases). Debian packages for each release are available as well.

## License

GNU GPL v3; see [LICENSE](LICENSE) in this repo for details.

## Author

Chris Dzombak
- [dzombak.com](https://www.dzombak.com)
- [GitHub @cdzombak](https://github.com/cdzombak)
