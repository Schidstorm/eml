package eml

import (
	"reflect"
	"testing"
)

type tokenTest struct {
	s string
	t []string
}

var tokenTests = []tokenTest{
	{``, []string{}},
	{`a`, []string{`a`}},
	{`af&' al43`, []string{`af&'`, `al43`}},
	{
		`"Joe Q. Public" <john.q.public@example.com>`,
		[]string{`"Joe Q. Public"`, `<`, `john.q.public`, `@`, `example.com`, `>`},
	},
	{
		`"Giant; \"Big\" Box" <sysservices@example.net>`,
		[]string{
			`"Giant; \"Big\" Box"`,
			`<`,
			`sysservices`,
			`@`,
			`example.net`,
			`>`,
		},
	},
}

func TestTokenize(t *testing.T) {
	for _, tt := range tokenTests {
		o, e := tokenize([]byte(tt.s))
		if e != nil {
			t.Errorf("tokenize returned error for %#v", tt.s)
		} else {
			rt := []string{}
			for _, tok := range o {
				rt = append(rt, string(tok))
			}
			if !reflect.DeepEqual(rt, tt.t) {
				t.Errorf("tokenize(%#v) gave %#v; expected %#v", tt.s, rt, tt.t)
			}
		}
	}
}
