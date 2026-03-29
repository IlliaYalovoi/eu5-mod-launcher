package main

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkSortLayoutModIDs_SequentialVsConcurrent(b *testing.B) {
	layoutTemplate, position := buildBenchmarkLayoutAndPosition(120, 60)
	sortedCount := len(position)

	b.Run("sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			layout := cloneBenchmarkLayout(layoutTemplate)
			sortLayoutModIDsSequential(&layout, position, sortedCount)
		}
	})

	b.Run("concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			layout := cloneBenchmarkLayout(layoutTemplate)
			sortLayoutModIDsConcurrent(&layout, position, sortedCount, 8)
		}
	})
}

func BenchmarkStartupArtifactLoad_SequentialVsConcurrent(b *testing.B) {
	loadFn := func() {
		payload := make([]byte, 0, 1024)
		for i := 0; i < 2000; i++ {
			payload = append(payload, byte(i%251))
		}
		_ = len(payload)
	}

	b.Run("sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			loadFn()
			loadFn()
			loadFn()
		}
	})

	b.Run("concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			wg.Add(3)
			go func() { defer wg.Done(); loadFn() }()
			go func() { defer wg.Done(); loadFn() }()
			go func() { defer wg.Done(); loadFn() }()
			wg.Wait()
		}
	})
}

func buildBenchmarkLayoutAndPosition(categories, modsPerCategory int) (LauncherLayout, map[string]int) {
	position := make(map[string]int, categories*modsPerCategory+modsPerCategory)
	layout := LauncherLayout{Ungrouped: []string{}, Categories: make([]LauncherCategory, 0, categories), Order: []string{defaultUngroupedCategoryID}}
	counter := 0
	for i := 0; i < modsPerCategory; i++ {
		id := fmt.Sprintf("u_%04d", i)
		layout.Ungrouped = append(layout.Ungrouped, id)
		position[id] = counter
		counter++
	}
	for c := 0; c < categories; c++ {
		cat := LauncherCategory{ID: fmt.Sprintf("category:c_%03d", c), Name: fmt.Sprintf("C%03d", c), ModIDs: make([]string, 0, modsPerCategory)}
		layout.Order = append(layout.Order, cat.ID)
		for m := modsPerCategory - 1; m >= 0; m-- {
			id := fmt.Sprintf("c%03d_m%03d", c, m)
			cat.ModIDs = append(cat.ModIDs, id)
			position[id] = counter
			counter++
		}
		layout.Categories = append(layout.Categories, cat)
	}
	return layout, position
}

func cloneBenchmarkLayout(layout LauncherLayout) LauncherLayout {
	out := LauncherLayout{
		Ungrouped:  append([]string(nil), layout.Ungrouped...),
		Categories: make([]LauncherCategory, len(layout.Categories)),
		Order:      append([]string(nil), layout.Order...),
		Collapsed:  map[string]bool{},
	}
	for i := range layout.Categories {
		out.Categories[i] = LauncherCategory{
			ID:     layout.Categories[i].ID,
			Name:   layout.Categories[i].Name,
			ModIDs: append([]string(nil), layout.Categories[i].ModIDs...),
		}
	}
	for id, value := range layout.Collapsed {
		out.Collapsed[id] = value
	}
	return out
}
