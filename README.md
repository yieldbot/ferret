## Ferret

[![Build Status][travis-image]][travis-url] [![GoDoc][godoc-image]][godoc-url] [![Release][release-image]][release-url]

Ferret is a search engine that unifies search results from different resources
such as Github, Slack, AnswerHub and more.

Distributed knowledge and avoiding context switching are very important for
efficiency. Ferret provides a unified search interface for retrieving and
accessing to information with minimal effort.

## IT'S STILL UNDER DEVELOPMENT


### Installation

| OSX | Linux | Win |
|:---:|:---:|:---:|
| [64bit][download-darwin-amd64-url] | [64bit][download-linux-amd64-url] | [64bit][download-windows-amd64-url] |

[See](#building-from-source) for building from source.


### Usage

_Make sure Ferret is [configured](#configuration) properly before use it._

#### Help

```
./ferret -h
```

#### Search Github

```
./ferret search github intent
./ferret search github intent+extension:md
```

[See](https://developer.github.com/v3/search/) for more Github search syntax.

#### Search Slack

```
./ferret search slack "meeting minutes"
```

#### Search AnswerHub

```
./ferret search answerhub vpn
```

#### Opening search results

```
# search for alerting on Github
./ferret search github alerting

# go to the second search result
./ferret search github alerting --goto 2
```


### Configuration

Keep environment files in `~/.bash_profile` file for persistence.

#### Global Configuration

- `FERRET_GOTO_CMD`:
  The command for opening links. This command is used for `--goto` argument. 
  (i.e. `ferret search PROVIDER KEYWORD --goto #1`). Default is `open`

#### Providers Configurations

Each search provider needs set of environment variables for operating. You can 
define environment variables for one or more search provider.

##### Github

  [See](https://help.github.com/articles/creating-an-access-token-for-command-line-use/)
  for getting an access token and set following environment variables;

  * `FERRET_GITHUB_URL` (i.e. `https://api.github.com/`)
  * `FERRET_GITHUB_TOKEN`
  * `FERRET_GITHUB_SEARCH_USER` (*optional*)

##### Slack

  [See](https://api.slack.com/docs/oauth-test-tokens?team_id=T025F5Q7Y)
  for getting an access token and set following environment variables;

  * `FERRET_SLACK_TOKEN`

##### AnswerHub

  [See](http://docs.answerhub.com/articles/1444/how-to-enable-and-grant-use-of-the-rest-api.html)
  for enabling the REST API and set following environment variables;

  * `FERRET_ANSWERHUB_URL` (i.e. `https://answerhub.example.com/`)
  * `FERRET_ANSWERHUB_USERNAME` (see *My Preferences->Authentication Modes* in AnswerHub site)
  * `FERRET_ANSWERHUB_PASSWORD` (see *My Preferences->Authentication Modes* in AnswerHub site)


### Building from source

```
go get -u -v github.com/yieldbot/ferret
cd $GOPATH/src/github.com/yieldbot/ferret
go build
```


### License

Licensed under The MIT License (MIT)  
For the full copyright and license information, please view the LICENSE.txt file.


[travis-url]: https://travis-ci.org/yieldbot/ferret
[travis-image]: https://travis-ci.org/yieldbot/ferret.svg?branch=master

[godoc-url]: https://godoc.org/github.com/yieldbot/ferret
[godoc-image]: https://godoc.org/github.com/yieldbot/ferret?status.svg

[release-url]: https://github.com/yieldbot/ferret/releases/latest
[release-image]: https://img.shields.io/badge/release-1.0.0-blue.svg

[download-darwin-amd64-url]: https://github.com/yieldbot/fetter/releases/download/v1.0.0/fetter-darwin-amd64.zip
[download-linux-amd64-url]: https://github.com/yieldbot/fetter/releases/download/v1.0.0/fetter-linux-amd64.zip
[download-windows-amd64-url]: https://github.com/yieldbot/fetter/releases/download/v1.0.0/fetter-windows-amd64.zip