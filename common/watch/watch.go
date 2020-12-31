package watch

//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	"github.com/ethereum/go-ethereum/accounts/keystore"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/ethclient"
//	"github.com/shopspring/decimal"
//	"github.com/spf13/cast"
//	"golang.org/x/crypto/sha3"
//	"io/ioutil"
//	"math/big"
//	"xdf/common/log"
//	"xdf/common/token"
//	"xdf/common/tools"
//	"xdf/model"
//	"os"
//	"regexp"
//	"strings"
//	"sync"
//	"time"
//)
//
////存储转账信息
//type TransferInfo struct {
//	From        string
//	To          string
//	Tx          string
//	Amount      float64
//	Nonce       uint64  //转账时的随机数  可用于加速订单
//	GasLimit    uint64  //GasLimit
//	GasPrice    float64 //GasPrice单位 Gwei
//	BlockNumber int64
//}
//
//type conf struct {
//	MainWalletsPath       string
//	MainWalletsAccountHex string
//	MainWalletsAccount    accounts.Account
//	MainWalletsKeyStore   *keystore.KeyStore
//	Password              string
//	KeyStoreDir           string
//	TmpKeyStoreDir        string
//}
//
//type transferPools struct {
//	AccountAddressPool  sync.Map        // 钱包地址池
//	ContractAddressPool map[string]uint // 合约地址地址池
//	TxPool              sync.Map        // 订单号池  主要指 转出ETH 和提笔USDT
//}
//type txInfo struct {
//	model.CapitalTrendsRecord
//	Path string
//}
//
//var (
//	//logType              common.Hash
//	pools           transferPools //转账信息池
//	Client          *ethclient.Client
//	SubscribeClient *ethclient.Client
//	Conf            conf //主钱包信息
//	USDTGasLimit    int64
//	Wei             = big.NewInt(100000000000000) //0.0001 eth
//	//暂时只能写死
//	UsdtWei              = big.NewInt(100)        //0.0001 usdt
//	Gwei                 = big.NewInt(1000000000) //1Gwei
//	FastGasPrice         = big.NewInt(0)
//	SaveGasPrice         = big.NewInt(0)
//	CapitalTrendsRecords chan model.CapitalTrendsRecord
//	NewBlockNumber       int64
//	mu                   sync.Mutex
//)
//
////初始化TPools 只初始化了 AccountAddressPool ContractAddressPool 剩下来个在启动监听前初始化
////func InitPool(c *transformer.Conf) {
////	//USDTGasReference.GasLimit = 50000
////	////gwei := big.NewInt(10000)
////	//USDTGasReference.GasPrice = *big.NewInt(1).Mul(Gwei, big.NewInt(1))
////	//USDTGasReference.GasUsed = 30000
////	//allAccountAddress:=[2]string{"0x26C9a10f147b4c0BBd1ab39bE6755B3D30C8dDcb","0x34C5170ac2805e8D6eFC8F8d2Ce9e5629A4E25dC"}
////	//初始化钱包地址池
////	Conf.Password = c.Data.SecretPassphrase
////	Conf.MainWalletsPath = c.Data.MainWalletsPath
////	Conf.KeyStoreDir = c.Data.KeyStoreDir
////	Conf.TmpKeyStoreDir = "./tmp"
////	pools.ContractAddressPool = make(map[string]uint)
////	CapitalTrendsRecords = make(chan model.CapitalTrendsRecord, 60)
////	//获取主钱包地址
////	ks := keystore.NewKeyStore(Conf.TmpKeyStoreDir, keystore.LightScryptN, keystore.LightScryptP)
////	jsonBytes, err := ioutil.ReadFile(Conf.MainWalletsPath)
////	if err != nil {
////		panic(err)
////	}
////	account, err := ks.Import(jsonBytes, Conf.Password, Conf.Password)
////	if err != nil {
////		panic(err)
////	}
////	err = ks.Unlock(account, Conf.Password)
////	if err != nil {
////		panic(err)
////	}
////	//删除额外创建的密匙文件
////	if err := os.Remove(account.URL.Path); err != nil {
////		panic(err)
////	}
////	Conf.MainWalletsAccountHex = account.Address.Hex()
////	Conf.MainWalletsAccount = account
////	Conf.MainWalletsKeyStore = ks
////	au := model.User{}
////	users := au.GetAllAccountAddress()
////	for _, u := range users {
////		InjectAccountAddressToPool(u)
////	}
////	//初始化合约地址地址池
////	ct := model.CurrencyType{}
////	allContractAddress := ct.GetCurrencyTypeMainNet()
////	for _, contractAddress := range allContractAddress {
////		InjectContractAddressPool(contractAddress)
////	}
////	cidInterface, _ := pools.ContractAddressPool["ETH"]
////	cid := cast.ToUint(cidInterface)
////	if cid == 0 {
////		panic("没有以太币数据！")
////	}
////	SubscribeClient, err = ethclient.Dial(c.Data.Url)
////	if err != nil {
////		panic(err)
////	}
////	Client, err = ethclient.Dial(c.Data.MyUrl)
////	if err != nil {
////		panic(err)
////	}
////	//初始化数据之后监听USDT 监听的同时启动灾难恢复
////	WatchUSDTStart()
////	go WatchTx()
////	go HandleTransferCapitalTrendsRecord()
////}
//
//func InjectContractAddressPool(c model.CurrencyType) {
//	pools.ContractAddressPool[c.ContractAddress] = c.ID
//}
//func IsInAccountAddressPool(accountAddress string) bool {
//	//address32 := common.LeftPadBytes(common.HexToAddress(accountAddress).Bytes(), 32)
//	//hexutil.Encode(address32)
//	_, ok := pools.AccountAddressPool.Load(strings.ToLower(accountAddress))
//	return ok
//}
//func GetETHCurrencyType() uint {
//	return pools.ContractAddressPool["ETH"]
//}
//
//func InjectTxPool(csr model.CapitalTrendsRecord) {
//	_, ok := pools.TxPool.Load(csr.Tx)
//	if !ok {
//		ti := &txInfo{}
//		u := model.User{}
//		u.GetUserById(csr.UserID)
//		if u.ID == 0 {
//			return
//		}
//		ti.CapitalTrendsRecord = csr
//		ti.Path = u.KeyStorePath
//		pools.TxPool.Store(csr.Tx, ti)
//	}
//}
//func DeleteTx(tx string) {
//	pools.TxPool.Delete(tx)
//}
//
////按队列 排队提交订单
//func HandleTransferCapitalTrendsRecord() {
//	for {
//		nowCtr := <-CapitalTrendsRecords
//		var transferInfo *TransferInfo
//		var err error
//		//提币
//		if nowCtr.Type == model.WithdrawMoney || nowCtr.Type == model.WithdrawMoneyToOwn {
//
//			if nowCtr.CurrencyTypeID == GetETHCurrencyType() {
//				nowCtr.TokenAddressHex = ""
//				eth := tools.FloatToWei(Wei, nowCtr.Amount)
//				//不需要知道gasprice 则设置为0
//				transferInfo, err = TransferETH(Client, Conf.MainWalletsPath, nowCtr.To, big.NewInt(0), eth)
//			} else {
//
//				ercAmount := tools.FloatToWei(UsdtWei, nowCtr.Amount)
//				transferInfo, err = TransferERC20(Client, nowCtr.TokenAddressHex,
//					Conf.MainWalletsPath,
//					nowCtr.To, uint64(USDTGasLimit), FastGasPrice, ercAmount)
//			}
//			if err != nil {
//				log.Error("转账错误", nowCtr.ID, err)
//				//修改订单状态
//				err := nowCtr.UpdateCapitalConvergenceRecordStatus(nowCtr.ID, 0, 0)
//				if err != nil {
//					log.Error(err)
//					return
//				}
//				var uw model.UserWallet
//				//提币失败退款
//				err = uw.RefundwMoney(nowCtr.ID)
//				if err != nil {
//					log.Error("提币失败退款错误,id为%d！%s", nowCtr.ID, err)
//				}
//				return
//			}
//
//		} else if nowCtr.Type == model.TransferOutETHServiceCharge {
//			//转出eth手续费
//			eth := tools.FloatToWei(Wei, nowCtr.Amount)
//			//不需要知道gasprice 则设置为0
//			if eth.Int64() == 0 {
//				log.Error("计算eth出现问题！")
//				return
//			}
//			transferInfo, err = TransferETH(Client, Conf.MainWalletsPath, nowCtr.To, big.NewInt(0), eth)
//			if err != nil {
//				log.Error(fmt.Errorf("订单ID为：%d,error为%s", nowCtr.ID, err))
//				err := nowCtr.UpdateCapitalConvergenceRecordStatus(nowCtr.ID, 0, 0)
//				if err != nil {
//					log.Error(err)
//					return
//				}
//				return
//			}
//		} else {
//			log.Println("不处理该种订单")
//			return
//		}
//		//修改 tx和noce
//		updateCsr, err := nowCtr.UpdateCapitalTrendsRecordInfo(nowCtr.ID, transferInfo.Tx, transferInfo.Nonce, transferInfo.GasLimit, transferInfo.GasPrice)
//		if err != nil {
//			log.Error("修改提币订单状态出现问题,id为%d！%s", nowCtr.ID, err)
//			return
//		}
//		//插入订单池
//		InjectTxPool(updateCsr)
//	}
//}
//
////灾难恢复以及监听启动
//func WatchUSDTStart() {
//	mnar := model.CapitalTrendsRecord{}
//	//获取当前数据库种订单最大块数作为恢复初始块
//	start := mnar.GetMaxBlockNumber()
//	header, err := Client.HeaderByNumber(context.Background(), nil)
//	if err != nil {
//		panic(err)
//	}
//	//查询最高快数
//	NewBlockNumber = header.Number.Int64()
//
//	// 从现在开始对第三方USDT转账进行监听（用户转账到钱包的USDT）
//	go watchNewHead()
//	//go WatchTripartiteTransfer(usdtAddress)
//	if start == 0 {
//		return
//	}
//	//如果数据库种不存在数据，则不需要恢复
//	go DataRecovery(start, NewBlockNumber)
//}
//func DataRecovery(start, end int64) {
//	log.Println("正在进行灾难恢复！！！！！，开始：", start, " ,结束：", end)
//
//	for x := start; x <= end; x++ {
//		HandleTransactions(x)
//	}
//	log.Println("灾难恢复完成！！！！！")
//}
//func GetGasLimit(logs *[]types.Log) {
//	for i := range *logs {
//		if len((*logs)[i].Data) < 3 || len((*logs)[i].Topics) < 3 {
//			continue
//		}
//		tx, _, err := Client.TransactionByHash(context.Background(), (*logs)[i].TxHash)
//		//防止更新过大的limit
//		if err == nil && len(tx.Data()) < 70 {
//			receipt, err := Client.TransactionReceipt(context.Background(), tx.Hash())
//			if err != nil {
//				continue
//			}
//			gasUsed := receipt.GasUsed
//			usdeRate := float64(gasUsed) / float64(tx.Gas())
//			if usdeRate < 0.55 || usdeRate > 0.9 {
//				continue
//			}
//
//			USDTGasLimit = int64(tx.Gas())
//
//			fmt.Println("gas过期，更新gas")
//			fmt.Println("使用率", usdeRate, "gasLimit", USDTGasLimit)
//			return
//		}
//
//	}
//
//	fmt.Println("无合适log")
//}
//func HandleTransactions(blockNumber int64) {
//	go handleUsdtTransfer(blockNumber)
//
//	//以下处理eth转账
//	retry := 3
//	var block *types.Block
//	for retry > 0 {
//		b, e := Client.BlockByNumber(context.Background(), big.NewInt(blockNumber))
//		if e != nil {
//			retry--
//			time.Sleep(5 * time.Second)
//			continue
//		}
//		block = b
//		break
//	}
//
//	if block == nil || len(block.Transactions()) == 0 {
//		return
//	}
//
//	wg := sync.WaitGroup{}
//	wg.Add(len(block.Transactions()))
//	for i := range block.Transactions() {
//		go func(tx *types.Transaction) {
//			defer wg.Done()
//			//var data []byte
//			//var cid uint
//
//			if tx.To() == nil || tx.Value().Int64() == 0 {
//				//合约创建 不需要管
//				return
//			}
//			uidInterface, ok := pools.AccountAddressPool.Load(strings.ToLower(tx.To().String()))
//			if !ok {
//				return
//			}
//
//			//eth转账
//			msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			amount := tx.Value()
//			cidInterface, _ := pools.ContractAddressPool["ETH"]
//			//处理eth代币
//			var txIn = TransferInfo{
//				From:        msg.From().String(),
//				To:          tx.To().String(),
//				Tx:          tx.Hash().String(),
//				Amount:      tools.BigIntToFloat(amount, Wei),
//				Nonce:       0,
//				BlockNumber: blockNumber,
//			}
//			fmt.Println("eth ", txIn.Tx, "   to:", txIn.To)
//			err = HandleTripartiteTransfers(txIn, cast.ToUint(uidInterface), cast.ToUint(cidInterface))
//			if err != nil {
//				log.Error(err)
//			}
//		}(block.Transactions()[i])
//
//	}
//	wg.Wait()
//}
//
//func handleUsdtTransfer(blockNumber int64) {
//
//	fq := ethereum.FilterQuery{}
//	usdtAdd := common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7")
//	fq.Addresses = []common.Address{usdtAdd}
//	fq.Topics = [][]common.Hash{{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")}}
//	fq.FromBlock = big.NewInt(blockNumber)
//	fq.ToBlock = big.NewInt(blockNumber)
//	logs, err := Client.FilterLogs(context.Background(), fq)
//	if err != nil {
//		fmt.Println("handleUsdtTransfer ERR:", err)
//		return
//	}
//	le := len(logs)
//	if le < 1 {
//		fmt.Println("handleUsdtTransfer No Logs in block:", blockNumber)
//		return
//	}
//
//	wg := sync.WaitGroup{}
//	wg.Add(le)
//
//	go GetGasLimit(&logs)
//	for i := range logs {
//		go func(l types.Log) {
//			defer wg.Done()
//			if len(l.Data) < 3 || len(l.Topics) < 3 {
//				return
//			}
//
//			from := common.HexToAddress(l.Topics[1].String()).String()
//
//			to := common.HexToAddress(l.Topics[2].String()).String()
//
//			uidInterface, ok := pools.AccountAddressPool.Load(strings.ToLower(to))
//			if !ok {
//				return
//			}
//			hh := common.Bytes2Hex(l.Data)[2:]
//			amount, flag := big.NewInt(1).SetString(hh, 16)
//			if !flag {
//				log.Errorf("金额转化失败,%s", l.TxHash.String())
//				return
//			}
//			var txIn = TransferInfo{
//				From:        from,
//				To:          to,
//				Tx:          l.TxHash.String(),
//				Amount:      tools.BigIntToFloat(amount, UsdtWei),
//				Nonce:       0,
//				BlockNumber: blockNumber,
//			}
//			fmt.Println("到账usdt : ", txIn.Tx, "   to:", to)
//			err = HandleTripartiteTransfers(txIn, cast.ToUint(uidInterface), 1)
//			if err != nil {
//				log.Error(err)
//			}
//		}(logs[i])
//	}
//	wg.Wait()
//}
//func watchNewHead() {
//	headers := make(chan *types.Header)
//	sub, err := SubscribeClient.SubscribeNewHead(context.Background(), headers)
//	if err != nil {
//		log.Error(err)
//	}
//	for {
//		select {
//		case err := <-sub.Err():
//			log.Error(err)
//		case header := <-headers:
//			//取当前快的前一个快
//			NewBlockNumber = header.Number.Int64() - 1
//
//			fmt.Printf("获取到新块号为：%d 中的数据\n", NewBlockNumber)
//			HandleTransactions(NewBlockNumber)
//		}
//	}
//}
//
////func WatchTripartiteTransfer(contractAddress common.Address) {
////	logTransferSig := []byte("Transfer(address,address,uint256)")
////	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
////	var ct model.CurrencyType
////	contractAddressHex := contractAddress.String()
////	cid := ct.GetCurrencyTypeByContractAddress(contractAddressHex).ID
////	if cid == 0 {
////		panic("不存在USDT数据")
////	}
////	var query ethereum.FilterQuery
////	query = ethereum.FilterQuery{
////		Addresses: []common.Address{contractAddress},
////		Topics:    [][]common.Hash{{logTransferSigHash}},
////	}
////	logs := make(chan types.Log)
////	sub, err := Client.SubscribeFilterLogs(context.Background(), query, logs)
////	if err != nil {
////		log.Error(err)
////		panic(err)
////	}
////	log.Println("TripartiteTransfer开始监听")
////	for {
////		select {
////		case err := <-sub.Err():
////			log.Error(err)
////		case vLog := <-logs:
////			//发送方再钱包池中 接收人为主钱包
////			to := vLog.Topics[2].String()
////			//从map中查看是否为存在该接收人 用户第三方转账到钱包
////			uidInterface, ok := pools.AccountAddressPool.Load(to)
////			NewBlockNumber = int64(vLog.BlockNumber)
////			var flag bool
////			if ok {
////				uid := cast.ToUint(uidInterface)
////				if uid == 0 {
////					log.Error("uid转换错误！")
////					continue
////				}
////				switch vLog.Topics[0] {
////				case logTransferSigHash:
////					err = HandleTripartiteTransfers(vLog, uid, cid)
////					if err != nil {
////						log.Error(err)
////					}
////				}
////			} else {
////				from := vLog.Topics[1].String()
////				if common.HexToAddress(from).String() == Conf.MainWalletsAccountHex || common.HexToAddress(to).String() == Conf.MainWalletsAccountHex {
////					flag = true
////				}
////			}
////			//当flag 为true 则说明该订单是我们的不用于更新 gas 和gaslimit
////			fmt.Println(flag)
////			if !flag {
////				gasLimit, gasPrice, gasUsed, err := GetGasLimit(vLog.TxHash)
////				if err != nil {
////					log.Errorf("GetGasLimit 获取失败,%s", err)
////					break
////				}
////				usdeRate := float64(gasUsed / gasLimit)
////				if usdeRate < 0.5 && usdeRate > 0.8 {
////					break
////				}
////				var usdtGas USDTGas
////				usdtGas.GasPrice = *gasPrice
////				usdtGas.GasLimit = gasLimit
////				usdtGas.GasUsed = gasUsed
////				USDTGasReference = usdtGas
////			}
////			//{
////			// "address":"0x499cb9b9ca5f323427860f984dc95b003b3016f1",  货币地址
////			// "topics":[
////			//			// "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
////			//			// "0x000000000000000000000000d9369bc44df691008b97d126d60f1e20b91175da", 发送方
////			//			// "0x00000000000000000000000026c9a10f147b4c0bbd1ab39be6755b3d30c8ddcb"],  接收方
////			// "data":"0x0000000000000000000000000000000000000000000000008ac7230489e80000",
////			// "blockNumber":"0x7cf92f",
////			// "transactionHash":"0xd9a8fe52499dbf211fe7c9f4c886273b551663a722fad23e9f08d932bcad3808",  订单号 交易哈希
////			// "transactionIndex":"0x0",
////			// "blockHash":"0x5c34bde87b244d6bf0f0a4186fa72bb77259b7b0b7b8d9b728e5692a5081b03d",
////			// "logIndex":"0x0",
////			// "removed": false
////			//}
////		}
////	}
////}
//
//func HandleTripartiteTransfers(vLog TransferInfo, uid, currencyTypeID uint) error {
//	tx := vLog.Tx //订单号
//	from := vLog.From
//	to := vLog.To
//	blockNumber := vLog.BlockNumber
//	amount := vLog.Amount
//	var ctr model.CapitalTrendsRecord
//	//充币处理
//	newCtr, isSkip, err := ctr.RechargeCurrency(uid, currencyTypeID, from, to, tx, blockNumber, amount)
//	if err != nil {
//		fmt.Println(err)
//		if isSkip {
//			return nil
//		}
//		return err
//	}
//	//加入订单池 等待确认
//	InjectTxPool(newCtr)
//	return nil
//}
//
////监控发出币的到账情况到账了 t=1转出ETH t=2提币
//func WatchTx() {
//	uw := &model.UserWallet{}
//	mcsr := model.CapitalTrendsRecord{}
//	//获取所以挂起和待确定状态的订单
//	csrs := mcsr.GetAllHangUpAccountRecord()
//	gathering := make(map[uint]uint)
//
//	//插入eth转账信息
//	for i, c := range csrs {
//		//通常这种情况为提币时崩溃，没有转出usdt
//		if c.Tx == "" && c.Status == 2 {
//			//没有转出则退还usdt
//			err := uw.RefundwMoney(c.ID)
//			if err != nil {
//				log.Error(err)
//			}
//		} else {
//			InjectTxPool(csrs[i])
//		}
//	}
//	expirationTime := NewBlockNumber + 5
//	for {
//		//if len()
//		if NewBlockNumber < expirationTime {
//			continue
//		}
//		expirationTime = NewBlockNumber + 6
//		pools.TxPool.Range(func(k, vi interface{}) bool {
//			v := vi.(*txInfo)
//			//处理数据
//			tr, isSuccess, err := TransactionIsSuccess(Client, common.HexToHash(v.Tx))
//			//订单失败
//			if err != nil {
//				log.Errorf("订单%s,出现下面问题:%s,当前最新块为：%d", v.Tx, err, NewBlockNumber)
//				//资金归集失败，则删除自己归集状态
//				if flag := checkType(v.Type); flag {
//					delete(gathering, v.UserID)
//				}
//				//提币失败退款
//				if v.Type == model.WithdrawMoney || v.Type == model.WithdrawMoneyToOwn {
//					err := uw.RefundwMoney(v.ID)
//					if err != nil {
//						log.Error(err)
//					}
//				}
//				err := v.UpdateCapitalConvergenceRecordStatus(v.ID, 0, 0)
//				if err != nil {
//					log.Error(err)
//				}
//				pools.TxPool.Delete(k)
//				return true
//			}
//			//订单为异常订单，或被删除了，从txpool中删除记录
//			if time.Now().Sub(v.CreatedAt) > time.Hour*24 {
//				//资金归集失败，则删除自己归集状态
//				if flag := checkType(v.Type); flag {
//					delete(gathering, v.UserID)
//				}
//				//订单异常 超过1天没完成
//				err := v.UpdateCapitalConvergenceRecordStatus(v.ID, 0, 4)
//				if err != nil {
//					log.Error(err)
//				}
//				pools.TxPool.Delete(k)
//				return true
//			}
//			//订单没成功直接返回
//			if !isSuccess {
//				return true
//			}
//			//给订单赋值BlockNumber 并且把状态修改为待确定
//			if v.BlockNumber == 0 && tr.BlockNumber.Int64() > 0 {
//				v.BlockNumber = tr.BlockNumber.Int64()
//				v.Status = 3
//				err := v.UpdateCapitalConvergenceRecordStatus(v.ID, tr.BlockNumber.Int64(), 3)
//				if err != nil {
//					log.Error(err)
//					return true
//				}
//				pools.TxPool.Store(k, v)
//			}
//			if NewBlockNumber-v.BlockNumber+1 < model.NeedConfirmationsNumber {
//				return true
//			}
//			//达到确认次数
//			switch v.Type {
//			case model.TransferOutETHServiceCharge:
//				//转账的eth到账 处理订单订单并且发送USDT v.Path转账者的keystorePath
//				//model.TransferOutETHServiceCharge 时候比较特别 是v.CurrencyTypeID是需要转账的代币
//				log.Println("手续费到账准备转usdt！")
//				err = HandleETHTransfer(v.CapitalTrendsRecord, tr, v.CurrencyTypeID, Conf.MainWalletsAccountHex, v.Path)
//				if err != nil {
//					//解除资金归集状态
//					delete(gathering, v.UserID)
//					log.Error(err)
//					err := v.UpdateCapitalConvergenceRecordStatus(v.ID, tr.BlockNumber.Int64(), 0)
//					if err != nil {
//						log.Error(err)
//					}
//					pools.TxPool.Delete(k)
//					return true
//				}
//			case model.WithdrawMoneyToOwn, model.TransferReceived:
//				//普通充币
//				uid := v.UserID
//				//提币到内部账户
//				if v.Type == model.WithdrawMoneyToOwn {
//					//当是转账给内部账户时候 uid需要获取的是to  地址所对应的uid
//					//to := hexutil.Encode(common.LeftPadBytes(common.HexToAddress(v.To).Bytes(), 32))
//					uidInterface, ok := pools.AccountAddressPool.Load(strings.ToLower(v.To))
//					uid = cast.ToUint(uidInterface)
//					if !ok || uid == 0 {
//						log.Errorf("该用户不是内部账户，AccountAddress：%s,tx:%s", v.To, v.Tx)
//						err := v.UpdateCapitalConvergenceRecordStatus(v.ID, tr.BlockNumber.Int64(), 0)
//						if err != nil {
//							log.Error(err)
//						}
//						pools.TxPool.Delete(k)
//						return true
//					}
//				}
//				err = uw.RechargeCurrency(v.ID, uid, v.CurrencyTypeID, int64(v.Amount))
//				if err != nil {
//					log.Error(err)
//					err := v.UpdateCapitalConvergenceRecordStatus(v.ID, tr.BlockNumber.Int64(), 0)
//					if err != nil {
//						log.Error(err)
//					}
//					pools.TxPool.Delete(k)
//					return true
//				}
//				log.Println("转账或提币到内部账号成功")
//				//在TransferCondition表中可以找到相应的数据则查询余额，余额足够则进行处理
//				//查询钱包余额 达到要求则调用资金合拢函数 HandleAmountSatisfiedUserWallet
//				//资金足够进行资金归集，添加到 gathering map中防止多次转eth
//				_, ok := gathering[uid]
//				if !ok {
//					if USDTGasLimit > 1 {
//						ct := model.CurrencyType{}.GetCurrencyTypeByID(v.CurrencyTypeID)
//						if v.CurrencyTypeID == GetETHCurrencyType() {
//							//判断到账的是否为eth
//							ct.ContractAddress = ""
//						}
//						balance, err := QueryBalance(Client, v.To, ct.ContractAddress)
//						if err != nil {
//							log.Errorf("获取余额失败，%s", err)
//						} else {
//							wei := Wei
//							if ct.ContractAddress != "" {
//								wei = UsdtWei
//							}
//							balance = balance.Div(balance, wei)
//							itc := &model.IniTransferCondition{}
//							itc = itc.GetTransferConditionByCurrencyTypeID(ct.ID)
//							//不存在则不转
//							if itc.ID == 0 {
//								log.Error("没有添加usdt资金归集条件！")
//							} else {
//								if balance.Int64() >= int64(itc.Threshold) {
//									uwi, err := uw.GetUserWalletInfoByID(uid, itc.CurrencyTypeID)
//									if err != nil || uwi.UserID == 0 {
//										log.Error(err, "或找不到用户钱包")
//									} else {
//										fmt.Printf("id:%d的用户代币足够，正在进行资金据集", uid)
//										gathering[uid] = 1
//										err = HandleAmountSatisfiedUserWallet(uwi, ct.ContractAddress, uwi.CurrencyTypeID)
//										if err != nil {
//											log.Error(err)
//										}
//									}
//								}
//							}
//						}
//					}
//				}
//			case model.RecoveryERC, model.WithdrawMoney, model.RecoveryETH:
//				//资金归集成功，删除记录
//				if flag := checkType(v.Type); flag {
//					delete(gathering, v.UserID)
//				}
//				//资金归集转账usdt 提笔
//				err := v.UpdateCapitalConvergenceRecordStatus(v.ID, tr.BlockNumber.Int64(), 1)
//				if err != nil {
//					log.Error(err)
//					return true
//				}
//			default:
//				log.Errorf("该订单类型错误，tx:%s ，type:%d", v.Tx, v.Type)
//			}
//			pools.TxPool.Delete(k)
//			return true
//		})
//	}
//}
//
////如果是资金汇集转入USDT 或者 转账eth 失败则删除，只要资金汇集转入USDT 也要删除
//func checkType(t uint) bool {
//	if t == model.RecoveryERC || t == model.TransferOutETHServiceCharge || t == model.RecoveryETH {
//		//资金归集成功，删除记录
//		return true
//	}
//	return false
//}
//
////ETH手续费到账处理 准备代币归集 toAddress 主钱包地址
//func HandleETHTransfer(ctr model.CapitalTrendsRecord, tr *types.Receipt, currencyType uint, toAddress, keystorePath string) error {
//	mct := model.CurrencyType{}
//	c := model.CapitalTrendsRecord{}
//	//ETH转账成功
//	//修改转账ETH记录状态
//	err := c.UpdateCapitalConvergenceRecordStatus(ctr.ID, tr.BlockNumber.Int64(), 1)
//	if err != nil {
//		log.Error(err)
//		return err
//	}
//	mct = mct.GetCurrencyTypeByID(currencyType)
//	tokenAddress := mct.ContractAddress
//	if tokenAddress == "" {
//		log.Error("不存在该币种，TX：%s", ctr.Tx)
//		return nil
//	}
//	// 获取指定币种余额
//	balence, err := QueryBalance(Client, ctr.To, tokenAddress)
//	if err != nil {
//		log.Error(err)
//		return err
//	}
//	itc := &model.IniTransferCondition{}
//	itc = itc.GetTransferConditionByCurrencyTypeID(mct.ID)
//	if itc.ID == 0 {
//		log.Error("没有添加usdt资金归集条件！")
//		return nil
//	}
//	newbalance := big.NewInt(1).Div(balence, UsdtWei)
//	if newbalance.Int64() < itc.Threshold {
//		return nil
//	}
//	//获取eth余额
//	ETHBalance, err := QueryBalance(Client, ctr.To, "")
//	if err != nil {
//		log.Error(err)
//		return err
//	}
//
//	for USDTGasLimit < 1 {
//		time.Sleep(10 * time.Second)
//	}
//	//  gasLimit要计算 gasprice使用最新的
//	//gasPrice := USDTGasReference.GasPrice
//
//	//这里余额除以limit，结果大于savePrice就可以用结果当price，limit-1保证余额充足
//	curPrice := big.NewInt(1).Div(ETHBalance, big.NewInt(USDTGasLimit))
//	//钱足够我则使用最新的price和计算的limit进行转钱
//	//判断比最近一次所需要的gas小则不转钱
//	if curPrice.Int64() < SaveGasPrice.Int64() {
//		log.Error("钱包%s，uid=%d,余额低于安全单价，停止归集！", ctr.To, ctr.UserID)
//		//余额不足 放弃这次资金归集
//		return nil
//	}
//	//TODO 手续费不够的处理方式  第一种 暂时处理 等待 第二种再转钱 不太推荐
//	//转指定币种（USDT） 通过用户的userid查到文件位置
//	transferInfo, err := TransferERC20(Client, tokenAddress, Conf.KeyStoreDir+"/"+keystorePath, toAddress, uint64(USDTGasLimit)-1, curPrice, balence)
//	if err != nil {
//		log.Errorf("用户%d转账USDT时出现问题！%s", ctr.UserID, err)
//		//eth转账还是到了 只是资金归集不了不是失败了 所以为nil
//		return nil
//	}
//	//插入转账记录 t=0 转入usdt currencyTypeID 指定币种
//	newCsr, err := c.CreateCapitalConvergenceRecord(ctr.UserID, model.RecoveryERC, ctr.CurrencyTypeID, 2,
//		0, transferInfo.Nonce, transferInfo.GasLimit, transferInfo.Amount, transferInfo.GasPrice,
//		"ETH手续费到账,向主账号转账erc20代币", transferInfo.From,
//		transferInfo.To, transferInfo.Tx, 0)
//	if err != nil {
//		log.Errorf("插入记录出现问题%s，订单号为%s", err, transferInfo.Tx)
//		return err
//	}
//	//向订单池插入数据
//	InjectTxPool(newCsr)
//	return nil
//}
//
////查询余额 tokenAddressHex  合约地址为空 查询ETH余额 否则查询相关代币
//func QueryBalance(client *ethclient.Client, AccountAddressHex, tokenAddressHex string) (balance *big.Int, err error) {
//	//账户地址
//	address := common.HexToAddress(AccountAddressHex)
//	if tokenAddressHex == "" {
//		balance, err = client.BalanceAt(context.Background(), address, nil)
//		if err != nil {
//			return balance, err
//		}
//		//fmt.Println(balance)
//	} else {
//		//合约地址
//		tokenAddress := common.HexToAddress(tokenAddressHex)
//		instance, err := token.NewToken(tokenAddress, client)
//		if err != nil {
//			return balance, err
//		}
//		balance, err = instance.BalanceOf(&bind.CallOpts{}, address)
//		if err != nil {
//			return balance, err
//		}
//
//	}
//	return balance, err
//}
//func TransactionIsSuccess(client *ethclient.Client, hash common.Hash) (tr *types.Receipt, ok bool, err error) {
//	tx, err := client.TransactionReceipt(context.Background(), hash)
//	if err != nil {
//		if err.Error() == "not found" {
//			return nil, false, nil
//		}
//		return nil, false, err
//	}
//	if tx.Status == 1 {
//		return tx, true, nil
//	} else {
//		return nil, false, errors.New("订单失败")
//	}
//}
//
////订单状态 ok 是否转账成功 panding 订单是否处于挂起状态
//func TransactionStatus(client *ethclient.Client, hash common.Hash) (tr *types.Receipt, ok, isPending bool, err error) {
//	_, pending, err := client.TransactionByHash(context.Background(), hash)
//	if err != nil {
//		return nil, false, false, err
//	}
//	if !pending {
//		tx, err := client.TransactionReceipt(context.Background(), hash)
//		if err != nil {
//			return nil, false, true, err
//		}
//		if tx.Status == 1 {
//			return tx, true, false, nil
//		}
//	} else {
//		return nil, false, true, nil
//	}
//	return nil, false, false, nil
//
//}
//
////ETH转账的燃气应设上限为“21000”单位。
//func TransferETH(client *ethclient.Client, keystorePath string, toAddressHex string, gasPrice, eth *big.Int) (*TransferInfo, error) {
//	info, err := transferAccounts(client, "", keystorePath, toAddressHex, uint64(21000), 0, gasPrice, eth, big.NewInt(0))
//	return info, err
//}
//func TransferERC20(client *ethclient.Client, tokenAddressHex string, keystorePath string, toAddressHex string, gasLimit uint64, gasPrice, erc20 *big.Int) (*TransferInfo, error) {
//	info, err := transferAccounts(client, tokenAddressHex, keystorePath, toAddressHex, gasLimit, 0, gasPrice, big.NewInt(0), erc20)
//	return info, err
//}
//
////加速转账
//func AccelerateTransfer(client *ethclient.Client, tokenAddressHex string, keystorePath string, toAddressHex string, gasLimit, nonce uint64, gasPrice, amount *big.Int) (*TransferInfo, error) {
//	var err error
//	var transferInfo *TransferInfo
//	if tokenAddressHex == "" {
//		transferInfo, err = transferAccounts(client, tokenAddressHex, keystorePath, toAddressHex, gasLimit, nonce, gasPrice, amount, big.NewInt(0))
//
//	} else {
//		transferInfo, err = transferAccounts(client, tokenAddressHex, keystorePath, toAddressHex, gasLimit, nonce, gasPrice, big.NewInt(0), amount)
//	}
//	return transferInfo, err
//}
//func transferAccounts(client *ethclient.Client, tokenAddressHex string, keystorePath string, toAddressHex string, gasLimit, nonce uint64, gasPrice, eth, erc20 *big.Int) (transferInfo *TransferInfo, err error) {
//	password := Conf.Password
//	////简单判断钱包地址和合约地址
//	if !IsETHAddress(toAddressHex) {
//		return nil, fmt.Errorf("请输入正确的合约地址和钱包地址！")
//	}
//	var ks *keystore.KeyStore
//	var account accounts.Account
//	if keystorePath == Conf.MainWalletsPath {
//		//如果是从主账号转出，可以直接使用 内存中的账户信息
//		ks = Conf.MainWalletsKeyStore
//		account = Conf.MainWalletsAccount
//	} else {
//		ks = keystore.NewKeyStore(Conf.TmpKeyStoreDir, keystore.LightScryptN, keystore.LightScryptP)
//		jsonBytes, err := ioutil.ReadFile(keystorePath)
//		if err != nil {
//			return nil, err
//		}
//		account, err = ks.Import(jsonBytes, password, password)
//		if err != nil {
//			return nil, err
//		}
//		err = ks.Unlock(account, password)
//		if err != nil {
//			return nil, err
//		}
//		if err := os.Remove(account.URL.Path); err != nil {
//			return nil, err
//		}
//	}
//	fromAddress := account.Address
//	if nonce == 0 {
//		nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
//		if err != nil {
//			return nil, err
//		}
//	}
//	toAddress := common.HexToAddress(toAddressHex)
//	transferFnSignature := []byte("transfer(address,uint256)")
//	hash := sha3.NewLegacyKeccak256()
//	hash.Write(transferFnSignature)
//	methodID := hash.Sum(nil)[:4]
//	var data []byte
//	//传入的gasPrice为0 则使用方法获取
//	if gasPrice.Cmp(big.NewInt(0)) == 0 {
//		//使用最快的·GasPrice
//		gasPrice = FastGasPrice
//	}
//	var tx *types.Transaction
//	//合约转账
//	if tokenAddressHex != "" {
//		tokenAddress := common.HexToAddress(tokenAddressHex)
//		bytecode, err := client.CodeAt(context.Background(), tokenAddress, nil) // nil is latest block
//		if err != nil {
//			return nil, err
//		}
//		if len(bytecode) <= 0 {
//			return nil, fmt.Errorf("请输入正确的合约地址！")
//		}
//		paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
//		paddedAmount := common.LeftPadBytes(erc20.Bytes(), 32)
//		data = append(data, methodID...)
//		data = append(data, paddedAddress...)
//		data = append(data, paddedAmount...)
//		tx = types.NewTransaction(nonce, tokenAddress, eth, gasLimit, gasPrice, data)
//	} else {
//		//eth转账
//		tx = types.NewTransaction(nonce, toAddress, eth, gasLimit, gasPrice, data)
//	}
//	chainID, err := client.NetworkID(context.Background())
//	if err != nil {
//		return nil, err
//	}
//	signedTx, err := ks.SignTx(account, tx, chainID)
//	if err != nil {
//		return nil, err
//	}
//	err = client.SendTransaction(context.Background(), signedTx)
//	if err != nil {
//		return nil, err
//	}
//	gwei := big.NewFloat(1).SetInt(Gwei)
//	ethWei := big.NewFloat(1).SetInt(Wei)
//	erc20Wei := big.NewFloat(1).SetInt(UsdtWei)
//	var amount float64
//	var amountString string
//	transferInfo = new(TransferInfo)
//	transferInfo.From = fromAddress.String()
//	transferInfo.To = toAddress.String()
//	transferInfo.Tx = signedTx.Hash().Hex()
//	transferInfo.Nonce = signedTx.Nonce()
//	transferInfo.GasLimit = gasLimit
//	gasPriceString := big.NewFloat(1).Quo(big.NewFloat(1).SetInt(gasPrice), gwei).String()
//	nowGasPrice, err := tools.StringTofloat(gasPriceString)
//	if err != nil {
//		log.Errorf("订单：%s,转换gas price出现问题，%s", transferInfo.Tx, err)
//		return nil, err
//	}
//	transferInfo.GasPrice = nowGasPrice
//	if eth.Cmp(big.NewInt(0)) == 0 {
//		amountString = big.NewFloat(1).Quo(big.NewFloat(1).SetInt(erc20), erc20Wei).String()
//	} else {
//		amountString = big.NewFloat(1).Quo(big.NewFloat(1).SetInt(eth), ethWei).String()
//	}
//	amount, err = tools.StringTofloat(amountString)
//	if err != nil {
//		log.Errorf("订单：%s,转换amount出现问题，%s", transferInfo.Tx, err)
//		return nil, err
//	}
//	transferInfo.Amount = amount
//	log.Printf("转账交易hash: %s", signedTx.Hash().Hex())
//	return transferInfo, nil
//}
//
////处理资金满足的钱包   contractAddress合约地址
//func HandleAmountSatisfiedUserWallet(w model.UserWalletInfo, contractAddress string, currencyTypeID uint) error {
//	csr := model.CapitalTrendsRecord{}
//	ETHBalance, err := QueryBalance(Client, w.AccountAddress, "")
//	if err != nil {
//		return fmt.Errorf("用户ID为：%d,error为%s", w.UserID, err)
//	}
//	//判断是否是eth资金归集
//	if contractAddress == "" && currencyTypeID == GetETHCurrencyType() {
//		//确定是资金归集eth 总金额减去手续费 210000*gasPrice
//		gasPrice, err := Client.SuggestGasPrice(context.Background())
//		if err != nil {
//			return err
//		}
//		limit := int64(21000)
//		fee := big.NewInt(1).Mul(gasPrice, big.NewInt(limit))
//		eth := ETHBalance.Sub(ETHBalance, fee)
//		if eth.Int64() <= 0 {
//			return nil
//		}
//		//获取用户的keystore
//		u := &model.User{}
//		u.GetUserById(w.UserID)
//		transferInfo, err := TransferETH(Client, Conf.KeyStoreDir+"/"+u.KeyStorePath, Conf.MainWalletsAccountHex, gasPrice, eth)
//		if err != nil {
//			return err
//		}
//		newCsr, err := csr.CreateCapitalConvergenceRecord(w.UserID, model.RecoveryETH, currencyTypeID, 2,
//			0, transferInfo.Nonce, transferInfo.GasLimit, transferInfo.Amount, transferInfo.GasPrice,
//			"向主账号转ETH", transferInfo.From,
//			transferInfo.To, transferInfo.Tx, 0)
//		if err != nil {
//			return fmt.Errorf("转账代币时，插入记录出现问题%s，订单号为%s", err, transferInfo.Tx)
//		}
//		//向订单池插入数据
//		InjectTxPool(newCsr)
//		return nil
//	}
//	for {
//		if USDTGasLimit == 0 {
//			time.Sleep(10 * time.Second)
//		} else {
//			break
//		}
//	}
//
//	//同gasPrice时计算账号中的gasLimit
//	curPrice := big.NewInt(1).Div(ETHBalance, big.NewInt(USDTGasLimit))
//
//	//计算手续费差多少
//	if curPrice.Int64() > SaveGasPrice.Int64() {
//
//		//如果当前价格大于最快价格，使用最快价格
//		if curPrice.Int64() > FastGasPrice.Int64() {
//			curPrice = FastGasPrice
//		}
//		//获取指定币余额
//		balance, err := QueryBalance(Client, w.AccountAddress, contractAddress)
//		if err != nil {
//			return fmt.Errorf("获取币种：%d余额失败，用户ID为：%d,error为%s", w.CurrencyTypeID, w.UserID, err)
//		}
//		//将余额全部转入主钱包
//		u := &model.User{}
//		u = u.GetUserById(w.UserID)
//		if u.ID == 0 {
//			return fmt.Errorf("用户%d转账USDT时出现问题！%s", w.UserID, "找不到该用户！")
//		}
//		if balance.Int64() == 0 {
//			return nil
//		}
//		transferInfo, err := TransferERC20(Client, contractAddress, Conf.KeyStoreDir+"/"+u.KeyStorePath, Conf.MainWalletsAccountHex, big.NewInt(1).Div(ETHBalance, curPrice).Uint64()-1, curPrice, balance)
//		if err != nil {
//			return fmt.Errorf("用户%d转账USDT时出现问题gasLimit:%d,gasPrice:%s,！%s", w.UserID, USDTGasLimit, curPrice.String(), err)
//		}
//		//因为手续费足够，直接转 usdt到主钱包 currencyTypeID 指定币种
//		newCsr, err := csr.CreateCapitalConvergenceRecord(w.UserID, model.RecoveryERC, currencyTypeID, 2,
//			0, transferInfo.Nonce, transferInfo.GasLimit, transferInfo.Amount, transferInfo.GasPrice,
//			"向主账号转erc20代币", transferInfo.From,
//			transferInfo.To, transferInfo.Tx, 0)
//		if err != nil {
//			return fmt.Errorf("转账代币时，插入记录出现问题%s，订单号为%s", err, transferInfo.Tx)
//		}
//		//向订单池插入数据
//		InjectTxPool(newCsr)
//		return nil
//	}
//	//计算eth的差值  转最新记录的1.3倍的以太去钱包
//	eth := new(big.Int).Mul(FastGasPrice, big.NewInt(USDTGasLimit))
//	eth = eth.Sub(eth, ETHBalance)
//	fEth := big.NewFloat(1).SetInt(eth)
//	//转以太需要用以太的单位
//	fWei := big.NewFloat(1).SetInt(Wei)
//	amountString := big.NewFloat(1).Quo(fEth, fWei).String()
//	amountD, err := decimal.NewFromString(amountString)
//	if err != nil {
//		log.Error(err)
//		return fmt.Errorf("eth转换错误")
//	}
//	dEth, _ := amountD.Float64()
//	//插入转账记录 t=1 转出手续费 currencyTypeID 需要归集的代币id  这里为特殊处理 其他情况  currencyTypeID都为转账的代币id
//	newCsr, err := csr.CreateInitCapitalConvergenceRecord(w.UserID, model.TransferOutETHServiceCharge, currencyTypeID,
//		Conf.MainWalletsAccountHex, w.AccountAddress, "转给钱包手续费", dEth)
//	if err != nil {
//		return fmt.Errorf("用户ID为：%d,出现错误%s", w.UserID, err)
//	}
//	//插入队列
//	go func() {
//		CapitalTrendsRecords <- newCsr
//	}()
//	return nil
//}
//
//func IsETHAddress(addressHex string) bool {
//	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
//	return re.MatchString(addressHex)
//}
