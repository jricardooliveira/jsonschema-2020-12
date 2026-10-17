package formats

import (
	"strconv"
	"testing"
)

// RFC 6570 sections 2.1-2.4.
func TestURITemplate(t *testing.T) {
	max := strconv.Itoa(maxTemplatePrefix)
	over := strconv.Itoa(maxTemplatePrefix + 1)
	table(t, "uri-template", map[string]bool{
		"":                             true,
		"/users/{id}{?page}":           true,
		"{+a}{#b}{.c}{/d}{;e}{?f}{&g}": true,
		"{a,b*,c:3}":                   true,
		"{a.b}{%41}{_x}":               true,
		"{a:" + max + "}":              true,
		"http://ƒøø/{a}":               true,
		"/a%20b":                       true,
		"{":                            false,
		"{a":                           false,
		"{a}}":                         false,
		"}":                            false,
		"{{a}}":                        false,
		"{@a}":                         false,
		"{=a}":                         false,
		"{a:0}":                        false,
		"{a:01}":                       false,
		"{a:" + over + "}":             false,
		"{a:99999999999999999}":        false,
		"{a:}":                         false,
		"{a*:3}":                       false,
		"{a b}":                        false,
		"{}":                           false,
		"{+}":                          false,
		"{a,}":                         false,
		"{.a}{a..b}":                   false,
		"{a.}":                         false,
		"{a%4}":                        false,
		"a b":                          false,
		"a\"b":                         false,
		"a%zz":                         false,
		"a\xff":                        false,
	})
}
