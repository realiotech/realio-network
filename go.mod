module github.com/realiotech/realio-network

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.45.4
	github.com/cosmos/ibc-go/v3 v3.0.0
	github.com/ethereum/go-ethereum v1.10.16
	github.com/gogo/protobuf v1.3.3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/ignite-hq/cli v0.20.3
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.11.0
	github.com/stretchr/testify v1.7.1
	github.com/tendermint/tendermint v0.34.19
	github.com/tendermint/tm-db v0.6.7
	github.com/tharsis/ethermint v0.15.0
	google.golang.org/genproto v0.0.0-20220525015930-6ca3db687a9d
	google.golang.org/grpc v1.46.2
)

require (
	github.com/google/go-cmp v0.5.8 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	golang.org/x/crypto v0.0.0-20220518034528-6f7dac969898 // indirect

)

replace (
	// Our cosmos-sdk branch is:  https://github.com/realiotech/cosmos-sdk v0.45.x-realio-beta-0.1
	github.com/cosmos/cosmos-sdk => github.com/realiotech/cosmos-sdk v0.45.2-0.20220510192910-9240cf6c999b
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
