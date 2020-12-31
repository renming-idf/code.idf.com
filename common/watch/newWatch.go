package watch

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cast"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"
	"xdf/common/log"
	"xdf/model"
	"xdf/transformer"
)

var (
	accountPool          sync.Map
	ethClient            *myEthClient
	mainWalletPk         *ecdsa.PrivateKey
	wei                  = big.NewInt(100000000000000) //0.0001 eth
	FastGasPrice         *big.Int
	SafeGasPrice         *big.Int
	AvgGasPrice          *big.Int
	usdtGasLimit         uint64 = 80000
	CapitalTrendsRecords        = make(chan *model.CapitalTrendsRecord, 128)
	mainWalletTxPool            = sync.Map{}
	checkSign                   = make(chan struct{}, 64)
	usdtContractAddr            = common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7")
	// mtt
	//usdtContractAddr = common.HexToAddress("0x499cb9b9ca5f323427860f984dc95b003b3016f1")
	cashSweepChan = make(chan string, 128)
	ethThreshold  = big.NewInt(100000000000000000)
	usdtThreshold = big.NewInt(0)
)

func IsETHAddress(addressHex string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(addressHex)
}
func IsInAccountAddressPool(accountAddress string) bool {
	//address32 := common.LeftPadBytes(common.HexToAddress(accountAddress).Bytes(), 32)
	//hexutil.Encode(address32)
	_, ok := accountPool.Load(strings.ToLower(accountAddress))
	return ok
}

func NewAccount() (pks, addr string, e error) {
	pk, e := crypto.GenerateKey()
	if e != nil {
		return "", "", e
	}
	return common.Bytes2Hex(crypto.FromECDSA(pk)), crypto.PubkeyToAddress(pk.PublicKey).String(), nil

}

func InitPool(conf *transformer.Conf) {
	usdtThreshold, _ = big.NewInt(0).SetString("100000000", 10)
	ethClient = &myEthClient{}
	e := ethClient.NewClient(conf.Data.MyUrl)
	if e != nil {
		panic(e)
	}
	mainWalletPk, e = ethClient.StringToPrivateKey(conf.Data.MainWalletsPath)
	if e != nil {
		panic(e)
	}
	accountPool = sync.Map{}
	u := &model.User{}
	users := u.GetAllAccountAddress()
	for _, u := range users {
		InjectAccountAddressToPool(u)
	}

	go watchChain()
	go watchMainWalletTX()
	go monitorMainWalletTX()
	go recovery()
	go cashSweep()

}

func recovery() {
	//todo 灾难恢复，读取处理的最后一块，在watchChain里每收到一块插一条到数据库
	mcsr := model.CapitalTrendsRecord{}
	csrs := mcsr.GetAllHangUpAccountRecord()
	//csrs := model.CapitalTrendsRecord{}.GetAllHangUpAccountRecord()
	for i := range csrs {
		x := csrs[i]
		mainWalletTxPool.Store(common.HexToHash(x.Tx), &x)
	}
}

func cashSweep() {
	for addr := range cashSweepChan {
		ad := common.HexToAddress(addr)
		eb := ethClient.GetBalance(&ad, nil)
		ub := ethClient.GetBalance(&ad, &usdtContractAddr)
		if eb.Cmp(ethThreshold) < 0 && ub.Cmp(usdtThreshold) < 0 {
			continue
		}
		user := model.User{}.GetUserByAccountAddress(addr)
		pk, e := ethClient.StringToPrivateKey(user.KeyStorePath)
		if user.KeyStorePath == "" || e != nil {
			continue
		}
		ma := crypto.PubkeyToAddress(mainWalletPk.PublicKey)
		if ub.Cmp(usdtThreshold) > 0 {
			//usdt归集
			gas := big.NewInt(0).Mul(FastGasPrice, big.NewInt(80000))
			ctr := &model.CapitalTrendsRecord{}
			if eb.Cmp(gas) < 0 {
				//手续费不足，额外给2.4w个单位，防止手续费到账后价格变高
				amount := big.NewInt(0).Mul(FastGasPrice, big.NewInt(80000+24000))
				nctr, e := ctr.CreateCapitalConvergenceRecord(user.ID, 1, 2, 3, 0, 0, 0, cast.ToFloat64(amount.String()), 0, "向子账号转ETH手续费", ma.String(), addr, "", 0)
				if e != nil {
					log.Error(e)
					continue
				}
				//塞到CapitalTrendsRecords里由通道统一往出转，防止nonce重复
				CapitalTrendsRecords <- &nctr
				continue
			}
			//手续费足够
			tx, err := ethClient.Transfer(&ma, &usdtContractAddr, ub, FastGasPrice, 80000, pk)
			if err != nil {
				log.Error(err)
				continue
			}
			nctr, e := ctr.CreateCapitalConvergenceRecord(user.ID, 5, 2, 3, 0, tx.Nonce(), 0, cast.ToFloat64(ub.Div(ub, big.NewInt(100)).String()), 0, "向主账号转USDT", addr, ma.String(), tx.Hash().String(), 0)
			if e != nil {
				log.Error(e)
				continue
			}
			mainWalletTxPool.Store(tx.Hash(), &nctr)
			//归集了u就不归集eth了，避免usdt转完下面的转账金额不够
			continue
		}

		if eb.Cmp(ethThreshold) > 0 {
			//eth归集
			gp := *FastGasPrice
			gas := big.NewInt(0).Mul(&gp, big.NewInt(21000))
			tx, err := ethClient.Transfer(&ma, nil, eb.Sub(eb, gas).Sub(eb, big.NewInt(10)), &gp, 21000, pk)
			if err != nil {
				log.Error(err)
				continue
			}
			ctr := &model.CapitalTrendsRecord{}
			nctr, e := ctr.CreateCapitalConvergenceRecord(user.ID, 5, 2, 3, 0, tx.Nonce(), 0, cast.ToFloat64(eb.Div(eb, wei).String()), 0, "向主账号转ETH", addr, ma.String(), tx.Hash().String(), 0)
			if e != nil {
				log.Error(e)
				continue
			}
			mainWalletTxPool.Store(tx.Hash(), &nctr)
		}

	}
}

