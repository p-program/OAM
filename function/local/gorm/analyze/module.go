package analyze

import "go.uber.org/fx"

var Modules = fx.Options(
	fx.Provide(NewMeta))
