package utils

import (
	"encoding/json"
	"eth2-exporter/types"
	"testing"

	capella "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

/*
mainnet: {"data":{"genesis_time":"1606824023","genesis_validators_root":"0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95","genesis_fork_version":"0x00000000"}}
prater: {"data":{"genesis_time":"1616508000","genesis_validators_root":"0x043db0d9a83813551ee2f33450d23797757d430911a9320530ad8a0eabc43efb","genesis_fork_version":"0x00001020"}}
sepolia: {"data":{"genesis_time":"1655733600","genesis_validators_root":"0xd8ea171f3c94aea21ebc42a1ed61052acf3f9209c00e4efbaaddac09ed9b8078","genesis_fork_version":"0x90000069"}}
zhejiang: {"data":{"genesis_time":"1675263600","genesis_validators_root":"0x53a92d8f2bb1d85f62d16a156e6ebcd1bcaba652d0900b2c2f387826f3481f6f","genesis_fork_version":"0x00000069"}}
*/
func TestVerifyBlsToExecutionChangeSignature(t *testing.T) {
	Config = &types.Config{}
	ReadConfig(Config, "")
	Config.Chain.ClConfig.GenesisForkVersion = "0x00000069"
	Config.Chain.GenesisValidatorsRoot = "0x53a92d8f2bb1d85f62d16a156e6ebcd1bcaba652d0900b2c2f387826f3481f6f"
	Config.Chain.DomainBLSToExecutionChange = "0x0A000000"
	msg := []byte(`[{"message":{"validator_index":"62019","from_bls_pubkey":"0x8562a3e163bfbc20bebbbea2643bdcb8823d36a481ae770ce7eb358c53a78ac5e5074ef1f2506fe1c42d7c36f9dc650f","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x80f445291542b582cbda37142621461328cdf960bc460e8e4de98530fab5d8cdc052a5b2cb47100a2ddf33555363d6db07b79921b7eddc4922e925a83d747befe02e384c12fb7f13386fe7295ec4330bdd959f130e95a4dc476e27fc31c1ce42"},{"message":{"validator_index":"62020","from_bls_pubkey":"0xaa592161caf20a7ea52892cffd9e5e7770aff561380e842eeacdfb5b205bcf64e5fadd86229fa67ced09d1613e2d854b","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x89544b15ed8148a2471fd6ca96bf0aebe16419d59b5fb9eedeeba262592cba601d09107f848af3e05c3d8c1c5405284d05aaf488c24fbe140d8f06c285de3f4db102bf66f21c176715bbb1a30bf6aa27e5caa7684caf92c4c5b8fc69a73f158c"},{"message":{"validator_index":"62021","from_bls_pubkey":"0xad2cac326e83f26fad0139aede95a7d0bd5ae4efa76741c0d28ccfced0fd7264581b1e088d3cdbd3cc372d86808cf153","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xb52d33f0c6ae211a7991d3c6e8efd6b912bcbe7ee7425e6ea81318944b48767c53a7d8699e468e01f6cdb4779397354812fb6a399191c4e65833646eaea7090c112c67ad5960d8b859d3b5a9b0fa59aed7748740cd2f9633740e4769d07c0497"},{"message":{"validator_index":"62022","from_bls_pubkey":"0xb853c4a9f7c22100d11d73d416d619312a68a300ed8abff6498802885964f3e8b6332d5d449a1aafb182cc0d8b50ecec","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x8495a5eb3e50653349a9752cf211f7abf245abd23f321fb4832d889098db748fbec33198d061bca0100858c949df2827196c2cdc97b6c65da4308cf5b0af403ca947a23488eba34a23f390b3348607f213b14e19484674147ea00ade2b14f08e"},{"message":{"validator_index":"62023","from_bls_pubkey":"0x8f6bdd3e479dff75e94d7563b6a95af7eb63f70c620d20f8b942587070e352a7f883e93fc2678dcdfca1fba0c01afa6f","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xa192de81fc920866a2201f7e96e760d753922ffecedda3d3af1af7db376ac11be28ee3014e022ba3dc43b23464c26f601281e9bb932b88884d16ed0ef19114bc45f8883448412b899bc79b97e06313d319c8a36c8888e7ebeb0c15e38d685a35"},{"message":{"validator_index":"62024","from_bls_pubkey":"0xaeaa4e0f525c5506f69c3670ab87aebc4a730405954e6e88ee98e2b3b588855fbf567426183c5c93afa0e042cf3f3833","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x8f1d25968a7e2b8fdde984008856ee8ef6a2d88184e3555291aca5d209fff71d713af7ef6d70f5cf5c5a10f18e23ceda124135b721ec1f5757da503629bd3fdfb914488ae8975daec4861ce3c0476e9a8d1ae29d16a7d41308740b44dcda458b"},{"message":{"validator_index":"62025","from_bls_pubkey":"0x9528a8adc5d544dd349480317005d26bda5025e676f25ef8d071555836103f0421c3191dec84c14ac9fc9050c2cd5f38","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x951aae295977053a015dbcb4a3cdf42aa19821c34b0d740dfc5d5465cbc474069d60cb4e16965714c7108867985a469701c7a5fcb2c6a350faa77bc5a2f8358c54a15b51c40ce052ff2bdc1669d0b4d270fec30ac740c02e8c9e1cdd3e788eae"},{"message":{"validator_index":"62026","from_bls_pubkey":"0xb03d944e673257de3fcd0bfd040892e12a3cd69ae17dcf69ac6c45a9858ed673d84ce0be7f41b80848662c883b053922","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x9478e181eabd6d687b565db04979ca0cc1460cf22449b7ea9001da1018cf5c58d5906150c7c42a68de0044d86fe2964e146bfbebbc3cc82410bdcf5fcd7c8a9998f139fcc1b41dd3fb5b0b0491c8b77601e2f6b7931d0b1a5dec94688b20f85d"},{"message":{"validator_index":"62027","from_bls_pubkey":"0x981552e4fae7fb52599f374564b64de26ba0fe48ff5700bb7964ff70b5e483945b6f2ecd4cbc5edca55b23f9d5d94549","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xa05e465140eacf20ceea6a9564c989c2144690a199623a03551045113552cd6f33e91f997382527af40c91e5d4cd9f7e14b56a5000e08719dc67499a9dfc80d44520396c632917c683c7bb4424178ef7e5155e9e92f431127945b432e3ad9c81"},{"message":{"validator_index":"62028","from_bls_pubkey":"0x96845a73c0c3380da5670caeac534c48a0d173e29695c4b42fa3fbf5e45bda0ffc29f5c1311b7e1e9a0783e438815398","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x8e12ec0c3651ca613b2e61220858677ae5bc4bec1007eb2392e127418d858312cf8776ea45a5ff2b7ba30440ec8060a90b9b13473a2d106efe074e8aeae30d61f77df5b401c4adcc3b560dc8b5239a9269cad755e7a2138999a9d1e68f511ad4"},{"message":{"validator_index":"62029","from_bls_pubkey":"0x80875a4e606d7127d2bb7807823b5c0421eac6d8cdcf72f0d759f66c4e74f033aca81145bdd1fdfd26b8f56bf04d68c7","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xb8c0b653ad078223d73543c0cfb36d228c4feea47d0ffe4f9c776b4f9adb8f0fe11c8046dfdc363c890d59ded10d80e60c6128bf92f0fa5cc2a891e1e8b7b75a72b16c9fe23e01651ccb5e8dc4370ea1ce5657f32eb4f86fc94329c32f70f280"},{"message":{"validator_index":"62030","from_bls_pubkey":"0x9948ca839a3005b4bae4371feea1f4e39e8ffc4fd9d01225540d0b3e6dcf034e24edadb7e226ffc95c4f019f13b5b588","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xb374bb975d03a65355bfe84ff53f3fffaf1ffb94a593d192263a3ee62396a7316bfb08c58b86979a56d1db47a80a8eb80a65145d4502ec3febaf41d6bef7b23976eacef106e783677cf368fbd2a38779e44f936394e7cb3aad1790773590cdfd"},{"message":{"validator_index":"62031","from_bls_pubkey":"0xaf220d1ccaf5599ec56834636ecc6a576eb67da88b18b12bd0fd2ba058f20313fffe564bbedb4e75b2e7c37bc205b0b4","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xa8c20636cbb57e2b2ea869536f25082878ecf6959d2b5384234b9bd3d9fad70fccf990507aa3bfe1f1ebd46682eb60b811f4c6372d1a7cf5ca4eab8e4c1481ff93f7f8023bd5855681da215887abb0cb3e1962f16922f979645c1ecae280f834"},{"message":{"validator_index":"62032","from_bls_pubkey":"0xa91c18b7e42a96d88d1d44809276593a982919e35d80ae6148fadd64a37795f935415ceb20553cc3f752f29ca1e2cfa9","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x8db810873b93b0cd0eae2a1f490391e2c304a9e25663f3faea66e6ad3733cfc17704461a5a8322a9981b1999ce62874416471410b8892c4edfef0c87e08be473b998c9b1ac351e3b7e4c5736c92782ed9788c61e429abbd9fda476fe951686cc"},{"message":{"validator_index":"62033","from_bls_pubkey":"0xb0ba14be3eb1929ee57bb57cfe37f50b4a5c9bbb568e38daca4a44d12c609a22fc35f6f24c49d7af7e1e486cd10ec01a","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xb5068dedf358abf19e1ad2c3116c561c4b2fb8368b6826a7571febcc54b806c52ba1fb4ad21092d59a944cf853a48e570b9826cbe702388a02593a8468c8835815c010d6992aac1727a5a8e35bdb9ddc282e45b488023b99394b2f66074efad3"},{"message":{"validator_index":"62034","from_bls_pubkey":"0x94a7e2625184e1985f680a3164603b6b1f1d6e47005084c68485940ce95e540f3dc4744379814ff941fb1daf80f463e8","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x87c199b58bafbb0c871d6adb3cee57f273e6cf3fa9892a9405480a91c884abd4f7857cd7d04dcbc7d6317bc61375de0d14e7d9027dedbd534cebf6aee9898b776644038f33bd3af358459a57c4814e76a797af7978efe01f957aafdaf19af681"},{"message":{"validator_index":"62035","from_bls_pubkey":"0x8ed5793fb5fc0b35aa85364c8d387fb7260d69a4abe62efb1d0b284b26eb26cc137c3c7401d4a078a0ffd17680a72003","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x8ad92ac300edb9ceb87546f8fea80c1ddafe60e4ff7d5eeb2f75e0996122a8eb006a5555ed1334d30b1d466ee2c3159916a1c35560b25c831eaffeefcce78cb95f7b3530918ac45e741638d404f460c04fee4b21334ed63d2d8acea7fbe08fc2"},{"message":{"validator_index":"62036","from_bls_pubkey":"0x8985c25bf74bfae1ca3222a6daf61ad172d1e842f169eb59e8a5030c3cb2521e6a6b6a273b7a288806f491d560233dc4","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xa08c3d98caff5234f0f2bc2f034dee99899f099ec9bd461c5898243b2e050c57c602f148726481b9a6a4ce4d176edb6514cdcac1bbfae7bf10d3274c943b3e54feabac26a8aed1aef85a108c36a7a00799a23f79d51780cfdfa9785454ae05ac"},{"message":{"validator_index":"62037","from_bls_pubkey":"0xacbd99d84d14711e27e6a3ed188a0cc8a8d9afdfc8b9395a711f82465d9fbf4c8f4672f4d107e0d076af7f1f0e2f2659","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0xb498428c2ddf49cfd03fbb57754ce41e236883747c569ac295e8d7d1d36f3b71579d92d006f4f234a8cf6d278a66cbc90056e12ad3ccec8ac758c591b6530e3a83e9dfd10fb4f74e196d713af83a25248c40ad1c87ea2a176d47a1efe735083e"},{"message":{"validator_index":"62038","from_bls_pubkey":"0xb8feafe7495c2b3f97724dd3c1eaa6c60adbeda10ef28c4932a0f9fc9b0f0f26fbadfd1525bbcdc05203958f01b59ab0","to_execution_address":"0x0bcededbeea88da966e98dd40796b802d54342cc"},"signature":"0x8e16f68787c7c0e445734dac652624de08a07f9d71e2ec5a1ec7c221b47ccef6efaf21e6faafa7709e67aa0b1bc756d014a36cb37d8e2ea0f281d633f816f82da4868bdc782960b099595f39be767d844244e85f01ebb2ef80b73099b8b0f836"}]`)
	var ops []*capella.SignedBLSToExecutionChange
	err := json.Unmarshal(msg, &ops)
	if err != nil {
		t.Errorf("failed unmarshaling msg: %v", err)
	}
	for _, op := range ops {
		err = VerifyBlsToExecutionChangeSignature(op)
		if err != nil {
			t.Errorf("failed verifying bls sig: %v", err)
		}
	}
}

func TestVerifyVoluntaryExitSignature(t *testing.T) {
	Config = &types.Config{}
	ReadConfig(Config, "")
	Config.Chain.DomainVoluntaryExit = "0x04000000"
	ZhejiangGenesisForkVersion := "0x00000069"
	ZhejiangCapellaForkVersion := "0x00000072"
	ZhejiangGenesisValidatorsRoot := "0x53a92d8f2bb1d85f62d16a156e6ebcd1bcaba652d0900b2c2f387826f3481f6f"
	PraterGenesisForkVersion := "0x00001020"
	PraterGenesisValidatorsRoot := "0x043db0d9a83813551ee2f33450d23797757d430911a9320530ad8a0eabc43efb"
	tests := []struct {
		CurrentForkVersion    string
		GenesisValidatorsRoot string
		Msg                   []byte
		Pubkey                []byte
		Valid                 bool
	}{
		{
			CurrentForkVersion:    ZhejiangCapellaForkVersion,
			GenesisValidatorsRoot: ZhejiangGenesisValidatorsRoot,
			Msg:                   []byte(`{"message":{"epoch":"3541","validator_index":"62019"},"signature":"0xa0f4ff61e01346b98acb7a8003df6dc5e61760adf54da5e16d138d5171cf64c2429787763973697aee47d9949108fac2106799ba96690e33263e23e079a3213d80c5617c76a350a9eb114a6afd77ec94f1cf230f38e6caae5d7209474b285fc8"}`),
			Pubkey:                MustParseHex("0x9305db483ed03b526f0f70c6201a359e8becd3f584fe6ae52242e44346a5b4f7a74c29f8dbd981cbe885a4ce6b842a11"),
			Valid:                 true,
		},
		{
			CurrentForkVersion:    ZhejiangGenesisForkVersion,
			GenesisValidatorsRoot: ZhejiangGenesisValidatorsRoot,
			Msg:                   []byte(`{"message":{"epoch":"3541","validator_index":"62019"},"signature":"0x963723375cc200ce005b284f03a07cf45775b84cca94955b28c1d4a8d7dfcfc5d9d953a8b3c9f01e337146bff6328e79067edc58cc9ea0f2ef3fa96b48a01264b5bd719ff4177fd3de37689d3df9a599c6906b7cd1952a15dcd2bbb3a2c54341"}`),
			Pubkey:                MustParseHex("0x9305db483ed03b526f0f70c6201a359e8becd3f584fe6ae52242e44346a5b4f7a74c29f8dbd981cbe885a4ce6b842a11"),
			Valid:                 true,
		},
		{
			CurrentForkVersion:    PraterGenesisForkVersion,
			GenesisValidatorsRoot: ZhejiangGenesisValidatorsRoot,
			Msg:                   []byte(`{"message":{"epoch":"3541","validator_index":"62019"},"signature":"0x963723375cc200ce005b284f03a07cf45775b84cca94955b28c1d4a8d7dfcfc5d9d953a8b3c9f01e337146bff6328e79067edc58cc9ea0f2ef3fa96b48a01264b5bd719ff4177fd3de37689d3df9a599c6906b7cd1952a15dcd2bbb3a2c54341"}`),
			Pubkey:                MustParseHex("0x9305db483ed03b526f0f70c6201a359e8becd3f584fe6ae52242e44346a5b4f7a74c29f8dbd981cbe885a4ce6b842a11"),
			Valid:                 false,
		},
		{
			CurrentForkVersion:    ZhejiangGenesisForkVersion,
			GenesisValidatorsRoot: PraterGenesisValidatorsRoot,
			Msg:                   []byte(`{"message":{"epoch":"3541","validator_index":"62019"},"signature":"0x963723375cc200ce005b284f03a07cf45775b84cca94955b28c1d4a8d7dfcfc5d9d953a8b3c9f01e337146bff6328e79067edc58cc9ea0f2ef3fa96b48a01264b5bd719ff4177fd3de37689d3df9a599c6906b7cd1952a15dcd2bbb3a2c54341"}`),
			Pubkey:                MustParseHex("0x9305db483ed03b526f0f70c6201a359e8becd3f584fe6ae52242e44346a5b4f7a74c29f8dbd981cbe885a4ce6b842a11"),
			Valid:                 false,
		},
	}
	for _, test := range tests {
		Config.Chain.GenesisValidatorsRoot = test.GenesisValidatorsRoot
		var op *phase0.SignedVoluntaryExit
		err := json.Unmarshal(test.Msg, &op)
		if err != nil {
			t.Errorf("failed unmarshaling msg: %v", err)
		}
		err = VerifyVoluntaryExitSignature(op, MustParseHex(test.CurrentForkVersion), test.Pubkey)
		if test.Valid != (err == nil) {
			t.Errorf("wrong verify, should be valid: %v, err: %v", test.Valid, err)
		}
	}
}
