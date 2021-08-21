package configs

// Init contains config of the Init command
type Init struct {
	WorkingDir             string `json:"workdir,omitempty"`
	Network                string `json:"network,omitempty"`
	RPCPort                int    `json:"rpc-port,omitempty"`
	RPCUser                string `json:"rpc-user,omitempty"`
	RPCPwd                 string `json:"rpc-pwd,omitempty"`
	Force                  bool   `json:"force,omitempty"`
	Peers                  string `json:"peers"`
	PastelExecDir          string `json:"pastelexecdir,omitempty"`
	Version                string `json:"nodeversion,omitempty"`
	RemoteWorkingDir       string `json:"remoteworkingdir,omitempty"`
	RemotePastelExecDir    string `json:"remotepastelexecdir,omitempty"`
	RemotePastelUtilityDir string `json:"remotepastelutilitydir,omitempty"`
	TransferLocal          bool   `json:"transfer-local,omitempty"`
	DisablePortConfig      bool   `json:"disable-port-config,omitempty"`
}

/*
// NewInit returns a new Init instance.
func NewInit() *Init {
	return &Init{}
}*/
