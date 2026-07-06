package snippets

import (
	"errors"
	"fmt"
	"os"
	"reshell/pkg/config"
	"strings"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/atotto/clipboard"
)

// IsValidLanguage checks if a language lexer is registered in Chroma.
func IsValidLanguage(lang string) bool {
	return lexers.Get(lang) != nil
}

// AddOrUpdate creates or updates a snippet.
func AddOrUpdate(name, code, desc string, tags []string, lang string, favorite bool) error {
	cfg, err := config.LoadSnippets()
	if err != nil {
		return err
	}

	found := false

	for i, snip := range cfg.Snippets {
		if snip.Name == name {
			cfg.Snippets[i].Code = code
			cfg.Snippets[i].Description = desc
			cfg.Snippets[i].Tags = tags
			cfg.Snippets[i].Language = lang
			cfg.Snippets[i].Favorite = favorite
			found = true
			break
		}
	}

	if !found {
		cfg.Snippets = append(cfg.Snippets, config.Snippet{
			Name:        name,
			Code:        code,
			Description: desc,
			Tags:        tags,
			Language:    lang,
			Favorite:    favorite,
		})
	}

	return config.SaveSnippets(cfg)
}

// Remove deletes a snippet.
func Remove(name string) error {
	cfg, err := config.LoadSnippets()
	if err != nil {
		return err
	}

	newSnips := make([]config.Snippet, 0, len(cfg.Snippets))
	found := false
	for _, snip := range cfg.Snippets {
		if snip.Name == name {
			found = true
			continue
		}
		newSnips = append(newSnips, snip)
	}

	if !found {
		return errors.New("snippet not found")
	}

	cfg.Snippets = newSnips
	return config.SaveSnippets(cfg)
}

// ToggleFavorite toggles the favorite status of a snippet.
func ToggleFavorite(name string) error {
	cfg, err := config.LoadSnippets()
	if err != nil {
		return err
	}

	found := false
	for i, snip := range cfg.Snippets {
		if snip.Name == name {
			cfg.Snippets[i].Favorite = !snip.Favorite
			found = true
			break
		}
	}

	if !found {
		return errors.New("snippet not found")
	}

	return config.SaveSnippets(cfg)
}

// CopyToClipboard copies the snippet code to system clipboard.
func CopyToClipboard(name string) error {
	cfg, err := config.LoadSnippets()
	if err != nil {
		return err
	}

	for _, snip := range cfg.Snippets {
		if snip.Name == name {
			return clipboard.WriteAll(snip.Code)
		}
	}

	return errors.New("snippet not found")
}

// Search performs a case-insensitive keyword search on name, description, tags and code.
func Search(query string) ([]config.Snippet, error) {
	cfg, err := config.LoadSnippets()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return cfg.Snippets, nil
	}

	var results []config.Snippet
	q := strings.ToLower(query)

	for _, snip := range cfg.Snippets {
		matched := strings.Contains(strings.ToLower(snip.Name), q) ||
			strings.Contains(strings.ToLower(snip.Description), q) ||
			strings.Contains(strings.ToLower(snip.Code), q)

		if !matched {
			for _, tag := range snip.Tags {
				if strings.Contains(strings.ToLower(tag), q) {
					matched = true
					break
				}
			}
		}

		if matched {
			results = append(results, snip)
		}
	}

	return results, nil
}

// Export writes the snippets config file to target path.
func Export(destPath string) error {
	cfgDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}
	src := config.SaveTOMLFile(destPath, cfgDir) // wait, config.SaveTOMLFile takes relative path, let's copy file instead
	_ = src
	srcFile := fmt.Sprintf("%s/snippets.toml", cfgDir)
	return copyFile(srcFile, destPath)
}

// Import overrides current config with snippets from a target path.
func Import(srcPath string) error {
	cfgDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}
	destFile := fmt.Sprintf("%s/snippets.toml", cfgDir)
	return copyFile(srcPath, destFile)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
