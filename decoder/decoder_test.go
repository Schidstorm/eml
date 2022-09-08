package decoder

import (
	"errors"
	"testing"
)

type testCase struct {
	actual        string
	expected      string
	expectedError error
}

var testCases = []testCase{
	{"", "", nil},
	{"text without encoding.", "text without encoding.", nil},
	{"=text with = equals", "=text with = equals", nil},
	{"=?text with invalid content", "", errors.New("invalid encoding format")},
	{"text with invalid=?content", "text with invalid", errors.New("invalid encoding format")},
	{"=?UTF-8?Q?german_=C3=BC_=26_=26_=2E?=", "german_ü_&_&_.", nil},
}

func stringOrNullString(e error) string {
	if e == nil {
		return "nil"
	}
	return e.Error()
}

func TestParse(t *testing.T) {
	for _, pt := range testCases {
		actual, err := Parse([]byte(pt.actual))
		if stringOrNullString(err) != stringOrNullString(pt.expectedError) {
			t.Errorf("unexpected error.\nactual: %s\n expected:%s\n", stringOrNullString(err), stringOrNullString(pt.expectedError))
			continue
		}

		if string(actual) != pt.expected {
			t.Errorf("unexpected result.\nactual: %s\n expected:%s\n", string(actual), pt.expected)
			continue
		}
	}
}
