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
Create the testnet directory
```
mkdir testnet
cd testnet
```
# Install the cbt tool
```
sudo apt remove google-cloud-cli
sudo apt install google-cloud-sdk-cbt
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
make misc
```
# Start postgres, redis & little_bigtable
```
cd ~/eth2-beaconchain-explorer/local-deployment/
docker compose up -d
```
Redis will be available on port 6379, little_bigtable on port 9000 and postgres on port 5432 (username `postgres`, password `pass`, db `db`)
# Initialize the bigtable schema
```
bash ~/eth2-beaconchain-explorer/local-deployment/init-bigtable.sh
```
# Start up the local testnet nodes
```
kurtosis run --enclave my-testnet github.com/kurtosis-tech/eth-network-package
```
Take note of the http API ports of the el & cl clients.
# Copy the example config file and add the el & cl client ports
```
cp ~/eth2-beaconchain-explorer/local-deployment/testnet-config-example.yml ~/testnet/explorer-config.yml
```
# Initialize the db schema
```
BIGTABLE_EMULATOR_HOST="127.0.0.1:9000" ~/eth2-beaconchain-explorer/bin/misc -config ~/testnet/explorer-config.yml -command applyDbSchema
```
# Start the indexer
```
BIGTABLE_EMULATOR_HOST="127.0.0.1:9000" ~/eth2-beaconchain-explorer/bin/explorer -config ~/testnet/explorer-config.yml
```
# Start the frontend-data-updater
```
BIGTABLE_EMULATOR_HOST="127.0.0.1:9000" ~/eth2-beaconchain-explorer/bin/frontend-data-updater -config ~/testnet/explorer-config.yml
```
# Start the frontend
```
BIGTABLE_EMULATOR_HOST="127.0.0.1:9000" ~/eth2-beaconchain-explorer/bin/frontend-data-updater -config ~/testnet/explorer-config.yml
```
