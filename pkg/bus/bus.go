package bus

type (
	// ConfigUpdate represents Config update event for some Subsystem.
	ConfigUpdate struct {
		Subsystem string
		Config    interface{}
	}
)

var (
	// Config represents configuration change event bus.
	Config = make(chan ConfigUpdate, 1)
)
