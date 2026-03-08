package domain

type SourceType string

const (
	SourceTypeHackerNews   SourceType = "hackernews"
	SourceTypeArXiv        SourceType = "arxiv"
	SourceTypeGitHub       SourceType = "github"
	SourceTypeProductHunt  SourceType = "producthunt"
	SourceTypeTechCrunch   SourceType = "techcrunch"
)

type Source struct {
	ID             string
	Name           string
	Type           SourceType
	URL            string
	Enabled        bool
	ScoreThreshold float64
}
