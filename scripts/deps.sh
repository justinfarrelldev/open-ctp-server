#!/bin/sh

# This file is meant to grab the necessary dependencies for this project (including Taskfile and protoc).
# You will need bash to run this file. Linux users should have bash installed by default as a GNU system
# util, but if you are on Windows you may need to download and work with Git Bash. If this does not work
# with your system, please feel free to open a GitHub issue!

ls ./bin

RESULT=$?

if [ $RESULT -gt 0 ]; then
    echo Creating bin folder to hold the binaries...
    mkdir ./bin
fi

ls ./proto_out

RESULT=$?

if [ $RESULT -gt 0 ]; then
    echo Creating proto_out folder to hold the built protos...
    mkdir -p ./proto_out/cpp
    mkdir -p ./proto_out/java
    mkdir -p ./proto_out/csharp
    mkdir -p ./proto_out/kotlin
    mkdir -p ./proto_out/python
    mkdir -p ./proto_out/go
fi

echo Getting the Taskfile install script...
curl -o ./bin/taskfile-install.sh https://taskfile.dev/install.sh

echo Making the Taskfile installer executable...
chmod +x ./bin/taskfile-install.sh

echo Executing the Taskfile installer...
./bin/taskfile-install.sh

echo Removing the Taskfile installer...
rm ./bin/taskfile-install.sh

echo Task is in ./bin. Either add this manually to your PATH variable or run the Task executable with "./bin/task [arguments]".

OS=$(uname)
case $OS in
  'Linux')
    OS='Linux'
    curl -L -o ./bin/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip
    ;;
  'WindowsNT')
    OS='Windows'
    curl -L -o ./bin/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-win64.zip
    ;;
  'Darwin') 
    OS='Mac'
    curl -L -o ./bin/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-osx-universal_binary.zip
    ;;
  *) 
    echo Please manually grab the most relevant release of Protoc v25.1 here and add it to your ./bin folder: https://github.com/protocolbuffers/protobuf/releases
    exit 0
    ;;
esac

unzip -o ./bin/protoc.zip -d ./bin/protoc-out

rm ./bin/protoc.zip

mv ./bin/protoc-out/bin/protoc ./bin/protoc

mkdir -p /usr/local/include/google

mv ./bin/protoc-out/include/google /usr/local/include/google

echo Making the Protoc installer executable...
chmod +x ./bin/protoc-out/bin/protoc

echo Protoc is in ./bin. Either add this manually to your PATH variable or run the Protoc executable with "./bin/protoc [arguments]".

echo Finished! It is recommended to add the bin folder to your PATH within your .bashrc, .zshrc, etc. to make these binaries available by their name. There is an explanation of this process in README.md.
