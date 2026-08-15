package layout //nolint:testpackage // tests unexported titleSimilarity

import "testing"

const similarityTestTitle = "Hello World"

func TestTitleSimilarity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		saved string
		live  string
		want  float64
	}{
		{name: "identical titles", saved: similarityTestTitle, live: similarityTestTitle, want: 1},
		{name: "reordered words", saved: similarityTestTitle, live: "World Hello", want: 1},
		{name: "case-only difference", saved: similarityTestTitle, live: "hello world", want: 1},
		{name: "no shared words", saved: similarityTestTitle, live: "Foo Bar", want: 0},
		{name: "saved title empty", saved: "", live: "Foo", want: 0},
		{name: "live title empty", saved: "Foo", live: "", want: 0},
		{name: "both titles empty", saved: "", live: "", want: 0},
		{
			// shared {alpha,beta}=2, union {alpha,beta,gamma,delta}=4
			name:  "partial word overlap",
			saved: "alpha beta gamma",
			live:  "alpha beta delta",
			want:  0.5,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := titleSimilarity(testCase.saved, testCase.live)
			if got != testCase.want {
				t.Fatalf(
					"titleSimilarity(%q, %q) = %v, want %v",
					testCase.saved, testCase.live, got, testCase.want,
				)
			}
		})
	}
}
