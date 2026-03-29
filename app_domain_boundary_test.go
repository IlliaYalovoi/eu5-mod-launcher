package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"eu5-mod-launcher/internal/graph"
	"eu5-mod-launcher/internal/repo"
	"eu5-mod-launcher/internal/service"
)

type failingLayoutRepo struct{}

func (r *failingLayoutRepo) Load(path string) (repo.LauncherLayoutData, error) {
	return repo.LauncherLayoutData{}, nil
}

func (r *failingLayoutRepo) Save(path string, layout repo.LauncherLayoutData) error {
	return errors.New("layout save failed")
}

func TestAddConstraintRejectsInvalidTargets(t *testing.T) {
	app := newReadyAppForLaunchTest(t)

	err := app.AddConstraint("", "modB")
	if err == nil {
		t.Fatalf("AddConstraint() error = nil, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "from target") {
		t.Fatalf("AddConstraint() error = %v, want from-target context", err)
	}
}

func TestAddConstraintRejectsMixedTypes(t *testing.T) {
	app := newReadyAppForLaunchTest(t)

	err := app.AddConstraint("category:graphics", "modB")
	if err == nil {
		t.Fatalf("AddConstraint() error = nil, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "categories can only constrain categories") {
		t.Fatalf("AddConstraint() error = %v, want mixed-type message", err)
	}
}

func TestSetLoadOrderRejectsCategoryIDs(t *testing.T) {
	app := newReadyAppForLaunchTest(t)

	err := app.SetLoadOrder([]string{"modA", "category:graphics"})
	if err == nil {
		t.Fatalf("SetLoadOrder() error = nil, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid mod id") {
		t.Fatalf("SetLoadOrder() error = %v, want invalid mod id context", err)
	}
}

func TestDeleteLauncherCategoryRejectsInvalidCategoryID(t *testing.T) {
	app := newReadyAppForLaunchTest(t)

	err := app.DeleteLauncherCategory("graphics")
	if err == nil {
		t.Fatalf("DeleteLauncherCategory() error = nil, want error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "delete launcher category") {
		t.Fatalf("DeleteLauncherCategory() error = %v, want category context", err)
	}
}

func TestAutosortRollsBackWhenLayoutSaveFails(t *testing.T) {
	app := newReadyAppForLaunchTest(t)
	app.initCoreServices()
	app.conGraph = graph.New()
	app.loState.OrderedIDs = []string{"mod2", "mod1"}
	app.launcherLayout = defaultLauncherLayout(app.loState.OrderedIDs)
	app.conGraph.Add("mod2", "mod1")

	app.layoutRepo = &failingLayoutRepo{}
	app.layoutSvc = service.NewLayoutService(normalizeLauncherLayout, func(layout LauncherLayout) error {
		return app.layoutRepo.Save(app.layoutPath, toRepoLayout(layout))
	})
	app.initConstraintsService()

	before := append([]string(nil), app.loState.OrderedIDs...)
	_, err := app.Autosort()
	if err == nil {
		t.Fatalf("Autosort() error = nil, want layout save failure")
	}
	if !reflect.DeepEqual(app.loState.OrderedIDs, before) {
		t.Fatalf("Autosort() state leaked: got %v, want rollback to %v", app.loState.OrderedIDs, before)
	}
}