func monitorMainWalletTX() {
	for {
		<-checkSign
		mainWalletTxPool.Range(func(key, value interface{}) bool {
			tx, _ := key.(common.Hash)
			_, pending, e := ethClient.C.TransactionByHash(context.Background(), tx)
			if e != nil || pending {
				return true
			}
			defer mainWalletTxPool.Delete(key)
			ctr, _ := value.(*model.CapitalTrendsRecord)
			e = ctr.UpdateCapitalConvergenceRecordStatus(ctr.ID, 0, 1)
			if e != nil {
				log.Error(e)
			}
			//switch ctr.Type {
			////0 资金汇集转入erc代币  1转出ETH当作手续费 2提币 3 第三方充值进账户 4提币到本平台账户 5资金归集eth
			//case 1:
			//	//手续费到账,进行归集
			//	cashSweepChan <- ctr.To
			//	break
			//	//case 2:
			//	//	if r.Status == 0 {
			//	//		//退回余额
			//	//		refund(ctr.ID)
			//	//	}
			//	//	break
			//}
			return true
		})
	}
}

func refund(ctrId uint) {
	ctr := &model.CapitalTrendsRecord{}
	err := ctr.UpdateCapitalConvergenceRecordStatus(ctrId, 0, 0)
	if err != nil {
		log.Error(err)
		return
	}
	var uw model.UserWallet
	//提币失败退款
	err = uw.RefundwMoney(ctrId)
	if err != nil {
		log.Error("提币失败退款错误,id为%d！%s", ctrId, err)
	}

}

func watchMainWalletTX() {
	for ctr := range CapitalTrendsRecords {
		if ctr.Type == 4 {
			uid, ok := accountPool.Load(strings.ToLower(ctr.To))
			if !ok {
				e := ctr.UpdateCapitalConvergenceRecordStatus(ctr.ID, 0, 0)
				if e != nil {
					log.Error(e)
				}
				log.Error("提币到内部账户失败，未找到" + ctr.To)
				continue
			}
			e := ctr.UpdateCapitalConvergenceRecordStatus(ctr.ID, 0, 1)
			if e != nil {
				log.Error(e)
			}
			e = model.UserWallet{}.RechargeCurrency(ctr.ID, cast.ToUint(uid), ctr.CurrencyTypeID, int64(ctr.Amount))
			if e != nil {
				log.Error(e)
			}
			continue
		}

		to := common.HexToAddress(ctr.To)
		e := errors.New("watchMainWalletTX error CurrencyTypeID")
		var tx *types.Transaction
		if ctr.CurrencyTypeID == 1 {
			//usdt
			a := big.NewInt(int64(ctr.Amount))
			a.Mul(a, big.NewInt(100))
			tx, e = ethClient.Transfer(&to, &usdtContractAddr, a, FastGasPrice, usdtGasLimit, mainWalletPk)
		} else if ctr.CurrencyTypeID == 2 {
			//eth
			a := big.NewInt(int64(ctr.Amount))
			if ctr.Type != 1 {
				a = big.NewInt(0).Mul(wei, big.NewInt(int64(ctr.Amount)))
			}
			tx, e = ethClient.Transfer(&to, nil, a, FastGasPrice, 21000, mainWalletPk)
		}
		if e != nil {
			log.Error(e)
			//退回余额
			if ctr.Type == 2 {
				refund(ctr.ID)
			}
			continue
		}
		if _, e = ctr.UpdateCapitalTrendsRecordInfo(ctr.ID, tx.Hash().String(), 0, 0, 0); e != nil {
			log.Error(e)
		}

		mainWalletTxPool.Store(tx.Hash(), ctr)
	}
}

