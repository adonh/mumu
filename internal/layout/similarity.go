package layout

import "strings"

// titleSimilarity scores how similar two window titles are as the Jaccard
// index (intersection over union) of their lowercased, whitespace-split
// word sets: word order and repeated words don't affect the score, only
// which distinct words appear. Two empty titles score 0, not 1 — no title
// information is "no signal," not a perfect match, so an empty title falls
// through to the tie-break cascade like any other zero-scoring candidate
// rather than being spuriously preferred (see design.md).
func titleSimilarity(saved, live string) float64 {
	savedWords := titleWordSet(saved)
	liveWords := titleWordSet(live)

	shared := 0

	for word := range savedWords {
		if liveWords[word] {
			shared++
		}
	}

	union := len(savedWords) + len(liveWords) - shared
	if union == 0 {
		return 0
	}

	return float64(shared) / float64(union)
}

// titleWordSet lowercases and whitespace-splits a title into the set of
// its distinct words.
func titleWordSet(title string) map[string]bool {
	words := map[string]bool{}

	for word := range strings.FieldsSeq(strings.ToLower(title)) {
		words[word] = true
	}

	return words
}
