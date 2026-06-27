package plugin

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed opencode_plugin.js
var opencodePluginJS []byte

type HostPlugin struct {
	Name        string
	DisplayName string
	ConfigPaths []string
	Install     func(socketPath string) error
	Uninstall   func() error
	IsInstalled func() bool
}

var hosts = []*HostPlugin{
	claudeCode(),
	codex(),
	cursor(),
	windsurf(),
	antigravity(),
	copilot(),
	aider(),
	zed(),
	opencode(),
	continu(),
	cline(),
	goose(),
	pearai(),
}

func List() []*HostPlugin {
	return hosts
}

func Get(name string) *HostPlugin {
	for _, h := range hosts {
		if strings.EqualFold(h.Name, name) {
			return h
		}
	}
	return nil
}

func Installed() []*HostPlugin {
	var installed []*HostPlugin
	for _, h := range hosts {
		if h.IsInstalled() {
			installed = append(installed, h)
		}
	}
	return installed
}

// --- Claude Code ---

func claudeCode() *HostPlugin {
	return &HostPlugin{
		Name:        "claude-code",
		DisplayName: "Claude Code",
		ConfigPaths: configPaths("claude", "settings.json"),
		Install:     installClaudeCode,
		Uninstall:   uninstallClaudeCode,
		IsInstalled: isClaudeCodeInstalled,
	}
}

func claudeConfigPath() (string, error) {
	home, err := userHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// claudeSettings mirrors the shape of ~/.claude/settings.json.
// Hooks use PascalCase keys matching Claude Code's hook event names.
type claudeSettings struct {
	Hooks map[string][]claudeHookEntry `json:"hooks,omitempty"`
}

type claudeHookEntry struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []claudeHookCmd `json:"hooks"`
}

type claudeHookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func slashHookEntries() []claudeHookEntry {
	return []claudeHookEntry{
		{Hooks: []claudeHookCmd{{Type: "command", Command: "slash hook"}}},
	}
}

func isClaudeCodeInstalled() bool {
	path, err := claudeConfigPath()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "slash hook")
}

func installClaudeCode(socketPath string) error {
	path, err := claudeConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	var s claudeSettings
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &s)
	}
	if s.Hooks == nil {
		s.Hooks = make(map[string][]claudeHookEntry)
	}
	s.Hooks["PreToolUse"] = slashHookEntries()
	s.Hooks["PostToolUse"] = slashHookEntries()

	data, err = json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func uninstallClaudeCode() error {
	path, err := claudeConfigPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var s claudeSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	delete(s.Hooks, "PreToolUse")
	delete(s.Hooks, "PostToolUse")
	if len(s.Hooks) == 0 {
		s.Hooks = nil
	}
	data, err = json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// --- Codex ---

func codex() *HostPlugin {
	return &HostPlugin{
		Name:        "codex",
		DisplayName: "Codex (OpenAI)",
		ConfigPaths: configPaths("codex", "hooks.json"),
		Install:     installCodex,
		Uninstall:   uninstallCodex,
		IsInstalled: isCodexInstalled,
	}
}

func codexConfigPath() (string, error) {
	home, err := userHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex", "hooks.json"), nil
}

type codexHooks struct {
	PreToolUse  string `json:"pre_tool_use"`
	PostToolUse string `json:"post_tool_use"`
}

func isCodexInstalled() bool {
	path, err := codexConfigPath()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "slash")
}

func installCodex(socketPath string) error {
	path, err := codexConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	h := codexHooks{PreToolUse: "slash hook", PostToolUse: "slash hook"}
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func uninstallCodex() error {
	path, err := codexConfigPath()
	if err != nil {
		return nil
	}
	return os.Remove(path)
}

// --- Cursor ---

func cursor() *HostPlugin {
	return &HostPlugin{
		Name:        "cursor",
		DisplayName: "Cursor",
		ConfigPaths: configPaths("Cursor", "User", "settings.json"),
		Install:     installCursorLike("Cursor"),
		Uninstall:   uninstallCursorLike("Cursor"),
		IsInstalled: isCursorLikeInstalled("Cursor"),
	}
}


// --- Windsurf ---

func windsurf() *HostPlugin {
	return &HostPlugin{
		Name:        "windsurf",
		DisplayName: "Windsurf",
		ConfigPaths: configPaths("Windsurf", "User", "settings.json"),
		Install:     installCursorLike("Windsurf"),
		Uninstall:   uninstallCursorLike("Windsurf"),
		IsInstalled: isCursorLikeInstalled("Windsurf"),
	}
}

func cursorLikeConfigPath(app string) (string, error) {
	home, err := userHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", app, "User", "settings.json"), nil
}

type vscodeSettings struct {
	Hooks *vscodeHooks `json:"hooks,omitempty"`
}

type vscodeHooks struct {
	PreToolUse  string `json:"preToolUse,omitempty"`
	PostToolUse string `json:"postToolUse,omitempty"`
}

func installCursorLike(app string) func(string) error {
	return func(socketPath string) error {
		path, err := cursorLikeConfigPath(app)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		var s vscodeSettings
		data, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(data, &s)
		}
		s.Hooks = &vscodeHooks{PreToolUse: "slash hook", PostToolUse: "slash hook"}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0600)
	}
}

