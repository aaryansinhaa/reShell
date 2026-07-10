package config

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type ParsedAlias struct {
	Name   string
	Value  string
	Source string
}

type ParsedEnvVar struct {
	Name   string
	Value  string
	Source string
}

type ParsedFunction struct {
	Name   string
	Code   string
	Source string
	Shell  string // "bash" or "fish"
}

type ParsedSnippet struct {
	Name        string
	Code        string
	Description string
	Tags        []string
	Source      string
}

type DiscoveryResults struct {
	Aliases   []ParsedAlias
	EnvVars   []ParsedEnvVar
	Functions []ParsedFunction
	Snippets  []ParsedSnippet
}

var (
	// Bash/Zsh: alias name='value' or alias name="value" or alias name=value
	bashAliasRegex = regexp.MustCompile(`^\s*alias\s+([a-zA-Z0-9_\-\.]+)\s*=\s*(.*)$`)
	// Fish: alias name 'value' or alias name "value" or alias name value
	fishAliasRegex = regexp.MustCompile(`^\s*alias\s+([a-zA-Z0-9_\-\.]+)\s+(.*)$`)

	// Bash/Zsh env: export NAME="value" or export NAME='value' or export NAME=value
	bashEnvRegex = regexp.MustCompile(`^\s*export\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*(.*)$`)
	// Fish env: set -gx NAME "value" or set -x NAME value or setenv NAME value
	fishSetEnvRegex = regexp.MustCompile(`^\s*set\s+(?:-gx|-x|--export)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+(.*)$`)
	fishSetenvRegex = regexp.MustCompile(`^\s*setenv\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+(.*)$`)

	// Function start regex
	// Bash/Zsh: function name() { or function name { or name() {
	bashFuncStartRegex = regexp.MustCompile(`^\s*(?:function\s+)?([a-zA-Z0-9_\-]+)\s*(?:\(\))?\s*\{\s*$`)
	// Fish: function name
	fishFuncStartRegex = regexp.MustCompile(`^\s*function\s+([a-zA-Z0-9_\-]+)\s*$`)
)

func stripQuotes(val string) string {
	val = strings.TrimSpace(val)
	if len(val) >= 2 {
		first := val[0]
		last := val[len(val)-1]
		if (first == '\'' && last == '\'') || (first == '"' && last == '"') {
			return val[1 : len(val)-1]
		}
	}
	return val
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			sb.WriteRune(r)
		} else if r == ' ' {
			sb.WriteRune('-')
		}
	}
	res := sb.String()
	res = regexp.MustCompile(`-+`).ReplaceAllString(res, "-")
	return strings.Trim(res, "-")
}

func IsSecret(name, val string) bool {
	lowerName := strings.ToLower(name)
	for _, term := range []string{"key", "secret", "password", "token", "auth", "pass", "credential"} {
		if strings.Contains(lowerName, term) {
			return true
		}
	}
	valTrim := strings.TrimSpace(val)
	valLower := strings.ToLower(valTrim)
	if strings.HasPrefix(valLower, "ghp_") ||
		strings.HasPrefix(valLower, "aws_") ||
		strings.HasPrefix(valLower, "akias") ||
		strings.HasPrefix(valLower, "sk_live") {
		return true
	}
	return false
}
func CleanLineForBracesAndKeywords(line string, inSingleQuote, inDoubleQuote *bool) string {
	var sb strings.Builder
	runes := []rune(line)
	n := len(runes)
	escaped := false

	for i := 0; i < n; i++ {
		r := runes[i]

		if escaped {
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if *inSingleQuote {
			if r == '\'' {
				*inSingleQuote = false
			}
			continue
		}

		if *inDoubleQuote {
			if r == '"' {
				*inDoubleQuote = false
			}
			continue
		}

		if r == '\'' {
			*inSingleQuote = true
			continue
		}

		if r == '"' {
			*inDoubleQuote = true
			continue
		}

		if r == '#' {
			// Start of comment, discard the rest of the line
			break
		}

		sb.WriteRune(r)
	}

	return sb.String()
}

func parseBashFile(filePath string, results *DiscoveryResults) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var inFunction bool
	var funcName string
	var funcBody []string
	var braceCount int
	var inSingleQuote bool
	var inDoubleQuote bool

	for scanner.Scan() {
		line := scanner.Text()

		if !inFunction {
			// Check for alias
			if matches := bashAliasRegex.FindStringSubmatch(line); matches != nil {
				results.Aliases = append(results.Aliases, ParsedAlias{
					Name:   matches[1],
					Value:  stripQuotes(matches[2]),
					Source: filePath,
				})
				continue
			}

			// Check for env var
			if matches := bashEnvRegex.FindStringSubmatch(line); matches != nil {
				results.EnvVars = append(results.EnvVars, ParsedEnvVar{
					Name:   matches[1],
					Value:  stripQuotes(matches[2]),
					Source: filePath,
				})
				continue
			}

			// Check for function start
			if matches := bashFuncStartRegex.FindStringSubmatch(line); matches != nil {
				inFunction = true
				funcName = matches[1]
				funcBody = []string{line}
				inSingleQuote = false
				inDoubleQuote = false
				cleaned := CleanLineForBracesAndKeywords(line, &inSingleQuote, &inDoubleQuote)
				braceCount = strings.Count(cleaned, "{") - strings.Count(cleaned, "}")
				if braceCount == 0 {
					results.Functions = append(results.Functions, ParsedFunction{
						Name:   funcName,
						Code:   strings.Join(funcBody, "\n"),
						Source: filePath,
						Shell:  "bash",
					})
					inFunction = false
					inSingleQuote = false
					inDoubleQuote = false
				}
				continue
			}
		} else {
			funcBody = append(funcBody, line)
			cleaned := CleanLineForBracesAndKeywords(line, &inSingleQuote, &inDoubleQuote)
			braceCount += strings.Count(cleaned, "{") - strings.Count(cleaned, "}")
			if braceCount <= 0 {
				results.Functions = append(results.Functions, ParsedFunction{
					Name:   funcName,
					Code:   strings.Join(funcBody, "\n"),
					Source: filePath,
					Shell:  "bash",
				})
				inFunction = false
				inSingleQuote = false
				inDoubleQuote = false
			}
		}
	}
	return scanner.Err()
}

