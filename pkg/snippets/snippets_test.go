package snippets

import (
	"os"
	"testing"
)

func setupTestHome(t *testing.T) string {
	tempHome, err := os.MkdirTemp("", "reshell-snippets-test-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)

	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tempHome)
	})

	return tempHome
}

func TestAddRemoveSnippet(t *testing.T) {
	setupTestHome(t)

	// Add new snippet
	err := AddOrUpdate("test-snip", "echo 'hello'", "A test snippet", []string{"test", "hello"}, "bash", false)
	if err != nil {
		t.Fatalf("AddOrUpdate failed: %v", err)
	}

	// Verify it can be searched
	results, err := Search("test")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "test-snip" {
		t.Errorf("expected snippet name 'test-snip', got %s", results[0].Name)
	}

	// Toggle favorite
	err = ToggleFavorite("test-snip")
	if err != nil {
		t.Fatalf("ToggleFavorite failed: %v", err)
	}

	results, _ = Search("test-snip")
	if !results[0].Favorite {
		t.Errorf("expected snippet to be favorite")
	}

	// Remove snippet
	err = Remove("test-snip")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	results, _ = Search("test-snip")
	if len(results) != 0 {
		t.Errorf("expected snippet to be removed, but search returned %d results", len(results))
	}
}

func TestSnippetSearch(t *testing.T) {
	setupTestHome(t)

	_ = AddOrUpdate("golang-run", "go run .", "Run go program", []string{"go", "run"}, "go", false)
	_ = AddOrUpdate("docker-build", "docker build -t test .", "Build container", []string{"docker", "build"}, "bash", false)

	tests := []struct {
		query string
		count int
	}{
		{"go", 1},
		{"docker", 1},
		{"run", 1},
		{"build", 1},
		{"container", 1},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			res, err := Search(tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			if len(res) != tt.count {
				t.Errorf("Search(%q) returned %d items, expected %d", tt.query, len(res), tt.count)
			}
		})
	}
}

func TestEditSnippetFields(t *testing.T) {
	setupTestHome(t)

	// Add initial snippet
	err := AddOrUpdate("test-edit", "echo 'orig'", "Orig desc", []string{"orig"}, "bash", false)
	if err != nil {
		t.Fatalf("AddOrUpdate failed: %v", err)
	}

	// Update the snippet fields (preserve code & favorite)
	err = AddOrUpdate("test-edit", "echo 'orig'", "New desc", []string{"new", "edit"}, "python", true)
	if err != nil {
		t.Fatalf("AddOrUpdate update failed: %v", err)
	}

	// Verify update
	results, err := Search("test-edit")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(results))
	}

	snip := results[0]
	if snip.Code != "echo 'orig'" {
		t.Errorf("expected code to be preserved, got %s", snip.Code)
	}
	if snip.Description != "New desc" {
		t.Errorf("expected description to be updated, got %s", snip.Description)
	}
	if len(snip.Tags) != 2 || snip.Tags[0] != "new" || snip.Tags[1] != "edit" {
		t.Errorf("expected tags to be updated, got %v", snip.Tags)
	}
	if snip.Language != "python" {
		t.Errorf("expected language to be updated, got %s", snip.Language)
	}
	if !snip.Favorite {
		t.Errorf("expected favorite to be preserved as true")
	}
}
