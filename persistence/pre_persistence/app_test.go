package pre_persistence

import (
	"bytes"
	"github.com/pokt-network/pocket/shared/types"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

func NewTestApp() typesGenesis.App {
	pub1, _ := crypto.GeneratePublicKey()
	addr1 := pub1.Address()
	addr2, _ := crypto.GenerateAddress()
	defaultMaxRelays := types.BigIntToString(big.NewInt(1000000))
	return typesGenesis.App{
		Address:         addr1,
		PublicKey:       pub1.Bytes(),
		Paused:          false,
		Status:          typesGenesis.DefaultStakeStatus,
		Chains:          typesGenesis.DefaultChains,
		MaxRelays:       defaultMaxRelays,
		StakedTokens:    typesGenesis.DefaultStake,
		PausedHeight:    0,
		UnstakingHeight: 0,
		Output:          addr2,
	}
}

func TestGetAppExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	addr2, _ := crypto.GenerateAddress()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetAppExists(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("actor that should exists does not")
	}
	exists, err = ctx.GetAppExists(addr2)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("actor that exists should not")
	}
}

func TestGetApp(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	got, err := ctx.(*PrePersistenceContext).GetApp(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actor.Address, got.Address) || !bytes.Equal(actor.PublicKey, got.PublicKey) {
		t.Fatalf("unexpected actor returned; expected %v got %v", actor, got)
	}
}

func TestGetAllApps(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor1 := NewTestApp()
	actor2 := NewTestApp()
	if err := ctx.InsertApplication(actor1.Address, actor1.PublicKey, actor1.Output, actor1.Paused, int(actor1.Status),
		actor1.MaxRelays, actor1.StakedTokens, actor1.Chains, int64(actor1.PausedHeight), actor1.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.InsertApplication(actor2.Address, actor2.PublicKey, actor2.Output, actor2.Paused, int(actor2.Status),
		actor2.MaxRelays, actor2.StakedTokens, actor2.Chains, int64(actor2.PausedHeight), actor2.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	apps, err := ctx.(*PrePersistenceContext).GetAllApps(0)
	if err != nil {
		t.Fatal(err)
	}
	got1, got2 := false, false
	for _, a := range apps {
		if bytes.Equal(a.Address, actor1.Address) {
			got1 = true
		}
		if bytes.Equal(a.Address, actor2.Address) {
			got2 = true
		}
	}
	if !got1 || !got2 {
		t.Fatal("not all actors returned")
	}
}

func TestUpdateApplication(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	zero := types.BigIntToString(&big.Int{})
	bigExpectedTokens := big.NewInt(1)
	one := types.BigIntToString(bigExpectedTokens)
	before, err := ctx.(*PrePersistenceContext).GetApp(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	tokens := before.StakedTokens
	bigBeforeTokens, err := types.StringToBigInt(tokens)
	if err != nil {
		t.Fatal(err)
	}
	err = ctx.UpdateApplication(actor.Address, zero, one, typesGenesis.DefaultChains)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ctx.(*PrePersistenceContext).GetApp(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	bigAfterTokens, err := types.StringToBigInt(got.StakedTokens)
	if err != nil {
		t.Fatal(err)
	}
	bigAfterTokens.Sub(bigAfterTokens, bigBeforeTokens)
	if bigAfterTokens.Cmp(bigExpectedTokens) != 0 {
		t.Fatal("incorrect after balance")
	}
}

func TestDeleteApplication(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	err := ctx.DeleteApplication(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.(*PrePersistenceContext).GetAppExists(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("actor exists when it shouldn't")
	}
}

func TestGetAppsReadyToUnstake(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetAppUnstakingHeightAndStatus(actor.Address, 0, 1); err != nil {
		t.Fatal(err)
	}
	unstakingApps, err := ctx.(*PrePersistenceContext).GetAppsReadyToUnstake(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unstakingApps[0].Address, actor.Address) {
		t.Fatalf("wrong actor returned, expected addr %v, got %v", unstakingApps[0].Address, actor.Address)
	}
}

func TestGetAppStatus(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetAppStatus(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if status != int(actor.Status) {
		t.Fatal("unequal status")
	}
}

func TestGetAppPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pauseHeight := 1
	err := ctx.SetAppPauseHeight(actor.Address, int64(pauseHeight))
	if err != nil {
		t.Fatal(err)
	}
	pauseBeforeHeight, err := ctx.GetAppPauseHeightIfExists(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if pauseHeight != int(pauseBeforeHeight) {
		t.Fatalf("incorrect pause height: expected %v, got %v", pauseHeight, pauseBeforeHeight)
	}
}

func TestSetAppsStatusAndUnstakingHeightPausedBefore(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, true, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	pauseBeforeHeight, unstakingHeight, status := int64(1), int64(10), 1
	err := ctx.SetAppsStatusAndUnstakingHeightPausedBefore(pauseBeforeHeight, unstakingHeight, status)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ctx.(*PrePersistenceContext).GetApp(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if got.UnstakingHeight != unstakingHeight {
		t.Fatalf("wrong unstaking height: expected %v, got %v", unstakingHeight, got.UnstakingHeight)
	}
	if int(got.Status) != status {
		t.Fatalf("wrong status: expected %v, got %v", status, got.Status)
	}
}

func TestGetAppOutputAddress(t *testing.T) {
	ctx := NewTestingPrePersistenceContext(t)
	actor := NewTestApp()
	if err := ctx.InsertApplication(actor.Address, actor.PublicKey, actor.Output, actor.Paused, int(actor.Status),
		actor.MaxRelays, actor.StakedTokens, actor.Chains, int64(actor.PausedHeight), actor.UnstakingHeight); err != nil {
		t.Fatal(err)
	}
	output, err := ctx.GetAppOutputAddress(actor.Address)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actor.Output, output) {
		t.Fatalf("incorrect output address expected %v, got %v", actor.Output, output)
	}
}
