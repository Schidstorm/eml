// Address parsing

package eml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Schidstorm/eml/decoder"
)

type Address interface {
	String() string
	Name() string
	Email() string
}

type MailboxAddr struct {
	name   string
	local  string
	domain string
}

func (ma MailboxAddr) Name() string {
	if ma.name == "" {
		return fmt.Sprintf("%s@%s", ma.local, ma.domain)
	}
	return ma.name
}

func (ma MailboxAddr) String() string {
	if ma.name == "" {
		return fmt.Sprintf("%s@%s", ma.local, ma.domain)
	}
	return fmt.Sprintf("%s <%s@%s>", ma.name, ma.local, ma.domain)
}

func (ma MailboxAddr) Email() string {
	return fmt.Sprintf("%s@%s", ma.local, ma.domain)
}

type GroupAddr struct {
	name  string
	boxes []MailboxAddr
}

func (ga GroupAddr) Name() string {
	return ga.name
}

func (ga GroupAddr) String() string {
	return ""
}

func (ga GroupAddr) Email() string {
	return ""
}

type DecodedAddress struct {
	name   string
	email  string
	string string
}

func CreateDecodedAddress(addr Address) DecodedAddress {

	return DecodedAddress{
		name:   tryDecode(addr.Name()),
		email:  tryDecode(addr.Email()),
		string: tryDecode(addr.String()),
	}
}

func tryDecode(in string) string {
	if res, err := decoder.Parse([]byte(in)); err == nil {
		return string(res)
	}
	return in
}

func (da DecodedAddress) Name() string {
	return da.name
}

func (da DecodedAddress) String() string {
	return da.string
}

func (da DecodedAddress) Email() string {
	return da.email
}

func ParseAddress(bs []byte) (Address, error) {
	toks, err := tokenize(bs)
	if err != nil {
		return nil, err
	}
	return parseAddress(toks)
}

func parseAddress(toks []token) (Address, error) {
	// If this is a group, it must end in a ";" token.
	ltok := toks[len(toks)-1]
	if len(ltok) == 1 && ltok[0] == ';' {
		ga := GroupAddr{}
		// we split on ':'
		nts, rest, err := splitOn(toks, []byte{':'})
		if err != nil {
			return nil, err
		}
		for _, nt := range nts {
			ga.name += string(nt) + " "
		}
		ga.name = strings.TrimSpace(ga.name)
		ga.boxes = []MailboxAddr{}

		last := 0
		something := false
		for i, t := range rest {
			if len(t) == 1 && (t[0] == ',' || t[0] == ';') && something {
				ma, err := parseMailboxAddr(rest[last:i])
				if err != nil {
					return nil, err
				}
				ga.boxes = append(ga.boxes, ma)
				last = i + 1
			}
			something = true
		}
		return CreateDecodedAddress(ga), nil
	}

	addr, err := parseMailboxAddr(toks)
	return CreateDecodedAddress(addr), err
}

func splitOn(ts []token, s token) ([]token, []token, error) {
	for i, t := range ts {
		if string(t) == string(s) {
			return ts[:i], ts[i+1:], nil
		}
	}
	return nil, nil, errors.New("split token not found")
}

func parseMailboxAddr(ts []token) (ma MailboxAddr, err error) {
	// We're either name-addr or an addr-spec. If we end in ">", then all
	// characters up to "<" constitute the name. Otherwise, there is no
	// name.
	ma = MailboxAddr{}
	ltok := ts[len(ts)-1]
	if len(ltok) == 1 && ltok[0] == '>' {
		var nts, ats []token
		nts, ats, err = splitOn(ts, []byte{'<'})
		if err != nil {
			return
		}
		for _, nt := range nts {
			ma.name += string(nt) + " "
		}
		ma.name = strings.TrimSpace(ma.name)
		ma.local, ma.domain, err = parseSimpleAddr(ats[:len(ats)-1])
		return
	}
	ma.local, ma.domain, err = parseSimpleAddr(ts)
	return
}

func parseSimpleAddr(ts []token) (l, d string, e error) {
	// The second token must be '@' - all further tokens are stuck in the domain.
	l = string(ts[0])
	if !(len(ts[1]) == 1 && ts[1][0] == '@') {
		return "", "", errors.New("invalid simpleAddr")
	}
	for _, dp := range ts[2:] {
		d += string(dp) + " "
	}
	d = strings.TrimSpace(d)
	return
}
