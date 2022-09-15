package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/test_artifacts"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/modules"
	sharedTest "github.com/pokt-network/pocket/shared/test_artifacts"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

var (
	DefaultChains     = []string{"0001"}
	ChainsToUpdate    = []string{"0002"}
	DefaultServiceUrl = "https://foo.bar"
	DefaultPoolName   = "TESTING_POOL"

	DefaultDeltaBig     = big.NewInt(100)
	DefaultAccountBig   = big.NewInt(1000000)
	DefaultStakeBig     = big.NewInt(1000000000000000)
	DefaultMaxRelaysBig = big.NewInt(1000000)

	DefaultAccountAmount = types.BigIntToString(DefaultAccountBig)
	DefaultStake         = types.BigIntToString(DefaultStakeBig)
	DefaultMaxRelays     = types.BigIntToString(DefaultMaxRelaysBig)
	StakeToUpdate        = types.BigIntToString((&big.Int{}).Add(DefaultStakeBig, DefaultDeltaBig))

	DefaultStakeStatus     = persistence.StakedStatus
	DefaultPauseHeight     = int64(-1)
	DefaultUnstakingHeight = int64(-1)

	OlshanskyURL    = "https://olshansky.info"
	OlshanskyChains = []string{"OLSH"}

	testSchema = "test_schema"
)
var testPersistenceMod modules.PersistenceModule // initialized in TestMain

// See https://github.com/ory/dockertest as reference for the template of this code
// Postgres example can be found here: https://github.com/ory/dockertest/blob/v3/examples/PostgreSQL.md
func TestMain(m *testing.M) {
	pool, resource, dbUrl := sharedTest.SetupPostgresDocker()
	testPersistenceMod = newTestPersistenceModule(dbUrl)
	m.Run()
	os.Remove(testingConfigFilePath)
	os.Remove(testingGenesisFilePath)
	sharedTest.CleanupPostgresDocker(m, pool, resource)
}

func NewTestPostgresContext(t *testing.T, height int64) *persistence.PostgresContext {
	ctx, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	db := &persistence.PostgresContext{
		Height: height,
		DB:     ctx.(persistence.PostgresContext).DB,
	}

	t.Cleanup(func() {
		require.NoError(t, db.Release())
		require.NoError(t, testPersistenceMod.ResetContext())
	})

	return db
}

// REFACTOR: Can we leverage using `NewTestPostgresContext`here by creating a common interface?
func NewFuzzTestPostgresContext(f *testing.F, height int64) *persistence.PostgresContext {
	ctx, err := testPersistenceMod.NewRWContext(height)
	if err != nil {
		log.Fatalf("Error creating new context: %s", err)
	}
	db := persistence.PostgresContext{
		Height: height,
		DB:     ctx.(persistence.PostgresContext).DB,
	}
	f.Cleanup(func() {
		db.Release()
		testPersistenceMod.ResetContext()
	})

	return &db
}

// TODO(andrew): Take in `t testing.T` as a parameter and error if there's an issue
func newTestPersistenceModule(databaseUrl string) modules.PersistenceModule {
	cfg := modules.Config{
		Persistence: &types.PersistenceConfig{
			PostgresUrl:    databaseUrl,
			NodeSchema:     testSchema,
			BlockStorePath: "",
		},
	}
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	createTestingGenesisAndConfigFiles(cfg, genesisState)
	persistenceMod, err := persistence.Create(testingConfigFilePath, testingGenesisFilePath)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return persistenceMod
}

