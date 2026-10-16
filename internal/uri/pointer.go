package uri

import (
	"fmt"
	"strings"
)

// ParsePointer parses a URI fragment (without '#') as a JSON Pointer. Each
// token is percent-decoded first, then ~1 and ~0 are unescaped, in that order.
func ParsePointer(frag string) ([]string, error) {
	if err := ValidatePercent(frag); err != nil {
		return nil, err
	}
	if frag == "" {
		return nil, nil
	}
	if frag[0] != '/' {
		return nil, fmt.Errorf("JSON Pointer %q must start with '/'", frag)
	}
	toks := strings.Split(frag[1:], "/")
	for i, t := range toks {
		var b strings.Builder
		for j := 0; j < len(t); j++ {
			if t[j] == '%' {
				b.WriteByte(unhex(t[j+1])<<4 | unhex(t[j+2]))
				j += 2
			} else {
				b.WriteByte(t[j])
			}
		}
		d := b.String()
		for k := 0; k < len(d); k++ {
			if d[k] == '~' && (k+1 >= len(d) || (d[k+1] != '0' && d[k+1] != '1')) {
				return nil, fmt.Errorf("invalid '~' escape in JSON Pointer %q", frag)
			}
		}
		toks[i] = strings.ReplaceAll(strings.ReplaceAll(d, "~1", "/"), "~0", "~")
	}
	return toks, nil
}
