package policy

import "testing"

func TestResolvePrecedence(t *testing.T) {
	if g := Resolve(Allow, "", ""); g != Allow {
		t.Fatalf("default allow: %s", g)
	}
	if g := Resolve(Allow, ApprovalRequired, ""); g != ApprovalRequired {
		t.Fatalf("manifest: %s", g)
	}
	if g := Resolve(Allow, ApprovalRequired, Deny); g != Deny {
		t.Fatalf("config overrides manifest: %s", g)
	}
	if g := Resolve(Deny, "", Allow); g != Allow {
		t.Fatalf("config override default: %s", g)
	}
}

func TestParseVerb(t *testing.T) {
	if _, err := ParseVerb("nope"); err == nil {
		t.Fatal("expected error")
	}
	v, err := ParseVerb("approval-required")
	if err != nil || v != ApprovalRequired {
		t.Fatalf("got %s %v", v, err)
	}
}

func TestTableVerb(t *testing.T) {
	tab := Table{Default: Allow, Caps: map[string]Verb{"shell.exec": ApprovalRequired}}
	if tab.Verb("shell.exec", Allow) != ApprovalRequired {
		t.Fatal("config map should win")
	}
	if tab.Verb("fs.read", "") != Allow {
		t.Fatal("unlisted cap inherits default")
	}
}
