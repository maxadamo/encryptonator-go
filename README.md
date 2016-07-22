# WORK IN PROGRESS

The applications starts consuming a RabbitMQ queue called `encryptonator`. When a message arrives, it moves the files to the `queued` directory, run a goroutine to create an AES key, splits the file in chunks, merges and encrypts all chunks with the AES key, and publishes a message again to a different queue on rabbitMQ.

The message will be grabbed by a python daemon that will create the MD5, signed with asymmetric encryption the AES Key and upload the files to Sftp site.

## Setup GOPATH

```
mkdir ~/gopath

# add to .bashrc/.profile/.zshrc/... and reload
export GOPATH=~/gopath
export PATH=$GOPATH/bin:$PATH
```

## Build

```
mkdir -p $GOPATH/src/github.com/ecg-shared-technology
git clone https://github.com/maxadamo/encryptonator-go.git $GOPATH/src/github.com/ecg-shared-technology
go install github.com/ecg-shared-technology/encryptonator-go
```

## Run

```
encryptonator-go
```