func uninstallCursorLike(app string) func() error {
	return func() error {
		path, err := cursorLikeConfigPath(app)
		if err != nil {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var s vscodeSettings
		if err := json.Unmarshal(data, &s); err != nil {
			return nil
		}
		s.Hooks = nil
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0600)
	}
}

func isCursorLikeInstalled(app string) func() bool {
	return func() bool {
		path, err := cursorLikeConfigPath(app)
		if err != nil {
			return false
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		return strings.Contains(string(data), "slash")
	}
}

// --- Antigravity (agy) ---

func antigravity() *HostPlugin {
	return &HostPlugin{
		Name:        "antigravity",
		DisplayName: "Antigravity (agy)",
		ConfigPaths: configPaths("antigravity", "hooks.json"),
		Install:     installJSONHook("antigravity"),
		Uninstall:   uninstallJSONHook("antigravity"),
		IsInstalled: isJSONHookInstalled("antigravity"),
	}
}

func antigravityConfigPath() string {
	home, _ := os.UserHomeDir() // safe: empty string causes stat/read to fail gracefully
	return filepath.Join(home, ".config", "agy", "hooks.json")
}

// --- Copilot CLI ---

func copilot() *HostPlugin {
	return &HostPlugin{
		Name:        "copilot",
		DisplayName: "Copilot CLI",
		ConfigPaths: configPaths("github-copilot", "hooks.json"),
		Install:     installJSONHook("github-copilot"),
		Uninstall:   uninstallJSONHook("github-copilot"),
		IsInstalled: isJSONHookInstalled("github-copilot"),
	}
}

func copilotConfigPath() string {
	home, _ := os.UserHomeDir() // safe: empty string causes stat/read to fail gracefully
	return filepath.Join(home, ".config", "github-copilot", "hooks.json")
}

type genericHooks struct {
	PreToolUse  string `json:"pre_tool_use,omitempty"`
	PostToolUse string `json:"post_tool_use,omitempty"`
}

func installJSONHook(app string) func(string) error {
	configPath := jsonHookConfigPath(app)
	return func(socketPath string) error {
		path := configPath()
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		h := genericHooks{PreToolUse: "slash hook", PostToolUse: "slash hook"}
		data, err := json.MarshalIndent(h, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0600)
	}
}

func uninstallJSONHook(app string) func() error {
	configPath := jsonHookConfigPath(app)
	return func() error {
		return os.Remove(configPath())
	}
}

func isJSONHookInstalled(app string) func() bool {
	configPath := jsonHookConfigPath(app)
	return func() bool {
		_, err := os.Stat(configPath())
		return err == nil
	}
}

// --- Aider ---

func aider() *HostPlugin {
	return &HostPlugin{
		Name:        "aider",
		DisplayName: "Aider",
		ConfigPaths: []string{"~/.aider.conf.yml", "~/.aider.env"},
		Install:     installAider,
		Uninstall:   uninstallAider,
		IsInstalled: isAiderInstalled,
	}
}

func aiderEnvPath() (string, error) {
	home, err := userHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aider.env"), nil
}

func isAiderInstalled() bool {
	path, err := aiderEnvPath()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "SLASH_HOOK")
}

func installAider(socketPath string) error {
	path, err := aiderEnvPath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n# Slash hook integration\nSLASH_HOOK=1\n")
	return err
}

func uninstallAider() error {
	path, err := aiderEnvPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.Contains(line, "SLASH_HOOK") && !strings.Contains(line, "Slash hook") {
			lines = append(lines, line)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}

// --- Zed ---

func zed() *HostPlugin {
	return &HostPlugin{
		Name:        "zed",
		DisplayName: "Zed",
		ConfigPaths: []string{".zed/settings.json"},
		Install:     installZed,
		Uninstall:   uninstallZed,
		IsInstalled: isZedInstalled,
	}
}

func zedConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".zed", "settings.json"),
	}
}

