package oracle

import (
	utils "github.com/YaleOpenLab/smartPropertyMVP/stellar/utils"
)

// PriceOracle returns the power tariffs and any data that we need to  certify
// that is in the real world. Right now, this is hardcoded since we need to come up
// with a construct to get the price data in a reliable way - this could be a website
// were poeple erport this or certified authorities can timestamp this on chain
// or similar. Web s craping governemnt websites might work, but that seems too
// overkill for what we're doing now.
func PriceOracle() (string, error) {
	// right now, community consensus look like the price of electricity is
	// $0.2 per kWH in Puerto Rico, so hardcoding that here.
	priceOfElectricity := 0.2
	// since solar is free, they just need to pay this and then in some x time (chosen
	// when the order is created / confirmed on the school side), they
	// can own the panel.
	// the average energy consumption in puerto rico seems to be 5,657 kWh or about
	// 471 kWH per household. lets take 600 accounting for a 20% error margin.
	averageConsumption := float64(600)
	avgString := utils.FloatToString(priceOfElectricity * averageConsumption)
	return avgString, nil
}

// PriceOracleInFloat does the same thing as PriceOracle, but returns the data
// as a float for use in appropriate places
func PriceOracleInFloat() float64 {
	priceOfElectricity := 0.2
	averageConsumption := float64(600)
	return priceOfElectricity * averageConsumption
}