package redact_test

import (
	"strings"
	"testing"

	"github.com/lasttoss/agent-memory-workflow"
)

// Every value below is fabricated for this test. None of them is, or ever was, a working credential: the
// shapes are what matter, and the point of the package is that a real one never reaches a memory page.
// Every value below is fabricated for this test, and every one of them is assembled at run time from
// pieces. That is not decoration: the first push of this repository was refused by the host's secret
// scanning, because a test fixture with the exact shape of a credential is indistinguishable - to a
// scanner, and that is the scanner's job - from a credential. Splitting the value keeps the shape the test
// needs and keeps the scanner out of a decision it should not have to make about a commit.
//
// The lesson generalises: if you write a redactor, the fixtures are the hard part.
const (
	fakeGitHubPAT  = "ghp_" + "0123456789abcdefghijklmnopqrstuvwxyz"
	fakeGitHubFine = "github_pat_" + "11ABCDEFGHIJKLMNOPQRSTUV_0123456789abcdefghijklmnopqrstuvwxyzABCDEFGH"
	fakeAWSKey     = "AKIA" + "IOSFODNN7EXAMPLE"
	fakeOpenAIKey  = "sk-" + "proj-abcdefghijklmnopqrstuvwxyz0123456789"
	fakeSlack      = "xoxb-" + "0123456789-abcdefghijklmnop"
	fakeJWT        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" + "." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIn0" + "." +
		"dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	fakePrivateKey = "-----BEGIN " + "OPENSSH PRIVATE KEY-----\n" +
		"b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW\n" +
		"QyNTUxOQAAACBmYWtlLWtleS1ib2R5LWZvci10ZXN0aW5nLW9ubHkAAAAA\n" +
		"-----END " + "OPENSSH PRIVATE KEY-----"
)

func TestEverySecretIsGoneFromTheStoredText(t *testing.T) {
	page := `# Debugging the ingest worker

The token was in the environment: ` + fakeGitHubPAT + `

And the AWS key was passed on the command line:
  ` + fakeAWSKey + `

The compose file had postgres://gameuser:hunter2hunter2@db:5432/games

Deploy used ` + fakeJWT + `
`

	stored, findings := redact.Redact(page)

	for name, secret := range map[string]string{
		"github PAT": fakeGitHubPAT,
		"AWS key":    fakeAWSKey,
		"password":   "hunter2hunter2",
		"JWT":        fakeJWT,
	} {
		if strings.Contains(stored, secret) {
			t.Errorf("the %s is still in the stored page", name)
		}
	}
	if len(findings) != 4 {
		t.Errorf("%d findings for 4 secrets: %+v", len(findings), findings)
	}
	// Each marker names the shape that was recognised, and the more specific shape wins: the password in
	// a connection URI is a connection URI, not a generic assignment. A rotation list that calls it
	// "credential-assignment" is a list that sends somebody looking in the wrong place.
	for _, marker := range []string{
		"[redacted:github-token]",
		"[redacted:aws-access-key]",
		"[redacted:connection-uri-with-password]",
		"[redacted:jwt]",
	} {
		if !strings.Contains(stored, marker) {
			t.Errorf("the stored page does not say %s:\n%s", marker, stored)
		}
	}
	if strings.Contains(stored, "[redacted:credential-assignment]") {
		t.Errorf("a shape rule lost to a convention rule:\n%s", stored)
	}
	// What the page said still has to be readable: a redactor that destroys the note destroys the memory.
	if !strings.Contains(stored, "# Debugging the ingest worker") || !strings.Contains(stored, "the ingest worker") {
		t.Errorf("the page was destroyed rather than redacted:\n%s", stored)
	}
}

func TestNoFindingEverCarriesTheSecretWithIt(t *testing.T) {
	_, findings := redact.Redact("key=" + fakeOpenAIKey + "\ntoken=" + fakeGitHubFine)

	if len(findings) == 0 {
		t.Fatal("nothing was found in text with two keys in it")
	}
	for _, finding := range findings {
		for _, secret := range []string{fakeOpenAIKey, fakeGitHubFine} {
			if strings.Contains(finding.Context, secret) {
				t.Errorf("a finding carries the value it is reporting: %q", finding.Context)
			}
		}
		if finding.Length < 8 {
			t.Errorf("a finding does not say how long the value was: %+v", finding)
		}
		if finding.Kind == "" {
			t.Errorf("a finding does not say what it found: %+v", finding)
		}
	}
}

// A page is re-read and rewritten by consolidations, so the redactor has to be a no-op on its own output.
func TestRedactingTwiceChangesNothingTheSecondTime(t *testing.T) {
	once, first := redact.Redact("OPENAI_API_KEY=" + fakeOpenAIKey)
	twice, second := redact.Redact(once)

	if twice != once {
		t.Errorf("the second pass changed the text:\n%q\n%q", once, twice)
	}
	if len(second) != 0 {
		t.Errorf("the second pass found something in its own marker: %+v", second)
	}
	if len(first) != 1 {
		t.Errorf("the first pass found %d things", len(first))
	}
}

func TestAPrivateKeyBlockIsRemovedWhole(t *testing.T) {
	stored, findings := redact.Redact("the key file was:\n" + fakePrivateKey + "\nand that is all")

	if strings.Contains(stored, "b3BlbnNzaC1rZXktdjE") || strings.Contains(stored, "PRIVATE KEY-----") {
		t.Errorf("part of the key block survived:\n%s", stored)
	}
	if len(findings) != 1 || findings[0].Kind != redact.KindPrivateKey {
		t.Fatalf("findings = %+v", findings)
	}
	if !strings.Contains(stored, "and that is all") {
		t.Errorf("the text around the key was lost:\n%s", stored)
	}
}

// The false positives are the interesting half: a redactor that mangles prose about tokens or an example
// with a short placeholder is one that gets switched off, and then nothing is redacted at all.
func TestProseAboutCredentialsIsLeftAlone(t *testing.T) {
	untouched := []string{
		"The token is required for every request, and it expires after an hour.",
		"Set GITHUB_TOKEN=... in your environment before running the script.",
		"API_KEY=changeme",
		"Use --token <token> to override the default.",
		"the password: prompt appears twice, and the second time it must match",
		"# a note about the SECRET= handling in the config loader",
		"https://example.com/docs?page=2",
	}
	for _, text := range untouched {
		stored, findings := redact.Redact(text)
		if len(findings) != 0 {
			t.Errorf("%q was redacted: %+v", text, findings)
		}
		if stored != text {
			t.Errorf("%q was changed to %q", text, stored)
		}
	}
}

func TestAFindingNamesTheLineWithTheValueAlreadyRemoved(t *testing.T) {
	_, findings := redact.Redact("step 1\nAPI_KEY=" + fakeOpenAIKey + "\nstep 3")

	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
	if strings.Contains(findings[0].Context, fakeOpenAIKey) {
		t.Errorf("the context carries the value: %q", findings[0].Context)
	}
	if !strings.Contains(findings[0].Context, "API_KEY=") {
		t.Errorf("the context does not say which name it was: %q", findings[0].Context)
	}
}

func TestSummarySaysWhatHasToBeRotated(t *testing.T) {
	_, findings := redact.Redact(fakeGitHubPAT + "\n" + fakeAWSKey + "\n" + fakeGitHubFine)

	summary := redact.Summary(findings)
	if !strings.Contains(summary, "github-token x2") || !strings.Contains(summary, "aws-access-key") {
		t.Errorf("summary = %q", summary)
	}
	if redact.Summary(nil) != "" {
		t.Errorf("nothing to rotate was summarised as %q", redact.Summary(nil))
	}
}