// IMPROVE(team): Extend this to more complex and variable test cases challenging & randomizing the state of persistence.
func fuzzSingleProtocolActor(
	f *testing.F,
	newTestActor func() (types.BaseActor, error),
	getTestActor func(db *persistence.PostgresContext, address string) (*types.BaseActor, error),
	protocolActorSchema types.ProtocolActorSchema) {

	db := NewFuzzTestPostgresContext(f, 0)

	err := db.DebugClearAll()
	require.NoError(f, err)

	actor, err := newTestActor()
	require.NoError(f, err)

	err = db.InsertActor(protocolActorSchema, actor)
	require.NoError(f, err)

	// IMPROVE(team): Extend this to make sure we have full code coverage of the persistence context operations.
	operations := []string{
		"UpdateActor",

		"GetActorsReadyToUnstake",
		"GetActorStatus",
		"GetActorPauseHeight",
		"GetActorOutputAddr",

		"SetActorUnstakingHeight",
		"SetActorPauseHeight",
		"SetPausedActorToUnstaking",

		"IncrementHeight"}
	numOperationTypes := len(operations)

	numDbOperations := 100
	for i := 0; i < numDbOperations; i++ {
		f.Add(operations[rand.Intn(numOperationTypes)])
	}

	f.Fuzz(func(t *testing.T, op string) {
		originalActor, err := getTestActor(db, actor.Address)
		require.NoError(t, err)

		addr, err := hex.DecodeString(originalActor.Address)
		require.NoError(t, err)

		switch op {
		case "UpdateActor":
			numParamUpdatesSupported := 3
			newStakedTokens := originalActor.StakedTokens
			newChains := originalActor.Chains
			newActorSpecificParam := originalActor.ActorSpecificParam

			iterations := rand.Intn(2)
			for i := 0; i < iterations; i++ {
				switch rand.Intn(numParamUpdatesSupported) {
				case 0:
					newStakedTokens = getRandomBigIntString()
				case 1:
					switch protocolActorSchema.GetActorSpecificColName() {
					case types.ServiceURLCol:
						newActorSpecificParam = getRandomServiceURL()
					case types.MaxRelaysCol:
						newActorSpecificParam = getRandomBigIntString()
					default:
						t.Error("Unexpected actor specific column name")
					}
				case 2:
					if protocolActorSchema.GetChainsTableName() != "" {
						newChains = getRandomChains()
					}
				}
			}
			updatedActor := types.BaseActor{
				Address:            originalActor.Address,
				PublicKey:          originalActor.PublicKey,
				StakedTokens:       newStakedTokens,
				ActorSpecificParam: newActorSpecificParam,
				OutputAddress:      originalActor.OutputAddress,
				PausedHeight:       originalActor.PausedHeight,
				UnstakingHeight:    originalActor.UnstakingHeight,
				Chains:             newChains,
			}
			err = db.UpdateActor(protocolActorSchema, updatedActor)
			require.NoError(t, err)

			newActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)

			require.ElementsMatch(t, newActor.Chains, newChains, "staked chains not updated")
			// TODO(andrew): Use `require.Contains` instead
			if strings.Contains(newActor.StakedTokens, "invalid") {
				fmt.Println("")
			}
			require.Equal(t, newActor.StakedTokens, newStakedTokens, "staked tokens not updated")
			require.Equal(t, newActor.ActorSpecificParam, newActorSpecificParam, "actor specific param not updated")
		case "GetActorsReadyToUnstake":
			unstakingActors, err := db.GetActorsReadyToUnstake(protocolActorSchema, db.Height)
			require.NoError(t, err)

			if originalActor.UnstakingHeight != db.Height { // Not ready to unstake
				require.Nil(t, unstakingActors)
			} else {
				idx := slices.IndexFunc(unstakingActors, func(a modules.IUnstakingActor) bool {
					return originalActor.Address == hex.EncodeToString(a.GetAddress())
				})
				require.NotEqual(t, idx, -1, fmt.Sprintf("actor that is unstaking was not found %+v", originalActor))
			}
		case "GetActorStatus":
			status, err := db.GetActorStatus(protocolActorSchema, addr, db.Height)
			require.NoError(t, err)

			switch {
			case originalActor.UnstakingHeight == DefaultUnstakingHeight:
				require.Equal(t, persistence.StakedStatus, status, "actor status should be staked")
			case originalActor.UnstakingHeight > db.Height:
				require.Equal(t, persistence.UnstakingStatus, status, "actor status should be unstaking")
			default:
				require.Equal(t, persistence.UnstakedStatus, status, "actor status should be unstaked")
			}
		case "GetActorPauseHeight":
			pauseHeight, err := db.GetActorPauseHeightIfExists(protocolActorSchema, addr, db.Height)
			require.NoError(t, err)

			require.Equal(t, originalActor.PausedHeight, pauseHeight, "pause height incorrect")
		case "SetActorUnstakingHeight":
			newUnstakingHeight := rand.Int63()

			err = db.SetActorUnstakingHeightAndStatus(protocolActorSchema, addr, newUnstakingHeight)
			require.NoError(t, err)

			newActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)

			require.Equal(t, newUnstakingHeight, newActor.UnstakingHeight, "setUnstakingHeight")
		case "SetActorPauseHeight":
			newPauseHeight := rand.Int63()

			err = db.SetActorPauseHeight(protocolActorSchema, addr, newPauseHeight)
			require.NoError(t, err)

			newActor, err := getTestActor(db, actor.Address)
			require.NoError(t, err)

			require.Equal(t, newPauseHeight, newActor.PausedHeight, "setPauseHeight")
		case "SetPausedActorToUnstaking":
			newUnstakingHeight := db.Height + int64(rand.Intn(15))
			err = db.SetActorStatusAndUnstakingHeightIfPausedBefore(protocolActorSchema, db.Height, newUnstakingHeight)
			require.NoError(t, err)

			newActor, err := getTestActor(db, originalActor.Address)
			require.NoError(t, err)

			if db.Height > originalActor.PausedHeight { // isPausedAndReadyToUnstake
				require.Equal(t, newActor.UnstakingHeight, newUnstakingHeight, "setPausedToUnstaking")
			}
		case "GetActorOutputAddr":
			outputAddr, err := db.GetActorOutputAddress(protocolActorSchema, addr, db.Height)
			require.NoError(t, err)

			require.Equal(t, originalActor.OutputAddress, hex.EncodeToString(outputAddr), "output address incorrect")
		case "IncrementHeight":
			db.Height++
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}
	})
}

