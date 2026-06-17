package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"math/big"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	wwwDirName            = "www"
	generatedDirName       = "generated"
	apacheGeneratedName    = "apache"
	apacheSitesDirName     = "sites-enabled"
	apacheAliasDirName     = "alias"
	mysqlGeneratedName     = "mysql"
	mysqlDataDirName       = "data"
	mysqlGeneratedConfig   = "my.ini"
	hostsFileRelativePath = `C:\Windows\System32\drivers\etc\hosts`
)

type ServiceBinaryConfig struct {
	Path    string   `yaml:"path" json:"path"`
	Args    []string `yaml:"args" json:"args"`
	Ports   []int    `yaml:"ports" json:"ports"`
	WorkDir string   `yaml:"workDir" json:"workDir"`
	Enabled bool     `yaml:"enabled" json:"enabled"`
}

type AppConfig struct {
	DefaultWebServer string             `yaml:"defaultWebServer" json:"defaultWebServer"`
	Apache           ServiceBinaryConfig `yaml:"apache" json:"apache"`
	Nginx            ServiceBinaryConfig `yaml:"nginx" json:"nginx"`
	MySQL            ServiceBinaryConfig `yaml:"mysql" json:"mysql"`
	PHP              ServiceBinaryConfig `yaml:"php" json:"php"`
	ConfigPath       string              `yaml:"-" json:"configPath"`
	BaseDir          string              `yaml:"-" json:"baseDir"`
}

type ServiceStatus struct {
	Running         bool   `json:"running"`
	Message         string `json:"message,omitempty"`
	ActiveWebServer string `json:"activeWebServer,omitempty"`
	ApachePID       int    `json:"apachePid,omitempty"`
	NginxPID        int    `json:"nginxPid,omitempty"`
	MySQLPID        int    `json:"mysqlPid,omitempty"`
	PHPPID          int    `json:"phpPid,omitempty"`
	ConfigPath      string `json:"configPath,omitempty"`
	WebServerConfig string `json:"webServerConfig,omitempty"`
}

type serviceHandle struct {
	name string
	cmd  *exec.Cmd
	pid  int
}

type ServiceManager struct {
	mu      sync.Mutex
	config  AppConfig
	handles map[string]*serviceHandle
	loadErr error
}

func NewServiceManager() (*ServiceManager, error) {
	config, err := loadAppConfig()
	manager := &ServiceManager{
		config:  config,
		handles: make(map[string]*serviceHandle),
		loadErr: err,
	}
	if err != nil {
		return manager, err
	}
	return manager, nil
}

func loadAppConfig() (AppConfig, error) {
	configPath, baseDir, err := resolveConfigPath()
	if err != nil {
		return AppConfig{}, err
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config.yaml: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return AppConfig{}, fmt.Errorf("parse config.yaml: %w", err)
	}

	config.ConfigPath = configPath
	config.BaseDir = baseDir
	config.normalize()
	return config, nil
}

func resolveConfigPath() (string, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	wdConfig := filepath.Join(wd, "config.yaml")
	if fileExists(wdConfig) {
		return wdConfig, wd, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", "", err
	}
	exeDir := filepath.Dir(exePath)
	exeConfig := filepath.Join(exeDir, "config.yaml")
	if fileExists(exeConfig) {
		return exeConfig, exeDir, nil
	}

	return "", "", fmt.Errorf("config.yaml not found in %s or %s", wd, exeDir)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (c *AppConfig) normalize() {
	if c.DefaultWebServer == "" {
		c.DefaultWebServer = "apache"
	}
	if c.BaseDir == "" {
		c.BaseDir = "."
	}
}

func (m *ServiceManager) StartAllServices() (ServiceStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loadErr != nil {
		return m.currentStatus("", false, ""), m.loadErr
	}
	if len(m.handles) > 0 {
		return m.currentStatus("Services already running", true, m.activeWebServer()), nil
	}

	webServerName := strings.ToLower(strings.TrimSpace(m.config.DefaultWebServer))
	var started []string

	if webServerName == "apache" {
		_ = exec.Command("taskkill", "/IM", "nginx.exe", "/F", "/T").Run()
	}

	if err := m.prepareWebEnvironment(); err != nil {
		return m.currentStatus("", false, webServerName), err
	}

	if err := m.checkConfig(); err != nil {
		return m.currentStatus("", false, ""), err
	}

	if err := m.startConfiguredService("php", m.config.PHP, &started); err != nil {
		m.stopStartedLocked(started)
		return m.currentStatus("", false, webServerName), err
	}

	if err := m.startConfiguredService("mysql", m.config.MySQL, &started); err != nil {
		m.stopStartedLocked(started)
		return m.currentStatus("", false, webServerName), err
	}

	webConfig := m.webServerConfig(webServerName)
	if err := m.startConfiguredService(webServerName, webConfig, &started); err != nil {
		m.stopStartedLocked(started)
		return m.currentStatus("", false, webServerName), err
	}

	status := m.currentStatus("Services started successfully", true, webServerName)
	status.WebServerConfig = webConfig.Path
	return status, nil
}

func (m *ServiceManager) StopAllServices() (ServiceStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.handles) == 0 {
		return m.currentStatus("No running services found", false, ""), nil
	}

	var stopErr error
	for name, handle := range m.handles {
		if err := stopProcess(handle); err != nil && stopErr == nil {
			stopErr = fmt.Errorf("stop %s: %w", name, err)
		}
		delete(m.handles, name)
	}

	status := m.currentStatus("Services stopped", false, "")
	return status, stopErr
}

