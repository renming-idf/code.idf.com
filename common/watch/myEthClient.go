package watch

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type myEthClient struct {
	C           *ethclient.Client
	chainID     *big.Int
	transferPre string
}

func (m *myEthClient) NewClient(rawUrl string) error {
	c, e := ethclient.Dial(rawUrl)
	if e != nil {
		return e
	}
	m.chainID, e = c.ChainID(context.Background())
	if e != nil {
		return e
	}
	m.C = c
	m.transferPre = "a9059cbb"
	return nil
}

//私钥或证书随便传一个，证书优先
func (m *myEthClient) signTransaction(tx *types.Transaction, privateKeyStr string, pk *ecdsa.PrivateKey) (*types.Transaction, error) {
	if pk == nil {
		p, err := m.StringToPrivateKey(privateKeyStr)
		if err != nil {
			return nil, err
		}
		pk = p
	}
	//1是主网
	return types.SignTx(tx, types.NewEIP155Signer(m.chainID), pk)

}

func (m *myEthClient) StringToPrivateKey(privateKeyStr string) (*ecdsa.PrivateKey, error) {
	if len(privateKeyStr) >= 2 && privateKeyStr[0] == '0' && (privateKeyStr[1] == 'x' || privateKeyStr[1] == 'X') {
		privateKeyStr = privateKeyStr[2:]
	}
	privateKeyByte, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.ToECDSA(privateKeyByte)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (m *myEthClient) PrivateKeyToString(pk *ecdsa.PrivateKey) string {
	b := crypto.FromECDSA(pk)
	return hex.EncodeToString(b)

}

func (m *myEthClient) NewAccount() (*common.Address, *ecdsa.PrivateKey, string, error) {
	pk, e := crypto.GenerateKey()
	if e != nil {
		return nil, nil, "", e
	}
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	return &addr, pk, m.PrivateKeyToString(pk), nil

}

func (m *myEthClient) Transfer(to, contractAddr *common.Address, amount, gasPrice *big.Int, gasLimit uint64, pk *ecdsa.PrivateKey) (*types.Transaction, error) {
	from := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := m.C.PendingNonceAt(context.Background(), from)
	if err != nil {
		return nil, err
	}
	tx := &types.Transaction{}

	if contractAddr != nil {
		//合约转账
		balance := m.GetBalance(&from, contractAddr)
		if balance.Cmp(amount) < 0 {
			return nil, errors.New("热钱包余额不足")
		}
		tx = types.NewTransaction(nonce, *contractAddr, nil, gasLimit, gasPrice, bytes.Join([][]byte{common.Hex2Bytes(m.transferPre), common.LeftPadBytes(to.Bytes(), 32), common.LeftPadBytes(amount.Bytes(), 32)}, []byte("")))
	} else {
		//eth转账
		balance := m.GetBalance(&from, nil)
		if balance.Sub(balance, big.NewInt(0).Mul(gasPrice, big.NewInt(21000))).Cmp(amount) < 0 {
			return nil, errors.New("热钱包余额不足")
		}
		tx = types.NewTransaction(nonce, *to, amount, 21000, gasPrice, nil)
	}

	tx, err = types.SignTx(tx, types.NewEIP155Signer(m.chainID), pk)
	if err != nil {
		return nil, err
	}
	return tx, m.C.SendTransaction(context.Background(), tx)
}

//合约地址为空则返回eth余额
func (m *myEthClient) GetBalance(addr, contract *common.Address) *big.Int {
	if contract == nil {
		b, _ := m.C.BalanceAt(context.Background(), *addr, nil)
		return b
	}

	c := ethereum.CallMsg{}
	c.From = *addr
	c.To = contract
	c.Data = common.Hex2Bytes("70a08231000000000000000000000000" + addr.String()[2:])
	r, e := m.C.CallContract(context.Background(), c, nil)
	if e != nil {
		fmt.Println(e)
		return big.NewInt(0)
	}
	return big.NewInt(0).SetBytes(r)

}
