package eml

import (
	"reflect"
	"testing"
)

type parseAddressListTest struct {
	ins []byte
	out []Address
}

var parseAddressListTests = []parseAddressListTest{
	{
		[]byte(``),
		[]Address{},
	},
	{
		[]byte(`"Joe Q. Public" <john.q.public@example.com>`),
		[]Address{
			MailboxAddr{`"Joe Q. Public"`, `john.q.public`, `example.com`},
		},
	},
	{
		[]byte(`"Joe Q. Public" <john.q.public@example.com>, <boss@nil.test>`),
		[]Address{
			MailboxAddr{`"Joe Q. Public"`, `john.q.public`, `example.com`},
			MailboxAddr{``, `boss`, `nil.test`},
		},
	},
}

func TestParseAddressList(t *testing.T) {
	for _, pt := range parseAddressListTests {
		o, e := parseAddressList(pt.ins)
		if e != nil {
			t.Errorf("parseAddressList returned error for %#v", pt.ins)
		} else if !reflect.DeepEqual(o, pt.out) {
			t.Errorf(
				"parseAddressList: incorrect result for %#v: %#v vs. %#v",
				string(pt.ins), o, pt.out)
		}
	}
}
