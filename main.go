package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	repoUrlRegexp = "^https://github\\.com/(.+)/(.+)"
	githubApiPath = "https://api.github.com"
	linkRegexp = "<(.+?)>"
)

var (
	letters   = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	linebreak = ""
)

func init() {
	if runtime.GOOS == "windows" {
		linebreak = "\r\n"
	} else {
		linebreak = "\n"
	}
}

func main() {
	app := cli.NewApp()
	app.Before = func(context *cli.Context) error {
		log.SetLevel(log.DebugLevel)
		log.SetOutput(os.Stdout)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:     "repourl",
			Usage:    "posts url",
			Required: true,
			Hidden:   false,
		},
		cli.StringFlag{
			Name:        "owner",
			Usage:       "repo user name",
			Required:    true,
			Hidden:false,
		},
	}
	app.Action = action
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func action(ctx *cli.Context) {
	url := ctx.String("repourl")
	owner := ctx.String("owner")
	repoOwnerName, repo := parseRepo(url)

	token := parseTokenFromEnv()

	i := &info{
		url,owner,repo, repoOwnerName,token,1,
	}
	iterator(i)
}

type info struct {
	url string
	owner string
	repo string
	repoOwnerName string
	token string
	currPage int
}

func getAPIPath(repoOwnerName string, repo string, owner string, token string, page int) string {
	return fmt.Sprintf("%s/repos/%s/%s/issues?state=open&creator=%s&page=%d", githubApiPath, repoOwnerName, repo, owner,page)
}

func(i *info) apiWithPage() string{
	return getAPIPath(i.repoOwnerName,i.repo,i.owner,i.token,i.currPage)
}

func iterator(i *info) {
	b,header := i.apirequest(i.apiWithPage())
	if b == nil {
		return
	}
	if link, ok := header["Link"] ; ok{
		r, err := regexp.Compile(linkRegexp)
		if err!= nil {
			panic(err)
		}
		result := r.FindAllStringSubmatch(link[0],-1)
		next, last := result[0], result[1];
		_ = last
		pageNum := regexp.MustCompile("page=(\\d+)").FindStringSubmatch(next[1])[1]
		num := Must2(strconv.Atoi(pageNum)).(int)
		process(b)
		if num < i.currPage {
			return
		}
		i.currPage = num
		iterator(i)
	}
}

func process(b []byte) {
	var issues []map[string]interface{}
	if err := json.Unmarshal(b, &issues); err != nil {
		log.Error(err)
	}
	currPath, err := os.Getwd()
	postspath := filepath.Join(currPath, "source/_posts")
	if err != nil {
		log.Error(postspath)
	}

	f, err := os.Stat(postspath)
	if os.IsNotExist(err) {
		log.Warnf("file path %s don't exists, _posts will be created", postspath)
		Must(os.MkdirAll(postspath, 0644))
	}

	if f != nil && !f.IsDir() {
		log.Errorf("%s is not a folder", postspath)
	}

	groups := sync.WaitGroup{}
	length := len(issues)
	groups.Add(length)
	rand.Seed(time.Now().Unix())
	for _, issue := range issues {
		random := make([]rune, 6)
		for count := 6; count > 0; count-- {
			random[count-1] = letters[rand.Intn(len(letters))]
		}
		go generateFile(&groups, postspath, string(random), issue)
	}
	groups.Wait()
}

// 创建文件时添加上随机数，避免冲突。
func generateFile(waitGroup *sync.WaitGroup, fp string, random string, issue map[string]interface{}) {
	defer waitGroup.Done()
	labels := issue["labels"].([]interface{})
	tags := make([]string, 0)

	for _, l := range labels {
		v := l.(map[string]interface{})["name"].(string)
		tags = append(tags, v)
	}

	title := issue["title"].(string)
	body := issue["body"].(string)
	date := issue["created_at"].(string)
	date = strings.Split(date, "T")[0]

	updateBy := issue["updated_at"].(string)
	updateBy = strings.Split(updateBy, "T")[0]
	url := issue["url"].(string)
	part := strings.Split(url, "/")
	issueId := part[len(part) - 1]
	generator := NewYamlGenerator()
	pageHeader := generator.WithKV("title", title).
		WithKV("date", date).
		WithKV("updated",updateBy).
		WithKV("issueid",issueId).
		WithArray("tags", tags).Done()

	buffer := bytes.Buffer{}
	buffer.WriteString("---\n")
	buffer.WriteString(pageHeader)
	buffer.WriteString("---\n")
	buffer.WriteString(body)
	f, err := os.OpenFile(filepath.Join(fp, issueId + "-" + title +".md"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		log.Error(err)
	}
	if _, err := f.WriteString(buffer.String()); err != nil {
		log.Error(err)
	}
}

func (i* info)apirequest(path string) ([]byte,http.Header){
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	req, err := http.NewRequest("GET",path,bytes.NewBufferString(""))
	if err != nil {
	log.Error(err)
	}
	if i.token != "" {
		req.Header.Add("Authorization","token "+ i.token)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Read from response failed, error: %v", err)
		return nil,nil
	}
	return bs,resp.Header
}

func parseRepo(url string) (string,string) {
	re := regexp.MustCompile(repoUrlRegexp)
	result := re.FindStringSubmatch(url)
	return result[1],result[2]
}

func parseTokenFromEnv() (token string) {
	envs := os.Environ()
	token = ""
	for _, v := range envs {
		if strings.Index(v, "GITHUB_TOKEN") == 0 {
			token = strings.Split(v, "=")[1]
			break
		}
	}
	return
}

func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
