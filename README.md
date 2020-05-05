# bnc-eth

Blockchain Network Composer for Ethereum.
This project helps bootstraping a Parity network.

## Getting Started

* Build docker image
```
docker build --target prod_install -t bnc-eth:1.0 $PWD
```

* Run a single node
```
docker run -it --rm --name node bnc-eth:1.0 bootstrap --enroll org1 --validators 1 --privateKey b45c2d049b489a5d7f5a1b5212a0c262472a28b241e73e3e465d3133036a1c2f
```

* Run one hub, enroll two validators and connect one node
```
docker network create --driver=bridge --subnet=172.19.0.0/16 eth_legacy
docker run -d --name node1 --hostname node1 --network eth_legacy bnc-eth:1.0 bootstrap --enroll org1 --validators 2
docker run -d --name node2 --hostname node2 --network eth_legacy bnc-eth:1.0 join --enroll org2 --node node1:8080
docker run -d --name node3 --hostname node3 --network eth_legacy bnc-eth:1.0 join --enroll org3 node1:8080
docker run -d --name node4 --hostname node4 --network eth_legacy bnc-eth:1.0 connect node1:5000
```

* Run a node and connect it to an existing blockchain
```
docker run -it --rm --name peer bnc-eth:1.0 run --peers enode://9138679a3e670ba9a4544055f01ad59eee63eaab085e034f5ddacb7cef49eedb47f55781cd0c9c9b4e85137508184042df3954a14d0ffc892100c92f7a5d3f6b@node:30303 genesis.json
```

## Development

* Build docker image
```
docker build --target dev_install -t bnc-eth-dev $PWD
```

* Run go env
```
docker run -it --rm --name dev -v $PWD:/go/src bnc-eth-dev bash
$ go run cmd/app/main.go bootstrap --enroll org1 --validators 1 --privateKey b45c2d049b489a5d7f5a1b5212a0c262472a28b241e73e3e465d3133036a1c2f
```

* Update config templates (check that a file internal/box/blob.go has been added)
```
go generate ./...
```

