package decoder

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"

	"encoding/base64"
	"mime/quotedprintable"

	"github.com/paulrosania/go-charset/charset"
)

func UTF8(cs string, data []byte) ([]byte, error) {
	if strings.ToUpper(cs) == "UTF-8" {
		return data, nil
	}

	r, err := charset.NewReader(cs, bytes.NewReader(data))
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(r)

}

func Parse(bstr []byte) ([]byte, error) {
	result, err := forEncodedParts(string(bstr), func(encodingName, encodingType, encodingContent string) (string, error) {
		res, err := Decode(encodingName, encodingType, encodingContent)
		return string(res), err
	})
	return []byte(result), err

}

type encodingState int

const (
	stateNone = encodingState(iota)
	stateBegin
	stateName
	stateType
	stateContent
	stateEnd
)

func forEncodedParts(input string, cb func(encodingName, encodingType, encodingContent string) (string, error)) (string, error) {
	state := stateNone
	result := strings.Builder{}
	tmp := []rune{}
	var name = ""
	var typeString = ""
	var content = ""
	for i, c := range input {
		switch state {
		case stateNone:
			if c == '=' {
				state = stateBegin
			} else {
				result.WriteRune(c)
			}
		case stateBegin:
			if c == '?' {
				state = stateName
			} else {
				result.WriteRune('=')
				result.WriteRune(c)
				state = stateNone
			}
		case stateName:
			if c == '?' {
				name = string(tmp)
				tmp = tmp[:0]
				state = stateType
			} else {
				tmp = append(tmp, c)
			}
		case stateType:
			if c == '?' {
				typeString = string(tmp)
				tmp = tmp[:0]
				state = stateContent
			} else {
				tmp = append(tmp, c)
			}
		case stateContent:
			if c == '?' && input[i+1] == '=' {
				content = string(tmp)
				tmp = tmp[:0]
				state = stateEnd
			} else {
				tmp = append(tmp, c)
			}
		case stateEnd:
			part, err := cb(name, typeString, content)
			if err != nil {
				return "", err
			}
			result.WriteString(part)
			state = stateNone
		}
	}

	if state != stateNone {
		return result.String(), errors.New("invalid encoding format")
	}

	return result.String(), nil
}

func Decode(encodingName, encodingType, encodingContent string) ([]byte, error) {
	var err error
	var decoded []byte
	switch strings.ToUpper(encodingType) {
	case "Q":
		decoded, err = ioutil.ReadAll(quotedprintable.NewReader(bytes.NewReader([]byte(encodingContent))))
	case "B":
		decoded, err = base64.StdEncoding.DecodeString(encodingContent)
	default:
		return nil, errors.New("missing encoding type")
	}

	if err != nil {
		return nil, err
	}

	return UTF8(encodingName, decoded)
}