type zedSettings struct {
	MCP       map[string]zedMCP `json:"mcp,omitempty"`
	Assistant *zedAssistant     `json:"assistant,omitempty"`
}

type zedMCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type zedAssistant struct {
	DefaultModel string   `json:"default_model,omitempty"`
	MCPTools     []string `json:"mcp_tools,omitempty"`
}

func isZedInstalled() bool {
	paths := zedConfigPaths()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "slash") {
			return true
		}
	}
	return false
}

func installZed(socketPath string) error {
	for _, path := range zedConfigPaths() {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		var s zedSettings
		data, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(data, &s)
		}
		if s.MCP == nil {
			s.MCP = make(map[string]zedMCP)
		}
		s.MCP["slash"] = zedMCP{
			Command: "slash",
			Args:    []string{"mcp", "--port", "8765"},
		}
		s.Assistant = &zedAssistant{
			DefaultModel: "claude-4",
			MCPTools:     []string{"slash"},
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func uninstallZed() error {
	for _, path := range zedConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s zedSettings
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		delete(s.MCP, "slash")
		if len(s.MCP) == 0 {
			s.MCP = nil
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		_ = os.WriteFile(path, data, 0644)
	}
	return nil
}

// --- Opencode ---

func opencode() *HostPlugin {
	return &HostPlugin{
		Name:        "opencode",
		DisplayName: "Opencode",
		ConfigPaths: opencodeConfigPaths(),
		Install:     installOpencode,
		Uninstall:   uninstallOpencode,
		IsInstalled: isOpencodeInstalled,
	}
}

func opencodeConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "opencode", "opencode.json"),
		".opencode/plugin/slash-plugin.js",
	}
}

func opencodePluginDir() string {
	return ".opencode/plugin"
}

func opencodePluginPath() string {
	return filepath.Join(opencodePluginDir(), "slash-plugin.js")
}

func isOpencodeInstalled() bool {
	_, err := os.Stat(opencodePluginPath())
	return err == nil
}

func installOpencode(socketPath string) error {
	pluginDir := opencodePluginDir()
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(opencodePluginPath(), opencodePluginJS, 0644)
}

func uninstallOpencode() error {
	return os.Remove(opencodePluginPath())
}

// --- Continue ---

func continu() *HostPlugin {
	return &HostPlugin{
		Name:        "continue",
		DisplayName: "Continue",
		ConfigPaths: continueConfigPaths(),
		Install:     installContinue,
		Uninstall:   uninstallContinue,
		IsInstalled: isContinueInstalled,
	}
}

func continueConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".continue", "config.json"),
	}
}

type continueConfig struct {
	Models       []interface{}          `json:"models,omitempty"`
	Experimental *continueExperimental  `json:"experimental,omitempty"`
}

type continueExperimental struct {
	MCPServers map[string]continueMCP `json:"mcpServers,omitempty"`
}

type continueMCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func isContinueInstalled() bool {
	paths := continueConfigPaths()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "slash") {
			return true
		}
	}
	return false
}

func installContinue(socketPath string) error {
	for _, path := range continueConfigPaths() {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		var s continueConfig
		data, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(data, &s)
		}
		if s.Experimental == nil {
			s.Experimental = &continueExperimental{}
		}
		if s.Experimental.MCPServers == nil {
			s.Experimental.MCPServers = make(map[string]continueMCP)
		}
		s.Experimental.MCPServers["slash"] = continueMCP{
			Command: "slash",
			Args:    []string{"mcp", "--port", "8765"},
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func uninstallContinue() error {
	for _, path := range continueConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s continueConfig
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		if s.Experimental != nil {
			delete(s.Experimental.MCPServers, "slash")
			if len(s.Experimental.MCPServers) == 0 {
				s.Experimental.MCPServers = nil
			}
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		_ = os.WriteFile(path, data, 0644)
	}
	return nil
}

// --- Cline ---

func cline() *HostPlugin {
	return &HostPlugin{
		Name:        "cline",
		DisplayName: "Cline",
		ConfigPaths: clineConfigPaths(),
		Install:     installCline,
		Uninstall:   uninstallCline,
		IsInstalled: isClineInstalled,
	}
}

func clineConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".cline", "mcp_settings.json"),
	}
}

type clineSettings struct {
	MCPServers map[string]clineMCP `json:"mcpServers,omitempty"`
}

type clineMCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func isClineInstalled() bool {
	paths := clineConfigPaths()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "slash") {
			return true
		}
	}
	return false
}

func installCline(socketPath string) error {
	for _, path := range clineConfigPaths() {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		var s clineSettings
		data, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(data, &s)
		}
		if s.MCPServers == nil {
			s.MCPServers = make(map[string]clineMCP)
		}
		s.MCPServers["slash"] = clineMCP{
			Command: "slash",
			Args:    []string{"mcp", "--port", "8765"},
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func uninstallCline() error {
	for _, path := range clineConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s clineSettings
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		delete(s.MCPServers, "slash")
		if len(s.MCPServers) == 0 {
			s.MCPServers = nil
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		_ = os.WriteFile(path, data, 0644)
	}
	return nil
}

// --- Goose ---

func goose() *HostPlugin {
	return &HostPlugin{
		Name:        "goose",
		DisplayName: "Goose",
		ConfigPaths: gooseConfigPaths(),
		Install:     installGoose,
		Uninstall:   uninstallGoose,
		IsInstalled: isGooseInstalled,
	}
}

func gooseConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "goose", "config.json"),
	}
}

type gooseConfig struct {
	MCP *gooseMCPConfig `json:"mcp,omitempty"`
}

type gooseMCPConfig struct {
	Servers map[string]gooseMCP `json:"servers,omitempty"`
}

type gooseMCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func isGooseInstalled() bool {
	paths := gooseConfigPaths()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "slash") {
			return true
		}
	}
	return false
}

func installGoose(socketPath string) error {
	for _, path := range gooseConfigPaths() {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		var s gooseConfig
		data, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(data, &s)
		}
		if s.MCP == nil {
			s.MCP = &gooseMCPConfig{}
		}
		if s.MCP.Servers == nil {
			s.MCP.Servers = make(map[string]gooseMCP)
		}
		s.MCP.Servers["slash"] = gooseMCP{
			Command: "slash",
			Args:    []string{"mcp", "--port", "8765"},
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func uninstallGoose() error {
	for _, path := range gooseConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s gooseConfig
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		if s.MCP != nil {
			delete(s.MCP.Servers, "slash")
			if len(s.MCP.Servers) == 0 {
				s.MCP.Servers = nil
			}
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		_ = os.WriteFile(path, data, 0644)
	}
	return nil
}

// --- PearAI ---

func pearai() *HostPlugin {
	return &HostPlugin{
		Name:        "pearai",
		DisplayName: "PearAI",
		ConfigPaths: pearaiConfigPaths(),
		Install:     installPearAI,
		Uninstall:   uninstallPearAI,
		IsInstalled: isPearAIInstalled,
	}
}

func pearaiConfigPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".pearai", "config.json"),
	}
}

type pearaiConfig struct {
	Models       []interface{}          `json:"models,omitempty"`
	Experimental *pearaiExperimental    `json:"experimental,omitempty"`
}

type pearaiExperimental struct {
	MCPServers map[string]pearaiMCP `json:"mcpServers,omitempty"`
}

type pearaiMCP struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func isPearAIInstalled() bool {
	paths := pearaiConfigPaths()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "slash") {
			return true
		}
	}
	return false
}

func installPearAI(socketPath string) error {
	for _, path := range pearaiConfigPaths() {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		var s pearaiConfig
		data, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(data, &s)
		}
		if s.Experimental == nil {
			s.Experimental = &pearaiExperimental{}
		}
		if s.Experimental.MCPServers == nil {
			s.Experimental.MCPServers = make(map[string]pearaiMCP)
		}
		s.Experimental.MCPServers["slash"] = pearaiMCP{
			Command: "slash",
			Args:    []string{"mcp", "--port", "8765"},
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func uninstallPearAI() error {
	for _, path := range pearaiConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s pearaiConfig
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		if s.Experimental != nil {
			delete(s.Experimental.MCPServers, "slash")
			if len(s.Experimental.MCPServers) == 0 {
				s.Experimental.MCPServers = nil
			}
		}
		data, err = json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		_ = os.WriteFile(path, data, 0644)
	}
	return nil
}

// --- helpers ---

func userHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return home, nil
}

func configPaths(parts ...string) []string {
	home, _ := os.UserHomeDir()
	return []string{filepath.Join(append([]string{home, ".config"}, parts...)...)}
}

func jsonHookConfigPath(app string) func() string {
	switch app {
	case "github-copilot":
		return copilotConfigPath
	default:
		return antigravityConfigPath
	}
}
