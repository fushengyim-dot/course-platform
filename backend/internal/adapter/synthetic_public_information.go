package adapter

import (
	"context"
	"time"

	"group-platform/internal/repository"
)

// SyntheticPublicInformation is an in-memory demonstration fixture. No record or URL is real.
type SyntheticPublicInformation struct {
	Items       []repository.PublicInformation
	Institutions []repository.PublicInstitution
}

func NewSyntheticPublicInformation() SyntheticPublicInformation {
	date := func(value string) *string { return &value }
	institutions := []repository.PublicInstitution{
		{ID: "20000000-0000-4000-8000-000000000001", Name: "示例大学", Aliases: []string{"示大"}, Region: "华东 · 演示地区", Description: "虚构演示院校，用于预览院校资料和招生信息的浏览方式。", Colleges: []string{"计算与智能学院"}, AdmissionEntries: []repository.PublicAdmissionEntry{{ID: "30000000-0000-4000-8000-000000000001", Title: "计算与智能学院夏令营通知（演示）", ExpectedPublicationOn: date("2026-11-12"), RegistrationURL: "https://admission.example.invalid/demo/summer-camp", SourceTitle: "虚构演示来源（非真实发布）", SourcePublishedOn: date("2026-05-25")}}},
		{ID: "20000000-0000-4000-8000-000000000002", Name: "东岭示范大学", Aliases: []string{"东岭大"}, Region: "华南 · 演示地区", Description: "虚构演示院校，展示校级资料和学院级信息入口。", Colleges: []string{"生命与环境学院"}, AdmissionEntries: []repository.PublicAdmissionEntry{{ID: "30000000-0000-4000-8000-000000000002", Title: "生命与环境学院开放日（演示）", ExpectedPublicationOn: date("2026-11-12"), RegistrationURL: "https://admission.example.invalid/demo/open-day", SourceTitle: "虚构演示来源（非真实发布）", SourcePublishedOn: date("2026-06-12")}}},
		{ID: "20000000-0000-4000-8000-000000000003", Name: "北川演示学院", Aliases: []string{"北川院"}, Region: "华北 · 演示地区", Description: "虚构演示院校，所有名称、链接和日期均为演示数据。", Colleges: []string{"电子工程系"}, AdmissionEntries: []repository.PublicAdmissionEntry{{ID: "30000000-0000-4000-8000-000000000003", Title: "电子工程系预推免说明（演示）", ExpectedPublicationOn: date("2026-11-20"), RegistrationURL: "https://admission.example.invalid/demo/pre-recommendation", SourceTitle: "虚构演示来源（非真实发布）", SourcePublishedOn: date("2026-05-17")}}},
	}
	items := []repository.PublicInformation{
		{ID: "10000000-0000-4000-8000-000000000001", InstitutionName: institutions[0].Name, InstitutionAliases: institutions[0].Aliases, CollegeNames: institutions[0].Colleges, CollegeAliases: []string{"计智学院", "计院"}, Title: "计算与智能学院夏令营通知（演示）", Body: "虚构演示信息：介绍计算与智能学院开放日和夏令营安排。", ExpectedPublicationOn: date("2026-11-12"), SourceTitle: "虚构演示来源（非真实发布）", SourcePublishedOn: date("2026-05-25"), RegistrationURL: institutions[0].AdmissionEntries[0].RegistrationURL, PublishedAt: time.Date(2026, 5, 25, 9, 0, 0, 0, time.UTC)},
		{ID: "10000000-0000-4000-8000-000000000002", InstitutionName: institutions[1].Name, InstitutionAliases: institutions[1].Aliases, CollegeNames: institutions[1].Colleges, CollegeAliases: []string{"生环学院"}, Title: "生命与环境学院开放日（演示）", Body: "虚构演示信息：介绍生命与环境学院的招生开放日。", ExpectedPublicationOn: date("2026-11-12"), SourceTitle: "虚构演示来源（非真实发布）", SourcePublishedOn: date("2026-06-12"), RegistrationURL: institutions[1].AdmissionEntries[0].RegistrationURL, PublishedAt: time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)},
		{ID: "10000000-0000-4000-8000-000000000003", InstitutionName: institutions[2].Name, InstitutionAliases: institutions[2].Aliases, CollegeNames: institutions[2].Colleges, CollegeAliases: []string{"电工系"}, Title: "电子工程系预推免说明（演示）", Body: "虚构演示信息：介绍电子工程系预推免说明活动。", ExpectedPublicationOn: date("2026-11-20"), SourceTitle: "虚构演示来源（非真实发布）", SourcePublishedOn: date("2026-05-17"), RegistrationURL: institutions[2].AdmissionEntries[0].RegistrationURL, PublishedAt: time.Date(2026, 5, 17, 9, 0, 0, 0, time.UTC)},
	}
	return SyntheticPublicInformation{Items: items, Institutions: institutions}
}

func (r SyntheticPublicInformation) ListPublished(ctx context.Context) ([]repository.PublicInformation, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	return append([]repository.PublicInformation(nil), r.Items...), nil
}

func (r SyntheticPublicInformation) GetPublished(ctx context.Context, id string) (repository.PublicInformation, bool, error) {
	if err := ctx.Err(); err != nil { return repository.PublicInformation{}, false, err }
	for _, item := range r.Items { if item.ID == id { return item, true, nil } }
	return repository.PublicInformation{}, false, nil
}

func (r SyntheticPublicInformation) ListPublicInstitutions(ctx context.Context) ([]repository.PublicInstitution, error) {
	if err := ctx.Err(); err != nil { return nil, err }
	return append([]repository.PublicInstitution(nil), r.Institutions...), nil
}

func (r SyntheticPublicInformation) GetPublicInstitution(ctx context.Context, id string) (repository.PublicInstitution, bool, error) {
	if err := ctx.Err(); err != nil { return repository.PublicInstitution{}, false, err }
	for _, institution := range r.Institutions { if institution.ID == id { return institution, true, nil } }
	return repository.PublicInstitution{}, false, nil
}
