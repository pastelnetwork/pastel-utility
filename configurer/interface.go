package configurer

import (
	"net/url"

	"github.com/pastelnetwork/pastelup/constants"
)

// IConfigurer returns a interface of Configurer
type IConfigurer interface {
	DefaultHomeDir() string
	WorkDir() string // get workdir without absolutePath
	DefaultWorkingDir() string
	DefaultZksnarkDir() string
	DefaultPastelExecutableDir() string
	DefaultArchiveDir() string
	GetSuperNodeLogFile(workingDir string) string
	GetHermesLogFile(workingDir string) string
	GetBridgeLogFile(workingDir string) string
	GetWalletNodeLogFile(workingDir string) string
	GetSuperNodeConfFile(workingDir string) string
	GetHermesConfFile(workingDir string) string
	GetBridgeConfFile(workingDir string) string
	GetWalletNodeConfFile(workingDir string) string
	GetRQServiceConfFile(workingDir string) string
	GetDownloadURL(network string, version string, tool constants.ToolType) (*url.URL, string, error)
}
