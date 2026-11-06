package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"group-platform/internal/repository"
)

type publicInfoRepo struct { items []repository.PublicInformation; fail bool }
func (r publicInfoRepo) ListPublished(context.Context) ([]repository.PublicInformation,error) { if r.fail{return nil,errors.New("private")};return r.items,nil }
func (r publicInfoRepo) GetPublished(_ context.Context,id string)(repository.PublicInformation,bool,error){ if r.fail{return repository.PublicInformation{},false,errors.New("private")};for _,item:=range r.items{if item.ID==id{return item,true,nil}};return repository.PublicInformation{},false,nil }

func testInfo() repository.PublicInformation { date:="2026-05-20";return repository.PublicInformation{ID:"10000000-0000-4000-8000-000000000001",InstitutionName:"示例大学",InstitutionAliases:[]string{"示大"},CollegeNames:[]string{"计算与智能学院"},CollegeAliases:[]string{"计智学院"},Title:"夏令营演示通知",ExpectedPublicationOn:&date,SourceTitle:"虚构来源",PublishedAt:time.Date(2026,5,25,0,0,0,0,time.UTC)} }

func TestPublicInformationMatchingExpectedDateAndStudentScope(t *testing.T){
	item:=testInfo();svc:=PublicInformation{Repository:publicInfoRepo{items:[]repository.PublicInformation{item}}}
	for _,query:=range []PublicInformationQuery{{Institution:"示例大学"},{Institution:"示大"},{Institution:"示例大校"},{College:"计智学院"}} { page,err:=svc.List(context.Background(),query);if err!=nil||len(page.Items)!=1{t.Errorf("query %+v: %d %v",query,len(page.Items),err)} }
	for _,query:=range []PublicInformationQuery{{Institution:"示例小校"},{College:"不存在"}} { page,err:=svc.List(context.Background(),query);if err!=nil||len(page.Items)!=0{t.Errorf("negative query %+v: %d %v",query,len(page.Items),err)} }
	for _,q:=range []PublicInformationQuery{{ExpectedPublishFrom:"2026-05-20",ExpectedPublishTo:"2026-05-20"},{ExpectedPublishFrom:"2026-05-19",ExpectedPublishTo:"2026-05-20"},{ExpectedPublishFrom:"2026-05-20"}} { page,err:=svc.List(context.Background(),q);if err!=nil||len(page.Items)!=1{t.Errorf("inclusive expected-publication range %+v: %d %v",q,len(page.Items),err)} }
	page,err:=svc.List(context.Background(),PublicInformationQuery{ExpectedPublishFrom:"2026-05-21"});if err!=nil||len(page.Items)!=0{t.Fatalf("valid no-match range: %d %v",len(page.Items),err)}
	for _,q:=range []PublicInformationQuery{{ExpectedPublishFrom:"2026-05-21",ExpectedPublishTo:"2026-05-20"},{ExpectedPublishFrom:"2026-02-30"}} { if _,err:=svc.List(context.Background(),q);err!=ErrInvalidPublicQuery{t.Errorf("invalid query accepted: %+v %v",q,err)} }
	missing:=item;missing.ExpectedPublicationOn=nil;svc.Repository=publicInfoRepo{items:[]repository.PublicInformation{missing}}
	page,err=svc.List(context.Background(),PublicInformationQuery{ExpectedPublishFrom:"2026-05-01"});if err!=nil||len(page.Items)!=0{t.Fatal("undated entry matched expected-publication filter")}
}

func TestPublicInformationCursorAndRepositoryFailure(t *testing.T){
	one,two:=testInfo(),testInfo();two.ID="10000000-0000-4000-8000-000000000002";two.PublishedAt=two.PublishedAt.Add(-time.Hour)
	svc:=PublicInformation{Repository:publicInfoRepo{items:[]repository.PublicInformation{one,two}},Now:func()time.Time{return time.Date(2026,11,1,0,0,0,0,time.UTC)}}
	first,err:=svc.List(context.Background(),PublicInformationQuery{PageSize:1});if err!=nil||len(first.Items)!=1||first.NextCursor==""{t.Fatalf("first page: %+v %v",first,err)}
	second,err:=svc.List(context.Background(),PublicInformationQuery{PageSize:1,Cursor:first.NextCursor});if err!=nil||len(second.Items)!=1||second.Items[0].Information.ID==first.Items[0].Information.ID{t.Fatalf("second page: %+v %v",second,err)}
	if _,err=svc.List(context.Background(),PublicInformationQuery{PageSize:1,Institution:"another",Cursor:first.NextCursor});err!=ErrInvalidPublicQuery{t.Fatal("cursor accepted for another query")}
	svc.Repository=publicInfoRepo{fail:true};if _,err=svc.List(context.Background(),PublicInformationQuery{});err!=ErrPublicInformationUnavailable{t.Fatal("repository failure leaked")}
}

func TestPublicInformationCalendarUsesEstimatedDateAndGroupsEverySchool(t *testing.T) {
	first, second := testInfo(), testInfo()
	second.ID = "10000000-0000-4000-8000-000000000002"
	second.InstitutionName = "东岭示范大学"
	second.Title = "开放日演示"
	second.RegistrationURL = "https://example.invalid/another"
	otherMonth := second
	otherMonth.ID = "10000000-0000-4000-8000-000000000003"
	date := "2026-06-01"
	otherMonth.ExpectedPublicationOn = &date
	undated := second
	undated.ID = "10000000-0000-4000-8000-000000000004"
	undated.ExpectedPublicationOn = nil
	svc := PublicInformation{Repository: publicInfoRepo{items: []repository.PublicInformation{first, second, otherMonth, undated}}}

	calendar, err := svc.Calendar(context.Background(), "2026-05")
	if err != nil || calendar.Month != "2026-05" || len(calendar.Days) != 1 || calendar.Days[0].Date != "2026-05-20" || len(calendar.Days[0].Items) != 2 {
		t.Fatalf("expected both schools grouped on estimated date: %+v %v", calendar, err)
	}
	if calendar.Days[0].Items[0].InstitutionName == calendar.Days[0].Items[1].InstitutionName {
		t.Fatal("calendar omitted a school")
	}
	for _, month := range []string{"2026-13", "2026-00", "26-05", "0000-01", "2026-5"} {
		if _, err := svc.Calendar(context.Background(), month); err != ErrInvalidPublicQuery {
			t.Errorf("invalid month %q: %v", month, err)
		}
	}
	svc.Repository = publicInfoRepo{fail: true}
	if _, err := svc.Calendar(context.Background(), "2026-05"); err != ErrPublicInformationUnavailable {
		t.Fatal("calendar repository failure leaked")
	}
}
