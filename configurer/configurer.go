package configurer

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-errors/errors"
	"github.com/pastelnetwork/pastelup/constants"
	"github.com/pastelnetwork/pastelup/utils"
)

const (
	templateDownloadURL = constants.DownloadBaseURL + "/%s/%s/%s"
)

type configurer struct {
	workingDir          string
	superNodeLogFile    string
	walletNodeLogFile   string
	superNodeConfFile   string
	hermesConfFile      string
	hermesLogFile       string
	bridgeLogFile       string
	bridgeConfFile      string
	walletNodeConfFile  string
	rqServiceConfFile   string
	zksnarkDir          string
	pastelExecutableDir string
	homeDir             string
	archiveDir          string
	architecture        constants.ArchitectureType
	osType              constants.OSType
}

// DefaultHomeDir returns the home path.
func (c *configurer) DefaultHomeDir() string {
	return c.homeDir
}

// WorkDir returns the working directory (i.e. "Pastel") without the absolute path
func (c *configurer) WorkDir() string {
	return c.workingDir
}

// DefaultWorkingDir returns the default config path.
func (c *configurer) DefaultWorkingDir() string {
	return filepath.Join(c.DefaultHomeDir(), filepath.FromSlash(getAppDataDir()), c.workingDir)
}

// GetSuperNodeLogFile returns the default supernode log file
func (c *configurer) GetSuperNodeLogFile(workingDir string) string {
	return filepath.Join(workingDir, c.superNodeLogFile)
}

// GetWalletNodeLogFile returns the default supernode log file
func (c *configurer) GetWalletNodeLogFile(workingDir string) string {
	return filepath.Join(workingDir, c.walletNodeLogFile)
}

// GetHermesLogFile returns the default hermes log file
func (c *configurer) GetHermesLogFile(workingDir string) string {
	return filepath.Join(workingDir, c.hermesLogFile)
}

// GetBridgeLogFile returns the default bridge log file
func (c *configurer) GetBridgeLogFile(workingDir string) string {
	return filepath.Join(workingDir, c.bridgeLogFile)
}

// GetSuperNodeConfFile returns the default supernode log file
func (c *configurer) GetSuperNodeConfFile(workingDir string) string {
	return filepath.Join(workingDir, c.superNodeConfFile)
}

// GetHermesConfFile returns the default Hermes log file
func (c *configurer) GetHermesConfFile(workingDir string) string {
	return filepath.Join(workingDir, c.hermesConfFile)
}

// GetBridgeConfFile returns the default Bridge log file
func (c *configurer) GetBridgeConfFile(workingDir string) string {
	return filepath.Join(workingDir, c.bridgeConfFile)
}

// GetWalletNodeConfFile returns the default supernode log file
func (c *configurer) GetWalletNodeConfFile(workingDir string) string {
	return filepath.Join(workingDir, c.walletNodeConfFile)
}

// GetRQServiceConfFile returns the default supernode log file
func (c *configurer) GetRQServiceConfFile(workingDir string) string {
	return filepath.Join(workingDir, c.rqServiceConfFile)
}

// DefaultZksnarkDir returns the default config path.
func (c *configurer) DefaultZksnarkDir() string {
	return filepath.Join(c.DefaultHomeDir(), filepath.FromSlash(getAppDataDir()), c.zksnarkDir)
}

// DefaultPastelExecutableDir returns the default pastel executable path.
func (c *configurer) DefaultPastelExecutableDir() string {
	return filepath.Join(c.DefaultHomeDir(), filepath.FromSlash(getAppDir()), c.pastelExecutableDir)
}

// DefaultArchiveDir returns the default pastel arhive path.
func (c *configurer) DefaultArchiveDir() string {
	return filepath.Join(c.DefaultHomeDir(), filepath.FromSlash(getAppDataDir()), c.archiveDir)
}

