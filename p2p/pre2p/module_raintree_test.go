package pre2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	mocksPre2P "github.com/pokt-network/pocket/p2p/pre2p/types/mocks"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

// IMPROVE(team): Looking into adding more tests and accounting for more edge cases.

// ### RainTree Unit Tests ###

func TestRainTreeCompleteOneNodes(t *testing.T) {
	// val_1
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1): {0, 0}, // {numReads, numWrites}
	}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}

func TestRainTreeCompleteTwoNodes(t *testing.T) {
	// val_1
	//   └───────┐
	// 	       val_2
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {1, 1},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}

func TestRainTreeCompleteThreeNodes(t *testing.T) {
	// 	          val_1
	// 	   ┌───────┴────┬─────────┐
	//   val_2        val_1     val_3
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {1, 1},
		validatorId(t, 3): {1, 1},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}

func TestRainTreeCompleteFourNodes(t *testing.T) {
	// Test configurations (visualization retrieved from simulator)
	// 	                val_1
	// 	  ┌───────────────┴────┬─────────────────┐
	//  val_2                val_1             val_3
	//    └───────┐            └───────┐         └───────┐
	// 		    val_3                val_2             val_4
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {2, 2},
		validatorId(t, 3): {2, 2},
		validatorId(t, 4): {1, 1},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}

func TestRainTreeCompleteNineNodes(t *testing.T) {
	// 	                              val_1
	// 	         ┌──────────────────────┴────────────┬────────────────────────────────┐
	//         val_4                               val_1                            val_7
	//   ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐
	// val_6        val_4     val_8        val_3        val_1     val_5     val_9        val_7     val_2
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {1, 1},
		validatorId(t, 3): {1, 1},
		validatorId(t, 4): {1, 1},
		validatorId(t, 5): {1, 1},
		validatorId(t, 6): {1, 1},
		validatorId(t, 7): {1, 1},
		validatorId(t, 8): {1, 1},
		validatorId(t, 9): {1, 1},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}

func TestRainTreeCompleteEighteenNodes(t *testing.T) {
	// 	                                                                                                              val_1
	// 	                                      ┌──────────────────────────────────────────────────────────────────────────┴─────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────┐
	//                                      val_7                                                                                                            val_1                                                                                                     val_13
	//             ┌──────────────────────────┴────────────┬────────────────────────────────────┐                                     ┌────────────────────────┴────────────┬──────────────────────────────────┐                                ┌────────────────────────┴──────────────┬────────────────────────────────────┐
	//           val_11                                   val_7                               val_15                                 val_5                                 val_1                              val_9                           val_17                                  val_13                                val_3
	//    ┌────────┴─────┬───────────┐             ┌───────┴────┬──────────┐           ┌────────┴─────┬──────────┐            ┌───────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴────┬─────────┐         ┌───────┴─────┬──────────┐             ┌────────┴─────┬───────────┐          ┌───────┴────┬──────────┐
	// val_13         val_11      val_16        val_9        val_7      val_12      val_17         val_15     val_8        val_7        val_5      val_10        val_3        val_1     val_6      val_11        val_9     val_2     val_1         val_17     val_4         val_15         val_13      val_18     val_5        val_3      val_14
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1):  {1, 1}, // Originator
		validatorId(t, 2):  {1, 1},
		validatorId(t, 3):  {2, 2},
		validatorId(t, 4):  {1, 1},
		validatorId(t, 5):  {2, 2},
		validatorId(t, 6):  {1, 1},
		validatorId(t, 7):  {2, 2},
		validatorId(t, 8):  {1, 1},
		validatorId(t, 9):  {2, 2},
		validatorId(t, 10): {1, 1},
		validatorId(t, 11): {2, 2},
		validatorId(t, 12): {1, 1},
		validatorId(t, 13): {2, 2},
		validatorId(t, 14): {1, 1},
		validatorId(t, 15): {2, 2},
		validatorId(t, 16): {1, 1},
		validatorId(t, 17): {2, 2},
		validatorId(t, 18): {1, 1},
	}
	// Note that the originator, `val_1` is also messaged by `val_17` outside of continuously
	// demoting itself.
	testRainTreeCalls(t, originatorNode, expectedCalls, true)
}

