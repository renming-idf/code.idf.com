package tools

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func PrivateKeyTOKeystore(privateKey, password string) (string, error) {
	ks := keystore.NewKeyStore(".", keystore.LightScryptN, keystore.LightScryptP)
	prk, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", err
	}
	accounts, err := ks.ImportECDSA(prk, password)
	if err != nil {
		return "", err
	}
	return accounts.URL.String(), nil
}

func KeystoreToPrivateKey(privateKeyFile, password string) (*ecdsa.PrivateKey, error) {
	keyjson, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		fmt.Println("read keyjson file failed：", err)
	}
	unlockedKey, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return nil, err
	}
	return unlockedKey.PrivateKey, nil
}
func IsValidAddress(iaddress interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := iaddress.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// ToDecimal wei to decimals
func ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}

// ToWei decimals to wei
func ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

// CalcGasCost calculate gas cost given gas limit (units) and gas price (wei)
func CalcGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimit))
	return gasLimitBig.Mul(gasLimitBig, gasPrice)
}

// SigRSV signatures R S V returned as arrays
func SigRSV(isig interface{}) ([32]byte, [32]byte, uint8) {
	var sig []byte
	switch v := isig.(type) {
	case []byte:
		sig = v
	case string:
		sig, _ = hexutil.Decode(v)
	}

	sigstr := common.Bytes2Hex(sig)
	rS := sigstr[0:64]
	sS := sigstr[64:128]
	R := [32]byte{}
	S := [32]byte{}
	copy(R[:], common.FromHex(rS))
	copy(S[:], common.FromHex(sS))
	vStr := sigstr[128:130]
	vI, _ := strconv.Atoi(vStr)
	V := uint8(vI + 27)

	return R, S, V
}
func StringTofloat(amountString string) (float64, error) {
	f, err := decimal.NewFromString(amountString)
	if err != nil {
		return 0, err
	}
	amount, _ := f.Float64()
	return amount, nil
}
func BytesToAddr(b []byte) common.Address {
	addr := common.HexToAddress(common.BytesToHash(b).String())
	return addr
}
func HandleData(data []byte) (to string, amount *big.Int, err error) {
	if len(data) != 68 {
		return "", nil, nil
	}
	if hexutil.Encode(data[:4]) != "0xa9059cbb" {
		return "", nil, fmt.Errorf("不是转账！")
	}
	to = strings.ToLower(BytesToAddr(data[4:36]).String())
	hh := BytesToAddr(data[36:]).String()[2:]
	amount, flag := big.NewInt(1).SetString(hh, 16)
	if !flag {
		return "", nil, fmt.Errorf("获取amount失败")
	}
	return to, amount, nil
}
func BigIntToFloat(n, wei *big.Int) float64 {
	f := big.NewFloat(1).SetInt(n)
	amount := f.Quo(f, big.NewFloat(1).SetInt(wei))
	a, _ := amount.Float64()
	return a
}
func FloatToWei(wei *big.Int, amount float64) *big.Int {
	f := big.NewFloat(1).SetInt(wei)
	i, _ := f.Mul(f, big.NewFloat(amount)).Int(big.NewInt(1))
	if i == nil {
		return big.NewInt(0)
	}
	return i
}

// ffmpeg进行视频处理
func GetVideoMp4(filename string) (string, error) {
	mp4FileName := fmt.Sprintf("%s", uuid.NewV4()) + ".mp4"
	fp := "ffmpeg"
	if err := exec.Command(fp, "-i", "./uploads/"+filename, "-b:v", "400K", "./uploads/"+mp4FileName).Run(); err != nil {
		return "", err
	}
	//h256FileName := fmt.Sprintf("%s", uuid.NewV4()) + ".h264"
	//if err := exec.Command(fp, "-i", "./uploads/"+mp4FileName, "-an", "-vcodec", "libx264", "-crf", "23", "./uploads/"+h256FileName).Run(); err != nil {
	//	return "", err
	//}
	return mp4FileName, nil
}

func GetCover(filename, coverTime string) (string, error) {
	// cmd := exec.Command("ffmpeg", "-i", filename, "-vframes", strconv.Itoa(index), "-s", fmt.Sprintf("%dx%d", width, height), "-f", "singlejpeg", "-")
	endFileName := "picture-" + fmt.Sprintf("%s", uuid.NewV4()) + ".jpg" //图片用uuid进行重命名
	//endFileName := "zzzztest.png" //图片用uuid进行重命名
	//ffmpeg-ss 00:01:00 -i video.mp4 -to 00:02:00 -c copy cut.mp4
	//cmd := exec.Command("ffmpeg", "-ss", beginTime, "-i", filename, "-c", "copy", "-t", duration, endFileName)
	//ffmpeg -ss 0:1:30 -t 0:0:20 -i input.avi -vcodec copy -acodec copy output.avi
	fp := "ffmpeg"
	/*
		-ss 表示开始时间
		-i 表示文件
		-t 表示共要多少时间
		-r 表示每一秒几帧
		-q:v 表示存储jpeg的图像质量，一般2是高质量。
	*/
	cmd := exec.Command(fp, "-ss", coverTime, "-i", "./uploads/"+filename, "-t", "1", "-r", "1", "-f", "image2", "./uploads/"+endFileName)
	//cmd := exec.Command(fp, "-ss", beginTime, "-i", "./uploads/"+filename, "-t", "1", "-c", "copy", "./uploads/"+endFileName)
	return endFileName, cmd.Run() //要用cmd.Run启动
}

func SavePicture(base64Info string) string {
	// 存储一个base64作为用户的实名图片
	dist, _ := base64.StdEncoding.DecodeString(base64Info)
	pictureName := fmt.Sprintf("%s", uuid.NewV4()) + ".png"
	f, _ := os.OpenFile("./uploads/"+pictureName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer f.Close()
	f.Write(dist)
	return pictureName
}
