// Package redact keeps credentials out of long-term agent memory.
//
// The problem this exists for: an agent writes a memory page after a debugging session, and the page
// contains the environment variable it was reading, or a curl command with the token in it. Memory is
// searched by every future session in the project, mirrored into a wiki, and - in the bad case - pushed to
// a public repository. A secret that reaches memory has already leaked; "we will find it later" is a
// cleanup task nobody does.
//
// So the redactor runs on the write path, before the page is stored, and it does two jobs: it returns text
// that is safe to store, and it reports what it took out so a human can rotate the credential. Deleting
// the value is not enough on its own - the leak already happened when the text was produced - which is why
// the findings carry a kind and a length and never the value itself.
package redact

import (
	"regexp"
	"sort"
	"strings"
)

// Kind is what a finding is, named so that a log line or a report says what has to be rotated.
type Kind string

const (
	KindPrivateKey    Kind = "private-key"
	KindGitHubToken   Kind = "github-token"
	KindAWSAccessKey  Kind = "aws-access-key"
	KindOpenAIKey     Kind = "openai-key"
	KindJWT           Kind = "jwt"
	KindSlackToken    Kind = "slack-token"
	KindAssignment    Kind = "credential-assignment"
	KindConnectionURI Kind = "connection-uri-with-password"
)

// Finding is one removed secret. Preview deliberately says how long the value was instead of showing any
// part of it: a preview is a second copy of the secret in a second place, and a reviewer reading a report
// does not need the first four characters to know that they must rotate a key.
type Finding struct {
	Kind    Kind
	Start   int
	End     int
	Length  int
	Context string // the line it was on, with the value already removed - for a human fixing the cause
}

// Marker is what replaces a removed value. It is bracketed and names the kind so that a reader of the
// stored page can tell that something was removed and why, rather than finding a suspicious gap.
func Marker(kind Kind) string { return "[redacted:" + string(kind) + "]" }

type rule struct {
	kind Kind
	re   *regexp.Regexp
	// value is the submatch to remove; 0 means the whole match.
	value int
	// minValue is the shortest value worth removing for this rule. A shape rule knows its own length; a
	// convention rule does not, and getting this wrong is how a redactor starts mangling examples.
	minValue int
}

// The rules are shapes, not words: "token = ..." is a word, and matching on words is how a redactor ends up
// mangling prose about tokens while missing the key on the next line.
var rules = []rule{
	// A private key block is replaced whole: half a key is still a key, and the body is not something a
	// line-based rule can see.
	{
		kind: KindPrivateKey,
		re:   regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`),
	},
	// GitHub: ghp_/gho_/ghu_/ghs_/ghr_ followed by 36, and the fine-grained github_pat_ form.
	{
		kind: KindGitHubToken,
		re:   regexp.MustCompile(`\b(?:gh[pousr]_[A-Za-z0-9]{36,}|github_pat_[A-Za-z0-9_]{22,})\b`),
	},
	{kind: KindAWSAccessKey, re: regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`)},
	{kind: KindOpenAIKey, re: regexp.MustCompile(`\bsk-(?:proj-)?[A-Za-z0-9_-]{20,}\b`)},
	{kind: KindSlackToken, re: regexp.MustCompile(`\bxox[abposr]-[A-Za-z0-9-]{10,}\b`)},
	// A JWT: three base64url segments, the first two of which decode to JSON. The second group is the
	// signature, and the third is what makes this a credential rather than a coincidence.
	{
		kind: KindJWT,
		re:   regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{10,}\b`),
	},
	// postgres://user:password@host/db - the password is the part that matters and the part that is
	// usually typed into a compose file and then pasted into a note.
	{
		kind:  KindConnectionURI,
		re:    regexp.MustCompile(`\b([a-z][a-z0-9+.-]*://[^:/@\s]+):([^@\s/]{3,})@`),
		value: 2,
	},
	// NAME=value, NAME: value and --flag=value for the names that are credentials by convention. The value
	// has to be long enough to be one: SHORT=values in an example are not secrets, and a redactor that
	// mangles examples is a redactor that gets turned off.
	{
		kind:     KindAssignment,
		minValue: 12,
		re:       regexp.MustCompile(`(?i)\b([A-Za-z0-9_-]*(?:SECRET|TOKEN|PASSWORD|PASSWD|API[_-]?KEY|ACCESS[_-]?KEY|PRIVATE[_-]?KEY|CREDENTIAL|BEARER)[A-Za-z0-9_-]*)\s*(?:=|:)\s*["']?([^\s"'#]{8,})["']?`),
		value:    2,
	},
}

