package repository

import (
	"context"
	"time"
)

// PublicInformation contains only fields available to anonymous readers.
type PublicInformation struct {
	ID                      string    `json:"id"`
	InstitutionName         string    `json:"institution_name"`
	InstitutionAliases      []string  `json:"institution_aliases"`
	CollegeNames            []string  `json:"college_names"`
	CollegeAliases          []string  `json:"college_aliases"`
	Title                   string    `json:"title"`
	Body                    string    `json:"body"`
	ExpectedPublicationOn   *string   `json:"expected_publication_on"`
	SourceTitle             string    `json:"source_title"`
	SourcePublishedOn       *string   `json:"source_published_on"`
	RegistrationURL         string    `json:"registration_url"`
	PublishedAt             time.Time `json:"published_at"`
}

type PublicAdmissionEntry struct {
	ID                    string  `json:"id"`
	Title                 string  `json:"title"`
	ExpectedPublicationOn *string `json:"expected_publication_on"`
	RegistrationURL       string  `json:"registration_url"`
	SourceTitle           string  `json:"source_title"`
	SourcePublishedOn     *string `json:"source_published_on"`
}

type PublicInstitution struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Aliases         []string               `json:"aliases"`
	Region          string                 `json:"region"`
	Description     string                 `json:"description"`
	Colleges        []string               `json:"colleges"`
	AdmissionEntries []PublicAdmissionEntry `json:"admission_entries"`
}

type PublicInformationRepository interface {
	ListPublished(context.Context) ([]PublicInformation, error)
	GetPublished(context.Context, string) (PublicInformation, bool, error)
}

type PublicInstitutionRepository interface {
	ListPublicInstitutions(context.Context) ([]PublicInstitution, error)
	GetPublicInstitution(context.Context, string) (PublicInstitution, bool, error)
}
