package uniform

type Provider interface {
	NewResource(string) (Resource, error)
}

type TranslatedProvider struct {
	BaseProvider  Provider
	BaseLocation  string
	JoinLocations func(base, rel string) string
}

func (me *TranslatedProvider) NewResource(rel string) (Resource, error) {
	return me.BaseProvider.NewResource(me.JoinLocations(me.BaseLocation, rel))
}
