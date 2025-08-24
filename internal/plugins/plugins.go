package plugins

import "goname/internal/models"

type Plugin interface {
	Name() string
	Version() string
	PreProcess(file *models.VideoFile) error
	PostProcess(result *models.RenameResult) error
}

type PluginManager struct {
	plugins []Plugin
}
