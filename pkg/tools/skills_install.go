package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/domeclaw/pkg/logger"
	"github.com/sipeed/domeclaw/pkg/skills"
	"github.com/sipeed/domeclaw/pkg/utils"
)

// InstallSkillTool allows the LLM agent to install skills from registries.
// It shares the same RegistryManager that FindSkillsTool uses,
// so all registries configured in config are available for installation.
type InstallSkillTool struct {
	registryMgr *skills.RegistryManager
	workspace   string
	mu          sync.Mutex
}

// NewInstallSkillTool creates a new InstallSkillTool.
// registryMgr is the shared registry manager (same instance as FindSkillsTool).
// workspace is the root workspace directory; skills install to {workspace}/skills/{slug}/.
func NewInstallSkillTool(registryMgr *skills.RegistryManager, workspace string) *InstallSkillTool {
	return &InstallSkillTool{
		registryMgr: registryMgr,
		workspace:   workspace,
		mu:          sync.Mutex{},
	}
}

func (t *InstallSkillTool) Name() string {
	return "install_skill"
}

func (t *InstallSkillTool) Description() string {
	return "Install a skill from a registry by slug OR from a direct URL (GitHub, GitLab, ZIP file). For registry: use slug+registry. For URL: use source='https://...' with registry='url'."
}

func (t *InstallSkillTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"slug": map[string]any{
				"type":        "string",
				"description": "The unique slug of the skill (for registry install) OR leave empty for URL install",
			},
			"version": map[string]any{
				"type":        "string",
				"description": "Specific version to install (optional, defaults to latest)",
			},
			"registry": map[string]any{
				"type":        "string",
				"description": "Registry name (e.g., 'clawhub') OR 'url' for direct URL install",
			},
			"source": map[string]any{
				"type":        "string",
				"description": "Direct URL to skill ZIP/archive (e.g., GitHub URL). Use with registry='url'",
			},
			"force": map[string]any{
				"type":        "boolean",
				"description": "Force reinstall if skill already exists (default false)",
			},
		},
		"required": []string{"registry"},
	}
}

func (t *InstallSkillTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	// Install lock to prevent concurrent directory operations.
	t.mu.Lock()
	defer t.mu.Unlock()

	registryName, _ := args["registry"].(string)
	version, _ := args["version"].(string)
	force, _ := args["force"].(bool)

	// Ensure skills directory exists.
	skillsDir := filepath.Join(t.workspace, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return ErrorResult(fmt.Sprintf("failed to create skills directory: %v", err))
	}

	// Check if this is a URL-based install
	sourceURL, _ := args["source"].(string)
	if registryName == "url" || sourceURL != "" {
		return t.installFromURL(ctx, sourceURL, version, force, skillsDir)
	}

	// Registry-based install (existing logic)
	slug, _ := args["slug"].(string)
	if slug == "" {
		return ErrorResult("slug is required for registry-based install")
	}

	if err := utils.ValidateSkillIdentifier(slug); err != nil {
		return ErrorResult(fmt.Sprintf("invalid slug %q: error: %s", slug, err.Error()))
	}

	if err := utils.ValidateSkillIdentifier(registryName); err != nil {
		return ErrorResult(fmt.Sprintf("invalid registry %q: error: %s", registryName, err.Error()))
	}

	targetDir := filepath.Join(skillsDir, slug)

	if !force {
		if _, err := os.Stat(targetDir); err == nil {
			return ErrorResult(
				fmt.Sprintf("skill %q already installed at %s. Use force=true to reinstall.", slug, targetDir),
			)
		}
	} else {
		os.RemoveAll(targetDir)
	}

	registry := t.registryMgr.GetRegistry(registryName)
	if registry == nil {
		return ErrorResult(fmt.Sprintf("registry %q not found", registryName))
	}

	result, err := registry.DownloadAndInstall(ctx, slug, version, targetDir)
	if err != nil {
		rmErr := os.RemoveAll(targetDir)
		if rmErr != nil {
			logger.ErrorCF("tool", "Failed to remove partial install",
				map[string]any{
					"tool":       "install_skill",
					"target_dir": targetDir,
					"error":      rmErr.Error(),
				})
		}
		return ErrorResult(fmt.Sprintf("failed to install %q: %v", slug, err))
	}

	// Moderation: block malware.
	if result.IsMalwareBlocked {
		rmErr := os.RemoveAll(targetDir)
		if rmErr != nil {
			logger.ErrorCF("tool", "Failed to remove partial install",
				map[string]any{
					"tool":       "install_skill",
					"target_dir": targetDir,
					"error":      rmErr.Error(),
				})
		}
		return ErrorResult(fmt.Sprintf("skill %q is flagged as malicious and cannot be installed", slug))
	}

	// Write origin metadata.
	if err := writeOriginMeta(targetDir, registry.Name(), slug, result.Version); err != nil {
		logger.ErrorCF("tool", "Failed to write origin metadata",
			map[string]any{
				"tool":     "install_skill",
				"error":    err.Error(),
				"target":   targetDir,
				"registry": registry.Name(),
				"slug":     slug,
				"version":  result.Version,
			})
		_ = err
	}

	// Build result with moderation warning if suspicious.
	var output string
	if result.IsSuspicious {
		output = fmt.Sprintf("⚠️ Warning: skill %q is flagged as suspicious (may contain risky patterns).\n\n", slug)
	}
	output += fmt.Sprintf("Successfully installed skill %q v%s from %s registry.\nLocation: %s\n",
		slug, result.Version, registry.Name(), targetDir)

	if result.Summary != "" {
		output += fmt.Sprintf("Description: %s\n", result.Summary)
	}
	output += "\nThe skill is now available and can be loaded in the current session."

	return SilentResult(output)
}

