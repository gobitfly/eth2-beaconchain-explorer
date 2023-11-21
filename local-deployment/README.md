This guide outlines how to deploy the explorer using a local lh-geth testnet. Utilized postgres, redis and little_bigtable as data storage

# Install docker
```
sudo apt update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
```
# Install kurtosis-cli
```
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt install kurtosis-cli
```
# Install golang
```
wget https://go.dev/dl/go1.20.7.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.20.7.linux-amd64.tar.gz
```
Add the golang binaries to the path by adding the following lines to your ~/.profile file and then logout & login again
```
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin
```
# Clone the explorer repository
```
cd ~/
git clone https://github.com/gobitfly/eth2-beaconchain-explorer.git
cd eth2-beaconchain-explorer
```
# Build the explorer binaries
```
sudo apt install build-essential
make all
```
# Start postgres, redis, little_bigtable & the eth test network
```
cd ~/eth2-beaconchain-explorer/local-deployment/
kurtosis clean -a && kurtosis run --enclave my-testnet . "$(cat network-params.json)"
```
# Generate the explorer config file for the deployed testnet
```
cd ~/eth2-beaconchain-explorer/local-deployment/
bash provision-explorer-config.sh
```
This will generate a config.yml to be used by the explorer and then create the bigtable & postgres schema

# Start the explorer modules
```
cd ~/eth2-beaconchain-explorer/local-deployment/
docker-compose up -d
```
You can start / stop the exporter submodules using `docker-compose`

# Convenience-script run.sh

The `run.sh` script in this directory can be used to start and stop everything. Just run `./run.sh start` to start the local chain and the explorer, then browse http://localhost:8080 to see it in action. You can run `./run.sh sql` to explore the sql-database. Everything can be stopped and cleaned up with `./run.sh stop`.

# Exit validators
Exiting individual validators can be done using the provided `exit_validator.sh` script. Requires [https://github.com/wealdtech/ethdo](ethdo) to be available on the path.
```
bash exit_validators.sh -i validator_index -m "memonic" -b "http://bn_api_host:bn_api_port"
```

# Enabling withdrawals
The script `add_withdrawal_address.sh` allows you to create & submit a bls to execution layer address change message in order to enable withdrawals for specific validators.. Requires [/github.com/protolambda/eth2-val-tools](eth2-val-tools) to be available on the path.
```
go get github.com/protolambda/eth2-val-tools@master
go install github.com/protolambda/eth2-val-tools@master
bash add_withdrawal_address.sh -a "EL address (0x1234...)" -i validator_index -m "memonic" -b "http://bn_api_host:bn_api_port"
```
