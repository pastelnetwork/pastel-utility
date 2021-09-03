package configurer

import (
	"net/url"

	"github.com/pastelnetwork/pastel-utility/constants"
)

// IConfigurer returns a interface of Configurer
type IConfigurer interface {
	GetHomeDir() string
	DefaultWorkingDir() string
	DefaultZksnarkDir() string
	DefaultPastelExecutableDir() string
	GetSuperNodeLogFile(workingDir string) string
	GetWalletNodeLogFile(workingDir string) string
	GetSuperNodeConfFile(workingDir string) string
	GetWalletNodeConfFile(workingDir string) string
	GetRQServiceConfFile(workingDir string) string
	GetDownloadURL(version string, tool constants.ToolType) (*url.URL, string, error)
	GetDownloadGitURL(version string, tool constants.ToolType) (*url.URL, string, error)
	GetDownloadGitCheckSumURL(version string, tool constants.ToolType) (*url.URL, string, error)
}
