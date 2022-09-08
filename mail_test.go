package eml

import (
	"reflect"
	"strings"
	"testing"
)

// Converts all newlines to CRLFs.
func crlf(s string) []byte {
	return []byte(strings.Replace(s, "\n", "\r\n", -1))

}

type parseRawTest struct {
	msg []byte
	ret RawMessage
}

var parseRawTests = []parseRawTest{
	{
		msg: crlf(`
`),
		ret: RawMessage{
			RawHeaders: []RawHeader{},
			Body:       crlf(""),
		},
	},
	{
		msg: crlf(`
ab
c
`),
		ret: RawMessage{
			RawHeaders: []RawHeader{},
			Body: crlf(`ab
c
`),
		},
	},
	{
		msg: crlf(`a: b

`),
		ret: RawMessage{
			RawHeaders: []RawHeader{{crlf("a"), crlf("b")}},
			Body:       crlf(""),
		},
	},
	{
		msg: crlf(`a: b
c: def
 hi

`),
		ret: RawMessage{
			RawHeaders: []RawHeader{
				{crlf("a"), crlf("b")},
				{crlf("c"), crlf("def hi")},
			},
			Body: crlf(``),
		},
	},
	{
		msg: crlf(`a: b
c: d fdsa
ef:  as

hello, world
`),
		ret: RawMessage{
			RawHeaders: []RawHeader{
				{crlf("a"), crlf("b")},
				{crlf("c"), crlf("d fdsa")},
				{crlf("ef"), crlf("as")},
			},
			Body: crlf(`hello, world
`),
		},
	},
	{
		msg: []byte(`a: b
c: d fdsa
ef:  as

hello, world
`),
		ret: RawMessage{
			RawHeaders: []RawHeader{
				{[]byte("a"), []byte("b")},
				{[]byte("c"), []byte("d fdsa")},
				{[]byte("ef"), []byte("as")},
			},
			Body: []byte(`hello, world
`),
		},
	},
}

func TestParseRaw(t *testing.T) {
	for _, pt := range parseRawTests {
		msg := pt.msg
		ret := pt.ret
		act, err := ParseRaw(msg)
		if err != nil {
			t.Errorf("ParseRaw returned error for %#v", string(msg))
		} else if !reflect.DeepEqual(act, ret) {
			t.Errorf("ParseRaw: incorrect result from %#v as %#v; expected %#v", string(msg), act, ret)
		}
	}
}

type processTest struct {
	name string
	raw  RawMessage
	ret  Message
}

var processTests = []processTest{}

func TestProcess(t *testing.T) {
	for _, pt := range processTests {
		act, err := Process(pt.raw)
		if err != nil {
			t.Errorf("Parse returned error for %s", pt.name)
		} else if !reflect.DeepEqual(act, pt.ret) {
			t.Errorf("Parse: incorrect result from %#v as %#v; expected %#v", pt.name, act, pt.ret)
		}
	}
}

type parseTest struct {
	msg []byte
	ret Message
}

var parseTests = []parseTest{
	{
		crlf(`

`),
		Message{
			HeaderInfo: HeaderInfo{
				FullHeaders: HeaderList{},
			},
			Text: "\r\n",
			Parts: []Part{
				{
					Type:    "text/plain",
					Charset: "",
					Data:    []uint8{0xd, 0xa},
				},
			},
		},
	},
	{
		crlf(`Subject: Hello, world

G'day, mate.
`),
		Message{
			HeaderInfo: HeaderInfo{
				FullHeaders: HeaderList{"Subject": {"Hello, world"}},
			},
			Text: "G'day, mate.\r\n",
			Parts: []Part{
				{
					Type:    "text/plain",
					Charset: "",
					Data:    []byte("G'day, mate.\r\n"),
				},
			},
		},
	},
	{
		crlf(`Subject: =?UTF-8?Q?german_=C3=BC_=26_=26_=2E?=

G'day, mate.
`),
		Message{
			HeaderInfo: HeaderInfo{
				FullHeaders: HeaderList{"Subject": {"german_ü_&_&_."}},
			},
			Text: "G'day, mate.\r\n",
			Parts: []Part{
				{
					Type:    "text/plain",
					Charset: "",
					Data:    []byte("G'day, mate.\r\n"),
				},
			},
		},
	},

	{
		crlf(`Subject: =?UTF-8?Q?german_=C3=BC_=26_=26_=2E?=
 =?UTF-8?Q?german_=C3=BC_=26_=26_=2E?=

G'day, mate.
`),
		Message{
			HeaderInfo: HeaderInfo{
				FullHeaders: HeaderList{"Subject": {"german_ü_&_&_. german_ü_&_&_."}},
			},
			Text: "G'day, mate.\r\n",
			Parts: []Part{
				{
					Type:    "text/plain",
					Charset: "",
					Data:    []byte("G'day, mate.\r\n"),
				},
			},
		},
	},
	{
		crlf(`Subject: Hello, world
Content-Type: text/plain
Content-Transfer-Encoding: base64

VGhpcyBpcyBhIHRlc3QgaW4gYmFzZTY0
VGhpcyBpcyBhIHRlc3QgaW4gYmFzZTY0
`),
		Message{
			HeaderInfo: HeaderInfo{
				FullHeaders: HeaderList{
					"Subject":                   {"Hello, world"},
					"Content-Type":              {"text/plain"},
					"Content-Transfer-Encoding": {"base64"},
				},
			},
			Text: "This is a test in base64This is a test in base64",
			Parts: []Part{
				{
					Type:    "text/plain",
					Charset: "",
					Data:    []byte("This is a test in base64This is a test in base64"),
				},
			},
		},
	},

	{
		crlf(`Content-Type: multipart/alternative; boundary="_----------=_MCPart_418513213"

--_----------=_MCPart_418513213
Content-Type: text/plain

Some text.
--_----------=_MCPart_418513213
Content-Type: text/plain

Some text.
--_----------=_MCPart_418513213
`),
		Message{
			HeaderInfo: HeaderInfo{
				FullHeaders: HeaderList{
					"Content-Type": {"multipart/alternative; boundary=\"_----------=_MCPart_418513213\""},
				},
			},
			Text: "Some text.",
			Parts: []Part{
				{
					"text/plain",
					"UTF-8",
					[]byte("Some text."),
					map[string][]string{
						"Content-Type": {
							"text/plain",
						},
					},
				},
				{
					"text/plain",
					"UTF-8",
					[]byte("Some text."),
					map[string][]string{
						"Content-Type": {
							"text/plain",
						},
					},
				},
			},
		},
	},
}

func TestParse(t *testing.T) {
	for _, pt := range parseTests {
		msg := pt.msg
		ret := pt.ret
		act, err := Parse(msg)
		if err != nil {
			t.Errorf("Parse returned error for %#v\n", string(msg))
			t.Errorf("Error: %s", err.Error())
		} else if !reflect.DeepEqual(act, ret) {
			t.Errorf("Parse: incorrect result from %#v \nas\n %#v; \nexpected\n %#v", string(msg), act, ret)
		}
	}
}