func updateFishDepth(line string, depth int) int {
	words := strings.Fields(line)
	for _, w := range words {
		w = strings.Trim(w, ";()")
		if w == "end" {
			depth--
		} else if w == "function" || w == "if" || w == "for" || w == "while" || w == "switch" || w == "begin" {
			depth++
		}
	}
	return depth
}

func parseFishFile(filePath string, results *DiscoveryResults) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var inFunction bool
	var funcName string
	var funcBody []string
	var depth int
	var inSingleQuote bool
	var inDoubleQuote bool

	for scanner.Scan() {
		line := scanner.Text()

		if !inFunction {
			// Check for alias
			if matches := fishAliasRegex.FindStringSubmatch(line); matches != nil {
				results.Aliases = append(results.Aliases, ParsedAlias{
					Name:   matches[1],
					Value:  stripQuotes(matches[2]),
					Source: filePath,
				})
				continue
			}

			// Check for env var: set -gx or setenv
			if matches := fishSetEnvRegex.FindStringSubmatch(line); matches != nil {
				results.EnvVars = append(results.EnvVars, ParsedEnvVar{
					Name:   matches[1],
					Value:  stripQuotes(matches[2]),
					Source: filePath,
				})
				continue
			}
			if matches := fishSetenvRegex.FindStringSubmatch(line); matches != nil {
				results.EnvVars = append(results.EnvVars, ParsedEnvVar{
					Name:   matches[1],
					Value:  stripQuotes(matches[2]),
					Source: filePath,
				})
				continue
			}

			// Check for function start
			if matches := fishFuncStartRegex.FindStringSubmatch(line); matches != nil {
				inFunction = true
				funcName = matches[1]
				funcBody = []string{line}
				depth = 1
				inSingleQuote = false
				inDoubleQuote = false
				continue
			}
		} else {
			funcBody = append(funcBody, line)
			cleaned := CleanLineForBracesAndKeywords(line, &inSingleQuote, &inDoubleQuote)
			depth = updateFishDepth(cleaned, depth)
			if depth <= 0 {
				results.Functions = append(results.Functions, ParsedFunction{
					Name:   funcName,
					Code:   strings.Join(funcBody, "\n"),
					Source: filePath,
					Shell:  "fish",
				})
				inFunction = false
				inSingleQuote = false
				inDoubleQuote = false
			}
		}
	}
	return scanner.Err()
}

type VSCodeSnippet struct {
	Prefix      interface{} `json:"prefix"`
	Body        interface{} `json:"body"`
	Description string      `json:"description"`
}

func parseVSCodeSnippetFile(filePath string, results *DiscoveryResults) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var snippetsMap map[string]VSCodeSnippet
	if err := json.Unmarshal(data, &snippetsMap); err != nil {
		return err
	}

	for name, snip := range snippetsMap {
		var bodyStr string
		switch b := snip.Body.(type) {
		case string:
			bodyStr = b
		case []interface{}:
			var lines []string
			for _, l := range b {
				if s, ok := l.(string); ok {
					lines = append(lines, s)
				}
			}
			bodyStr = strings.Join(lines, "\n")
		default:
			continue
		}

		if bodyStr == "" {
			continue
		}

		desc := snip.Description
		if desc == "" {
			desc = name
		}

		results.Snippets = append(results.Snippets, ParsedSnippet{
			Name:        slugify(name),
			Code:        bodyStr,
			Description: desc,
			Tags:        []string{"vscode-imported"},
			Source:      filePath,
		})
	}
	return nil
}

