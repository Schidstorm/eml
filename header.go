// Header parsing functionality.

package eml

func split(ts []token, s token) [][]token {
	r, l := [][]token{}, 0
	for i, t := range ts {
		if string(t) == string(s) {
			r = append(r, ts[l:i])
			l = i + 1
		}
	}
	if l != len(ts) {
		r = append(r, ts[l:])
	}
	return r
}

// BUG: We don't currently support domain literals with commas.
func parseAddressList(s []byte) ([]Address, error) {
	al := []Address{}
	ts, e := tokenize(s)
	if e != nil {
		return al, e
	}
	tts := split(ts, []byte{','})
	for _, ts := range tts {
		a, e := parseAddress(ts)
		if e != nil {
			return al, e
		}
		al = append(al, a)
	}
	return al, nil
}
