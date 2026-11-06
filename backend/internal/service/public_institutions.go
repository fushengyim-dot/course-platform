package service

import (
	"context"

	"group-platform/internal/repository"
)

type PublicInstitutionDirectory struct{ Repository repository.PublicInstitutionRepository }

func (s PublicInstitutionDirectory) List(ctx context.Context) ([]repository.PublicInstitution, error) {
	if s.Repository == nil { return nil, ErrPublicInstitutionUnavailable }
	items, err := s.Repository.ListPublicInstitutions(ctx)
	if err != nil { return nil, ErrPublicInstitutionUnavailable }
	return items, nil
}

func (s PublicInstitutionDirectory) Get(ctx context.Context, id string) (repository.PublicInstitution, error) {
	if !validUUID(id) { return repository.PublicInstitution{}, ErrPublicInstitutionNotFound }
	if s.Repository == nil { return repository.PublicInstitution{}, ErrPublicInstitutionUnavailable }
	item, found, err := s.Repository.GetPublicInstitution(ctx, id)
	if err != nil { return repository.PublicInstitution{}, ErrPublicInstitutionUnavailable }
	if !found { return repository.PublicInstitution{}, ErrPublicInstitutionNotFound }
	return item, nil
}
