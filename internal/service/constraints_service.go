package service

import (
	"errors"
	"eu5-mod-launcher/internal/domain"
	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/repo"
	"fmt"
	"strings"
)

var (
	errNoModsResolved             = errors.New("no mods resolved from target")
	errLoadFirstModIDEmpty        = errors.New("add load-first: mod id must not be empty")
	errLoadLastModIDEmpty         = errors.New("add load-last: mod id must not be empty")
	errConflictAlreadyLoadFirst   = errors.New("conflict: target is already marked load first")
	errConflictAlreadyLoadLast    = errors.New("conflict: target is already marked load last")
	errConflictHasOutgoingAfter   = errors.New("conflict: target has loads-after dependencies")
	errConflictHasIncomingAfter   = errors.New("conflict: target has incoming constraints")
	errConflictModLoadFirst       = errors.New("conflict: mod is already marked load first")
	errConflictModLoadLast        = errors.New("conflict: mod is already marked load last")
	errConflictModOutgoingAfter   = errors.New("conflict: mod has loads-after dependencies")
	errConflictModIncomingAfter   = errors.New("conflict: mod has incoming constraints")
	errSourceAndTargetMustDiffer  = errors.New("source and target must differ")
	errCrossCategoryModConstraint = errors.New("mod constraints must stay in same category")
	errSaveConstraints            = errors.New("save constraints")
)

const errTypeMismatchMsg = "categories can only constrain categories, and mods can only constrain mods"

type ConstraintsService struct {
	graph       *graph.Graph
	repo        repo.ConstraintsRepository
	path        string
	expand      func(string) []string
	isCategory  func(string) bool
	modCategory func(string) string
}

func NewConstraintsService(
	constraintGraph *graph.Graph,
	constraintsPath string,
	repository repo.ConstraintsRepository,
	expand func(string) []string,
	isCategory func(string) bool,
	modCategory func(string) string,
) *ConstraintsService {
	if expand == nil {
		expand = func(string) []string { return nil }
	}
	if isCategory == nil {
		isCategory = domain.IsCategoryID
	}
	if repository == nil {
		repository = repo.NewFileConstraintsRepository()
	}
	if modCategory == nil {
		modCategory = func(string) string { return "" }
	}
	return &ConstraintsService{
		graph:       constraintGraph,
		path:        constraintsPath,
		repo:        repository,
		expand:      expand,
		isCategory:  isCategory,
		modCategory: modCategory,
	}
}

func (s *ConstraintsService) All() []graph.Constraint {
	if s == nil || s.graph == nil {
		return []graph.Constraint{}
	}
	return s.graph.All()
}

func (s *ConstraintsService) AddConstraint(from, target string) error {
	if err := s.validateConstraintTargets(from, target); err != nil {
		return err
	}

	fromCategory := s.isCategory(from)
	toCategory := s.isCategory(target)
	if fromCategory != toCategory {
		return fmt.Errorf("%w: %s", domain.ErrTypeMismatch, errTypeMismatchMsg)
	}

	if fromCategory {
		if err := s.applyWithRollback(func() error {
			return s.addAfterConstraintSingle(from, target)
		}, "save constraints after add %q -> %q", from, target); err != nil {
			return fmt.Errorf("add category constraint %q -> %q: %w", from, target, err)
		}
		return nil
	}

	fromIDs := s.expand(from)
	toIDs := s.expand(target)
	if len(fromIDs) == 0 || len(toIDs) == 0 {
		return fmt.Errorf("add constraint %q -> %q: %w", from, target, errNoModsResolved)
	}
	if err := s.validateSameCategory(fromIDs, toIDs); err != nil {
		return fmt.Errorf("add constraint %q -> %q: %w", from, target, err)
	}

	return s.applyWithRollback(func() error {
		for _, fromID := range fromIDs {
			for _, toID := range toIDs {
				if fromID == toID {
					continue
				}
				if err := s.addAfterConstraintSingle(fromID, toID); err != nil {
					return fmt.Errorf(
						"add constraint %q -> %q expanded as %q -> %q: %w",
						from,
						target,
						fromID,
						toID,
						err,
					)
				}
			}
		}
		return nil
	}, "save constraints after add %q -> %q", from, target)
}

func (s *ConstraintsService) RemoveConstraint(from, target string) error {
	if err := s.validateConstraintTargets(from, target); err != nil {
		return err
	}

	fromCategory := s.isCategory(from)
	toCategory := s.isCategory(target)
	if fromCategory != toCategory {
		return fmt.Errorf("%w: %s", domain.ErrTypeMismatch, errTypeMismatchMsg)
	}

	if fromCategory {
		return s.applyWithRollback(func() error {
			s.graph.Remove(from, target)
			return nil
		}, "save constraints after remove %q -> %q", from, target)
	}

	fromIDs := s.expand(from)
	toIDs := s.expand(target)
	return s.applyWithRollback(func() error {
		for _, fromID := range fromIDs {
			for _, toID := range toIDs {
				s.graph.Remove(fromID, toID)
			}
		}
		return nil
	}, "save constraints after remove %q -> %q", from, target)
}

