package service

import (
	"context"
	"sort"
	"time"
)

type PublicInformationCalendar struct {
	Month string                         `json:"month"`
	Days  []PublicInformationCalendarDay `json:"days"`
}

type PublicInformationCalendarDay struct {
	Date  string                           `json:"date"`
	Items []PublicInformationCalendarEntry `json:"items"`
}

type PublicInformationCalendarEntry struct {
	ID                string   `json:"id"`
	InstitutionName   string   `json:"institution_name"`
	CollegeNames      []string `json:"college_names"`
	Title             string   `json:"title"`
	ExpectedPublishOn string   `json:"expected_publication_on"`
	RegistrationURL   string   `json:"registration_url"`
}

func (s PublicInformation) Calendar(ctx context.Context, month string) (PublicInformationCalendar, error) {
	parsed, err := time.Parse("2006-01", month)
	if err != nil || parsed.Year() < 1 || parsed.Format("2006-01") != month {
		return PublicInformationCalendar{}, ErrInvalidPublicQuery
	}
	if s.Repository == nil {
		return PublicInformationCalendar{}, ErrPublicInformationUnavailable
	}
	items, err := s.Repository.ListPublished(ctx)
	if err != nil {
		return PublicInformationCalendar{}, ErrPublicInformationUnavailable
	}
	lastDay := parsed.AddDate(0, 1, -1).Day()
	byDate := make(map[string][]PublicInformationCalendarEntry)
	for _, item := range items {
		if item.ExpectedPublicationOn == nil || len(*item.ExpectedPublicationOn) != len("2006-01-02") || (*item.ExpectedPublicationOn)[:7] != month {
			continue
		}
		if _, err := time.Parse("2006-01-02", *item.ExpectedPublicationOn); err != nil {
			continue
		}
		byDate[*item.ExpectedPublicationOn] = append(byDate[*item.ExpectedPublicationOn], PublicInformationCalendarEntry{
			ID: item.ID, InstitutionName: item.InstitutionName, CollegeNames: append([]string(nil), item.CollegeNames...),
			Title: item.Title, ExpectedPublishOn: *item.ExpectedPublicationOn, RegistrationURL: item.RegistrationURL,
		})
	}
	days := make([]PublicInformationCalendarDay, 0, len(byDate))
	for day := 1; day <= lastDay; day++ {
		date := parsed.AddDate(0, 0, day-1).Format("2006-01-02")
		entries := byDate[date]
		if len(entries) == 0 {
			continue
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].InstitutionName != entries[j].InstitutionName {
				return entries[i].InstitutionName < entries[j].InstitutionName
			}
			return entries[i].Title < entries[j].Title
		})
		days = append(days, PublicInformationCalendarDay{Date: date, Items: entries})
	}
	return PublicInformationCalendar{Month: month, Days: days}, nil
}