// originMeta tracks which registry a skill was installed from.
type originMeta struct {
	Version          int    `json:"version"`
	Registry         string `json:"registry"`
	Slug             string `json:"slug"`
	InstalledVersion string `json:"installed_version"`
	InstalledAt      int64  `json:"installed_at"`
}

func writeOriginMeta(targetDir, registryName, slug, version string) error {
	meta := originMeta{
		Version:          1,
		Registry:         registryName,
		Slug:             slug,
		InstalledVersion: version,
		InstalledAt:      time.Now().UnixMilli(),
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(targetDir, ".skill-origin.json"), data, 0o644)
}

// installFromURL downloads and installs a skill from a direct URL (GitHub, GitLab, ZIP, etc.)
func (t *InstallSkillTool) installFromURL(ctx context.Context, sourceURL, version string, force bool, skillsDir string) *ToolResult {
	if sourceURL == "" {
		return ErrorResult("source URL is required for URL-based install. Use source='https://...'")
	}

	// Validate URL
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return ErrorResult(fmt.Sprintf("invalid URL: %v", err))
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrorResult("only http/https URLs are allowed")
	}

	// Extract skill name from URL
	slug := extractSlugFromURL(sourceURL)
	if slug == "" {
		return ErrorResult("could not determine skill name from URL")
	}

	targetDir := filepath.Join(skillsDir, slug)

	// Check if already installed
	if !force {
		if _, err := os.Stat(targetDir); err == nil {
			return ErrorResult(
				fmt.Sprintf("skill %q already installed at %s. Use force=true to reinstall.", slug, targetDir),
			)
		}
	} else {
		os.RemoveAll(targetDir)
	}

	logger.InfoCF("tool", "Installing skill from URL",
		map[string]any{
			"tool": "install_skill",
			"url":  sourceURL,
			"slug": slug,
		})

	// Download the file
	tempFile := filepath.Join(skillsDir, fmt.Sprintf("%s_temp.zip", slug))
	if err := downloadFile(ctx, sourceURL, tempFile); err != nil {
		os.RemoveAll(tempFile)
		return ErrorResult(fmt.Sprintf("failed to download from URL: %v", err))
	}
	defer os.Remove(tempFile)

	// Extract the archive
	if err := extractArchive(tempFile, targetDir); err != nil {
		os.RemoveAll(targetDir)
		return ErrorResult(fmt.Sprintf("failed to extract archive: %v", err))
	}

	// Validate skill structure
	if err := validateSkillStructure(targetDir); err != nil {
		os.RemoveAll(targetDir)
		return ErrorResult(fmt.Sprintf("invalid skill structure: %v", err))
	}

	// Write origin metadata
	meta := originMeta{
		Version:          1,
		Registry:         "url",
		Slug:             slug,
		InstalledVersion: version,
		InstalledAt:      time.Now().UnixMilli(),
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		logger.ErrorCF("tool", "Failed to marshal origin metadata",
			map[string]any{
				"tool":  "install_skill",
				"error": err.Error(),
			})
	} else {
		if err := os.WriteFile(filepath.Join(targetDir, ".skill-origin.json"), data, 0o644); err != nil {
			logger.ErrorCF("tool", "Failed to write origin metadata",
				map[string]any{
					"tool":  "install_skill",
					"error": err.Error(),
				})
		}
	}

	output := fmt.Sprintf("Successfully installed skill %q from URL.\nLocation: %s\n", slug, targetDir)
	output += "\nThe skill is now available and can be loaded in the current session."

	return SilentResult(output)
}