func watchChain() {
	c := make(chan *types.Header, 256)
	_, e := ethClient.C.SubscribeNewHead(context.Background(), c)
	if e != nil {
		panic(e)
	}
	for h := range c {

		if !h.EmptyReceipts() {
			go handle_ETH_transfer(h.Number)
			go handle_USDT_transfer(h.Number)
			go func() {
				checkSign <- struct{}{}
			}()
		}
	}

}

func InjectAccountAddressToPool(u model.User) {
	accountPool.Store(strings.ToLower(u.AccountAddress), u.ID)
}

func handle_ETH_transfer(headerNum *big.Int) {
	x := 0
	var block *types.Block
	var e error
	for x < 10 {
		block, e = ethClient.C.BlockByNumber(context.Background(), headerNum)
		if e != nil {
			x++
			time.Sleep(time.Duration(5*x) * time.Second)
			continue
		}
		break
	}
	if e != nil {
		log.Error(e)
		return
	}
	wg := sync.WaitGroup{}
	for i := range block.Transactions() {
		wg.Add(1)
		go func(tx *types.Transaction) {
			defer wg.Done()
			if tx.To() == nil || tx.Value() == nil || tx.Value().Int64() == 0 {
				return
			}
			uid, ok := accountPool.Load(strings.ToLower(tx.To().String()))
			if !ok {
				return
			}
			//r, e := ethClient.C.TransactionReceipt(context.Background(), tx.Hash())
			//if e != nil || r.Status == 0 {
			//	return
			//}
			go func() {
				cashSweepChan <- tx.To().String()
			}()

			ctr, _, e := model.CapitalTrendsRecord{}.RechargeCurrency(cast.ToUint(uid), 2, "", tx.To().String(), tx.Hash().String(), headerNum.Int64(), cast.ToFloat64(tx.Value().Int64()/wei.Int64()))
			if e != nil {
				log.Error(e)
				return
			}

			e = model.UserWallet{}.RechargeCurrency(ctr.ID, cast.ToUint(uid), 2, int64(ctr.Amount))
			if e != nil {
				log.Error(e)
			}

		}(block.Transactions()[i])
	}
	wg.Wait()
}

func handle_USDT_transfer(headerNum *big.Int) {
	x := 0
	var logs []types.Log
	var e error
	fq := ethereum.FilterQuery{}
	fq.Addresses = []common.Address{usdtContractAddr}
	fq.FromBlock = headerNum
	fq.ToBlock = headerNum
	fq.Topics = [][]common.Hash{{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")}}
	for x < 10 {
		logs, e = ethClient.C.FilterLogs(context.Background(), fq)
		if e != nil {
			x++
			time.Sleep(time.Duration(5*x) * time.Second)
			continue
		}
		break
	}
	if e != nil {
		log.Error(e)
		return
	}
	if logs == nil || len(logs) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	for i := range logs {
		wg.Add(1)
		go func(l types.Log) {
			defer wg.Done()
			amount := big.NewInt(0)
			if amount.SetBytes(l.Data).Int64() == 0 || len(l.Topics) != 3 {
				return
			}
			toAddr := common.HexToAddress(l.Topics[2].String())
			uid, ok := accountPool.Load(strings.ToLower(toAddr.String()))
			if !ok {
				return
			}
			//log.Println("bingo ", toAddr.String())
			//r, e := ethClient.C.TransactionReceipt(context.Background(), l.TxHash)
			//if e != nil || r.Status == 0 {
			//	log.Error(e, "==>", l.TxHash.String())
			//	return
			//}
			go func() {
				cashSweepChan <- toAddr.String()
			}()
			//充值进入
			a := big.NewInt(0).SetBytes(l.Data)
			ctr, _, e := model.CapitalTrendsRecord{}.RechargeCurrency(cast.ToUint(uid), 1, common.HexToAddress(l.Topics[1].String()).String(), toAddr.String(), l.TxHash.String(), headerNum.Int64(), float64(a.Div(a, big.NewInt(100)).Int64()))
			//ctr, _, e := model.CapitalTrendsRecord{}.RechargeCurrency(cast.ToUint(uid), 1, common.HexToAddress(l.Topics[1].String()).String(), toAddr.String(), l.TxHash.String(), headerNum.Int64(), float64(a.Div(a, big.NewInt(100000000000000)).Int64()))
			if e != nil {
				log.Error(e)
				return
			}

			e = model.UserWallet{}.RechargeCurrency(ctr.ID, cast.ToUint(uid), 1, int64(ctr.Amount))
			if e != nil {
				log.Error(e)
			}
		}(logs[i])
	}
	wg.Wait()
}
