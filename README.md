# gh-bbs-analyzer
GitHub CLI extension for analyzing BitBucket Server to get migration statistics.

[![build](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/build.yaml/badge.svg)](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/build.yaml)
[![release](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/release.yaml/badge.svg)](https://github.com/mona-actions/gh-bbs-analyzer/actions/workflows/release.yaml)

## Under Construction
This is a work-in-progress. Currently the tool will report:

- Repository disk sizes
- Pull request counts
- Comment counts
- If repository is archived
- If repository is public

Tested on BitBucket Server versions:
- 8.6.1

Future plans include:
- Attachment counts
- LFS detection
- Large file scanning (above 100MB)

## Prerequisites
- [GitHub CLI](https://cli.github.com/manual/installation) installed.

## Permissions Required
In order for the analysis to work correctly, you will need to utilize a systems admin account so that all project and repository endpoints respond appropriately. If you don't, you will only receive limited information (the projects & repositories the user has access to).

The endpoints currently called are:
- `<server>/<api-url>/projects`
- `<server>/<api-url>/projects/<project-key>/repos`
- `<server>/projects/<project-key>/repos/<repo-slug>/sizes` (non-API endpoint hack)
- `<server>/<api-url>/projects/<project-key>/repos/<repo-slug>/pull_requests`

## Install

```bash
$ gh extension install mona-actions/gh-bbs-analyzer
```

## Upgrade
```bash
$ gh extension upgrade bbs-analyzer
```

## Usage

```txt
$ gh bbs-analyzer [flags]
```

```txt
GitHub CLI extension to analyze BitBucket Server for migration statistics

Usage:
  gh bbs-analyzer [flags]

Flags:
  -p, --bbs-password string     The Bitbucket password of the user specified by --bbs-username. If not set will be read from BBS_PASSWORD environment variable.
      --bbs-project string      A specific Bitbucket project instead of analyzing all proejcts.
  -s, --bbs-server-url string   The full URL of the Bitbucket Server/Data Center to migrate from. E.g. http://bitbucket.contoso.com:7990
  -u, --bbs-username string     The Bitbucket username of a user with site admin privileges. If not set will be read from BBS_USERNAME environment variable.
  -h, --help                    help for gh
      --no-ssl-verify           Disables SSL verification when communicating with your Bitbucket Server/Data Center instance. All other migration steps will continue to verify SSL. If your Bitbucket instance has a self-signed SSL certificate then setting this flag will allow the migration archive to be exported.
  -o, --output-file string      The file to output the results to. (default "results.csv")
  -t, --threads int             Number of threads to process concurrently. (default 3)
  -v, --version                 version for gh
```