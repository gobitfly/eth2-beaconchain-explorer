package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/utils"
)

func GetValidatorOnlineThresholdSlot() uint64 {
	latestProposedSlot := services.LatestProposedSlot()
	var validatorOnlineThresholdSlot uint64
	if latestProposedSlot < 1 {
		validatorOnlineThresholdSlot = 0
	} else {
		validatorOnlineThresholdSlot = latestProposedSlot - utils.Config.Chain.SlotsPerEpoch*2
	}

	return validatorOnlineThresholdSlot
}
