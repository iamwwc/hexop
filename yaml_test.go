package main

import (
	"fmt"
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