func (m *ServiceManager) checkConfig() error {
	for name, cfg := range map[string]ServiceBinaryConfig{
		"php":   m.config.PHP,
		"mysql": m.config.MySQL,
	} {
		if err := validateServiceConfig(name, cfg, m.config.BaseDir); err != nil {
			return err
		}
	}

	serverName := strings.ToLower(strings.TrimSpace(m.config.DefaultWebServer))
	return validateServiceConfig(serverName, m.webServerConfig(serverName), m.config.BaseDir)
}

func (m *ServiceManager) prepareWebEnvironment() error {
	wwwRoot := m.wwwRoot()
	if err := os.MkdirAll(wwwRoot, 0o755); err != nil {
		return fmt.Errorf("create www directory: %w", err)
	}
	if err := os.MkdirAll(m.generatedApacheAliasDir(), 0o755); err != nil {
		return fmt.Errorf("create apache alias directory: %w", err)
	}

	phpPort, err := m.selectPhpPort()
	if err != nil {
		return err
	}
	m.config.PHP.Ports = []int{phpPort}
	m.config.PHP.Args = []string{"-b", net.JoinHostPort("127.0.0.1", strconv.Itoa(phpPort))}

	if err := m.ensureMySQLConfig(); err != nil {
		return err
	}

	if err := m.ensureWelcomePage(wwwRoot); err != nil {
		return err
	}

	projectDirs, err := m.discoverProjectFolders(wwwRoot)
	if err != nil {
		return err
	}

	if err := m.generateApacheVHosts(wwwRoot, projectDirs); err != nil {
		return err
	}

	if err := m.generateApachePhpProxyConfig(phpPort); err != nil {
		return err
	}

	if err := m.syncHostsEntries(projectDirs); err != nil {
		return err
	}

	if err := m.ensureApacheSslMaterial(wwwRoot); err != nil {
		return err
	}

	return nil
}

func (m *ServiceManager) wwwRoot() string {
	return filepath.Join(m.config.BaseDir, wwwDirName)
}

func (m *ServiceManager) generatedApacheSitesDir() string {
	return filepath.Join(m.config.BaseDir, generatedDirName, apacheGeneratedName, apacheSitesDirName)
}

func (m *ServiceManager) generatedApacheAliasDir() string {
	return filepath.Join(m.config.BaseDir, generatedDirName, apacheGeneratedName, apacheAliasDirName)
}

func (m *ServiceManager) generatedApacheSslDir() string {
	return filepath.Join(m.config.BaseDir, generatedDirName, apacheGeneratedName, "ssl")
}

func (m *ServiceManager) generatedApachePhpProxyPath() string {
	return filepath.Join(m.config.BaseDir, generatedDirName, apacheGeneratedName, "php-fcgi.conf")
}

func (m *ServiceManager) generatedMysqlConfigPath() string {
	return filepath.Join(m.config.BaseDir, generatedDirName, mysqlGeneratedName, mysqlGeneratedConfig)
}

func (m *ServiceManager) generatedMysqlDataDir() string {
	return filepath.Join(m.config.BaseDir, generatedDirName, mysqlGeneratedName, mysqlDataDirName)
}

