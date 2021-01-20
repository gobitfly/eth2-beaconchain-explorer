package utils

// A map of the european tax rates according to https://ec.europa.eu/taxation_customs/sites/taxation/files/resources/documents/taxation/vat/how_vat_works/rates/vat_rates_en.pdf
var Rates map[string]int = map[string]int{
	"AT": 20, // Austria
	"BE": 21, // Belgium
	"BG": 20, // Bulgaria
	"CY": 19, // Cyprus
	"CZ": 21, // Czech Republic
	"DE": 19, // Germany
	"DK": 25, // Denmark
	"EE": 20, // Estonia
	"EL": 24, // Greece
	"ES": 21, // Spain
	"FI": 24, // Finland
	"FR": 20, // France
	"GB": 20, // United Kingdom
	"HR": 25, // Croatia
	"HU": 27, // Hungary
	"IE": 23, // Ireland
	"IT": 22, // Italy
	"LT": 21, // Lithuania
	"LU": 17, // Luxembourg
	"LV": 21, // Latvia
	"MT": 18, // Malta
	"NL": 21, // Netherlands
	"PL": 23, // Poland
	"PT": 23, // Portugal
	"RO": 19, // Romania
	"SE": 25, // Sweden
	"SI": 22, // Slovenia
	"SK": 20, // Slovak Republic
}

// A map of the european tax rates according to https://ec.europa.eu/taxation_customs/sites/taxation/files/resources/documents/taxation/vat/how_vat_works/rates/vat_rates_en.pdf
var StripeRates map[string]string = map[string]string{
	"AT": "txr_1HqcFcBiORp9oTlKnyNWVp4r", // Austria
	"BE": "txr_1I9pNwBiORp9oTlKxTmOz7a1", // Belgium
	"BG": "txr_1I9pOKBiORp9oTlKfMcVou1L", // Bulgaria
	"CY": "txr_1I9pP0BiORp9oTlKy5CpfXQR", // Cyprus
	"CZ": "txr_1I9pPzBiORp9oTlKWnfxgOw1", // Czech Republic
	"DE": "txr_1HqdWaBiORp9oTlKkij8L6dU", // Germany
	"DK": "txr_1I9pQsBiORp9oTlK4UTiJJTN", // Denmark
	"EE": "txr_1I9pR7BiORp9oTlKgaTRDucB", // Estonia
	"EL": "txr_1I9pRKBiORp9oTlKudk6Zbf5", // Greece
	"ES": "txr_1I9pRYBiORp9oTlKlFCuCxDv", // Spain
	"FI": "txr_1I9pRsBiORp9oTlKabg2US6z", // Finland
	"FR": "txr_1I9pS5BiORp9oTlKK6b9bi1n", // France
	"GB": "txr_1I9pSNBiORp9oTlKCO3Of9YI", // United Kingdom
	"HR": "txr_1I9pSdBiORp9oTlKsdb5E2eO", // Croatia
	"HU": "txr_1I9pSuBiORp9oTlKW7OdEDln", // Hungary
	"IE": "txr_1I9pTCBiORp9oTlK4LAH8ZAQ", // Ireland
	"IT": "txr_1I9pUOBiORp9oTlKrissM9GJ", // Italy
	"LT": "txr_1I9pUgBiORp9oTlKCXqz67MM", // Lithuania
	"LU": "txr_1I9pUzBiORp9oTlK8zsYnVgS", // Luxembourg
	"LV": "txr_1I9pVCBiORp9oTlKUGQ94mXO", // Latvia
	"MT": "txr_1I9pVPBiORp9oTlKDaogHVHn", // Malta
	"NL": "txr_1I9pVeBiORp9oTlKbn6ZdXvh", // Netherlands
	"PL": "txr_1I9pVvBiORp9oTlKJuyxDeDD", // Poland
	"PT": "txr_1I9pW8BiORp9oTlKEiGKbdcq", // Portugal
	"RO": "txr_1I9pWNBiORp9oTlKRvGIm5mx", // Romania
	"SE": "txr_1I9pWbBiORp9oTlKqGEvmQ0H", // Sweden
	"SI": "txr_1I9pWrBiORp9oTlKljnvpOqI", // Slovenia
	"SK": "txr_1I9pX9BiORp9oTlKDtkJ8A09", // Slovak Republic
}

func strToPointer(st string) *string {
	return &st
}

var StripeDynamicRates = []*string{
	strToPointer("txr_1HqcFcBiORp9oTlKnyNWVp4r"),
	strToPointer("txr_1I9pNwBiORp9oTlKxTmOz7a1"),
	strToPointer("txr_1I9pOKBiORp9oTlKfMcVou1L"),
	strToPointer("txr_1I9pP0BiORp9oTlKy5CpfXQR"),
	strToPointer("txr_1I9pPzBiORp9oTlKWnfxgOw1"),
	strToPointer("txr_1HqdWaBiORp9oTlKkij8L6dU"),
	strToPointer("txr_1I9pQsBiORp9oTlK4UTiJJTN"),
	strToPointer("txr_1I9pR7BiORp9oTlKgaTRDucB"),
	strToPointer("txr_1I9pRKBiORp9oTlKudk6Zbf5"),
	strToPointer("txr_1I9pRYBiORp9oTlKlFCuCxDv"),
	strToPointer("txr_1I9pRsBiORp9oTlKabg2US6z"),
	strToPointer("txr_1I9pS5BiORp9oTlKK6b9bi1n"),
	strToPointer("txr_1I9pSNBiORp9oTlKCO3Of9YI"),
	strToPointer("txr_1I9pSdBiORp9oTlKsdb5E2eO"),
	strToPointer("txr_1I9pSuBiORp9oTlKW7OdEDln"),
	strToPointer("txr_1I9pTCBiORp9oTlK4LAH8ZAQ"),
	strToPointer("txr_1I9pUOBiORp9oTlKrissM9GJ"),
	strToPointer("txr_1I9pUgBiORp9oTlKCXqz67MM"),
	strToPointer("txr_1I9pUzBiORp9oTlK8zsYnVgS"),
	strToPointer("txr_1I9pVCBiORp9oTlKUGQ94mXO"),
	strToPointer("txr_1I9pVPBiORp9oTlKDaogHVHn"),
	strToPointer("txr_1I9pVeBiORp9oTlKbn6ZdXvh"),
	strToPointer("txr_1I9pVvBiORp9oTlKJuyxDeDD"),
	strToPointer("txr_1I9pW8BiORp9oTlKEiGKbdcq"),
	strToPointer("txr_1I9pWNBiORp9oTlKRvGIm5mx"),
	strToPointer("txr_1I9pWbBiORp9oTlKqGEvmQ0H"),
	strToPointer("txr_1I9pWrBiORp9oTlKljnvpOqI"),
	strToPointer("txr_1I9pX9BiORp9oTlKDtkJ8A09"),
}
