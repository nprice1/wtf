package gmail

import (
	"github.com/olebedev/config"
	"github.com/wtfutil/wtf/cfg"
)

const (
	defaultFocusable = true
	defaultTitle     = "Gmail"
)

// Settings defines the configuration options for this module
type Settings struct {
	common *cfg.Common

	mailCount     int    `help:"The number of mail messages to fetch, e.g. 5" optional:"true"`
	secretFile    string `help:"Your Google client secret JSON file." values:"A string representing a file path to the JSON secret file."`
	searchQuery   string `help:"The search query for displaying messages, e.g. is:unread label:INBOX." optional:"true"`
	messageFolder string `help:"The folder where messages are stored temporarily.`
}

// NewSettingsFromYAML creates and returns an instance of Settings with configuration options populated
func NewSettingsFromYAML(name string, ymlConfig *config.Config, globalConfig *config.Config) *Settings {
	settings := Settings{
		common: cfg.NewCommonSettingsFromModule(name, defaultTitle, defaultFocusable, ymlConfig, globalConfig),

		mailCount:     ymlConfig.UInt("mailCount", 5),
		secretFile:    ymlConfig.UString("secretFile", ""),
		searchQuery:   ymlConfig.UString("searchQuery", "is:unread label:INBOX"),
		messageFolder: ymlConfig.UString("messageFolder", ""),
	}

	return &settings
}