// extractSlugFromURL extracts a skill name from a URL
func extractSlugFromURL(sourceURL string) string {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return ""
	}

	path := parsedURL.Path
	path = strings.TrimSuffix(path, ".zip")
	path = strings.TrimSuffix(path, ".tar.gz")
	path = strings.TrimSuffix(path, ".tgz")

	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && parts[i] != "archive" && parts[i] != "-" {
			name := parts[i]
			name = strings.TrimSuffix(name, "-main")
			name = strings.TrimSuffix(name, "-master")
			name = strings.TrimSuffix(name, "-develop")
			return strings.ToLower(name)
		}
	}

	return ""
}

// downloadFile downloads a file from URL to destination
func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "DomeClaw-Skill-Installer/1.0")

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractArchive extracts a ZIP or TAR.GZ archive to destination
func extractArchive(archive, dest string) error {
	if strings.HasSuffix(archive, ".zip") {
		return extractZip(archive, dest)
	} else if strings.HasSuffix(archive, ".tar.gz") || strings.HasSuffix(archive, ".tgz") {
		return extractTarGz(archive, dest)
	}
	return fmt.Errorf("unsupported archive format")
}

// extractZip extracts a ZIP file
func extractZip(zipFile, dest string) error {
	cmd := exec.Command("unzip", "-o", zipFile, "-d", dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unzip failed: %v, output: %s", err, string(output))
	}

	entries, err := os.ReadDir(dest)
	if err != nil {
		return err
	}

	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(dest, entries[0].Name())
		files, err := os.ReadDir(subDir)
		if err != nil {
			return err
		}

		for _, file := range files {
			oldPath := filepath.Join(subDir, file.Name())
			newPath := filepath.Join(dest, file.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}
		os.Remove(subDir)
	}

	return nil
}

// extractTarGz extracts a TAR.GZ file
func extractTarGz(tarFile, dest string) error {
	cmd := exec.Command("tar", "-xzf", tarFile, "-C", dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar failed: %v, output: %s", err, string(output))
	}

	entries, err := os.ReadDir(dest)
	if err != nil {
		return err
	}

	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(dest, entries[0].Name())
		files, err := os.ReadDir(subDir)
		if err != nil {
			return err
		}

		for _, file := range files {
			oldPath := filepath.Join(subDir, file.Name())
			newPath := filepath.Join(dest, file.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}
		os.Remove(subDir)
	}

	return nil
}

// validateSkillStructure validates that a skill has required files
func validateSkillStructure(skillDir string) error {
	files := []string{"skill.json", "manifest.json", "README.md"}
	found := false

	for _, file := range files {
		if _, err := os.Stat(filepath.Join(skillDir, file)); err == nil {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("skill must have skill.json, manifest.json, or README.md")
	}

	return nil
}
