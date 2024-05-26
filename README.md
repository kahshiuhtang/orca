# orcanet-go

Main Repo for Orcanet-Go 🐳

Orcanet is a peer-to-peer file-sharing service. The project allows users to send and recieve file chunks, identified through file hashes, in exchange for the OrcaCoin cryptocurrency.

## Background Information

From the base directory, there are three folders: coin, peer and protos. The coin folder contains the blockchain node and blockchain wallet node code that will allow us to communicate with the blockchain. The peer folder contains code for the peer node aspect. Running the peer node will automatically start the two blockchain-related nodes. When the peer node runs, a CLI will appear. This allows for you to transfer files around without needing a front-end application. Running the peer node will also expose some HTTP API routes that allow the front-end GUI to perform interact with the network.

## Dependency Installation

* Make sure golang is installed: The version used for this project is 1.22.3, but any later version should work. The steps can be found [HERE](https://go.dev/doc/install), but as a quick guide for Linux users:

```bash
wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz;

rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz;
```

To add go to the PATH environment variable, dependening on your terminal, will look like:

```bash
nano $HOME/.profile
```

Then, paste this line at the bottom of the file:

```bash
export PATH=$PATH:/usr/local/go/bin
```

Finally, you will do CTRL-X, Y, ENTER to escape nano

Then do:

```bash
source $HOME/.profile

go version
```
This should display the correct go version, which if you followed these steps, should be amd1.22.3. To ensure that everything is setup correctly, close your terminal and run ```bash go version ``` again.



* Information about installing proto buffer compiler is found [HERE](https://grpc.io/docs/protoc-installation/)

The basic steps are:

```bash
apt install -y protobuf-compiler make

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

export PATH="$PATH:$(go env GOPATH)/bin"
```

## Running

First generate the gRPC files required. Make sure you are in the peer folder of the project and run the command below.

``` bash
cd peer;

protoc --go_out=./internal/fileshare \
  --go_opt=paths=source_relative \
  --go-grpc_out=./internal/fileshare \
  --go-grpc_opt=paths=source_relative \
  --proto_path=../protos/fileshare \
  file_share.proto
```

Second, make sure you create the executable for the bitcoin nodes. It can be done from the root directory as follows:

```bash
cd ../coin

./build.sh
```

If you run into some issues about the build file not being exectuable, you can always run the following command and retry running the shell file.

```bash
chmod +x ./build.sh
```

At the end of this script, you will be prompted to create a wallet for the blockchain. It doesnt really matter which options you select, just remember your wallet seed and the passphrase you use.

Finally, return the the peer folder and run make. This will run the Go backend peer node.

```bash
cd ../peer

make all
```

This will start up the peer node. You should see output in the terminal. You will need to enter in <i>three</i> numbers into the terminal before the peer node is fully running. These three numbers will be the port numbers used by the peer node to connect with various services. There is no agrred upon port number, but currently, these three ports can be the official

## TODO

Here are some things that can be implemented in the future / improved upon the existing repository.

1. **DHT Bad Address Connection** 
    - Implement the Sea Dolphins method of trying to reconnect to a peer on a bad address 3 times and then removing it from the peer's address book. 
2. **NAT Address Translation** 
    - Right now the peer node will join the DHT in client mode by default, which will only allow it to send out queries and not respond to them. The desired functionality is to join the DHT as a server node automatically if it can be determined that we can reach the node behind the NAT.
3. **NAT Holepunch / Relay**
    - Peer to peer functionality does not work due to hosts being behind a NAT. Maybe one of these libariees [1](https://github.com/malcolmseyd/natpunch-go) or [2](https://github.com/shawwwn/Gole). I think there was also an attempt to use relays
4. **Lightning Network**
    - Need to implement a network on top of the existing blockchain to speed up transactions. 
5. **Coin Transfer Verification**
    - There is not any system in place that will check to see if a transaction has gone through during the transfer of a file. 
6. **Reputation System**
    - Keep track of the good and bad actors in the network. 
7. **Producer Peer Disconnect**
    - Not sure how to title this, but when one peer disconnects, there should a mechanism where the connection automatically switches to another peer (preferably the cheapest priced option) to continue the file transfers
8. **Documentation Update**
    - Fix any inconsistencies in API routes or CLI. In addition, it would be nice to document the internal functions.
9. **Switch to use a CLI library**
    - Cobra CLI is an option. There is just a for loop that takes input. This is not ideal.
10. **Tracking activity**
    - It would be nice for the front-end teams to be able to display some visualization of what cryptocurrency has been sent around from this peer. In addition, more work could be done to improve the tracking of what type of files are being sent around but this could be tricky, as files as sent in hashes.
11. **Add unit tests**
    - We was lackadaisical on creating unit tests for the functions. Hopefully we will add in new functions to help ensure the validity of our code.