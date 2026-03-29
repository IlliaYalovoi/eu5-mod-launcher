package service

import (
	"fmt"
	"strings"

	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/repo"
)

type ConstraintsService struct {
	graph      *graph.Graph
	repo       repo.ConstraintsRepository
	path       string
	expand     func(string) []string
	isCategory func(string) bool
}

func NewConstraintsService(g *graph.Graph, constraintsPath string, repository repo.ConstraintsRepository, expand func(string) []string, isCategory func(string) bool) *ConstraintsService {
	if expand == nil {
		expand = func(string) []string { return nil }
	}
	if isCategory == nil {
		isCategory = domain.IsCategoryID
	}
	if repository == nil {
		repository = repo.NewFileConstraintsRepository()
	}
	return &ConstraintsService{graph: g, path: constraintsPath, repo: repository, expand: expand, isCategory: isCategory}
}

func (s *ConstraintsService) All() []graph.Constraint {
	if s == nil || s.graph == nil {
		return []graph.Constraint{}
	}
	return s.graph.All()
}

func (s *ConstraintsService) AddConstraint(from, to string) error {
	if err := s.validateConstraintTargets(from, to); err != nil {
		return err
	}

	fromCategory := s.isCategory(from)
	toCategory := s.isCategory(to)
	if fromCategory != toCategory {
		return fmt.Errorf("%w: categories can only constrain categories, and mods can only constrain mods", domain.ErrTypeMismatch)
	}

	if fromCategory {
		if err := s.applyWithRollback(func() error {
			return s.addAfterConstraintSingle(from, to)
		}, "save constraints after add %q -> %q", from, to); err != nil {
			return fmt.Errorf("add category constraint %q -> %q: %w", from, to, err)
		}
		return nil
	}

	fromIDs := s.expand(from)
	toIDs := s.expand(to)
	if len(fromIDs) == 0 || len(toIDs) == 0 {
		return fmt.Errorf("add constraint %q -> %q: no mods resolved from target", from, to)
	}

	if err := s.applyWithRollback(func() error {
		for _, fromID := range fromIDs {
			for _, toID := range toIDs {
				if fromID == toID {
					continue
				}
				if err := s.addAfterConstraintSingle(fromID, toID); err != nil {
					return fmt.Errorf("add constraint %q -> %q expanded as %q -> %q: %w", from, to, fromID, toID, err)
				}
			}
		}
		return nil
	}, "save constraints after add %q -> %q", from, to); err != nil {
		return err
	}
	return nil
}

func (s *ConstraintsService) RemoveConstraint(from, to string) error {
	if err := s.validateConstraintTargets(from, to); err != nil {
		return err
	}

	fromCategory := s.isCategory(from)
	toCategory := s.isCategory(to)
	if fromCategory != toCategory {
		return fmt.Errorf("%w: categories can only constrain categories, and mods can only constrain mods", domain.ErrTypeMismatch)
	}

	if fromCategory {
		return s.applyWithRollback(func() error {
			s.graph.Remove(from, to)
			return nil
		}, "save constraints after remove %q -> %q", from, to)
	}

	fromIDs := s.expand(from)
	toIDs := s.expand(to)
	return s.applyWithRollback(func() error {
		for _, fromID := range fromIDs {
			for _, toID := range toIDs {
				s.graph.Remove(fromID, toID)
			}
		}
		return nil
	}, "save constraints after remove %q -> %q", from, to)
}

func (s *ConstraintsService) AddLoadFirst(target string) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("add load-first: mod id must not be empty")
	}
	if s.isCategory(target) {
		if _, err := domain.ParseCategoryID(target); err != nil {
			return fmt.Errorf("add load-first %q: %w", target, err)
		}
		if s.graph.HasLast(target) {
			return fmt.Errorf("add load-first %q: conflict: target is already marked load last", target)
		}
		if s.graph.HasOutgoingAfter(target) {
			return fmt.Errorf("add load-first %q: conflict: target has 'loads after' dependencies", target)
		}
		return s.applyWithRollback(func() error {
			s.graph.AddFirst(target)
			return nil
		}, "save constraints after add load-first %q", target)
	}

	if _, err := domain.ParseModID(target); err != nil {
		return fmt.Errorf("add load-first %q: %w", target, err)
	}
	ids := s.expand(target)
	if len(ids) == 0 {
		return fmt.Errorf("add load-first %q: no mods resolved from target", target)
	}
	return s.applyWithRollback(func() error {
		for _, id := range ids {
			if s.graph.HasLast(id) {
				return fmt.Errorf("add load-first %q: conflict: mod %q is already marked load last", target, id)
			}
			if s.graph.HasOutgoingAfter(id) {
				return fmt.Errorf("add load-first %q: conflict: mod %q has 'loads after' dependencies", target, id)
			}
			s.graph.AddFirst(id)
		}
		return nil
	}, "save constraints after add load-first %q", target)
}

