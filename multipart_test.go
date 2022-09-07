package eml

import (
	"reflect"
	"testing"
)

type parseBodyTest struct {
	ct   string
	body []byte
	rps  []Part
}

var parseBodyTests = []parseBodyTest{
	{
		ct:   "text/plain; charset=UTF-8",
		body: []byte(`This is some text.`),
		rps: []Part{
			{"text/plain; charset=UTF-8", "UTF-8", []byte("This is some text."), nil},
		},
	},
	{
		ct: "multipart/alternative; boundary=90e6ba1efd30b0013a04b8d4970f",
		body: []byte(`--90e6ba1efd30b0013a04b8d4970f
Content-Type: text/plain; charset=ISO-8859-1

Some text.
--90e6ba1efd30b0013a04b8d4970f
Content-Type: text/html; charset=ISO-8859-1
Content-Transfer-Encoding: quoted-printable

Some other text.
--90e6ba1efd30b0013a04b8d4970f--
`),
		rps: []Part{
			{
				"text/plain; charset=ISO-8859-1",
				"ISO-8859-1",
				[]byte("Some text."),
				map[string][]string{
					"Content-Type": {
						"text/plain; charset=ISO-8859-1",
					},
				},
			},
			{
				"text/html; charset=ISO-8859-1",
				"ISO-8859-1",
				[]byte("Some other text."),
				map[string][]string{
					"Content-Type": {
						"text/html; charset=ISO-8859-1",
					},
					"Content-Transfer-Encoding": {
						"quoted-printable",
					},
				},
			},
		},
	},
}

func TestParseBody(t *testing.T) {
	for _, pt := range parseBodyTests {
		parts, e := parseMultipartBody(pt.ct, pt.body)
		if e != nil {
			t.Errorf("parseBody returned error for %#v: %#v", pt, e)
		} else if !reflect.DeepEqual(parts, pt.rps) {
			t.Errorf(
				"parseBody: incorrect result for %#v: \n%#v\nvs.\n%#v",
				pt, parts, pt.rps)
		}
	}
}
