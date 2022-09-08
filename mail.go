// Package mail implements a parser for electronic mail messages as specified
// in RFC5322.
//
// We allow both CRLF and LF to be used in the input, possibly mixed.
package eml

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"regexp"
	"strings"

	"github.com/Schidstorm/eml/decoder"
)

type HeaderInfo struct {
	FullHeaders HeaderList // all headers

}

type Message struct {
	HeaderInfo
	Body        []byte
	Text        string
	Html        string
	Attachments []Attachment
	Parts       []Part
}

type Attachment struct {
	Filename string
	Data     []byte
}

type Header struct {
	Key, Value string
}

func Parse(s []byte) (m Message, e error) {
	r, e := ParseRaw(s)
	if e != nil {
		return
	}
	return Process(r)
}

func Process(r RawMessage) (m Message, e error) {
	m.FullHeaders = HeaderList{}
	for _, rh := range r.RawHeaders {
		v, err := decoder.Parse(rh.Value)
		if err != nil {
			v = rh.Value
		}
		m.FullHeaders.Add(string(rh.Key), string(v))
	}

	var parts []Part
	var er error

	// is multipart with base64 encoding valid?
	if m.FullHeaders.MediaType().Type == "multipart/alternative" {
		parts, er = parseMultipartBody(m.FullHeaders.ContentType(), r.Body)
		if er != nil {
			e = er
			return
		}
	} else {
		body := r.Body
		if hdr, ok := m.HeaderInfo.FullHeaders.FirstByKey("Content-Transfer-Encoding"); ok {
			body, er = decodeByTransferEncoding(body, hdr)
			if er != nil {
				e = er
				return
			}
		}

		_, ps, err := mime.ParseMediaType(m.HeaderInfo.FullHeaders.ContentType())
		if err != nil {
			e = err
			return
		}

		if charset, ok := ps["charset"]; ok && charset != "" {
			body = encodeData(body, charset)
		}

		parts = append(parts, Part{m.FullHeaders.ContentType(), ps["charset"], body, nil})
	}

	for _, part := range parts {
		switch {
		case strings.Contains(part.Type, "text/plain"):
			m.Text = string(part.Data)
		case strings.Contains(part.Type, "text/html"):
			m.Html = string(part.Data)

		default:
			if cd, ok := part.Headers["Content-Disposition"]; ok {
				if strings.Contains(cd[0], "attachment") {
					filename := regexp.MustCompile("(?msi)name=\"(.*?)\"").FindStringSubmatch(cd[0]) //.FindString(cd[0])
					if len(filename) < 2 {
						fmt.Println("failed get filename from header content-disposition")
						break
					}

					dfilename, err := decoder.Parse([]byte(filename[1]))
					if err != nil {
						fmt.Println("Failed decode filename of attachment", err)
					} else {
						filename[1] = string(dfilename)
					}

					m.Attachments = append(m.Attachments, Attachment{filename[1], part.Data})

				}
			}
		}
	}

	m.Parts = parts
	m.Text = string(parts[0].Data)
	return
}

func decodeByTransferEncoding(body []byte, transferEncoding string) ([]byte, error) {
	var reader io.Reader = bytes.NewBuffer(body)
	switch transferEncoding {
	case "quoted-printable":
		reader = quotedprintable.NewReader(reader)
	case "base64":
		reader = base64.NewDecoder(base64.StdEncoding, reader)
	}

	return io.ReadAll(reader)
}

func encodeData(data []byte, charset string) []byte {
	decodedData, err := decoder.UTF8(charset, data)
	if err != nil {
		return data
	} else {
		return decodedData
	}
}

type RawHeader struct {
	Key, Value []byte
}

type RawMessage struct {
	RawHeaders []RawHeader
	Body       []byte
}

func isWSP(b byte) bool {
	return b == ' ' || b == '\t'
}

func ParseRaw(s []byte) (m RawMessage, e error) {
	// parser states
	const (
		READY = iota
		HKEY
		HVWS
		HVAL
	)

	const (
		CR = '\r'
		LF = '\n'
	)
	CRLF := []byte{CR, LF}

	state := READY
	kstart, kend, vstart := 0, 0, 0
	done := false

	m.RawHeaders = []RawHeader{}

	for i := 0; i < len(s); i++ {
		b := s[i]
		switch state {
		case READY:
			if b == CR && i < len(s)-1 && s[i+1] == LF {
				// we are at the beginning of an empty header
				m.Body = s[i+2:]
				done = true
				goto Done
			}
			if b == LF {
				m.Body = s[i+1:]
				done = true
				goto Done
			}
			// otherwise this character is the first in a header
			// key
			kstart = i
			state = HKEY
		case HKEY:
			if b == ':' {
				kend = i
				state = HVWS
			}
		case HVWS:
			if !isWSP(b) {
				vstart = i
				state = HVAL
			}
		case HVAL:
			if b == CR && i < len(s)-2 && s[i+1] == LF && !isWSP(s[i+2]) {
				v := bytes.Replace(s[vstart:i], CRLF, nil, -1)
				hdr := RawHeader{s[kstart:kend], v}
				m.RawHeaders = append(m.RawHeaders, hdr)
				state = READY
				i++
			} else if b == LF && i < len(s)-1 && !isWSP(s[i+1]) {
				v := bytes.Replace(s[vstart:i], CRLF, nil, -1)
				hdr := RawHeader{s[kstart:kend], v}
				m.RawHeaders = append(m.RawHeaders, hdr)
				state = READY
			}
		}
	}
Done:
	if !done {
		e = errors.New("unexpected EOF")
	}
	return
}
