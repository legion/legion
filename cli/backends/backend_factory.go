package backends

type BackendFactory struct {
	Backends []Backend
}

func (f *BackendFactory) init() {
	f.Backends = append(f.Backends, NewAzureContainerInstancesBackend())
}

func (f *BackendFactory) GetBackend(name string) Backend {
	for _, backend := range f.Backends {
		if backend.Name() == name {
			return backend
		}
	}

	return nil
}

func NewBackendFactory() BackendFactory {
	factory := BackendFactory{}
	factory.init()

	return factory
}
