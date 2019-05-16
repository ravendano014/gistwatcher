# gistwatcher

[![Build Status](https://travis-ci.org/tanaikech/gistwatcher.svg?branch=master)](https://travis-ci.org/tanaikech/gistwatcher)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENCE)

<a name="top"></a>

# Overview

This is a CLI tool for retrieving the number of comments, stars and forks of Gists.

# Demo

![](images/demo.gif)

In this demonstration, a Gist is retrieved by an URL. You can see that the number of comments, stars and forks can be retrieved.

# Install

You can install this using `go get` as follows.

```bash
$ go get -u github.com/tanaikech/gistwatcher
```

# Usage

## Requirement

gistwatcher uses the account name and password of GitHub. This is used for using GitHub API. The values are retrieved by the API with the account name and password.

Also you can use the access token instead of the account name and password. In that case, please retrieve the access token by yourself.

## Methods

For all methods, JSON object is returned.

### 1. Retrieve the number of comments, stars and forks of Gists from URLs and Gists IDs

```bash
$ gistwatcher -s -u https://gist.github.com/#####/#####
```

- You can retrieve values using the user name. When user name is used, all Gists of the user are retrieved.

### 2. Retrieve the number of comments, stars and forks of all Gists of an user

```bash
$ gistwatcher -s -user #####
```

- You can retrieve values using the user name. When user name is used, all Gists of the user are retrieved.

### 3. Retrieve the number of comments, stars and forks of all Gists from a file including Gists URLs

```bash
$ gistwatcher -s -f filename.txt
```

The sample file is as follows.

```
https://gist.github.com/#####/#####
https://gist.github.com/#####/#####
https://gist.github.com/#####/#####
```

### Options

| Options            | Description                                                                                          |
| :----------------- | :--------------------------------------------------------------------------------------------------- |
| --name, --n        | Login name of GitHub.                                                                                |
| --password, --p    | Login password of GitHub.                                                                            |
| --accesstoken, --a | Access token of GitHub. If you have this, please use this instead of 'name' and 'password'.          |
| --getstars, --s    | If you want to also retrieve the number of stars and forks, please use this.                         |
| --username, --user | User name of Gist you want to get. If you want to retrieve a specific user's Gists, please use this. |
| --url, --u         | URL of Gists you want to retrieve. You can also use Gist's ID instead of URL.                        |
| --file, --f        | Filename including URLs of Gists you want to retrieve.                                               |
| --help             | Show help.                                                                                           |

1. You can also use the environment variables for the account name and password, and the access token.
   - Environment variables for account name and password are `GISTWATCHER_NAME`, `GISTWATCHER_PASS`, respectively.
   - Environment variables for access token is `GISTWATCHER_ACCESSTOKEN`.

# Note

- In the current stage, in order to retrieve the number of stars and forks, it is required to directly scrape Gists. So I decided to use the bool flag when the number of stars and forks are retrieved.

---

<a name="licence"></a>

# Licence

[MIT](LICENCE)

<a name="author"></a>

# Author

[Tanaike](https://tanaikech.github.io/about/)

If you have any questions and commissions for me, feel free to tell me.

<a name="updatehistory"></a>

# Update History

- v1.0.0 (May 16, 2019)

  1. Initial release.

[TOP](#top)
