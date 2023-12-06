This guide outlines how to deploy the explorer using a local lh-geth testnet. Utilized postgres, redis and little_bigtable as data storage

# Install docker
If you never worked with Docker, [this short video](https://www.youtube.com/watch?v=rOTqprHv1YE) gives an overview to understand roughly what we will do with it.

Now, let us install it:
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

# Install kurtosis-cli
Kurtosis is a software which will launch inside your computer the different parts of a test network and Bitfly's explorer, using Docker. You will not have to deal with it (nor with Docker), because automating the launch of interdependant modules with Docker and configuring them is the point of Kurtosis. [This short video](https://www.loom.com/share/4256e2b84e5840d3a0a941a80037aebe) gives an overview if it is your first time.

Now, let us install it:
```
echo "deb [trusted=yes] https://apt.fury.io/kurtosis-tech/ /" | sudo tee /etc/apt/sources.list.d/kurtosis.list
sudo apt update
sudo apt install kurtosis-cli
```

# Install golang
You will find the last version of Go [on this page](https://go.dev/doc/install). The commands that you will type to install it will look like this:

```
wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
```
Add the golang binaries to the path by adding the following lines to your _~/.profile_ file and then logout & login again.
```
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin
```
The second line is not mentionned in the installation instructions of Go's website but will be necessary for our system.
Before continuing, restarting your computer now might save you from unexplained errors during the next steps.

# Clone the explorer repository
```
cd ~
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
Later in your developer life (after having started Kurtosis and stopped it a few times), if you encounter an error at this step, you might need to clean up bugged cache files from previous runs that Kurtosis or Docker left behind.
The `./stop` script [in this repository](https://github.com/thib-wien/scripts-localnetworkandexplorer) gathers cleaning commands which worked for their author (it might save you hours of browsing Stack Overflow and GitHub's issues).

# Generate the explorer config file for the deployed testnet
```
cd ~/eth2-beaconchain-explorer/local-deployment/
bash provision-explorer-config.sh
```
This will generate a _config.yml_ to be used by the explorer and then create the bigtable & postgres schema.

# Start the explorer modules
```
cd ~/eth2-beaconchain-explorer/local-deployment/
docker compose up -d
```
You can start / stop the exporter submodules using `docker compose`

# Convenience-script run.sh
Above, we have started / stopped the local chain + the explorer manually. The `run.sh` script in this directory can be used to start and stop everything automatically. Just run `./run.sh start` to start the whole system, wait a bit and browse http://localhost:8080 to see it in action. You can run `./run.sh sql` to explore the sql-database. Everything can be stopped and cleaned up with `./run.sh stop`.

# Exit validators
Exiting individual validators can be done using the provided `exit_validator.sh` script. Requires [https://github.com/wealdtech/ethdo](ethdo) to be available on the path.
```
bash exit_validators.sh -i validator_index -m "memonic" -b "http://bn_api_host:bn_api_port"
```

# Enabling withdrawals
First, install _JQ_ and _eth2-val-tools_:
```
sudo apt install jq
go get github.com/protolambda/eth2-val-tools@master
go install github.com/protolambda/eth2-val-tools@master
```
To enable withdrawals for specific validators in your local network, we provide the script `add_withdrawal_address.sh`. It creates and submits a BLS-to-execution-layer-address-change message.
The script needs some arguments: 
```
cd ~/eth2-beaconchain-explorer/local-deployment/scripts
bash add_withdrawal_address.sh -a "ETH address" -m "mnemonic" -b "URL" -i validator_index
```
To fill the address field, you will find address generators with Google but you can also peek any existing address, like your own to enjoy watching it receive huge rewards locally.
To fill the mnemonic field, copy it from our file _network-params.json_.
Replace `bn_api_port` with the port that Kurtosis maps the http port of the consensus client to. This port is written on the console by Kurtosis when it starts, on this line: `cl-1-lighthouse-geth http: ...../tcp -> http://127.0.0.1:XXXXX` (what you want is represented by XXXXX).
Finally, `validator_index` identifies the local validator that you want to activate withdrawals for. Give a number between 1 and 64.

Here is an example:
```
cd ~/eth2-beaconchain-explorer/local-deployment/scripts
bash add_withdrawal_addr.sh -a "0x0701BF988309bf45a6771afaa6B8802Ba3E24090" -m "giant issue aisle success illegal bike spike question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy very lucky have athlete" -b "http://localhost:32779" -i 42
```