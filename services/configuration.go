package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"fmt"
)

const (
	ConfigurationCategorySlotViz types.ExplorerConfigurationCategory = "SlotVizOnMainPage"
)
const (
	ConfigurationKeyVisibleFromEpoch types.ExplorerConfigurationKey = "VisibleFromEpoch"
	ConfigurationKeyVisibleToEpoch   types.ExplorerConfigurationKey = "VisibleToEpoch"
	ConfigurationKeyHardforkEpoch    types.ExplorerConfigurationKey = "HardforkEpoch"
	ConfigurationKeyHardforkName     types.ExplorerConfigurationKey = "HardforkName"
)

/*
**
This is the list of possible configurations that can be changed in the explorer administration
Per default these values will be taken, overridden by the values from the db, if they exist.
**
*/
var DefaultExplorerConfiguration types.ExplorerConfigurationMap = types.ExplorerConfigurationMap{
	ConfigurationCategorySlotViz: {
		ConfigurationKeyVisibleFromEpoch: {Value: "0", DataType: "int"},
		ConfigurationKeyHardforkEpoch:    {Value: "0", DataType: "int"},
		ConfigurationKeyVisibleToEpoch:   {Value: "0", DataType: "int"},
		ConfigurationKeyHardforkName:     {Value: "", DataType: "string"},
	},
}

func GetExplorerConfigurationsWithDefaults() (types.ExplorerConfigurationMap, error) {

	result, err := db.GetExplorerConfigurations()

	if err != nil {
		return nil, err
	}

	configs := types.ExplorerConfigurationMap{}

	// first we clone the defaults
	for category, keyMap := range DefaultExplorerConfiguration {
		configs[category] = make(types.ExplorerConfigurationKeyMap)
		for key, configValue := range keyMap {
			configs[category][key] = types.ExplorerConfigValue{
				Value:    fmt.Sprintf("%v", configValue.Value),
				DataType: configValue.DataType,
			}
		}
	}

	// now let's fill in the values from the db (if exists)
	for _, row := range result {
		keyMap, ok := configs[row.Category]
		if ok {
			explorerValue, ok := keyMap[row.Key]
			if ok {
				configs[row.Category][row.Key] = types.ExplorerConfigValue{Value: row.Value, DataType: explorerValue.DataType}
			}
		}
	}

	return configs, nil
}