func (m *ServiceManager) ensureMySQLConfig() error {
	configPath := m.generatedMysqlConfigPath()
	dataDir := m.generatedMysqlDataDir()
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create mysql data directory: %w", err)
	}

	resolvedMysqlExe := resolveRelativePath(m.config.BaseDir, m.config.MySQL.Path)
	if !fileExists(resolvedMysqlExe) {
		return fmt.Errorf("mysql binary not found at %s", resolvedMysqlExe)
	}
	mysqlBaseDir := filepath.Dir(filepath.Dir(resolvedMysqlExe))

	content := fmt.Sprintf(`[client]
port=3306

[mysqld]
basedir="%s"
datadir="%s"
port=3306
explicit_defaults_for_timestamp=1
default_authentication_plugin=mysql_native_password
secure-file-priv=""

[mysqldump]
quick
max_allowed_packet=512M
`, mysqlBaseDir, dataDir)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create mysql config directory: %w", err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write mysql config: %w", err)
	}

	if !fileExists(filepath.Join(dataDir, "mysql")) {
		initCmd := exec.Command(resolvedMysqlExe,
			"--initialize-insecure",
			"--basedir="+mysqlBaseDir,
			"--datadir="+dataDir,
		)
		initCmd.Dir = mysqlBaseDir
		initCmd.Stdout = os.Stdout
		initCmd.Stderr = os.Stderr
		initCmd.Env = os.Environ()
		if err := initCmd.Run(); err != nil {
			return fmt.Errorf("initialize mysql data directory: %w", err)
		}
	}

	m.config.MySQL.Args = []string{"--defaults-file=" + configPath}
	return nil
}

func (m *ServiceManager) ensureWelcomePage(wwwRoot string) error {
	welcomePath := filepath.Join(wwwRoot, "index.html")
	if fileExists(welcomePath) {
		return nil
	}

	content := `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>goDev</title>
  <style>
    :root { color-scheme: dark; }
    body { margin: 0; font-family: Segoe UI, Arial, sans-serif; background: linear-gradient(135deg, #0f172a, #111827 50%, #0b1120); color: #e5eefc; min-height: 100vh; display: grid; place-items: center; }
    .card { width: min(860px, calc(100vw - 32px)); background: rgba(15, 23, 42, 0.86); border: 1px solid rgba(148, 163, 184, 0.18); border-radius: 24px; padding: 40px; box-shadow: 0 24px 80px rgba(0, 0, 0, 0.35); }
    h1 { margin: 0 0 12px; font-size: 44px; }
    p { line-height: 1.6; color: #bfdbfe; }
    .links { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 24px; }
    a { color: #0f172a; text-decoration: none; background: #7dd3fc; padding: 10px 14px; border-radius: 999px; font-weight: 600; }
    code { background: rgba(148, 163, 184, 0.16); padding: 2px 8px; border-radius: 999px; }
    ul { margin: 18px 0 0; padding-left: 18px; }
  </style>
</head>
<body>
  <main class="card">
    <h1>goDev</h1>
    <p>Your local PHP workspace is running. Drop a folder into <code>www/</code> and the app will generate a matching <code>.test</code> host for it.</p>
    <div class="links">
      <a href="http://127.0.0.1:8080">Open localhost</a>
      <a href="http://127.0.0.1:8080">Help</a>
    </div>
    <ul>
      <li>Create a folder in <code>www/</code> to get a <code>folder.test</code> site.</li>
      <li>Put an <code>index.html</code> or <code>index.php</code> inside that folder.</li>
      <li>Restart services after adding or removing project folders.</li>
    </ul>
  </main>
</body>
</html>`

	if err := os.WriteFile(welcomePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write welcome page: %w", err)
	}

	return nil
}

func (m *ServiceManager) discoverProjectFolders(wwwRoot string) ([]string, error) {
	entries, err := os.ReadDir(wwwRoot)
	if err != nil {
		return nil, fmt.Errorf("scan www directory: %w", err)
	}

	projects := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == generatedDirName {
			continue
		}
		projects = append(projects, name)
	}

	return projects, nil
}

func (m *ServiceManager) generateApacheVHosts(wwwRoot string, projectDirs []string) error {
	sitesDir := m.generatedApacheSitesDir()
	if err := os.MkdirAll(sitesDir, 0o755); err != nil {
		return fmt.Errorf("create apache sites directory: %w", err)
	}

	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		return fmt.Errorf("read apache sites directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".conf") {
			_ = os.Remove(filepath.Join(sitesDir, entry.Name()))
		}
	}

	port := m.webServerPort()
	for _, project := range projectDirs {
		confPath := filepath.Join(sitesDir, project+".conf")
		projectRoot := filepath.Join(wwwRoot, project)
		serverName := project + ".test"
		content := fmt.Sprintf(`<VirtualHost *:%d>
    ServerName %s
    ServerAlias www.%s
    DocumentRoot "%s"
    <Directory "%s">
        AllowOverride All
        Require all granted
    </Directory>
    ErrorLog "logs/%s-error.log"
    CustomLog "logs/%s-access.log" common
</VirtualHost>
`, port, serverName, serverName, projectRoot, projectRoot, project, project)
		if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write apache vhost for %s: %w", project, err)
		}
	}

	return nil
}

