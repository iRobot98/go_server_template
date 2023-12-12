package base

import (
	"fmt"
	"regexp"
	"strings"
)

// Trim:  " .`'-?!" then\
// Regexp match rgex
func ValidateString(strin string, rgex string) (str string, err error) {

	nomen := strings.Trim(strin, " .`'-?!")
	if len(nomen) == 0 {
		return
	}

	ok, err := regexp.MatchString(rgex, nomen)

	if !ok || err != nil {
		if err == nil && !ok {
			err = fmt.Errorf("didn't match string")
		}
		return
	}

	str = nomen
	return
}

type Str struct {
	s string
}

// Replaces the underlying string with a new string
func (s *Str) Swap(s1 string) *Str {
	s.s = s1
	return s
}

// Returns New Str object from string
func NewStr(s string) (S Str) {
	S.s = s
	return
}

// Returns New Str object from formatted string
func NewStrF(s string, vars ...any) (S Str) {
	S.s = fmt.Sprintf(s, vars...)
	return
}

// Returns length of string
func (s *Str) Len() int {
	return len(s.s)
}

// Returns true if Str.s contains s1
func (s *Str) Contains(substr string) bool {
	return strings.Contains(s.s, substr)
}

// Returns true if Str.s starts with s1
func (s *Str) StartsWith(s1 string) bool {
	l1 := len(s1)
	l := len(s.s)
	if l < 1 || l1 > l {
		return false
	}
	if l1 < 1 {
		return true
	}

	return s.s[0:l1-1] == s1[0:l1-1]
}

// Returns true if Str.s contains s1.s
func (s *Str) StrContains(s1 Str) bool {
	return s.Contains(s1.s)
}

// Returns true if Str.s starts with s1.s
func (s *Str) StrStartsWith(s1 Str) bool {
	return s.StartsWith(s1.s)
}

// validates if Str.s matches regex, or error
// assume valid string returns nil error
func (s *Str) Validate(regex string) (err error) {
	s.s, err = ValidateString(s.s, regex)
	return
}

// Accesses underlying string
func (s *Str) Str() *string {
	return &s.s
}

// Returns copy of underlying string
func (s *Str) String() string {
	return s.s
}

// Returns a lower case copy of Str.s
func (s *Str) ToLower() string {
	return strings.ToLower(s.s)
}

// Returns a new Str in lower case
func (s *Str) ToLowerStr() Str {
	return NewStr(s.ToLower())
}

// Returns a new Str in lower case
func (s *Str) LowerCase() *Str {
	s.s = strings.ToLower(s.s)
	return s
}

// Returns a string array consisting of s split by 'spl'
func (s *Str) Split(spl string) []string {
	return strings.Split(s.s, spl)
}

// Returns a pointer to the Str object after trimming
func (s *Str) Trim() *Str {

	s.s = strings.Trim(s.s, " \r\n\t")
	return s
}

// Returns a channel of the runes in the underlying string
// to range over
func (s *Str) Map() <-chan rune {
	c := make(chan rune, len(s.s))
	for _, ch := range s.s {
		c <- ch
	}
	close(c)
	return c
}

func (s *Str) EndsWith(sub string) bool {
	l1 := s.Len()
	l2 := len(sub)
	if l2 > l1 {
		return false
	}
	if (l2 - l1 - 1) < 0 {
		return false
	}
	return s.s[l2-l1-1:] == sub
}
