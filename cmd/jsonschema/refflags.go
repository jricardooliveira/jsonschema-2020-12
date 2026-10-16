package main

import (
	"errors"
	"net/url"
	"strings"
)

// refHelp documents the --ref / --ref-dir value rule.
const refHelp = "split at the first '=': <uri>=<path> is a mapping only when the text before it is an absolute URI " +
	"(scheme of 2+ characters, no whitespace); otherwise the whole value is a path (prefix './' to force a path). " +
	"A URI containing '=' cannot be mapped here; use the library option"

// listFlag is a repeatable string flag.
type listFlag []string

func (l *listFlag) String() string     { return strings.Join(*l, ",") }
func (l *listFlag) Set(s string) error { *l = append(*l, s); return nil }

// parseRef splits a --ref or --ref-dir value at its first '='. mapped is
// false when the whole value is a path.
func parseRef(v string) (uri, path string, mapped bool, err error) {
	if v == "" {
		return "", "", false, errors.New("empty value")
	}
	pre, post, found := strings.Cut(v, "=")
	if !found {
		return "", v, false, nil
	}
	u, perr := url.Parse(pre)
	if perr != nil || !u.IsAbs() || len(u.Scheme) < 2 || strings.ContainsAny(pre, " \t\r\n") {
		return "", v, false, nil
	}
	if post == "" {
		return "", "", false, errors.New("mapping has an empty path")
	}
	return pre, post, true, nil
}
