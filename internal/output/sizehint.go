package output

import "github.com/neoface/jsonschema/internal/jsonvalue"

// valueSize estimates the compact JSON size of an annotation value without
// allocating (escaping is ignored, so it is a close lower bound).
func valueSize(v any) int {
	switch x := v.(type) {
	case nil:
		return 4
	case bool:
		if x {
			return 4
		}
		return 5
	case string:
		return len(x) + 2
	case *jsonvalue.Number:
		return len(x.String())
	case []string:
		n := 2
		for _, e := range x {
			n += len(e) + 3
		}
		return n
	case []any:
		n := 2
		for _, e := range x {
			n += valueSize(e) + 1
		}
		return n
	case map[string]any:
		n := 2
		for k, e := range x {
			n += len(k) + 4 + valueSize(e)
		}
		return n
	}
	return 0
}

// Fixed bytes around one basic unit's strings (keys, quotes, braces, comma).
const (
	errUnitOverhead = len(`{"instanceLocation":"","keywordLocation":"","error":""},`)
	annUnitOverhead = len(`{"valid":true,"instanceLocation":"","keywordLocation":"","annotation":},`)
	absKeyOverhead  = len(`"absoluteKeywordLocation":"",`)
)

// sizeHint returns the capacity to reserve for a basic report: the size
// derived from the location lengths (a lower bound of the report, since
// escaping only adds bytes), capped at the remaining report-byte limit, plus
// 1 byte for the CLI's trailing newline. It only sizes a buffer; it never
// affects byte counting or the limit check.
func sizeHint(r Report, remaining int) int {
	n := 0
	if r.Valid {
		if len(r.Annotations) == 0 {
			n = len(`{"valid":true}`)
		} else {
			n = len(`{"valid":true,"annotations":[]}`) - 1 // no comma after the last unit
		}
		for i := range r.Annotations {
			a := &r.Annotations[i]
			n += annUnitOverhead + len(a.InstanceLocation) + len(a.KeywordLocation) + valueSize(a.Value)
			if a.AbsoluteKeywordLocation != "" {
				n += absKeyOverhead + len(a.AbsoluteKeywordLocation)
			}
		}
	} else {
		n = len(`{"valid":false,"errors":[]}`)
		if len(r.Units) > 0 {
			n-- // no comma after the last unit
		}
		for i := range r.Units {
			u := &r.Units[i]
			n += errUnitOverhead + len(u.InstanceLocation) + len(u.KeywordLocation) + len(u.Error)
			if u.AbsoluteKeywordLocation != "" {
				n += absKeyOverhead + len(u.AbsoluteKeywordLocation)
			}
		}
	}
	if remaining < 0 {
		remaining = 0
	}
	return min(n, remaining) + 1
}