type PetSnippet struct {
	Description string   `toml:"description"`
	Command     string   `toml:"command"`
	Tag         []string `toml:"tag"`
}

type PetConfig struct {
	Snippets []PetSnippet `toml:"snippets"`
}

func parsePetSnippetFile(filePath string, results *DiscoveryResults) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var petCfg PetConfig
	if err := toml.Unmarshal(data, &petCfg); err != nil {
		return err
	}

	for _, snip := range petCfg.Snippets {
		if snip.Command == "" {
			continue
		}

		name := slugify(snip.Description)
		if name == "" {
			name = "pet-imported-snippet"
		}

		results.Snippets = append(results.Snippets, ParsedSnippet{
			Name:        name,
			Code:        snip.Command,
			Description: snip.Description,
			Tags:        append([]string{"pet-imported"}, snip.Tag...),
			Source:      filePath,
		})
	}
	return nil
}

func WalkAndParse(root string) (DiscoveryResults, error) {
	var results DiscoveryResults

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == ".venv" ||
				name == ".cargo" || name == ".idea" || name == ".vscode" ||
				name == "cache" || name == "caches" || name == "tmp" || name == "temp" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		base := filepath.Base(path)

		switch ext {
		case ".bashrc", ".zshrc", ".profile", ".bash_profile", ".bash_aliases", ".sh":
			_ = parseBashFile(path, &results)
		case ".fish":
			_ = parseFishFile(path, &results)
		case ".json":
			pathLower := strings.ToLower(path)
			if strings.Contains(pathLower, "code/user/snippets") ||
				strings.Contains(pathLower, "vscodium/user/snippets") {
				_ = parseVSCodeSnippetFile(path, &results)
			}
		case ".code-snippets":
			_ = parseVSCodeSnippetFile(path, &results)
		case ".toml":
			if base == "snippet.toml" || base == "pet.toml" {
				_ = parsePetSnippetFile(path, &results)
			}
		}

		if base == "config.fish" {
			_ = parseFishFile(path, &results)
		}

		return nil
	})

	return results, err
}

func DiscoverDefault(homeDir string) (DiscoveryResults, error) {
	var results DiscoveryResults

	// 1. Scan home shell files
	shellFiles := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".profile"),
		filepath.Join(homeDir, ".bash_profile"),
		filepath.Join(homeDir, ".bash_aliases"),
		filepath.Join(homeDir, ".config", "fish", "config.fish"),
	}

	for _, file := range shellFiles {
		if _, err := os.Stat(file); err == nil {
			if strings.HasSuffix(file, ".fish") || strings.Contains(file, "config.fish") {
				_ = parseFishFile(file, &results)
			} else {
				_ = parseBashFile(file, &results)
			}
		}
	}

	// 2. Scan VS Code snippets globally
	vscodeDirs := []string{
		filepath.Join(homeDir, ".config", "Code", "User", "snippets"),
		filepath.Join(homeDir, ".config", "Code - Insiders", "User", "snippets"),
		filepath.Join(homeDir, ".config", "VSCodium", "User", "snippets"),
	}

	for _, dir := range vscodeDirs {
		if _, err := os.Stat(dir); err == nil {
			_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				ext := filepath.Ext(path)
				if ext == ".json" || ext == ".code-snippets" {
					_ = parseVSCodeSnippetFile(path, &results)
				}
				return nil
			})
		}
	}

	// 3. Scan Pet snippets globally
	petFiles := []string{
		filepath.Join(homeDir, ".config", "pet", "snippet.toml"),
		filepath.Join(homeDir, ".config", "pet", "pet.toml"),
	}
	for _, file := range petFiles {
		if _, err := os.Stat(file); err == nil {
			_ = parsePetSnippetFile(file, &results)
		}
	}

	// 4. Scan ~/.config recursively (excluding large/cache dirs)
	configDir := filepath.Join(homeDir, ".config")
	if _, err := os.Stat(configDir); err == nil {
		_ = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "Code" ||
					name == "Code - Insiders" || name == "VSCodium" || name == "pet" ||
					name == "cache" || name == "caches" || name == "tmp" || name == "temp" {
					return filepath.SkipDir
				}
				return nil
			}

			ext := filepath.Ext(path)
			base := filepath.Base(path)

			switch ext {
			case ".sh":
				_ = parseBashFile(path, &results)
			case ".fish":
				_ = parseFishFile(path, &results)
			case ".code-snippets":
				_ = parseVSCodeSnippetFile(path, &results)
			case ".toml":
				if base == "snippet.toml" || base == "pet.toml" {
					_ = parsePetSnippetFile(path, &results)
				}
			}
			return nil
		})
	}

	return results, nil
}

// ReadConfigFromReader loads configuration data from an io.Reader
func ReadConfigFromReader(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, v)
}