func (m *ServiceManager) ensureApacheSslMaterial(wwwRoot string) error {
	sslDir := m.generatedApacheSslDir()
	if err := os.MkdirAll(sslDir, 0o755); err != nil {
		return fmt.Errorf("create apache ssl directory: %w", err)
	}

	certPath := filepath.Join(sslDir, "server.crt")
	keyPath := filepath.Join(sslDir, "server.key")
	if !fileExists(certPath) || !fileExists(keyPath) {
		if err := generateSelfSignedCertificate(certPath, keyPath); err != nil {
			return err
		}
	}

	sslConfPath := filepath.Join(m.config.BaseDir, generatedDirName, apacheGeneratedName, "httpd-ssl.conf")
	content := fmt.Sprintf(`# goDev generated SSL config
Listen 443

<IfModule ssl_module>
<VirtualHost _default_:443>
    DocumentRoot "%s"
    ServerName localhost:443
    ErrorLog "logs/ssl-error.log"
    CustomLog "logs/ssl-access.log" common
    SSLEngine on
    SSLCertificateFile "%s"
    SSLCertificateKeyFile "%s"
    <Directory "%s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>
</IfModule>
`, wwwRoot, certPath, keyPath, wwwRoot)

	if err := os.WriteFile(sslConfPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write apache ssl config: %w", err)
	}

	return nil
}

func (m *ServiceManager) generateApachePhpProxyConfig(phpPort int) error {
	proxyPath := m.generatedApachePhpProxyPath()
	content := fmt.Sprintf(`<FilesMatch "\\.php$">
    SetHandler "proxy:fcgi://127.0.0.1:%d"
</FilesMatch>
`, phpPort)
	if err := os.MkdirAll(filepath.Dir(proxyPath), 0o755); err != nil {
		return fmt.Errorf("create apache php proxy directory: %w", err)
	}
	if err := os.WriteFile(proxyPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write apache php proxy config: %w", err)
	}
	return nil
}

func (m *ServiceManager) selectPhpPort() (int, error) {
	preferred := 0
	if len(m.config.PHP.Ports) > 0 {
		preferred = m.config.PHP.Ports[0]
	}
	if preferred <= 0 {
		preferred = 9001
	}
	for port := preferred; port < preferred+100; port++ {
		if err := checkPortAvailable(port); err == nil {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free php port found starting at %d", preferred)
}

func generateSelfSignedCertificate(certPath, keyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate ssl key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("generate cert serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"goDev"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid:  true,
		DNSNames:              []string{"localhost", "*.test", "godev.test"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("create self-signed certificate: %w", err)
	}

	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("create cert file: %w", err)
	}
	defer certFile.Close()
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("write cert file: %w", err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("create key file: %w", err)
	}
	defer keyFile.Close()
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}); err != nil {
		return fmt.Errorf("write key file: %w", err)
	}

	return nil
}

func (m *ServiceManager) webServerPort() int {
	if len(m.config.Apache.Ports) > 0 {
		return m.config.Apache.Ports[0]
	}
	if len(m.config.Nginx.Ports) > 0 {
		return m.config.Nginx.Ports[0]
	}
	return 8080
}

func (m *ServiceManager) syncHostsEntries(projectDirs []string) error {
	if len(projectDirs) == 0 {
		return nil
	}

	hostsPath := hostsFileRelativePath
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil
		}
		return fmt.Errorf("read hosts file: %w", err)
	}

	existing := string(content)
	var additions []string
	for _, project := range projectDirs {
		hostLine := fmt.Sprintf("127.0.0.1 %s.test", project)
		if strings.Contains(existing, hostLine) {
			continue
		}
		additions = append(additions, hostLine)
	}

	if len(additions) == 0 {
		return nil
	}

	f, err := os.OpenFile(hostsPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return nil
		}
		return fmt.Errorf("update hosts file: %w", err)
	}
	defer f.Close()

	for _, line := range additions {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return fmt.Errorf("append hosts entry %q: %w", line, err)
		}
	}

	return nil
}

