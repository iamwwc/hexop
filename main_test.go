package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestGenerator(t *testing.T) {
	p := fmt.Sprintf("%s/repos/%s/%s/issues/15?state=open&creator=iamwwc",githubApiPath,"iamwwc","blogsuepost")
	result := apirequest(p)
	var r interface{}
	Must(json.Unmarshal(result, &r))
	m := r.(map[string]interface{})
	group := &sync.WaitGroup{}
	cwd := Must2(os.Getwd()).(string)
	fp := cwd + "/_posts/"
	Must(os.MkdirAll(fp,0644))
	defer Must(os.RemoveAll(fp))
	go generateFile(group, fp,"1234",m)
	group.Wait()
}
