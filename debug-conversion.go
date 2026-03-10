package main

import (
	"fmt"
	"math/big"
	"strconv"
)

func parseAmount(amountStr string) (float64, error) {
	return strconv.ParseFloat(amountStr, 64)
}

func convertToWei(amount float64, decimals int) *big.Int {
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountBig := new(big.Float).SetFloat64(amount)
	multiplierBig := new(big.Float).SetInt(multiplier)
	resultBig := new(big.Float).Mul(amountBig, multiplierBig)
	result := new(big.Int)
	resultBig.Int(result)
	return result
}

func main() {
	// Test with ClawSwift chain (decimal 6)
	amountStr := "0.001"
	decimals := 6
	
	amount, err := parseAmount(amountStr)
	if err != nil {
		fmt.Printf("Error parsing amount: %v\n", err)
		return
	}
	
	weiAmount := convertToWei(amount, decimals)
	fmt.Printf("Amount %s CLAW with decimal %d converts to: %s wei\n", amountStr, decimals, weiAmount.String())
	
	// Test with Ethereum (decimal 18) for comparison
	weiAmountEth := convertToWei(amount, 18)
	fmt.Printf("Amount %s ETH with decimal %d converts to: %s wei\n", amountStr, 18, weiAmountEth.String())
}
