package notes

import (
	"github.com/olebedev/config"
	"github.com/wtfutil/wtf/cfg"
)

const (
	defaultFocusable = true
	defaultTitle     = "Notes"
)

// Settings defines the configuration options for this module
type Settings struct {
	common *cfg.Common

	folder string `help:"The folder where notes are stored.`
}

// NewSettingsFromYAML creates and returns an instance of Settings with configuration options populated
func NewSettingsFromYAML(name string, ymlConfig *config.Config, globalConfig *config.Config) *Settings {
	settings := Settings{
		common: cfg.NewCommonSettingsFromModule(name, defaultTitle, defaultFocusable, ymlConfig, globalConfig),

		folder: ymlConfig.UString("folder", ""),
	}

	return &settings
}