// TODO(olshansky): Make these functions & variables more functional to avoid having "unexpected"
//                  side effects and making it clearer to the reader.
const (
	testingGenesisFilePath = "genesis.json"
	testingConfigFilePath  = "config.json"
)

func createTestingGenesisAndConfigFiles(cfg modules.Config, genesisState modules.GenesisState) {
	config, err := json.Marshal(cfg.Persistence)
	if err != nil {
		log.Fatal(err)
	}
	genesis, err := json.Marshal(genesisState.PersistenceGenesisState)
	if err != nil {
		log.Fatal(err)
	}
	genesisFile := make(map[string]json.RawMessage)
	configFile := make(map[string]json.RawMessage)
	persistenceModuleName := new(persistence.PersistenceModule).GetModuleName()
	genesisFile[test_artifacts.GetGenesisFileName(persistenceModuleName)] = genesis
	configFile[persistenceModuleName] = config
	genesisFileBz, err := json.MarshalIndent(genesisFile, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	configFileBz, err := json.MarshalIndent(configFile, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(testingGenesisFilePath, genesisFileBz, 0777); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(testingConfigFilePath, configFileBz, 0777); err != nil {
		log.Fatal(err)
	}
}

func getRandomChains() (chains []string) {
	setRandomSeed()

	charOptions := "ABCDEF0123456789"
	numCharOptions := len(charOptions)

	chainsMap := make(map[string]struct{})
	for i := 0; i < rand.Intn(14)+1; i++ {
		b := make([]byte, 4)
		for i := range b {
			b[i] = charOptions[rand.Intn(numCharOptions)]
		}
		if _, found := chainsMap[string(b)]; found {
			i--
			continue
		}
		chainsMap[string(b)] = struct{}{}
		chains = append(chains, string(b))
	}
	return
}

func getRandomServiceURL() string {
	setRandomSeed()

	charOptions := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numCharOptions := len(charOptions)

	b := make([]byte, rand.Intn(12))
	for i := range b {
		b[i] = charOptions[rand.Intn(numCharOptions)]
	}

	return fmt.Sprintf("https://%s.com", string(b))
}

func getRandomBigIntString() string {
	return types.BigIntToString(big.NewInt(rand.Int63()))
}

func setRandomSeed() {
	rand.Seed(time.Now().UnixNano())
}