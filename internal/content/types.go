package content

import (
	"fmt"
	"html/template"
	"time"

	"gopkg.in/yaml.v3"
)

type BlogPost struct {
	Slug        string
	Title       string        `yaml:"title"`
	Date        time.Time     `yaml:"date"`
	Description string        `yaml:"description"`
	Tags        []string      `yaml:"tags"`
	Content     template.HTML // rendered markdown
}

type Project struct {
	Slug        string
	Title       string        `yaml:"title"`
	Date        time.Time     `yaml:"date"`
	Description string        `yaml:"description"`
	Tags        []string      `yaml:"tags"`
	Repo        string        `yaml:"repo"`
	URL         string        `yaml:"url"`
	Status      string        `yaml:"status"`
	Featured    bool          `yaml:"featured"`
	Content     template.HTML // rendered markdown
}

type Resume struct {
	Name       string        `yaml:"name"`
	Tagline    string        `yaml:"tagline"`
	RawSummary string        `yaml:"summary"`
	Summary    template.HTML `yaml:"-"`

	Experience    []ResumeEntry        `yaml:"experience"`
	Education     []ResumeEntry        `yaml:"education"`
	Skills        []ResumeSkill        `yaml:"skills"`
	Research      []ResumeEntry        `yaml:"research"`
	Awards        []string             `yaml:"awards"`
	Presentations []ResumePresentation `yaml:"presentations"`
	Publications  []ResumePubSection   `yaml:"publications"`
	OpenSource    []ResumeOSSSection   `yaml:"opensource"`
}

type ResumeDate struct {
	Year  int `yaml:"year"`
	Month int `yaml:"month,omitempty"`
}

type ResumeEntry struct {
	Title        string         `yaml:"title"`
	Organization string         `yaml:"organization"`
	Location     string         `yaml:"location"`
	Start        ResumeDate     `yaml:"start"`
	End          *ResumeDate    `yaml:"end,omitempty"`
	Note         string         `yaml:"note,omitempty"`
	Bullets      []ResumeBullet `yaml:"bullets,omitempty"`
	DateRange    string         `yaml:"-"`
}

type ResumeBullet struct {
	Text    template.HTML  `yaml:"-"`
	RawText string         `yaml:"-"`
	Sub     []ResumeBullet `yaml:"-"`
	RawSub  []ResumeBullet `yaml:"-"`
}

type ResumeSkill struct {
	Category string `yaml:"category"`
	Detail   string `yaml:"detail"`
}

type ResumePresentation struct {
	Title         string        `yaml:"title"`
	RawVenue      string        `yaml:"venue"`
	Venue         template.HTML `yaml:"-"`
	Date          ResumeDate    `yaml:"date"`
	DateFormatted string        `yaml:"-"`
}

type ResumePubSection struct {
	Section  string          `yaml:"section"`
	RawItems []string        `yaml:"items"`
	Items    []template.HTML `yaml:"-"`
}

type ResumeOSSSection struct {
	Section  string             `yaml:"section"`
	Projects []ResumeOSSProject `yaml:"projects"`
}

type ResumeOSSProject struct {
	Name       string          `yaml:"name"`
	Tagline    string          `yaml:"tagline"`
	RawBullets []string        `yaml:"bullets,omitempty"`
	Bullets    []template.HTML `yaml:"-"`
	Links      []ResumeLink    `yaml:"links,omitempty"`
}

type ResumeLink struct {
	Text string `yaml:"text"`
	URL  string `yaml:"url"`
}

// UnmarshalYAML allows ResumeBullet to be either a plain string or an object
// with "text" and optional "sub" fields. Sub-items are recursively parsed.
func (b *ResumeBullet) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		b.RawText = value.Value
		return nil
	}

	var obj struct {
		Text string         `yaml:"text"`
		Sub  []ResumeBullet `yaml:"sub"`
	}
	if err := value.Decode(&obj); err != nil {
		return fmt.Errorf("decoding bullet: %w", err)
	}
	b.RawText = obj.Text
	b.RawSub = obj.Sub
	return nil
}

var shortMonths = [...]string{
	1: "Jan.", 2: "Feb.", 3: "Mar.", 4: "Apr.",
	5: "May", 6: "June", 7: "July", 8: "Aug.",
	9: "Sept.", 10: "Oct.", 11: "Nov.", 12: "Dec.",
}

// FormatDate renders a ResumeDate as "Month Year" or just "Year".
func (d ResumeDate) FormatDate() string {
	if d.Month >= 1 && d.Month <= 12 {
		return shortMonths[d.Month] + " " + fmt.Sprint(d.Year)
	}
	return fmt.Sprint(d.Year)
}

// FormatDateRange renders "Start – End" or "Start – Present".
func FormatDateRange(start ResumeDate, end *ResumeDate) string {
	s := start.FormatDate()
	if end == nil {
		return s + " – Present"
	}
	return s + " – " + end.FormatDate()
}

type ContentStore struct {
	Posts       []BlogPost
	PostsBySlug map[string]*BlogPost
	PostsByTag  map[string][]*BlogPost

	Projects       []Project
	ProjectsBySlug map[string]*Project

	Resume *Resume
}
