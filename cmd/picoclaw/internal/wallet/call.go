package wallet

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sipeed/picoclaw/pkg/wallet"

	"github.com/spf13/cobra"
	"math/big"
)

func newCallCommand(walletServiceFn func() (*wallet.Service, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call <contract_address> <abi_type> <method> [parameters]",
		Short: "Call a read-only smart contract function",
		Long: `Calls a read-only smart contract function on ClawSwift network.
		
Examples:
  /wallet call 0x20c0000000000000000000000000000000000000 erc20 balanceOf 0x44c2db1fc0986ca3c173403701c909874badc0d0
  /wallet call 0x20c0000000000000000000000000000000000000 erc20 symbol
  /wallet call 0x20c0000000000000000000000000000000000000 erc20 decimals`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := walletServiceFn()
			if err != nil {
				return err
			}

			contractAddr := common.HexToAddress(args[0])
			abiType := args[1]
			method := args[2]

			var params []interface{}
			if len(args) > 3 {
				for _, p := range args[3:] {
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

			ctx := cmd.Context()
			result, err := w.CallContractMethod(ctx, contractAddr, abiType, method, params...)
			if err != nil {
				return err
			}

			fmt.Printf("Result: %v\n", result)
			return nil
		},
	}
	return cmd
}