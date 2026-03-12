package wallet

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sipeed/picoclaw/pkg/wallet"

	"github.com/spf13/cobra"
	"math/big"
)

func newWriteCommand(walletServiceFn func() (*wallet.Service, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write <contract_address> <abi_type> <method> <value> <password> [parameters]",
		Short: "Execute a write smart contract function",
		Long: `Executes a write smart contract function on ClawSwift network.
		
Examples:
  /wallet write 0x20c0000000000000000000000000000000000000 erc20 transfer 0 1234 0xA3570FCDA303F55e0978be450f87F885d80a3758 1000000000000000000`,
		Args: cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := walletServiceFn()
			if err != nil {
				return err
			}

			contractAddr := common.HexToAddress(args[0])
			abiType := args[1]
			method := args[2]

			value := new(big.Int)
			value.SetString(args[3], 10)

			password := args[4]

			var params []interface{}
			if len(args) > 5 {
				for _, p := range args[5:] {
					if strings.HasPrefix(strings.ToLower(p), "0x") && len(p) == 42 {
						params = append(params, common.HexToAddress(p))
					} else {
						// Try to parse as number
						parsed, ok := new(big.Int).SetString(p, 10)
						if ok {
							params = append(params, parsed)
						} else {
							params = append(params, p)
						}
					}
				}
			}

			// Get accounts from keystore
			accounts := w.GetAccounts()
			if len(accounts) == 0 {
				return fmt.Errorf("no wallets found. Create one first with /wallet create [password]")
			}
			from := accounts[0].Address

			ctx := cmd.Context()
			tx, err := w.ExecuteContractMethod(ctx, from, contractAddr, abiType, method, value, password, params...)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction sent! Hash: %s\n", tx.Hash().Hex())
			return nil
		},
	}
	return cmd
}