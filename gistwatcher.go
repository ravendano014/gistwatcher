// Package main (gistwatcher.go) :
// These are the methods for gistwatcher.
package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	goquery "github.com/PuerkitoBio/goquery"
	fetchall "github.com/tanaikech/go-fetchall"
	"github.com/urfave/cli"
)

const (
	appname     = "gistwatcher"              // Application name
	nameenv     = "GISTWATCHER_NAME"         // Environment variable for the log in name of GitHub account.
	passenv     = "GISTWATCHER_PASS"         // Environment variable for the log in password of GitHub account.
	tokenenv    = "GISTWATCHER_ACCESSTOKEN"  // Environment variable for the access token of GitHub.
	baseAPIURL  = "https://api.github.com/"  // Base API URL of GitHub API
	baseGistURL = "https://gist.github.com/" // Base Gist URL
	perPage     = "100"                      // Items per a page
)

type (
	// para : Basic parameters for running this script.
	para struct {
		accountName      string
		accountPass      string
		accessToken      string
		apiURL           string
		client           *http.Client
		filename         string
		getStarsAndForks bool
		result           Result
		startTime        time.Time
		token            string
		URLFromArgs      string
		URLs             []string
		workDir          string
	}

	// gistAPIRes : Response from GitHub API.
	gistAPIRes struct {
		URL         string    `json:"url"`
		ID          string    `json:"id"`
		HTMLURL     string    `json:"html_url"`
		Public      bool      `json:"public"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Description string    `json:"description"`
		Comments    int       `json:"comments"`
		Owner       struct {
			Login string `json:"login"`
		} `json:"owner"`
		Error string `json:"error,omitempty"`
	}

	// pageToken : Pagetoken for retrieving gists.
	pageToken struct {
		Rel   string
		URL   string
		Query url.Values
	}

	// Result : Output values.
	Result struct {
		Items          []ResultItems `json:"items"`
		TotalItems     int           `json:"total_items"`
		ExpirationTime float64       `json:"expiration_time"`
		StartTime      string        `json:"start_time"`
	}

	// ResultItems : Output values.
	ResultItems struct {
		Comments  *int   `json:"comments,omitempty"`
		CreatedAt string `json:"create_at,omitempty"`
		Forks     *int   `json:"forks,omitempty"`
		Public    bool   `json:"public,omitempty"`
		Stars     *int   `json:"stars,omitempty"`
		Title     string `json:"title,omitempty"`
		UpdatedAt string `json:"updated_at,omitempty"`
		URL       string `json:"url,omitempty"`
		GistID    string `json:"gist_id,omitempty"`
		Error     string `json:"error,omitempty"`
	}
)

// putResultItems : Put values to element of result.
func putResultItems(g gistAPIRes) *ResultItems {
	if g.Error != "" {
		return &ResultItems{
			Error:  g.Error,
			GistID: g.ID,
		}
	}
	return &ResultItems{
		Comments:  &g.Comments,
		CreatedAt: g.CreatedAt.In(time.Local).Format("20060102 15:04:05 MST"),
		Public:    g.Public,
		Title:     g.Description,
		UpdatedAt: g.UpdatedAt.In(time.Local).Format("20060102 15:04:05 MST"),
		URL:       g.HTMLURL,
	}
}

// displayResult : Display result.
func (p *para) displayResult() error {
	if p.result.TotalItems == 0 {
		return errors.New("no Gists were retrieved")
	}
	p.result.ExpirationTime = math.Trunc(time.Now().Sub(p.startTime).Seconds()*1000) / 1000
	result, err := json.Marshal(p.result)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", result)
	return nil
}

// displayItems : Display items.
func (p *para) displayItems(g []gistAPIRes) {
	if !p.getStarsAndForks {
		for _, e := range g {
			p.result.Items = append(p.result.Items, *putResultItems(e))
		}
	}
}

// initParams : Initialize basic parameters.
func initParams(c *cli.Context) (*para, error) {
	var err error
	workdir, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	p := &para{
		accountName: c.String("name"),
		accountPass: c.String("password"),
		accessToken: c.String("accesstoken"),
		apiURL: func(u string) string {
			if u != "" {
				return baseAPIURL + "users/" + u + "/gists"
			}
			return baseAPIURL + "gists"
		}(c.String("username")),
		filename:         strings.TrimSpace(c.String("file")),
		getStarsAndForks: c.Bool("getstars"),
		workDir:          workdir,
		client: &http.Client{
			Timeout: time.Duration(30) * time.Second,
		},
		startTime:   time.Now(),
		URLFromArgs: c.String("url"),
	}
	p.result.StartTime = p.startTime.In(time.Local).Format("20060102 15:04:05 MST")
	if p.accessToken == "" {
		nameEnv := os.Getenv(nameenv)
		passEnb := os.Getenv(passenv)
		if (p.accountName == "" && p.accountPass == "") && (nameEnv != "" && passEnb != "") {
			p.accountName = strings.TrimSpace(nameEnv)
			p.accountPass = strings.TrimSpace(passEnb)
		}
		bAuth := p.accountName + ":" + p.accountPass
		p.token = "Basic " + base64.StdEncoding.EncodeToString([]byte(bAuth))
	} else {
		p.token = "token " + p.accessToken
	}
	if p.accountName == "" && p.accountPass == "" && p.accessToken == "" {
		return nil, errors.New("Please use account name and password for GitHub or your access token for using GitHub API")
	}
	return p, nil
}

// getItemsFromGists : Get all items from own Gists.
func (p *para) getItemsFromGists() ([]gistAPIRes, error) {
	urlVal := p.apiURL + "?page=1&per_page=" + perPage
	req, err := http.NewRequest("GET", urlVal, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", p.token)
	r := &fetchall.Request{
		Request: req,
		Client:  p.client,
	}
	params1 := &fetchall.Params{
		Workers: 10,
	}
	params1.Requests = append(params1.Requests, *r)
	items := fetchall.Do(params1)
	reg := regexp.MustCompile(`<(\w.+?)>; rel=\"(\w.+?)\"`)
	body, err := ioutil.ReadAll(items[0].Response.Body)
	if err != nil {
		return nil, err
	}
	defer items[0].Response.Body.Close()
	g := []gistAPIRes{}
	json.Unmarshal(body, &g)
	headerLink := items[0].Response.Header.Get("Link")
	links := strings.Split(headerLink, ",")
	pageTokenList := []pageToken{}
	for _, e := range links {
		link := reg.FindStringSubmatch(e)
		u, _ := url.Parse(link[1])
		query := u.Query()
		pt := &pageToken{
			URL:   link[1],
			Rel:   link[2],
			Query: query,
		}
		pageTokenList = append(pageTokenList, *pt)
	}
	plen := len(pageTokenList)
	if pageTokenList[plen-1].Rel == "last" {
		params2 := &fetchall.Params{
			Workers: 10,
		}
		lastPage, _ := strconv.Atoi(pageTokenList[plen-1].Query.Get("page"))
		for i := 1; i < lastPage; i++ {
			purl := p.apiURL + "?page=" + strconv.Itoa(i+1) + "&per_page=" + perPage
			req, err := http.NewRequest("GET", purl, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", p.token)
			r := &fetchall.Request{
				Request: req,
				Client:  p.client,
			}
			params2.Requests = append(params2.Requests, *r)
		}
		addedItems := fetchall.Do(params2)
		for _, e := range addedItems {
			g2 := []gistAPIRes{}
			body, err := ioutil.ReadAll(e.Response.Body)
			if err != nil {
				return nil, err
			}
			defer e.Response.Body.Close()
			if e.Response.StatusCode == 200 {
				json.Unmarshal(body, &g2)
			} else {
				ge := &gistAPIRes{}
				ge.Error = string(body)
				g2 = append(g2, *ge)
			}
			json.Unmarshal(body, &g2)
			g = append(g, g2...)
		}
	}
	p.result.TotalItems = len(g)

	for i, j := 0, len(g)-1; i < j; i, j = i+1, j-1 { // Ref https://stackoverflow.com/a/19239850
		g[i], g[j] = g[j], g[i]
	}

	if !p.getStarsAndForks {
		p.displayItems(g)
	}

	// Output : g is a slice which has the oldest and latest gists are the top and last element, respectively.
	return g, nil
}

// parseGistIDs : Parse Gist IDs.
func (p *para) parseGistIDs(val []string) {
	for _, e := range val {
		u := func(v string) string {
			if strings.Contains(v, "gist.github.com") {
				temp := strings.Split(v, "/")
				return temp[len(temp)-1]
			} else if len(v) == 32 {
				return v
			}
			return ""
		}(strings.TrimSpace(e))
		p.URLs = append(p.URLs, u)
	}
}

// parseURLsFromFile : Parse URLs from file.
func (p *para) parseURLsFromFile() error {
	f, err := os.Open(p.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	urls := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		temp := strings.TrimSpace(scanner.Text())
		if temp == "end" {
			break
		}
		if temp != "" {
			urls = append(urls, temp)
		}
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	p.parseGistIDs(urls)
	return nil
}

// parseURLsFromArgs : Parse URLs from args.
func (p *para) parseURLsFromArgs() {
	arg := strings.Split(p.URLFromArgs, ",")
	p.parseGistIDs(arg)
}

// getItemsFromGistsByURL : Get items from Gists by manually inputted URLs or IDs.
func (p *para) getItemsFromGistsByURLs() ([]gistAPIRes, error) {
	params2 := &fetchall.Params{
		Workers: 10,
	}
	for _, e := range p.URLs {
		if e != "" {
			req, err := http.NewRequest("GET", baseAPIURL+"gists/"+e, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", p.token)
			r := &fetchall.Request{
				Request: req,
				Client:  p.client,
			}
			params2.Requests = append(params2.Requests, *r)
		}
	}
	res := fetchall.Do(params2)
	g := []gistAPIRes{}
	for i, e := range res {
		g2 := gistAPIRes{}
		body, err := ioutil.ReadAll(e.Response.Body)
		if err != nil {
			return nil, err
		}
		defer e.Response.Body.Close()
		if e.Response.StatusCode == 200 {
			json.Unmarshal(body, &g2)
		} else {
			g2.Error = string(body)
			g2.ID = p.URLs[i]
		}
		g = append(g, g2)
	}
	p.result.TotalItems = len(g)
	if !p.getStarsAndForks {
		p.displayItems(g)
	}
	return g, nil
}

// getStargazersAndForks : Get startgazers and forks from Gists.
func (p *para) getStargazersAndForks(g []gistAPIRes) error {
	params := &fetchall.Params{
		Workers: 10,
	}
	for _, e := range g {
		gistURL := baseGistURL + e.Owner.Login + "/" + e.ID + "/stargazers"
		req, err := http.NewRequest("GET", gistURL, nil)
		if err != nil {
			return err
		}
		r := &fetchall.Request{
			Request: req,
			Client:  p.client,
		}
		params.Requests = append(params.Requests, *r)
	}
	htmlRes := fetchall.Do(params)
	for i, e := range htmlRes {
		ri := putResultItems(g[i])
		if e.Response.StatusCode == 200 {
			doc, err := goquery.NewDocumentFromReader(e.Response.Body)
			if err != nil {
				return err
			}
			doc.Find("ul[class='pagehead-actions float-none']").Each(func(_ int, s *goquery.Selection) {
				s.Find("li").Each(func(_ int, s *goquery.Selection) {
					identifier := func(str string) string {
						if strings.Contains(str, "star") {
							return "Star"
						} else if strings.Contains(str, "fork") {
							return "Fork"
						}
						return ""
					}(strings.ToLower(s.Text()))
					if identifier != "" {
						s.Find(".social-count").Each(func(_ int, s *goquery.Selection) {
							count := s.Text()
							cnt, _ := strconv.Atoi(strings.TrimSpace(count))
							if identifier == "Star" {
								ri.Stars = &cnt
							} else if identifier == "Fork" {
								ri.Forks = &cnt
							}
						})
					}
				})
			})
		} else {
			ri.Error = "Error: Couldn't find gist."
			ri.GistID = g[i].ID
		}
		p.result.Items = append(p.result.Items, *ri)
	}
	return nil
}

// handler : Handler of this script.
func handler(c *cli.Context) error {
	var err error
	p, err := initParams(c)
	if err != nil {
		return err
	}
	g, err := func() ([]gistAPIRes, error) {
		if c.String("url") == "" && c.String("file") == "" {
			return p.getItemsFromGists()
		} else if c.String("url") != "" && c.String("file") == "" {
			p.parseURLsFromArgs()
		} else if c.String("url") == "" && c.String("file") != "" {
			if err = p.parseURLsFromFile(); err != nil {
				return nil, err
			}
		}
		return p.getItemsFromGistsByURLs()
	}()
	if err != nil {
		return err
	}

	// in the current stage, in order to retrieve the number of stars and forks,
	// it is required to directly scrape Gists.
	// so I decided to use the bool flag when the number of stars and forks are retrieved.
	if p.getStarsAndForks {
		if err = p.getStargazersAndForks(g); err != nil {
			return err
		}
	}

	if err = p.displayResult(); err != nil {
		return err
	}
	return nil
}

// createHelp : Create help document.
func createHelp() *cli.App {
	a := cli.NewApp()
	a.Name = appname
	a.Authors = []*cli.Author{
		{Name: "tanaike [ https://github.com/tanaikech/" + appname + " ] ", Email: "tanaike@hotmail.com"},
	}
	a.UsageText = "Get comments, stars and forks from own Gists."
	a.Version = "1.0.1"
	a.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "name, n",
			Aliases: []string{"n"},
			Usage:   "Login name of GitHub.",
		},
		&cli.StringFlag{
			Name:    "password, p",
			Aliases: []string{"p"},
			Usage:   "Login password of GitHub.",
		},
		&cli.StringFlag{
			Name:    "accesstoken, a",
			Aliases: []string{"a"},
			Usage:   "Access token of GitHub. If you have this, please use this instead of 'name' and 'password'.",
		},
		&cli.BoolFlag{
			Name:    "getstars, s",
			Aliases: []string{"s"},
			Usage:   "If you want to also retrieve the number of stars and forks, please use this.",
		},
		&cli.StringFlag{
			Name:    "username, user",
			Aliases: []string{"user"},
			Usage:   "User name of Gist you want to get. If you want to retrieve a specific user's Gists, please use this.",
		},
		&cli.StringFlag{
			Name:    "url, u",
			Aliases: []string{"u"},
			Usage:   "URL of Gists you want to retrieve. You can also use Gist's ID instead of URL.",
		},
		&cli.StringFlag{
			Name:    "file, f",
			Aliases: []string{"f"},
			Usage:   "Filename including URLs of Gists you want to retrieve.",
		},
	}
	return a
}

// main : Main of this script
func main() {
	a := createHelp()
	a.Action = handler
	err := a.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