func (s *ConstraintsService) AddLoadFirst(target string) error {
	if strings.TrimSpace(target) == "" {
		return errLoadFirstModIDEmpty
	}
	if s.isCategory(target) {
		if _, err := domain.ParseCategoryID(target); err != nil {
			return fmt.Errorf("add load-first %q: %w", target, err)
		}
		if s.graph.HasLast(target) {
			return fmt.Errorf("add load-first %q: %w", target, errConflictAlreadyLoadLast)
		}
		if s.graph.HasOutgoingAfter(target) {
			return fmt.Errorf("add load-first %q: %w", target, errConflictHasOutgoingAfter)
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
		return fmt.Errorf("add load-first %q: %w", target, errNoModsResolved)
	}
	return s.applyWithRollback(func() error {
		for _, id := range ids {
			if s.graph.HasLast(id) {
				return fmt.Errorf("add load-first %q (mod %q): %w", target, id, errConflictModLoadLast)
			}
			if s.graph.HasOutgoingAfter(id) {
				return fmt.Errorf("add load-first %q (mod %q): %w", target, id, errConflictModOutgoingAfter)
			}
			s.graph.AddFirst(id)
		}
		return nil
	}, "save constraints after add load-first %q", target)
}

func (s *ConstraintsService) AddLoadLast(target string) error {
	if strings.TrimSpace(target) == "" {
		return errLoadLastModIDEmpty
	}
	if s.isCategory(target) {
		if _, err := domain.ParseCategoryID(target); err != nil {
			return fmt.Errorf("add load-last %q: %w", target, err)
		}
		if s.graph.HasFirst(target) {
			return fmt.Errorf("add load-last %q: %w", target, errConflictAlreadyLoadFirst)
		}
		if s.graph.HasIncomingAfter(target) {
			return fmt.Errorf("add load-last %q: %w", target, errConflictHasIncomingAfter)
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
		return fmt.Errorf("add load-last %q: %w", target, errNoModsResolved)
	}
	return s.applyWithRollback(func() error {
		for _, id := range ids {
			if s.graph.HasFirst(id) {
				return fmt.Errorf("add load-last %q (mod %q): %w", target, id, errConflictModLoadFirst)
			}
			if s.graph.HasIncomingAfter(id) {
				return fmt.Errorf("add load-last %q (mod %q): %w", target, id, errConflictModIncomingAfter)
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

func (s *ConstraintsService) validateSameCategory(fromIDs, toIDs []string) error {
	for _, fromID := range fromIDs {
		fromCategory := strings.TrimSpace(s.modCategory(fromID))
		if fromCategory == "" {
			continue
		}
		for _, toID := range toIDs {
			toCategory := strings.TrimSpace(s.modCategory(toID))
			if toCategory == "" {
				continue
			}
			if fromCategory != toCategory {
				return fmt.Errorf(
					"%w: %q (%s) -> %q (%s)",
					errCrossCategoryModConstraint,
					fromID,
					fromCategory,
					toID,
					toCategory,
				)
			}
		}
	}
	return nil
}

func (s *ConstraintsService) save(format string, args ...any) error {
	if err := s.repo.Save(s.path, s.graph); err != nil {
		context := fmt.Sprintf(format, args...)
		return fmt.Errorf("%w: %s: %w", errSaveConstraints, context, err)
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
	clearConstraintGraph(s.graph)
	applyConstraintSnapshot(s.graph, snapshot)
}

func clearConstraintGraph(constraintGraph *graph.Graph) {
	current := constraintGraph.All()
	for i := range current {
		removeConstraint(constraintGraph, current[i])
	}
}

func applyConstraintSnapshot(constraintGraph *graph.Graph, snapshot []graph.Constraint) {
	for i := range snapshot {
		addConstraint(constraintGraph, snapshot[i])
	}
}

func removeConstraint(constraintGraph *graph.Graph, constraint graph.Constraint) {
	switch normalizeConstraintType(constraint.Type) {
	case graph.ConstraintTypeFirst:
		constraintGraph.RemoveFirst(constraint.ModID)
	case graph.ConstraintTypeLast:
		constraintGraph.RemoveLast(constraint.ModID)
	default:
		constraintGraph.Remove(constraint.From, constraint.To)
	}
}

func addConstraint(constraintGraph *graph.Graph, constraint graph.Constraint) {
	switch normalizeConstraintType(constraint.Type) {
	case graph.ConstraintTypeFirst:
		if constraint.ModID != "" {
			constraintGraph.AddFirst(constraint.ModID)
		}
	case graph.ConstraintTypeLast:
		if constraint.ModID != "" {
			constraintGraph.AddLast(constraint.ModID)
		}
	default:
		if constraint.From != "" && constraint.To != "" {
			constraintGraph.Add(constraint.From, constraint.To)
		}
	}
}

func normalizeConstraintType(constraintType string) string {
	if constraintType == "" {
		return graph.ConstraintTypeAfter
	}

	return constraintType
}

func (s *ConstraintsService) addAfterConstraintSingle(from, target string) error {
	if from == target {
		return errSourceAndTargetMustDiffer
	}
	if s.graph.HasFirst(from) {
		return fmt.Errorf("%w: %q", errConflictAlreadyLoadFirst, from)
	}
	if s.graph.HasLast(target) {
		return fmt.Errorf("%w: %q", errConflictAlreadyLoadLast, target)
	}
	s.graph.Add(from, target)
	return nil
}

func (s *ConstraintsService) validateConstraintTargets(from, target string) error {
	fromCategory := s.isCategory(from)
	toCategory := s.isCategory(target)

	if fromCategory {
		if _, err := domain.ParseCategoryID(from); err != nil {
			return fmt.Errorf("from target %q: %w", from, err)
		}
	} else if _, err := domain.ParseModID(from); err != nil {
		return fmt.Errorf("from target %q: %w", from, err)
	}

	if toCategory {
		if _, err := domain.ParseCategoryID(target); err != nil {
			return fmt.Errorf("to target %q: %w", target, err)
		}
	} else if _, err := domain.ParseModID(target); err != nil {
		return fmt.Errorf("to target %q: %w", target, err)
	}

	if fromCategory != toCategory {
		return fmt.Errorf("%w: %s", domain.ErrTypeMismatch, errTypeMismatchMsg)
	}

	return nil
}
