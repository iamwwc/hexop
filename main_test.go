package main

import (
	"testing"
)

func TestGenerator(t *testing.T) {
	i := &info{token:"",owner:"iamwwc",repo:"articles",repoOwnerName:"iamwwcposts",url:"https://github.com/iamwwcposts/articles",currPage:1}
	iterator(i)
}
