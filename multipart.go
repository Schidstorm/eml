// Handle multipart messages.

package eml

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"regexp"
)

type Part struct {
	Type    string
	Charset string
	Data    []byte
	Headers map[string][]string
}

// Parse the body of a message, using the given content-type. If the content
// type is multipart, the parts slice will contain an entry for each part
// present; otherwise, it will contain a single entry, with the entire (raw)
// message contents.
func parseBody(ct string, body []byte) (parts []Part, err error) {
	_, ps, err := mime.ParseMediaType(ct)
	if err != nil {
		return
	}

	boundary, ok := ps["boundary"]
	if !ok {
		parts = append(parts, Part{ct, ps["charset"], body, nil})
		return
	}
	r := multipart.NewReader(bytes.NewReader(body), boundary)
	p, err := r.NextRawPart()
	for err != io.EOF {
		data, _ := ioutil.ReadAll(decodeByTransferEncoding(p, p.Header.Get("Content-Transfer-Encoding"))) // ignore error

		var subparts []Part
		subparts, err = parseBody(p.Header["Content-Type"][0], data)
		for i := range subparts {
			subparts[i].Headers = p.Header
		}

		//if err == nil then body have sub multipart, and append him
		if err == nil {
			parts = append(parts, subparts...)
		} else {
			contenttype := regexp.MustCompile("(?is)charset=(.*)").FindStringSubmatch(p.Header["Content-Type"][0])
			charset := "UTF-8"
			if len(contenttype) > 1 {
				charset = contenttype[1]
			}
			part := Part{p.Header["Content-Type"][0], charset, data, p.Header}
			parts = append(parts, part)
		}
		p, err = r.NextRawPart()
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func decodeByTransferEncoding(reader io.Reader, transferEncoding string) io.Reader {
	const cte = "Content-Transfer-Encoding"
	switch transferEncoding {
	case "quoted-printable":
		return quotedprintable.NewReader(reader)
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, reader)
	default:
		return reader
	}
}
