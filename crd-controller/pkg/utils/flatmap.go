package utils

/*
This files uses source code from the files map.go and flatmap.go to expose map interface.
https://github.com/devopsfaith/flatmap

//TODO open issue/PR to expose map.M to outside package
*/

import (
	"fmt"
	"strings"
)

// Tokenizer is an interface for key manipulators
type Tokenizer interface {
	Separator() string
	Token([]string) string
	Keys(string) []string
}

// StringTokenizer is a Tokenizer using the string as separator
type StringTokenizer string

// Token returns a token joining all the keys with s as separator
func (s StringTokenizer) Token(ks []string) string { return strings.Join(ks, string(s)) }

// Keys returns the keys contained in the received token
func (s StringTokenizer) Keys(ks string) []string { return strings.Split(ks, string(s)) }

// Separator returns the separator
func (s StringTokenizer) Separator() string { return string(s) }

// DefaultTokenizer is a tokenizer using a dot or fullstop
var DefaultTokenizer = StringTokenizer(".")

// Flatten takes a hierarchy and flatten it using the tokenizer supplied
func Flatten(m map[string]interface{}, tokenizer Tokenizer) (*Map, error) {
	result, err := newMap(tokenizer)
	if err != nil {
		return nil, err
	}
	flatten(m, []string{}, func(ks []string, v interface{}) {
		result.M[tokenizer.Token(ks)] = v
	})
	return result, nil
}

type updateFunc func([]string, interface{})

func flatten(i interface{}, ks []string, update updateFunc) {
	switch v := i.(type) {
	case map[string]interface{}:
		flattenMap(v, ks, update)
	case []interface{}:
		flattenSlice(v, ks, update)
	default:
		update(ks, v)
	}
}

func flattenMap(m map[string]interface{}, ks []string, update updateFunc) {
	for k, v := range m {
		flatten(v, append(ks, k), update)
	}
}

func flattenSlice(vs []interface{}, ks []string, update updateFunc) {
	update(append(ks, "#"), len(vs))
	for i, v := range vs {
		flatten(v, append(ks, fmt.Sprintf("%d", i)), update)
	}
}
