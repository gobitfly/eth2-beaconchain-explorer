This guide outlines how to deploy the explorer using a local lh-geth testnet. Utilized postgres, redis and little_bigtable as data storage

# Install docker
```
sudo apt update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
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
# Clone the lh repo
```
git clone https://github.com/sigp/lighthouse.git 
cd lighthouse
```
```
# setup rust dev environment
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh 
source "$HOME/.cargo/env"
```
# Install lh build deps
```
sudo apt install -y git gcc g++ make cmake pkg-config llvm-dev libclang-dev clang protobuf-compiler jq
```
# Build & install lighthouse
```
make
```
# Build & install lcli
```
make install-lcli
```
# Download and install geth & bootnode binary
```
cd ~
wget https://gethstore.blob.core.windows.net/builds/geth-alltools-linux-amd64-1.12.2-bed84606.tar.gz
tar -xf geth-alltools-linux-amd64-1.12.2-bed84606.tar.gz
sudo cp geth-alltools-linux-amd64-1.12.2-bed84606/geth /usr/bin/
sudo cp geth-alltools-linux-amd64-1.12.2-bed84606/bootnode /usr/bin/
```
# Clone the explorer repository
```
cd ~/
git clone https://github.com/gobitfly/eth2-beaconchain-explorer.git
cd eth2-beaconchain-explorer
```
# Build the explorer binaries
```
make all
make misc
```
# Clone the little_bigtable repository
```
cd ~/
git clone https://github.com/gobitfly/little_bigtable.git
cd little_bigtable
```
# Build the little_bigtable binary
```
make
```
# Start postgres & redis
```
cd ~/eth2-beaconchain-explorer/local-deployment/
docker compose up -d
```
Redis will be available on port 6379 and postgres on port 5432 (username `postgres`, password `pass`, db `db`)
# Start little_bigtable
```
# in a new terminal
~/little_bigtable/build/little_bigtable -db-file ~/testnet/bigtable.db
```
lbt will be available on http://127.0.0.1:9000
# Initialize the bigtable schema
```
bash ~/eth2-beaconchain-explorer/local-deployment/init-bigtable.sh
```
# Start up the local testnet nodes
## Switch to the lighthous scripts directory
```
cd testnet/lighthouse/scripts/local_testnet/
```
## Start the local testnet
```
cd testnet/lighthouse/scripts/local_testnet/
./start_local_testnet.sh -v 2 genesis.json
```
# Initialize the db schema
```
BIGTABLE_EMULATOR_HOST="127.0.0.1:9000" ~/eth2-beaconchain-explorer/bin/misc -config ~/eth2-beaconchain-explorer/local-deployment/testnet-config.yml -command applyDbSchema
```
# Start the indexer
```
BIGTABLE_EMULATOR_HOST="127.0.0.1:9000" ~/eth2-beaconchain-explorer/bin/explorer -config ~/eth2-beaconchain-explorer/local-deployment/testnet-config.yml
```
