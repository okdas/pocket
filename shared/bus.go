package shared

import (
	"log"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
)

type bus struct {
	modules.Bus

	channel modules.EventsChannel

	persistence modules.PersistenceModule
	p2p         modules.P2PModule
	utility     modules.UtilityModule
	consensus   modules.ConsensusModule
	telemetry   modules.TelemetryModule
}

const (
	DefaultPocketBusBufferSize = 100
)

func CreateBus(
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
	telemetry modules.TelemetryModule,
) (modules.Bus, error) {
	bus := &bus{
		channel:     make(modules.EventsChannel, DefaultPocketBusBufferSize),
		persistence: persistence,
		p2p:         p2p,
		utility:     utility,
		consensus:   consensus,
		telemetry:   telemetry,
	}

	panicOnMissingMod := func(name string, module modules.Module) {
		if module == nil {
			log.Fatalf("Bus Error: the provided %s module is nil, Please use CreateBusWithOptionalModules if you intended it to be nil.", name)
		}
	}

	modules := map[string]modules.Module{
		"persistence": persistence,
		"consensus":   consensus,
		"p2p":         p2p,
		"utility":     utility,
		"telemetry":   telemetry,
	}

	// checks if modules are not nil and sets their bus to this bus instance.
	// will not carry forward if one of the modules is nil
	for modName, mod := range modules {
		panicOnMissingMod(modName, mod)
		mod.SetBus(bus)
	}

	return bus, nil
}

// This is a version of CreateBus that accepts nil modules.
// This function allows you to use a specific module in isolation of other modules by providing a bus with nil modules.
//
// Example of usage: `app/client/main.go`
//
//    We want to use the pre2p module in isolation to communicate with nodes in the network.
//    The pre2p module expects to retrieve a telemetry module through the bus to perform instrumentation, thus we need to inject a bus that can retrieve a telemetry module.
//    However, we don't need telemetry for the dev client.
//    Using `CreateBusWithOptionalModules`, we can create a bus with only pre2p and a NOOP telemetry module
//    so that we can the pre2p module without any issues.
//
func CreateBusWithOptionalModules(
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
	telemetry modules.TelemetryModule,
) modules.Bus {
	bus := &bus{
		channel:     make(modules.EventsChannel, DefaultPocketBusBufferSize),
		persistence: nil,
		p2p:         nil,
		utility:     nil,
		consensus:   nil,
		telemetry:   nil,
	}

	if persistence != nil {
		bus.persistence = persistence
		persistence.SetBus(bus)
	}

	if p2p != nil {
		bus.p2p = p2p
		p2p.SetBus(bus)
	}

	if utility != nil {
		bus.utility = utility
		utility.SetBus(bus)
	}

	if consensus != nil {
		bus.consensus = consensus
		consensus.SetBus(bus)
	}

	if telemetry != nil {
		bus.telemetry = telemetry
		telemetry.SetBus(bus)
	}

	return bus
}

func (m *bus) PublishEventToBus(e *types.PocketEvent) {
	m.channel <- *e
}

func (m *bus) GetBusEvent() *types.PocketEvent {
	e := <-m.channel
	return &e
}

func (m *bus) GetEventBus() modules.EventsChannel {
	return m.channel
}

func (m *bus) GetPersistenceModule() modules.PersistenceModule {
	return m.persistence
}

func (m *bus) GetP2PModule() modules.P2PModule {
	return m.p2p
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	return m.utility
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	return m.consensus
}

func (m *bus) GetTelemetryModule() modules.TelemetryModule {
	return m.telemetry
}
