package eml

import (
	"fmt"
	"strings"
	"time"

	"github.com/Schidstorm/eml/decoder"
)

type HeaderList map[string][]string

func (h HeaderList) FirstByKey(key string) (string, bool) {
	if list, ok := h[key]; ok && len(list) > 0 {
		return list[0], true
	}
	return "", false
}

func (h HeaderList) Add(key, value string) {
	if _, ok := h[key]; !ok {
		h[key] = []string{}
	}

	h[key] = append(h[key], value)
}

func (h HeaderList) Set(key string, value []string) {
	h[key] = value
}

func (h HeaderList) ContentType() string {
	if header, ok := h.FirstByKey("Content-Type"); ok {
		return header
	}
	return "text/plain"
}

func (h HeaderList) MessageId() string {
	if header, ok := h.FirstByKey("Message-ID"); ok {
		return strings.Trim(header, `<>`)
	}
	return ""
}

func (h HeaderList) InReply() []string {
	if header, ok := h.FirstByKey("In-Reply-To"); ok {
		ids := strings.Fields(string(header))
		var inReply []string
		for _, id := range ids {
			inReply = append(inReply, strings.Trim(id, `<> `))
		}
		return inReply
	}
	return nil
}

func (h HeaderList) References() []string {
	if header, ok := h.FirstByKey("References"); ok {
		ids := strings.Fields(string(header))
		var references []string
		for _, id := range ids {
			references = append(references, strings.Trim(id, `<> `))
		}
		return references
	}
	return nil
}

func (h HeaderList) Date() time.Time {
	if header, ok := h.FirstByKey("Date"); ok {
		return ParseDate(header)
	}
	return time.Unix(0, 0)
}

func (h HeaderList) From() []Address {
	if header, ok := h.FirstByKey("From"); ok {
		if l, err := parseAddressList([]byte(header)); err == nil {
			return l
		}
	}
	return nil
}

func (h HeaderList) Sender() Address {
	if header, ok := h.FirstByKey("Sender"); ok {
		if l, err := ParseAddress([]byte(header)); err == nil {
			if l == nil {
				from := h.From()
				if len(from) > 0 {
					return from[0]
				}
			}
			return l
		}
	}
	return nil
}

func (h HeaderList) ReplyTo() []Address {
	if header, ok := h.FirstByKey("Reply-To"); ok {
		if l, err := parseAddressList([]byte(header)); err == nil {
			return l
		}
	}
	return nil
}

func (h HeaderList) To() []Address {
	if header, ok := h.FirstByKey("To"); ok {
		if l, err := parseAddressList([]byte(header)); err == nil {
			return l
		}
	}
	return nil
}

func (h HeaderList) Cc() []Address {
	if header, ok := h.FirstByKey("Cc"); ok {
		if l, err := parseAddressList([]byte(header)); err == nil {
			return l
		}
	}
	return nil
}

func (h HeaderList) Bcc() []Address {
	if header, ok := h.FirstByKey("Bcc"); ok {
		if l, err := parseAddressList([]byte(header)); err == nil {
			return l
		}
	}
	return nil
}

func (h HeaderList) Subject() string {
	if header, ok := h.FirstByKey("Subject"); ok {
		subject, err := decoder.Parse([]byte(header))
		if err != nil {
			fmt.Println("Failed decode subject", err)
		}
		return string(subject)
	}
	return ""
}

func (h HeaderList) Comments() []string {
	if header, ok := h.FirstByKey("Comments"); ok {
		return []string{header}
	}
	return nil
}

func (h HeaderList) Keywords() []string {
	if header, ok := h.FirstByKey("Keywords"); ok {
		ks := strings.Split(header, ",")
		var keywords []string
		for _, k := range ks {
			keywords = append(keywords, strings.TrimSpace(k))
		}
		return keywords
	}
	return nil
}
