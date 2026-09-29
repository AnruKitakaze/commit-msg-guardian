package parser

import "testing"

func TestSplitMessageCommentsAndScissors(t *testing.T) {
	message := `# Git editor hint
feat: add feature

First line

 # This indented line is content
# This line is an editor comment
# ------------------------ >8 ------------------------
diff --git a/file b/file`
	want := Message{
		Header: "feat: add feature",
		Body:   "First line\n\n # This indented line is content",
	}
	if got := SplitMessage(message); got != want {
		t.Fatalf("SplitMessage() = %#v, want %#v", got, want)
	}
}

func TestSplitMessageStoredAlsoIgnoresComments(t *testing.T) {
	message := "feat: add feature\n\n# actual body\n# ------------------------ >8 ------------------------\nMore body\n"
	want := Message{
		Header: "feat: add feature",
	}
	if got := SplitMessage(message); got != want {
		t.Fatalf("SplitMessage() = %#v, want %#v", got, want)
	}
}

func TestSplitMessageCRLFScissors(t *testing.T) {
	message := "feat: add feature\r\n\r\nbody\r\n# ------------------------ >8 ------------------------\r\ndiff --git a/file b/file\r\n"
	want := Message{Header: "feat: add feature", Body: "body"}
	if got := SplitMessage(message); got != want {
		t.Fatalf("SplitMessage() = %#v, want %#v", got, want)
	}
}

func TestSplitMessageTrimsOnlyBodyEdges(t *testing.T) {
	want := Message{Header: "feat: add feature", Body: "First line\n\nSecond line"}
	if got := SplitMessage("feat: add feature\r\n\n First line\n\nSecond line \n"); got != want {
		t.Fatalf("SplitMessage() = %#v, want %#v", got, want)
	}
}

func TestSplitMessageKeepsBodyAndFinalTrailersSeparate(t *testing.T) {
	for _, tc := range []struct {
		message, body, trailers string
	}{
		{"feat: change\n\nExplanation\n\nBREAKING CHANGE: migrate to v2", "Explanation", "BREAKING CHANGE: migrate to v2"},
		{"feat: change\n\nBREAKING CHANGE: migrate to v2", "", "BREAKING CHANGE: migrate to v2"},
		{"feat: change\n\nExplanation\n\nRefs: TASK-1\nReviewed-by: Alice", "Explanation", "Refs: TASK-1\nReviewed-by: Alice"},
		{"feat: change\n\nExplanation\nRefs: TASK-1", "Explanation\nRefs: TASK-1", ""},
	} {
		got := SplitMessage(tc.message)
		if got.Body != tc.body || got.Trailers != tc.trailers {
			t.Errorf("SplitMessage(%q) = %#v, want body %q, trailers %q", tc.message, got, tc.body, tc.trailers)
		}
	}
}
