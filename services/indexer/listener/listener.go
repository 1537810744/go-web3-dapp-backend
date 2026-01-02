package listener

import (
	"context"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"

	"go-web3-dapp-backend/services/indexer/models"
)

var ContractAddress common.Address
var ContractABI abi.ABI

const contractABIJSON = `[
    {"inputs":[],"name":"deposit","outputs":[],"stateMutability":"payable","type":"function"},
    {"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"withdraw","outputs":[],"stateMutability":"nonpayable","type":"function"},
    {"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"getDeposit","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
    {"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"user","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Deposit","type":"event"},
    {"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"user","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Withdraw","type":"event"}
]`

func InitContract(addressHex string) {
	ContractAddress = common.HexToAddress(addressHex)
	parsedABI, err := abi.JSON(strings.NewReader(contractABIJSON))
	if err != nil {
		log.Fatal("abi parse failed:", err)
	}
	ContractABI = parsedABI
}

func StartListener(client *ethclient.Client, db *gorm.DB) {
	var status models.SyncStatus
	if err := db.First(&status, 1).Error; err != nil {
		status = models.SyncStatus{ID: 1, LastBlock: 0}
		db.Create(&status)
	}
	fromBlock := status.LastBlock + 1

	for {
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		toBlock := header.Number.Uint64()
		if toBlock < fromBlock {
			time.Sleep(time.Second * 5)
			continue
		}
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(toBlock)),
			Addresses: []common.Address{ContractAddress},
		}
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		for _, vLog := range logs {
			// find matching event by ID
			if len(vLog.Topics) == 0 {
				continue
			}
			if vLog.Topics[0] == ContractABI.Events["Deposit"].ID {
				user := common.HexToAddress(vLog.Topics[1].Hex())
				data, _ := ContractABI.Unpack("Deposit", vLog.Data)
				amount := data[0].(*big.Int)
				tx := models.Transaction{
					TxHash:      vLog.TxHash.Hex(),
					EventType:   "deposit",
					UserAddress: user.Hex(),
					ToAddress:   "",
					Amount:      amount.String(),
					BlockNumber: vLog.BlockNumber,
				}
				db.Create(&tx)
			} else if vLog.Topics[0] == ContractABI.Events["Withdraw"].ID {
				user := common.HexToAddress(vLog.Topics[1].Hex())
				to := common.HexToAddress(vLog.Topics[2].Hex())
				data, _ := ContractABI.Unpack("Withdraw", vLog.Data)
				amount := data[0].(*big.Int)
				tx := models.Transaction{
					TxHash:      vLog.TxHash.Hex(),
					EventType:   "withdraw",
					UserAddress: user.Hex(),
					ToAddress:   to.Hex(),
					Amount:      amount.String(),
					BlockNumber: vLog.BlockNumber,
				}
				db.Create(&tx)
			}
			status.LastBlock = vLog.BlockNumber
			db.Save(&status)
		}
		fromBlock = toBlock + 1
		time.Sleep(time.Second * 5)
	}
}
