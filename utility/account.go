package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) HandleMessageSend(message *typesUtil.MessageSend) types.Error {
	// convert the amount to big.Int
	amount, err := types.StringToBigInt(message.Amount)
	if err != nil {
		return err
	}
	// get the sender's account amount
	fromAccountAmount, err := u.GetAccountAmount(message.FromAddress)
	if err != nil {
		return err
	}
	// subtract that amount from the sender
	fromAccountAmount.Sub(fromAccountAmount, amount)
	// if they go negative, they don't have sufficient funds
	// NOTE: we don't use the u.SubtractAccountAmount() function because Utility needs to do this check
	if fromAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmountError()
	}
	// add the amount to the recipient's account
	if err := u.AddAccountAmount(message.ToAddress, amount); err != nil {
		return err
	}
	// set the sender's account amount
	if err := u.SetAccountAmount(message.FromAddress, fromAccountAmount); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetMessageSendSignerCandidates(msg *typesUtil.MessageSend) ([][]byte, types.Error) {
	// only the from address is a proper signer candidate
	return [][]byte{msg.FromAddress}, nil
}

func (u *UtilityContext) GetAccountAmount(address []byte) (*big.Int, types.Error) {
	store := u.Store()
	amount, err := store.GetAccountAmount(address)
	if err != nil {
		return nil, types.ErrGetAccountAmount(err)
	}
	return types.StringToBigInt(amount)
}

func (u *UtilityContext) AddAccountAmount(address []byte, amountToAdd *big.Int) types.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, types.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) AddAccountAmountString(address []byte, amountToAdd string) types.Error {
	store := u.Store()
	if err := store.AddAccountAmount(address, amountToAdd); err != nil {
		return types.ErrAddAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) AddPoolAmount(name string, amountToAdd *big.Int) types.Error {
	store := u.Store()
	if err := store.AddPoolAmount(name, types.BigIntToString(amountToAdd)); err != nil {
		return types.ErrAddPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) SubPoolAmount(name string, amountToSub string) types.Error {
	store := u.Store()
	if err := store.SubtractPoolAmount(name, amountToSub); err != nil {
		return types.ErrSubPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) GetPoolAmount(name string) (*big.Int, types.Error) {
	store := u.Store()
	tokens, er := store.GetPoolAmount(name)
	if er != nil {
		return nil, types.ErrGetPoolAmount(name, er)
	}
	amount, err := types.StringToBigInt(tokens)
	if err != nil {
		return nil, err
	}
	return amount, nil
}

func (u *UtilityContext) InsertPool(name string, address []byte, amount string) types.Error {
	store := u.Store()
	if err := store.InsertPool(name, address, amount); err != nil {
		return types.ErrSetPool(name, err)
	}
	return nil
}

func (u *UtilityContext) SetPoolAmount(name string, amount *big.Int) types.Error {
	store := u.Store()
	if err := store.SetPoolAmount(name, types.BigIntToString(amount)); err != nil {
		return types.ErrSetPoolAmount(name, err)
	}
	return nil
}

func (u *UtilityContext) SetAccountWithAmountString(address []byte, amount string) types.Error {
	store := u.Store()
	if err := store.SetAccountAmount(address, amount); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) SetAccountAmount(address []byte, amount *big.Int) types.Error {
	store := u.Store()
	if err := store.SetAccountAmount(address, types.BigIntToString(amount)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}

func (u *UtilityContext) SubtractAccountAmount(address []byte, amountToSubtract *big.Int) types.Error {
	store := u.Store()
	if err := store.SubtractAccountAmount(address, types.BigIntToString(amountToSubtract)); err != nil {
		return types.ErrSetAccountAmount(err)
	}
	return nil
}