func (s *ConstraintsService) AddLoadLast(target string) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("add load-last: mod id must not be empty")
	}
	if s.isCategory(target) {
		if _, err := domain.ParseCategoryID(target); err != nil {
			return fmt.Errorf("add load-last %q: %w", target, err)
		}
		if s.graph.HasFirst(target) {
			return fmt.Errorf("add load-last %q: conflict: target is already marked load first", target)
		}
		if s.graph.HasIncomingAfter(target) {
			return fmt.Errorf("add load-last %q: conflict: target has incoming constraints", target)
		}
		return s.applyWithRollback(func() error {
			s.graph.AddLast(target)
			return nil
		}, "save constraints after add load-last %q", target)
	}

	if _, err := domain.ParseModID(target); err != nil {
		return fmt.Errorf("add load-last %q: %w", target, err)
	}
	ids := s.expand(target)
	if len(ids) == 0 {
		return fmt.Errorf("add load-last %q: no mods resolved from target", target)
	}
	return s.applyWithRollback(func() error {
		for _, id := range ids {
			if s.graph.HasFirst(id) {
				return fmt.Errorf("add load-last %q: conflict: mod %q is already marked load first", target, id)
			}
			if s.graph.HasIncomingAfter(id) {
				return fmt.Errorf("add load-last %q: conflict: mod %q has incoming constraints", target, id)
			}
			s.graph.AddLast(id)
		}
		return nil
	}, "save constraints after add load-last %q", target)
}

func (s *ConstraintsService) RemoveLoadFirst(target string) error {
	if s.isCategory(target) {
		if _, err := domain.ParseCategoryID(target); err != nil {
			return fmt.Errorf("remove load-first %q: %w", target, err)
		}
		return s.applyWithRollback(func() error {
			s.graph.RemoveFirst(target)
			return nil
		}, "save constraints after remove load-first %q", target)
	}
	if _, err := domain.ParseModID(target); err != nil {
		return fmt.Errorf("remove load-first %q: %w", target, err)
	}
	return s.applyWithRollback(func() error {
		for _, id := range s.expand(target) {
			s.graph.RemoveFirst(id)
		}
		return nil
	}, "save constraints after remove load-first %q", target)
}

func (s *ConstraintsService) RemoveLoadLast(target string) error {
	if s.isCategory(target) {
		if _, err := domain.ParseCategoryID(target); err != nil {
			return fmt.Errorf("remove load-last %q: %w", target, err)
		}
		return s.applyWithRollback(func() error {
			s.graph.RemoveLast(target)
			return nil
		}, "save constraints after remove load-last %q", target)
	}
	if _, err := domain.ParseModID(target); err != nil {
		return fmt.Errorf("remove load-last %q: %w", target, err)
	}
	return s.applyWithRollback(func() error {
		for _, id := range s.expand(target) {
			s.graph.RemoveLast(id)
		}
		return nil
	}, "save constraints after remove load-last %q", target)
}

func (s *ConstraintsService) save(format string, args ...any) error {
	if err := s.repo.Save(s.path, s.graph); err != nil {
		return fmt.Errorf(format+": %w", append(args, err)...)
	}
	return nil
}

func (s *ConstraintsService) applyWithRollback(mutate func() error, saveFormat string, saveArgs ...any) error {
	snapshot := s.graph.All()
	if err := mutate(); err != nil {
		return err
	}
	if err := s.save(saveFormat, saveArgs...); err != nil {
		s.restoreFrom(snapshot)
		return err
	}
	return nil
}

func (s *ConstraintsService) restoreFrom(snapshot []graph.Constraint) {
	current := s.graph.All()
	for _, c := range current {
		typ := c.Type
		if typ == "" || typ == graph.ConstraintTypeAfter {
			s.graph.Remove(c.From, c.To)
			continue
		}
		if typ == graph.ConstraintTypeFirst {
			s.graph.RemoveFirst(c.ModID)
			continue
		}
		if typ == graph.ConstraintTypeLast {
			s.graph.RemoveLast(c.ModID)
		}
	}

	for _, c := range snapshot {
		typ := c.Type
		if typ == "" || typ == graph.ConstraintTypeAfter {
			if c.From != "" && c.To != "" {
				s.graph.Add(c.From, c.To)
			}
			continue
		}
		if typ == graph.ConstraintTypeFirst {
			if c.ModID != "" {
				s.graph.AddFirst(c.ModID)
			}
			continue
		}
		if typ == graph.ConstraintTypeLast {
			if c.ModID != "" {
				s.graph.AddLast(c.ModID)
			}
		}
	}
}

func (s *ConstraintsService) addAfterConstraintSingle(from, to string) error {
	if from == to {
		return fmt.Errorf("source and target must differ")
	}
	if s.graph.HasFirst(from) {
		return fmt.Errorf("conflict: %q is marked load first", from)
	}
	if s.graph.HasLast(to) {
		return fmt.Errorf("conflict: %q is marked load last", to)
	}
	s.graph.Add(from, to)
	return nil
}

func (s *ConstraintsService) validateConstraintTargets(from, to string) error {
	fromCategory := s.isCategory(from)
	toCategory := s.isCategory(to)

	if fromCategory {
		if _, err := domain.ParseCategoryID(from); err != nil {
			return fmt.Errorf("from target %q: %w", from, err)
		}
	} else if _, err := domain.ParseModID(from); err != nil {
		return fmt.Errorf("from target %q: %w", from, err)
	}

	if toCategory {
		if _, err := domain.ParseCategoryID(to); err != nil {
			return fmt.Errorf("to target %q: %w", to, err)
		}
	} else if _, err := domain.ParseModID(to); err != nil {
		return fmt.Errorf("to target %q: %w", to, err)
	}

	if fromCategory != toCategory {
		return fmt.Errorf("%w: categories can only constrain categories, and mods can only constrain mods", domain.ErrTypeMismatch)
	}

	return nil
}
