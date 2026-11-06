package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"group-platform/internal/repository"
)

var (
	ErrInvalidPublicQuery = errors.New("invalid public information query")
	ErrPublicInformationNotFound = errors.New("public information not found")
	ErrPublicInformationUnavailable = errors.New("public information unavailable")
	ErrPublicInstitutionNotFound = errors.New("public institution not found")
	ErrPublicInstitutionUnavailable = errors.New("public institution directory unavailable")
)

type PublicInformationQuery struct {
	Institution       string
	College           string
	ExpectedPublishFrom string
	ExpectedPublishTo   string
	Sort              string
	PageSize          int
	Cursor            string
}

type PublicInformationResult struct {
	Information repository.PublicInformation `json:"information"`
	MatchedTerms []string                    `json:"matched_terms"`
}

type PublicInformationPage struct {
	Items      []PublicInformationResult `json:"items"`
	NextCursor string                    `json:"next_cursor,omitempty"`
}

type PublicInformation struct {
	Repository repository.PublicInformationRepository
	Now        func() time.Time
}

type cursorState struct {
	Offset      int    `json:"o"`
	Fingerprint string `json:"f"`
	ExpiresAt   int64  `json:"e"`
}

type rankedInformation struct {
	result PublicInformationResult
	score  int
}

func (s PublicInformation) List(ctx context.Context, q PublicInformationQuery) (PublicInformationPage, error) {
	q, fingerprint, err := normalizePublicQuery(q)
	if err != nil {
		return PublicInformationPage{}, err
	}
	if s.Repository == nil {
		return PublicInformationPage{}, ErrPublicInformationUnavailable
	}
	items, err := s.Repository.ListPublished(ctx)
	if err != nil {
		return PublicInformationPage{}, ErrPublicInformationUnavailable
	}
	ranked := make([]rankedInformation, 0, len(items))
	for _, item := range items {
		if q.ExpectedPublishFrom != "" || q.ExpectedPublishTo != "" {
			if item.ExpectedPublicationOn == nil || (q.ExpectedPublishFrom != "" && *item.ExpectedPublicationOn < q.ExpectedPublishFrom) || (q.ExpectedPublishTo != "" && *item.ExpectedPublicationOn > q.ExpectedPublishTo) {
				continue
			}
		}
		score, terms, matches := scoreInformation(item, q)
		if matches {
			ranked = append(ranked, rankedInformation{result: PublicInformationResult{Information: item, MatchedTerms: terms}, score: score})
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		a, b := ranked[i], ranked[j]
		if q.Sort == "title_asc" && a.result.Information.Title != b.result.Information.Title {
			return a.result.Information.Title < b.result.Information.Title
		}
		if q.Sort == "published_at_asc" && !a.result.Information.PublishedAt.Equal(b.result.Information.PublishedAt) {
			return a.result.Information.PublishedAt.Before(b.result.Information.PublishedAt)
		}
		if q.Sort == "published_at_desc" && !a.result.Information.PublishedAt.Equal(b.result.Information.PublishedAt) {
			return a.result.Information.PublishedAt.After(b.result.Information.PublishedAt)
		}
		if q.Sort == "relevance" && a.score != b.score {
			return a.score < b.score
		}
		if !a.result.Information.PublishedAt.Equal(b.result.Information.PublishedAt) {
			return a.result.Information.PublishedAt.After(b.result.Information.PublishedAt)
		}
		return a.result.Information.ID < b.result.Information.ID
	})
	offset := 0
	if q.Cursor != "" {
		state, err := s.decodeCursor(q.Cursor, fingerprint)
		if err != nil {
			return PublicInformationPage{}, ErrInvalidPublicQuery
		}
		offset = state.Offset
	}
	if offset > len(ranked) {
		return PublicInformationPage{}, ErrInvalidPublicQuery
	}
	end := min(offset+q.PageSize, len(ranked))
	page := PublicInformationPage{Items: make([]PublicInformationResult, 0, end-offset)}
	for _, item := range ranked[offset:end] {
		page.Items = append(page.Items, item.result)
	}
	if end < len(ranked) {
		page.NextCursor = s.encodeCursor(cursorState{Offset: end, Fingerprint: fingerprint, ExpiresAt: s.now().Add(15 * time.Minute).Unix()})
	}
	return page, nil
}

func (s PublicInformation) Get(ctx context.Context, id string) (repository.PublicInformation, error) {
	if !validUUID(id) {
		return repository.PublicInformation{}, ErrPublicInformationNotFound
	}
	if s.Repository == nil {
		return repository.PublicInformation{}, ErrPublicInformationUnavailable
	}
	item, found, err := s.Repository.GetPublished(ctx, id)
	if err != nil {
		return repository.PublicInformation{}, ErrPublicInformationUnavailable
	}
	if !found {
		return repository.PublicInformation{}, ErrPublicInformationNotFound
	}
	return item, nil
}