// Scan reports every secret it can see, without changing anything. Use it when the text is being kept
// somewhere else and all that is wanted is the alarm.
func Scan(text string) []Finding {
	findings := make([]Finding, 0, 4)
	for _, r := range rules {
		for _, match := range r.re.FindAllStringSubmatchIndex(text, -1) {
			start, end := match[0], match[1]
			if r.value > 0 {
				pair := 2 * r.value
				if len(match) <= pair+1 || match[pair] < 0 {
					continue
				}
				start, end = match[pair], match[pair+1]
			}
			if r.minValue > 0 && end-start < r.minValue {
				continue
			}
			if placeholder(text[start:end]) {
				continue
			}
			findings = append(findings, Finding{
				Kind:    r.kind,
				Start:   start,
				End:     end,
				Length:  end - start,
				Context: lineOf(text, start, end),
			})
		}
	}
	return dedupe(findings)
}

// placeholder is the escape hatch for idempotency and for examples. It is deliberately about the shape of
// the value rather than a list of words: the marker this package writes must not look like the thing this
// package looks for, and `<token>` in a document is a reader being told what to type.
func placeholder(value string) bool {
	return strings.HasPrefix(value, "[redacted:") || strings.HasPrefix(value, "<")
}

// dedupe keeps one finding per region of text, the more specific one. A key that is recognisable by shape
// is also, by convention, an assignment - reporting it twice makes a rotation list that says "two tokens"
// about one token, which is how a person learns to skim the report.
func dedupe(findings []Finding) []Finding {
	sort.SliceStable(findings, func(i, j int) bool { return findings[i].Start < findings[j].Start })
	kept := findings[:0]
	last := -1
	for _, finding := range findings {
		if finding.Start < last {
			continue
		}
		kept = append(kept, finding)
		last = finding.End
	}
	return kept
}

// Redact returns the text with every secret replaced by a marker, and the findings that were removed.
// Applying it twice changes nothing the second time: the markers it writes do not look like the things it
// looks for, which matters because a page can be re-read, re-summarised and written again.
func Redact(text string) (string, []Finding) {
	findings := Scan(text)
	if len(findings) == 0 {
		return text, nil
	}

	// Back to front, so that the offsets of the findings still to be applied are still valid.
	var builder strings.Builder
	builder.Grow(len(text))
	last := 0
	for _, finding := range findings {
		if finding.Start < last {
			// Overlapping rules matched the same bytes; the first one wins and the second is noise.
			continue
		}
		builder.WriteString(text[last:finding.Start])
		builder.WriteString(Marker(finding.Kind))
		last = finding.End
	}
	builder.WriteString(text[last:])
	return builder.String(), findings
}

// Summary is what goes into the memory page's metadata: what was removed, in a form that can be read
// without reading the secret.
func Summary(findings []Finding) string {
	if len(findings) == 0 {
		return ""
	}
	counts := make(map[Kind]int, len(findings))
	order := make([]Kind, 0, len(findings))
	for _, finding := range findings {
		if counts[finding.Kind] == 0 {
			order = append(order, finding.Kind)
		}
		counts[finding.Kind]++
	}
	parts := make([]string, 0, len(order))
	for _, kind := range order {
		if counts[kind] == 1 {
			parts = append(parts, string(kind))
			continue
		}
		parts = append(parts, string(kind)+" x"+itoa(counts[kind]))
	}
	return strings.Join(parts, ", ")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits [20]byte
	index := len(digits)
	for n > 0 {
		index--
		digits[index] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[index:])
}

// lineOf is the context a human needs to fix the cause: the line, with the value already gone.
func lineOf(text string, start, end int) string {
	lineStart := strings.LastIndexByte(text[:start], '\n') + 1
	lineEnd := strings.IndexByte(text[end:], '\n')
	if lineEnd < 0 {
		lineEnd = len(text)
	} else {
		lineEnd += end
	}
	return strings.TrimSpace(text[lineStart:start] + "[...]" + text[end:lineEnd])
}
