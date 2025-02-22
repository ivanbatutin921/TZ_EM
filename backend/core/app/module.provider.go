package app

import (
	song_module "root/module/song"
)

type moduleProvider struct {
	song *song_module.SongModule

	app *App
}

func NewModuleProvider(app *App) (*moduleProvider, error) {
	provider := &moduleProvider{
		app: app,
	}

	err := provider.initDeps()
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func (p *moduleProvider) initDeps() error {
	inits := []func() error{
		p.SongModule,
	}
	for _, init := range inits {
		err := init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *moduleProvider) SongModule() error {
	p.song = song_module.NewSongModule(p.app.logger, p.app.config, p.app.db)
	return nil
}