// GetDownloadURL returns download url of the pastel executables.
func (c *configurer) GetDownloadURL(version string, tool constants.ToolType) (*url.URL, string, error) {
	var name string
	switch tool {
	case constants.WalletNode:
		name = constants.WalletNodeExecName[c.osType]
		tool = constants.GoNode
	case constants.RQService:
		name = constants.PastelRQServiceExecName[c.osType]
	case constants.PastelD:
		name = constants.PastelExecArchiveName[c.osType]
	case constants.SuperNode:
		name = constants.SuperNodeExecName[c.osType]
		tool = constants.GoNode
	case constants.DDService:
		name = constants.DupeDetectionArchiveName
		tool = constants.DDService
	case constants.Pastelup:
		name = constants.PastelUpExecName[c.osType]
		tool = constants.Pastelup
	case constants.Hermes:
		name = constants.HermesExecName[c.osType]
		tool = constants.GoNode
	case constants.Bridge:
		name = constants.BridgeExecName[c.osType]
		tool = constants.GoNode
	default:
		return nil, "", errors.Errorf("unknown tool: %s", tool)
	}

	urlString := fmt.Sprintf(
		templateDownloadURL,
		constants.GetVersionSubURL(version),
		tool,
		name)
	// need special handling b/c not embedded within subdir
	if tool == constants.Pastelup {
		urlString = fmt.Sprintf("%v/%v/%v",
			constants.DownloadBaseURL,
			constants.GetVersionSubURL(version),
			name)
	}

	url, err := url.Parse(urlString)
	if err != nil {
		return nil, "", errors.Errorf("failed to parse url: %v", err)
	}
	tokens := strings.Split(urlString, "/")
	return url, tokens[len(tokens)-1], nil
}

func newLinuxConfigurer(homeDir string) IConfigurer {
	return &configurer{
		workingDir:          ".pastel",
		superNodeLogFile:    "supernode.log",
		hermesLogFile:       "hermes.log",
		bridgeLogFile:       "bridge.log",
		walletNodeLogFile:   "walletnode.log",
		superNodeConfFile:   "supernode.yml",
		hermesConfFile:      "hermes.yml",
		bridgeConfFile:      "bridge.yml",
		walletNodeConfFile:  "walletnode.yml",
		rqServiceConfFile:   "rqservice.toml",
		zksnarkDir:          ".pastel-params",
		pastelExecutableDir: "pastel",
		homeDir:             homeDir,
		archiveDir:          ".pastel_archives",
		architecture:        constants.AMD64,
		osType:              constants.Linux,
	}
}

func newDarwinConfigurer(homeDir string) IConfigurer {
	return &configurer{
		workingDir:          "Pastel",
		superNodeLogFile:    "supernode.log",
		walletNodeLogFile:   "walletnode.log",
		superNodeConfFile:   "supernode.yml",
		hermesConfFile:      "hermes.yml",
		bridgeConfFile:      "bridge.yml",
		walletNodeConfFile:  "walletnode.yml",
		rqServiceConfFile:   "rqservice.toml",
		zksnarkDir:          "PastelParams",
		pastelExecutableDir: "PastelWallet",
		homeDir:             homeDir,
		archiveDir:          "PastelArchives",
		architecture:        constants.AMD64,
		osType:              constants.Mac,
	}
}

func newWindowsConfigurer(homeDir string) IConfigurer {
	return &configurer{
		workingDir:          "Pastel",
		superNodeLogFile:    "supernode.log",
		walletNodeLogFile:   "walletnode.log",
		superNodeConfFile:   "supernode.yml",
		hermesConfFile:      "hermes.yml",
		bridgeConfFile:      "bridge.yml",
		walletNodeConfFile:  "walletnode.yml",
		rqServiceConfFile:   "rqservice.toml",
		zksnarkDir:          "PastelParams",
		pastelExecutableDir: "PastelWallet",
		homeDir:             homeDir,
		archiveDir:          "PastelArchives",
		architecture:        constants.AMD64,
		osType:              constants.Windows,
	}
}

// NewConfigurer return a new configurer instance
// Returns:
//
//	$HOME for MacOS and Linux
//	%userprofile% for Windows
func NewConfigurer() (IConfigurer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Errorf("failed to get user home dir: %v", err)
	}
	osType := utils.GetOS()
	switch osType {
	case constants.Linux:
		return newLinuxConfigurer(homeDir), nil
	case constants.Mac:
		return newDarwinConfigurer(homeDir), nil
	case constants.Windows:
		return newWindowsConfigurer(homeDir), nil
	default:
		return nil, errors.New("unknown os")
	}
}
