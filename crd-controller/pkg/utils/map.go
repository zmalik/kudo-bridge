package utils

/*
This files uses source code from the files map.go and flatmap.go to expose map interface.
https://github.com/devopsfaith/flatmap

//TODO open issue/PR to expose map.M to outside package
*/

import (
	"regexp"
	"strconv"
	"strings"
)

var defaultCollectionPattern = regexp.MustCompile(`\.\*\.`)

func newMap(t Tokenizer) (*Map, error) {
	sep := t.Separator()
	var hasWildcard *regexp.Regexp
	var err error
	if sep == "." {
		hasWildcard = defaultCollectionPattern
	} else {
		hasWildcard, err = regexp.Compile(sep + `\*` + sep)
	}
	if err != nil {
		return nil, err
	}
	return &Map{
		M:  make(map[string]interface{}),
		t:  t,
		re: hasWildcard,
	}, nil
}

// Map is a flatten map
type Map struct {
	M  map[string]interface{}
	t  Tokenizer
	re *regexp.Regexp
}

// Move makes changes in the flatten hierarchy moving contents from origin to newKey
func (m *Map) Move(original, newKey string) {
	if v, ok := m.M[original]; ok {
		m.M[newKey] = v
		delete(m.M, original)
		return
	}

	if m.re.MatchString(original) {
		m.moveSliceAttribute(original, newKey)
		return
	}

	sep := m.t.Separator()

	for k := range m.M {
		if !strings.HasPrefix(k, original) {
			continue
		}

		if k[len(original):len(original)+1] != sep {
			continue
		}

		m.M[newKey+sep+k[len(original)+1:]] = m.M[k]
		delete(m.M, k)
	}
}

// Del deletes a key out of the map with the given prefix
func (m *Map) Del(prefix string) {
	if _, ok := m.M[prefix]; ok {
		delete(m.M, prefix)
		return
	}

	if m.re.MatchString(prefix) {
		m.delSliceAttribute(prefix)
		return
	}

	sep := m.t.Separator()

	for k := range m.M {
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		if k[len(prefix):len(prefix)+1] != sep {
			continue
		}

		delete(m.M, k)
	}
}

func (m *Map) delSliceAttribute(prefix string) {
	i := strings.Index(prefix, "*")
	sep := m.t.Separator()
	prefixRemainder := prefix[i+1:]
	recursive := strings.Index(prefixRemainder, "*") > -1 //nolint:gosimple

	for k := range m.M {
		if len(k) < i+2 {
			continue
		}

		if !strings.HasPrefix(k, prefix[:i]) {
			continue
		}

		if recursive {
			// TODO: avoid recursive calls by managing nested collections in a single key evaluation
			newPref := k[:i+1+strings.Index(k[i+1:], sep)] + prefixRemainder
			m.Del(newPref)
			continue
		}

		keyRemainder := k[i+1+strings.Index(k[i+1:], sep):]
		if keyRemainder == prefixRemainder {
			delete(m.M, k)
			continue
		}

		if !strings.HasPrefix(keyRemainder, prefixRemainder+sep) {
			continue
		}

		delete(m.M, k)
	}
}

func (m *Map) moveSliceAttribute(original, newKey string) {
	i := strings.Index(original, "*")
	sep := m.t.Separator()
	originalRemainder := original[i+1:]
	recursive := strings.Index(originalRemainder, "*") > -1 //nolint:gosimple

	newKeyOffset := strings.Index(newKey, "*")
	newKeyRemainder := newKey[newKeyOffset+1:]
	newKeyPrefix := newKey[:newKeyOffset]

	for k := range m.M {
		if len(k) <= i+2 {
			continue
		}

		if !strings.HasPrefix(k, original[:i]) {
			continue
		}

		remainder := k[i:]
		idLen := strings.Index(remainder, sep)
		cleanRemainder := k[i+idLen:]
		keyPrefix := newKeyPrefix + k[i:i+idLen]

		if recursive {
			// TODO: avoid recursive calls by managing nested collections in a single key evaluation
			m.Move(k[:i+idLen]+originalRemainder, keyPrefix+newKeyRemainder)
			continue
		}

		if cleanRemainder == originalRemainder[1:] {
			m.M[keyPrefix+newKeyRemainder] = m.M[k]
			delete(m.M, k)
			continue
		}

		rPrefix := originalRemainder[1:] + sep

		if cleanRemainder != sep+originalRemainder[1:] && !strings.HasPrefix(cleanRemainder, sep+rPrefix) {
			continue
		}

		m.M[keyPrefix+newKeyRemainder+cleanRemainder[len(rPrefix):]] = m.M[k]
		delete(m.M, k)
	}
}

// Expand expands the Map into a more complex structure. This is the reverse of the Flatten operation.
func (m *Map) Expand() map[string]interface{} {
	res := map[string]interface{}{}
	hasCollections := false
	for k, v := range m.M {
		ks := m.t.Keys(k)
		tr := res

		if ks[len(ks)-1] == "#" {
			hasCollections = true
		}
		for _, tk := range ks[:len(ks)-1] {
			trnew, ok := tr[tk]
			if !ok {
				trnew = make(map[string]interface{})
				tr[tk] = trnew
			}
			tr = trnew.(map[string]interface{})
		}
		tr[ks[len(ks)-1]] = v
	}

	if !hasCollections {
		return res
	}

	return m.expandNestedCollections(res).(map[string]interface{})
}

func (m *Map) expandNestedCollections(original map[string]interface{}) interface{} {
	for k, v := range original {
		if t, ok := v.(map[string]interface{}); ok {
			original[k] = m.expandNestedCollections(t)
		}
	}

	size, ok := original["#"]
	if !ok {
		return original
	}

	col := make([]interface{}, size.(int))
	for k := range col {
		col[k] = original[strconv.Itoa(k)]
	}
	return col
}
