package eml

import (
	"reflect"
	"testing"
)

// We use examples from RFC5322 as our test suite.

type parseAddressTest struct {
	addrStr string
	addrRes Address
}

var parseAddressTests = []parseAddressTest{
	{
		`"Joe Q. Public" <john.q.public@example.com>`,
		MailboxAddr{`"Joe Q. Public"`, `john.q.public`, `example.com`},
	},
	{
		`Mary Smith <mary@x.test>`,
		MailboxAddr{`Mary Smith`, `mary`, `x.test`},
	},
	{
		`jdoe@example.org`,
		MailboxAddr{``, `jdoe`, `example.org`},
	},
	{
		`Who? <one@y.test>`,
		MailboxAddr{`Who?`, `one`, `y.test`},
	},
	{
		`<boss@nil.test>`,
		MailboxAddr{``, `boss`, `nil.test`},
	},
	{
		`"Giant; \"Big\" Box" <sysservices@example.net>`,
		MailboxAddr{`"Giant; \"Big\" Box"`, `sysservices`, `example.net`},
	},
	{
		`Pete <pete@silly.example>`,
		MailboxAddr{`Pete`, `pete`, `silly.example`},
	},
	{
		`A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;`,
		GroupAddr{
			`A Group`,
			[]MailboxAddr{
				{`Ed Jones`, `c`, `a.test`},
				{``, `joe`, `where.test`},
				{`John`, `jdoe`, `one.test`},
			},
		},
	},
	{
		`Undisclosed recipients:;`,
		GroupAddr{`Undisclosed recipients`, []MailboxAddr{}},
	},
	{
		`Undisclosed recipients:      ;`,
		GroupAddr{`Undisclosed recipients`, []MailboxAddr{}},
	},
}

func TestParseAddress(t *testing.T) {
	for _, pt := range parseAddressTests {
		address, err := ParseAddress([]byte(pt.addrStr))
		if err != nil {
			t.Errorf("ParseAddress returned error for %#v", pt.addrStr)
		} else if !reflect.DeepEqual(address, pt.addrRes) {
			t.Errorf(
				"ParseAddress: incorrect result for %#v: gave %#v; expected %#v",
				pt.addrStr, address, pt.addrRes)
		}
	}
}