func TestRainTreeCompleteTwentySevenNodes(t *testing.T) {
	// 	                                                                                                                    val_1
	// 	                                     ┌────────────────────────────────────────────────────────────────────────────────┴───────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────┐
	//                                    val_10                                                                                                                   val_1                                                                                                       val_19
	//            ┌──────────────────────────┴──────────────┬──────────────────────────────────────┐                                         ┌────────────────────────┴────────────┬──────────────────────────────────┐                                  ┌────────────────────────┴──────────────┬────────────────────────────────────┐
	//          val_16                                    val_10                                 val_22                                     val_7                                 val_1                             val_13                             val_25                                  val_19                                val_4
	//   ┌────────┴─────┬───────────┐              ┌────────┴─────┬───────────┐           ┌────────┴─────┬───────────┐              ┌────────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴─────┬──────────┐         ┌───────┴─────┬──────────┐             ┌────────┴─────┬───────────┐          ┌───────┴────┬──────────┐
	// val_20         val_16      val_24         val_14         val_10      val_18      val_26         val_22      val_12         val_11        val_7      val_15        val_5        val_1     val_9      val_17         val_13     val_3     val_2         val_25     val_6         val_23         val_19      val_27     val_8        val_4      val_21
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeCommConfig{
		validatorId(t, 1):  {0, 0}, // Originator
		validatorId(t, 2):  {1, 1},
		validatorId(t, 3):  {1, 1},
		validatorId(t, 4):  {1, 1},
		validatorId(t, 5):  {1, 1},
		validatorId(t, 6):  {1, 1},
		validatorId(t, 7):  {1, 1},
		validatorId(t, 8):  {1, 1},
		validatorId(t, 9):  {1, 1},
		validatorId(t, 10): {1, 1},
		validatorId(t, 11): {1, 1},
		validatorId(t, 12): {1, 1},
		validatorId(t, 13): {1, 1},
		validatorId(t, 14): {1, 1},
		validatorId(t, 15): {1, 1},
		validatorId(t, 16): {1, 1},
		validatorId(t, 17): {1, 1},
		validatorId(t, 18): {1, 1},
		validatorId(t, 19): {1, 1},
		validatorId(t, 20): {1, 1},
		validatorId(t, 21): {1, 1},
		validatorId(t, 22): {1, 1},
		validatorId(t, 23): {1, 1},
		validatorId(t, 24): {1, 1},
		validatorId(t, 25): {1, 1},
		validatorId(t, 26): {1, 1},
		validatorId(t, 27): {1, 1},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls, false)
}

// ### RainTree Unit Helpers - To remove redundancy of code in the unit tests ###

func testRainTreeCalls(t *testing.T, origNode string, testCommConfig TestRainTreeCommConfig, isOriginatorPinged bool) {
	// Network configurations
	numValidators := len(testCommConfig)
	configs := createConfigs(t, numValidators)

	// Test configurations
	var messageHandeledWaitGroup sync.WaitGroup
	if isOriginatorPinged {
		messageHandeledWaitGroup.Add(numValidators)
	} else {
		messageHandeledWaitGroup.Add(numValidators - 1) // -1 because the originator node implicitly handles the message
	}

	// Network initialization
	connMocks := make(map[string]typesPre2P.Transport)
	busMocks := make(map[string]modules.Bus)
	for valId, expectedCall := range testCommConfig {
		connMocks[valId] = prepareConnMock(t, expectedCall.numNetworkReads, expectedCall.numNetworkWrites)
		busMocks[valId] = prepareBusMock(t, &messageHandeledWaitGroup)
	}

	// Module injection
	p2pModules := prepareP2PModules(t, configs)
	for validatorId, mod := range p2pModules {
		mod.listener = connMocks[validatorId]
		mod.SetBus(busMocks[validatorId])
		for _, peer := range mod.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
		mod.Start()
		defer mod.Stop()
	}

	// Trigger originator message
	p := &anypb.Any{}
	p2pMod := p2pModules[origNode]
	p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC)

	// Wait for completion
	done := make(chan struct{})
	go func() {
		messageHandeledWaitGroup.Wait()
		close(done)
	}()

	// Timeout or succeed
	select {
	case <-done:
	// All done!
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message to be handled")
	}
}

// ### RainTree Unit Utils - Configurations & constants and such ###

const (
	genesisConfigSeedStart = 42
	maxNumKeys             = 42 // The number of keys generated for all the unit tests. Optimization to avoid regenerating every time.
	serviceUrlFormat       = "val_%d"
	testChannelSize        = 10000
)

// TODO(olshansky): Add configurations tests for dead and partially visible nodes
type TestRainTreeCommConfig map[string]struct {
	numNetworkReads  uint16
	numNetworkWrites uint16
}

var keys []cryptoPocket.PrivateKey

func init() {
	keys = generateKeys(nil, maxNumKeys)
}

func generateKeys(_ *testing.T, numValidators int) []cryptoPocket.PrivateKey {
	keys := make([]cryptoPocket.PrivateKey, numValidators)

	for i := range keys {
		seedInt := genesisConfigSeedStart + i
		seed := make([]byte, ed25519.PrivateKeySize)
		binary.LittleEndian.PutUint32(seed, uint32(seedInt))
		pk, err := cryptoPocket.NewPrivateKeyFromSeed(seed)
		if err != nil {
			panic(err)
		}
		keys[i] = pk
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Address().String() < keys[j].Address().String()
	})
	return keys
}

