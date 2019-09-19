package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestSlice(t *testing.T) {
	b := make([]byte,2,4)
	c := make([]byte,0,2)
	b = append(b, 1)
	c = append(c,1)
}

func TestYamlGenerator_Done(t *testing.T) {
	generator := NewYamlGenerator()
	result := generator.WithKV("title","iamtitle").
				WithArray("tags",[]string{"t1","t2","t3"}).
				Done()
	fmt.Printf(result)
}

func TestIndex(t *testing.T) {
	s := "GITHUB_TOKEN=9ace205cbd67dbee0cf29cccfd26d625a03bdd4b"
	strings.Index(s,"GITHUB_TOKEN")
}