func normalizePublicQuery(q PublicInformationQuery) (PublicInformationQuery, string, error) {
	for _, value := range []string{q.Institution, q.College} {
		if utf8Length(value) > 80 {
			return q, "", ErrInvalidPublicQuery
		}
	}
	q.Institution = normalizeTerm(q.Institution)
	q.College = normalizeTerm(q.College)
	for _, date := range []string{q.ExpectedPublishFrom, q.ExpectedPublishTo} {
		if date != "" {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return q, "", ErrInvalidPublicQuery
			}
		}
	}
	if q.ExpectedPublishFrom != "" && q.ExpectedPublishTo != "" && q.ExpectedPublishFrom > q.ExpectedPublishTo {
		return q, "", ErrInvalidPublicQuery
	}
	if q.Sort == "" {
		if q.Institution != "" || q.College != "" {
			q.Sort = "relevance"
		} else {
			q.Sort = "published_at_desc"
		}
	}
	if q.Sort != "relevance" && q.Sort != "published_at_desc" && q.Sort != "published_at_asc" && q.Sort != "title_asc" {
		return q, "", ErrInvalidPublicQuery
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	if q.PageSize < 1 || q.PageSize > 100 || len(q.Cursor) > 2048 {
		return q, "", ErrInvalidPublicQuery
	}
	filter := q
	filter.Cursor = ""
	encoded, err := json.Marshal(filter)
	if err != nil {
		return q, "", ErrInvalidPublicQuery
	}
	digest := sha256.Sum256(encoded)
	return q, hex.EncodeToString(digest[:]), nil
}

func scoreInformation(item repository.PublicInformation, q PublicInformationQuery) (int, []string, bool) {
	score := 0
	terms := make([]string, 0, 5)
	criteria := []struct {
		query   string
		names   []string
		aliases []string
	}{
		{q.Institution, []string{item.InstitutionName}, item.InstitutionAliases},
		{q.College, item.CollegeNames, item.CollegeAliases},
	}
	for _, criterion := range criteria {
		if criterion.query == "" {
			continue
		}
		tier, term, matched := bestTermMatch(criterion.query, criterion.names, criterion.aliases)
		if !matched {
			return 0, nil, false
		}
		score += tier
		if term != "" {
			terms = append(terms, term)
		}
	}
	return score, terms, true
}

func bestTermMatch(query string, names, aliases []string) (int, string, bool) {
	for _, value := range names {
		if normalizeTerm(value) == query {
			return 0, value, true
		}
	}
	for _, value := range aliases {
		if normalizeTerm(value) == query {
			return 1, value, true
		}
	}
	for _, candidates := range [][]string{names, aliases} {
		for _, value := range candidates {
			normalized := normalizeTerm(value)
			if strings.HasPrefix(normalized, query) {
				return 2, value, true
			}
		}
	}
	for _, candidates := range [][]string{names, aliases} {
		for _, value := range candidates {
			if strings.Contains(normalizeTerm(value), query) {
				return 3, value, true
			}
		}
	}
	if utf8Length(query) >= 4 {
		for _, candidates := range [][]string{names, aliases} {
			for _, value := range candidates {
				if oneEditWindow(query, normalizeTerm(value)) {
					return 4, value, true
				}
			}
		}
	}
	return 0, "", false
}

func oneEditWindow(query, candidate string) bool {
	q := []rune(query)
	c := []rune(candidate)
	minSize, maxSize := max(1, len(q)-1), min(len(c), len(q)+1)
	for size := minSize; size <= maxSize; size++ {
		for start := 0; start+size <= len(c); start++ {
			if editDistanceAtMostOne(q, c[start:start+size]) {
				return true
			}
		}
	}
	return false
}

func editDistanceAtMostOne(a, b []rune) bool {
	if math.Abs(float64(len(a)-len(b))) > 1 {
		return false
	}
	if len(a) == len(b) {
		first := -1
		for i := range a {
			if a[i] != b[i] {
				if first < 0 {
					first = i
				} else if i == first+1 && a[first] == b[i] && a[i] == b[first] {
					for j := i + 1; j < len(a); j++ {
						if a[j] != b[j] {
							return false
						}
					}
					return true
				} else {
					return false
				}
			}
		}
		return first >= 0
	}
	longer, shorter := a, b
	if len(a) < len(b) {
		longer, shorter = b, a
	}
	i, j, skipped := 0, 0, false
	for i < len(longer) && j < len(shorter) {
		if longer[i] == shorter[j] {
			i++
			j++
			continue
		}
		if skipped {
			return false
		}
		skipped = true
		i++
	}
	return true
}

func normalizeTerm(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func utf8Length(value string) int { return len([]rune(value)) }

func (s PublicInformation) encodeCursor(state cursorState) string {
	encoded, _ := json.Marshal(state)
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func (s PublicInformation) decodeCursor(value, fingerprint string) (cursorState, error) {
	var state cursorState
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || json.Unmarshal(data, &state) != nil || state.Offset < 0 || state.Fingerprint != fingerprint || state.ExpiresAt < s.now().Unix() {
		return cursorState{}, ErrInvalidPublicQuery
	}
	return state, nil
}

func (s PublicInformation) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for i, r := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !unicode.IsDigit(r) && !(r >= 'a' && r <= 'f') && !(r >= 'A' && r <= 'F') {
			return false
		}
	}
	return true
}

func (q PublicInformationQuery) String() string {
	return fmt.Sprintf("public-information query (terms=%d)", countTerms(q))
}

func countTerms(q PublicInformationQuery) int {
	n := 0
	for _, term := range []string{q.Institution, q.College} {
		if strings.TrimSpace(term) != "" {
			n++
		}
	}
	return n
}