// A mock of the application specific to know if a message was sent to be handled by the application
// INVESTIGATE(olshansky): Double check that how the expected calls are counted is accurate per the
//                         expectation with RainTree by comparing with Telemetry after updating specs.
func prepareBusMock(t *testing.T, wg *sync.WaitGroup) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)
	busMock.EXPECT().PublishEventToBus(gomock.Any()).Do(func(e *types.PocketEvent) {
		wg.Done()
		fmt.Println("App specific bus mock publishing event to bus")
	}).MaxTimes(1) // Using `MaxTimes` rather than `Times` because originator node implicitly handles the message
	return busMock
}

// The reason with use `MaxTimes` instead of `Times` here is because we could have gotten full coverage
// while a message was still being sent that would have later been dropped due to de-duplication. There
// is a race condition here, but it is okay because our goal is to achieve max coverage with an upper limit
// on the number of expected messages propagated.
// INVESTIGATE(olshansky): Double check that how the expected calls are counted is accurate per the
//                         expectation with RainTree by comparing with Telemetry after updating specs.
func prepareConnMock(t *testing.T, expectedNumNetworkReads, expectedNumNetworkWrites uint16) typesPre2P.Transport {
	testChannel := make(chan []byte, testChannelSize)
	ctrl := gomock.NewController(t)
	connMock := mocksPre2P.NewMockTransport(ctrl)

	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		data := <-testChannel
		return data, nil
	}).MaxTimes(int(expectedNumNetworkReads + 1)) // INVESTIGATE(olshansky): The +1 is necessary because there is one extra read of empty data by every channel...

	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		testChannel <- data
		return nil
	}).MaxTimes(int(expectedNumNetworkWrites))

	connMock.EXPECT().Close().Return(nil).Times(1)

	return connMock
}

func prepareP2PModules(t *testing.T, configs []*config.Config) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(configs))
	for i, config := range configs {
		_ = typesGenesis.GetNodeState(config)
		p2pMod, err := Create(config)
		require.NoError(t, err)
		p2pModules[validatorId(t, i+1)] = p2pMod.(*p2pModule)
		// HACK(olshansky): I hate that we have to do this, but it is outside the scope of this change...
		// Cleanup once we get rid of the singleton
		typesGenesis.ResetNodeState(t)
	}
	return
}

func createConfigs(t *testing.T, numValidators int) (configs []*config.Config) {
	configs = make([]*config.Config, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	validatorConfigs := genesisValidatorConfig(t, valKeys)

	for i := range configs {
		configs[i] = &config.Config{
			RootDir: "",
			Genesis: genesisJson(t, numValidators, validatorConfigs),

			PrivateKey: valKeys[i].(cryptoPocket.Ed25519PrivateKey),

			Pre2P: &config.Pre2PConfig{
				ConsensusPort:  8080,
				UseRainTree:    true,
				ConnectionType: config.EmptyConnection,
			},
			P2P:            &config.P2PConfig{},
			Consensus:      &config.ConsensusConfig{},
			PrePersistence: &config.PrePersistenceConfig{},
			Persistence:    &config.PersistenceConfig{},
			Utility:        &config.UtilityConfig{},
		}
	}
	return
}

func validatorId(_ *testing.T, i int) string {
	return fmt.Sprintf(serviceUrlFormat, i)
}

// TECHDEBT(olshansky): The fact that we are passing in a genesis string rather than a properly
// configured struct is a bit of legacy. Need to fix this sooner rather than later.
func genesisJson(_ *testing.T, numValidators int, validatorConfigs string) string {
	return fmt.Sprintf(`{
		"genesis_state_configs": {
			"num_validators": %d,
			"num_applications": 0,
			"num_fisherman": 0,
			"num_servicers": 0,
			"keys_seed_start": %d
		},
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash",
		"validators": [%s]
	}`, numValidators, genesisConfigSeedStart, validatorConfigs)
}

func genesisValidatorConfig(t *testing.T, valKeys []cryptoPocket.PrivateKey) string {
	s := strings.Builder{}
	for i, valKey := range valKeys {
		if i != 0 {
			s.WriteString(",")
		}
		addr := valKey.Address().String()
		s.WriteString(fmt.Sprintf(`{
			"status": 2,
			"service_url": "%s",
			"staked_tokens": "1000000000000000",
			"address": "%s",
			"output": "%s",
			"public_key": "%s"
		}`, validatorId(t, i+1), addr, addr, valKey.PublicKey().String()))
	}
	return s.String()
}
