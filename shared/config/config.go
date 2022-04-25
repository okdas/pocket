package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type JsonConfig struct {
	RootDir string `json:"root_dir"`
	Genesis string `json:"genesis"`

	PrivateKey string `json:"private_key"`

	Pre2P          Pre2PConfig          `json:"pre2p"` // TODO(derrandz): delete this once P2P is ready.
	P2P            P2PConfig            `json:"p2p"`
	Consensus      ConsensusConfig      `json:"consensus"`
	PrePersistence PrePersistenceConfig `json:"pre_persistence"`
	Persistence    PersistenceConfig    `json:"persistence"`
	Utility        UtilityConfig        `json:"utility"`
}
type Config struct {
	RootDir string `json:"root_dir"`
	Genesis string `json:"genesis"`

	PrivateKey cryptoPocket.Ed25519PrivateKey `json:"private_key"`

	Pre2P          *Pre2PConfig          `json:"pre2p"` // TODO(derrandz): delete this once P2P is ready.
	P2P            *P2PConfig            `json:"p2p"`
	Consensus      *ConsensusConfig      `json:"consensus"`
	PrePersistence *PrePersistenceConfig `json:"pre_persistence"`
	Persistence    *PersistenceConfig    `json:"persistence"`
	Utility        *UtilityConfig        `json:"utility"`
}

// TODO(derrandz): delete this once P2P is ready.
type Pre2PConfig struct {
	ConsensusPort uint32 `json:"consensus_port"`
}

type PrePersistenceConfig struct {
	Capacity        int `json:"capacity"`
	MempoolMaxBytes int `json:"mempool_max_bytes"`
	MempoolMaxTxs   int `json:"mempool_max_txs"`
}

type P2PConfig struct {
	Protocol         string   `json:"protocol,omitempty"`
	Address          string   `json:"address,omitempty"`
	ExternalIp       string   `json:"external_ip,omitempty"`
	Peers            []string `json:"peers,omitempty"`
	BufferSize       uint     `json:"connection_buffer_size,omitempty"`
	WireHeaderLength uint     `json:"max_wire_header_length,omitempty"`
	TimeoutInMs      uint     `json:"timeout_in_ms,omitempty"`
	ID               int      `json:"id,omitempty"`
	Redundancy       bool     `json:"redundancy,omitempty"`
}

type PacemakerConfig struct {
	TimeoutMsec               uint64 `json:"timeout_msec,omitempty"`
	Manual                    bool   `json:"manual,omitempty"`
	DebugTimeBetweenStepsMsec uint64 `json:"debug_time_between_steps_msec,omitempty"`
}

type ConsensusConfig struct {
	// Mempool
	MaxMempoolBytes uint64 `json:"max_mempool_bytes"` // TODO(olshansky): add unit tests for this

	// Block
	MaxBlockBytes uint64 `json:"max_block_bytes"` // TODO(olshansky): add unit tests for this

	// Pacemaker
	Pacemaker *PacemakerConfig `json:"pacemaker"`
}

type PersistenceConfig struct {
	DataDir string `json:"datadir"`
}

type UtilityConfig struct {
}

// TODO(insert tooling issue # here): Re-evaluate how load configs should be handeled.
func LoadConfig(file string) (c *Config) {
	c = &Config{}

	jsonFile, err := os.Open(file)
	if err != nil {
		log.Fatalln("Error opening config file: ", err)
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalln("Error reading config file: ", err)
	}
	jsonCfg := &JsonConfig{}
	if err = json.Unmarshal(bytes, jsonCfg); err != nil {
		log.Fatalln("Error parsing config file: ", err)
	}

	c = jsonCfg.toConfig()

	if err := c.ValidateAndHydrate(); err != nil {
		log.Fatalln("Error validating or completing config: ", err)
	}

	return
}

func (c *Config) ValidateAndHydrate() error {
	if len(c.PrivateKey) == 0 {
		return fmt.Errorf("private key in config file cannot be empty")
	}

	if len(c.Genesis) == 0 {
		return fmt.Errorf("must specify a genesis file or string")
	}
	c.Genesis = rootify(c.Genesis, c.RootDir)

	if err := c.Consensus.ValidateAndHydrate(); err != nil {
		log.Fatalln("Error validating or completing consensus config: ", err)
	}

	if err := c.P2P.ValidateAndHydrate(); err != nil {
		log.Fatalln("Error validating or completing P2P config: ", err)
	}

	return nil
}

func (c *P2PConfig) ValidateAndHydrate() error {
	return nil
}

func (c *ConsensusConfig) ValidateAndHydrate() error {
	if err := c.Pacemaker.ValidateAndHydrate(); err != nil {
		log.Fatalf("Error validating or completing Pacemaker configs")
	}

	if c.MaxMempoolBytes <= 0 {
		return fmt.Errorf("MaxMempoolBytes must be a positive integer")
	}

	if c.MaxBlockBytes <= 0 {
		return fmt.Errorf("MaxBlockBytes must be a positive integer")
	}

	return nil
}

func (jc *JsonConfig) toConfig() *Config {
	pc := &Config{
		RootDir: jc.RootDir,
		Genesis: jc.Genesis,

		PrivateKey: cryptoPocket.Ed25519PrivateKey([]byte(jc.PrivateKey)),

		P2P: &P2PConfig{
			Protocol:         jc.P2P.Protocol,
			Address:          jc.P2P.Address,
			ExternalIp:       jc.P2P.ExternalIp,
			Peers:            jc.P2P.Peers,
			BufferSize:       jc.P2P.BufferSize,
			WireHeaderLength: jc.P2P.WireHeaderLength,
			TimeoutInMs:      jc.P2P.TimeoutInMs,
		},
		Consensus: &ConsensusConfig{
			MaxMempoolBytes: jc.Consensus.MaxMempoolBytes,
			MaxBlockBytes:   jc.Consensus.MaxBlockBytes,
			Pacemaker: &PacemakerConfig{
				TimeoutMsec:               jc.Consensus.Pacemaker.TimeoutMsec,
				Manual:                    jc.Consensus.Pacemaker.Manual,
				DebugTimeBetweenStepsMsec: jc.Consensus.Pacemaker.DebugTimeBetweenStepsMsec,
			},
		},
		PrePersistence: &PrePersistenceConfig{
			Capacity:        jc.PrePersistence.Capacity,
			MempoolMaxBytes: jc.PrePersistence.MempoolMaxBytes,
			MempoolMaxTxs:   jc.PrePersistence.MempoolMaxTxs,
		},
		Persistence: &PersistenceConfig{
			DataDir: jc.Persistence.DataDir,
		},
		Utility: &UtilityConfig{},
	}
	return pc
}

func (c *PacemakerConfig) ValidateAndHydrate() error {
	return nil
}

// Helper function to make config creation independent of root dir
func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
