package pre_p2p

import (
	"fmt"
	"log"
	"pocket/p2p/pre_p2p/types"
	"pocket/shared/config"
	"pocket/shared/crypto"
	"sync"
)

// TODO: Return this as a singleton!
// Consider using the singleton pattern? https://medium.com/golang-issue/how-singleton-pattern-works-with-golang-2fdd61cd5a7f

type TestState struct {
	BlockHeight      uint64
	AppHash          string       // TODO: Why not call this a BlockHash or StateHash? SHould it be a []byte or string?
	ValidatorMap     types.ValMap // TODO: Need to update this on every validator pause/stake/unstake/etc.
	TotalVotingPower uint64       // TODO: Need to update this on every send transaction.

	PublicKey crypto.PublicKey
	Address   string

	Config config.Config // TODO: Should we store this here?
}

// The pocket state singleton.
var state *TestState

// Used to load the state when the singleton is created.
var once sync.Once

// Used to update the state. All exported functions should lock this when they are called and defer an unlock.
var lock = &sync.Mutex{}

func GetTestState() *TestState {
	once.Do(func() {
		state = &TestState{}
	})

	return state
}

func (ps *TestState) LoadStateFromConfig(cfg *config.Config) {
	lock.Lock()
	defer lock.Unlock()

	persistenceConfig := cfg.Persistence
	if persistenceConfig == nil || len(persistenceConfig.DataDir) == 0 {
		log.Println("[TODO] Load p2p state from persistence. Only supporting loading p2p state from genesis file for now.")
		ps.loadStateFromGenesis(cfg)
		return
	}

	log.Fatalf("[TODO] Config must not be nil when initializing the pocket state. ...")
}

func (ps *TestState) loadStateFromGenesis(cfg *config.Config) {
	genesis, err := PocketGenesisFromFileOrJSON(cfg.Genesis)
	if err != nil {
		log.Fatalf("Failed to load genesis: %v", err)
	}

	if cfg.PrivateKey == "" {
		log.Fatalf("[TODO] Private key must be set when initializing the pocket state. ...")
	}
	pk, err := crypto.NewPrivateKey(cfg.PrivateKey)
	if err != nil {
		panic(err)
	}
	*ps = TestState{
		BlockHeight:  0,
		ValidatorMap: types.ValidatorListToMap(genesis.Validators),

		PublicKey: pk.PublicKey(),
		Address:   pk.Address().String(),

		Config: *cfg,
	}

	ps.recomputeTotalVotingPower()
}

func (ps *TestState) recomputeTotalVotingPower() {
	ps.TotalVotingPower = 0
	for _, v := range ps.ValidatorMap {
		ps.TotalVotingPower += v.UPokt
	}
}

func (ps *TestState) PrintGlobalState() {
	fmt.Printf("\tGLOBAL STATE: (BlockHeight, PrevAppHash, # Validators, TotalVotingPower) is: (%d, %s, %d, %d)\n", ps.BlockHeight, ps.AppHash, len(ps.ValidatorMap), ps.TotalVotingPower)
}

func (ps *TestState) UpdateAppHash(appHash string) {
	ps.AppHash = appHash
}

func (ps *TestState) UpdateBlockHeight(blockHeight uint64) {
	ps.BlockHeight = blockHeight
}