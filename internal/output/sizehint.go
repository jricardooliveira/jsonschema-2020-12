package output

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
		for i := range r.Annotations {
			a := &r.Annotations[i]
			n += annUnitOverhead + len(a.InstanceLocation) + len(a.KeywordLocation)
			if a.AbsoluteKeywordLocation != "" {
				n += absKeyOverhead + len(a.AbsoluteKeywordLocation)
			}
		}
	} else {
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
