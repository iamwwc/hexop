package main

import (
	"fmt"
	"strings"
)

type yamlGenerator struct {
	src map[string] interface{}
}
func NewYamlGenerator() *yamlGenerator {
	return &yamlGenerator{src:make(map[string]interface{})}
}


func (y *yamlGenerator) WithKV(k, v string) *yamlGenerator {
	y.src[k] = v
	return y
}

func (y *yamlGenerator) WithArray(k string, v []string) *yamlGenerator {
	y.src[k] = v
	return y
}

func (y *yamlGenerator) Done() string{
	builder := strings.Builder{}
	for k, v := range y.src {
		switch value := v.(type) {
		case string:
			builder.WriteString(fmt.Sprintf("%s: %s\n",k, value))
		case []string:
			builder.WriteString(fmt.Sprintf("%s:\n",k))
			for _, a := range value {
				builder.WriteString(fmt.Sprintf("\t- %s: \n",a))
			}
		default:
			panic(fmt.Sprintf("Unknown type of %T",value))
		}
	}
	return builder.String()
}