func validateServiceConfig(name string, cfg ServiceBinaryConfig, baseDir string) error {
	if !cfg.Enabled {
		return nil
	}
	if strings.TrimSpace(cfg.Path) == "" {
		return fmt.Errorf("%s path is not configured", name)
	}
	resolved := resolveRelativePath(baseDir, cfg.Path)
	if !fileExists(resolved) {
		return fmt.Errorf("%s binary not found at %s", name, resolved)
	}
	for _, port := range cfg.Ports {
		if port <= 0 {
			continue
		}
		if err := checkPortAvailable(port); err != nil {
			return fmt.Errorf("%s port %d: %w", name, port, err)
		}
	}
	return nil
}

func checkPortAvailable(port int) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("port is already in use or unavailable: %w", err)
	}
	return listener.Close()
}

func (m *ServiceManager) webServerConfig(name string) ServiceBinaryConfig {
	switch name {
	case "nginx":
		return m.config.Nginx
	default:
		return m.config.Apache
	}
}

func (m *ServiceManager) startConfiguredService(name string, cfg ServiceBinaryConfig, started *[]string) error {
	if !cfg.Enabled {
		return nil
	}
	resolvedPath := resolveRelativePath(m.config.BaseDir, cfg.Path)
	if !fileExists(resolvedPath) {
		return fmt.Errorf("%s binary not found at %s", name, resolvedPath)
	}

	args := append([]string(nil), cfg.Args...)
	args = resolveServiceArgs(m.config.BaseDir, args)
	if strings.EqualFold(name, "php") && len(args) == 0 && len(cfg.Ports) > 0 {
		args = []string{"-b", net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.Ports[0]))}
	}

	cmd := exec.Command(resolvedPath, args...)
	cmd.Dir = resolveWorkingDir(m.config.BaseDir, cfg.WorkDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	m.handles[name] = &serviceHandle{name: name, cmd: cmd, pid: cmd.Process.Pid}
	*started = append(*started, name)
	return nil
}

func resolveServiceArgs(baseDir string, args []string) []string {
	resolved := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-f", "--defaults-file":
			resolved = append(resolved, arg)
			if i+1 < len(args) {
				resolved = append(resolved, resolveRelativePath(baseDir, args[i+1]))
				i++
			}
		default:
			if strings.HasPrefix(arg, "--defaults-file=") {
				parts := strings.SplitN(arg, "=", 2)
				resolved = append(resolved, parts[0]+"="+resolveRelativePath(baseDir, parts[1]))
			} else {
				resolved = append(resolved, arg)
			}
		}
	}
	return resolved
}

func resolveRelativePath(baseDir, candidate string) string {
	if filepath.IsAbs(candidate) {
		return candidate
	}
	if baseDir == "" {
		return candidate
	}
	return filepath.Clean(filepath.Join(baseDir, candidate))
}

func resolveWorkingDir(baseDir, candidate string) string {
	if strings.TrimSpace(candidate) != "" {
		return resolveRelativePath(baseDir, candidate)
	}
	return baseDir
}

func (m *ServiceManager) stopStartedLocked(started []string) {
	for i := len(started) - 1; i >= 0; i-- {
		name := started[i]
		if handle, ok := m.handles[name]; ok {
			_ = stopProcess(handle)
			delete(m.handles, name)
		}
	}
}

func stopProcess(handle *serviceHandle) error {
	if handle == nil || handle.cmd == nil || handle.cmd.Process == nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(handle.pid))
		kill.Stdout = os.Stdout
		kill.Stderr = os.Stderr
		if err := kill.Run(); err != nil {
			if processErr := handle.cmd.Process.Kill(); processErr != nil {
				return errors.Join(err, processErr)
			}
		}
		return nil
	}

	return handle.cmd.Process.Kill()
}

func (m *ServiceManager) currentStatus(message string, running bool, webServer string) ServiceStatus {
	status := ServiceStatus{
		Running:         running,
		Message:         message,
		ActiveWebServer: webServer,
		ConfigPath:      m.config.ConfigPath,
	}
	if handle, ok := m.handles["apache"]; ok {
		status.ApachePID = handle.pid
	}
	if handle, ok := m.handles["nginx"]; ok {
		status.NginxPID = handle.pid
	}
	if handle, ok := m.handles["mysql"]; ok {
		status.MySQLPID = handle.pid
	}
	if handle, ok := m.handles["php"]; ok {
		status.PHPPID = handle.pid
	}
	return status
}

func (m *ServiceManager) activeWebServer() string {
	if m.handles == nil {
		return ""
	}
	if _, ok := m.handles["nginx"]; ok {
		return "nginx"
	}
	if _, ok := m.handles["apache"]; ok {
		return "apache"
	}
	return ""
}