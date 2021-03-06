package messapi

import "github.com/maruel/mess/internal/model"

// ServerDetailsResponse is /server/details (GET).
type ServerDetailsResponse struct {
	ServerVersion string `json:"server_version"`
	BotVersion    string `json:"bot_version"`
	// DEPRECATED: MachineProviderTemplate  string `json:"machine_provider_template"`
	// DEPRECATED: DisplayServerURLTemplate string `json:"display_server_url_template"`
	LUCIConfig      string `json:"luci_config"`
	CASViewerServer string `json:"cas_viewer_server"`
}

// ServerPermissionsRequest is /server/permissions (GET).
type ServerPermissionsRequest struct {
	BotID  string
	TaskID model.TaskID
	Tags   []string
}

// ServerPermissionsResponse is /server/permissions (GET).
type ServerPermissionsResponse struct {
	DeleteBot    bool `json:"delete_bot"`
	DeleteBots   bool `json:"delete_bots"`
	TerminateBot bool `json:"terminate_bot"`
	// DEPRECATED: GetConfigs   bool `json:"get_configs"`
	// DEPRECATED: PutConfigs   bool `json:"put_configs"`
	// Cancel one single task
	CancelTask        bool `json:"cancel_task"`
	GetBootstrapToken bool `json:"get_bootstrap_token"`
	// Cancel multiple tasks at once, usually in emergencies.
	CancelTasks bool     `json:"cancel_tasks"`
	ListBots    []string `json:"list_bots"`
	ListTasks   []string `json:"list_tasks"`
}
