package output

import "encoding/json"

// basicUnit is one error unit. absoluteKeywordLocation is left out until an
// absolute base URI is known (stories 186/187).
type basicUnit struct {
	InstanceLocation string `json:"instanceLocation"`
	KeywordLocation  string `json:"keywordLocation"`
	Error            string `json:"error"`
}

// annotationUnit is one annotation unit of a passing report. Its annotation
// comes from the schema; the only instance-derived content is the evaluated
// property NAMES, which already appear in instanceLocation. No instance
// value is ever written.
type annotationUnit struct {
	Valid                   bool   `json:"valid"`
	InstanceLocation        string `json:"instanceLocation"`
	KeywordLocation         string `json:"keywordLocation"`
	AbsoluteKeywordLocation string `json:"absoluteKeywordLocation,omitempty"`
	Annotation              any    `json:"annotation"`
}

func put(w *Budget, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func puts(w *Budget, s string) error { _, err := w.Write([]byte(s)); return err }

// writeList writes `{"valid":<v>,"<key>":[unit,...]}`, one unit at a time so
// the byte cap is enforced before the whole report is built.
func writeList(w *Budget, valid bool, key string, n int, unit func(i int) any) error {
	head := `{"valid":false,"` + key + `":[`
	if valid {
		head = `{"valid":true,"` + key + `":[`
	}
	if err := puts(w, head); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if i > 0 {
			if err := puts(w, ","); err != nil {
				return err
			}
		}
		if err := put(w, unit(i)); err != nil {
			return err
		}
	}
	return puts(w, "]}")
}

func init() {
	RegisterReport("basic", func(r Report, w *Budget) error {
		if r.Valid {
			if len(r.Annotations) == 0 {
				return put(w, struct {
					Valid bool `json:"valid"`
				}{true})
			}
			return writeList(w, true, "annotations", len(r.Annotations), func(i int) any {
				a := r.Annotations[i]
				return annotationUnit{true, a.InstanceLocation, a.KeywordLocation, a.AbsoluteKeywordLocation, a.Value}
			})
		}
		return writeList(w, false, "errors", len(r.Units), func(i int) any {
			u := r.Units[i]
			return basicUnit{u.InstanceLocation, u.KeywordLocation, u.Error}
		})
	})
}
