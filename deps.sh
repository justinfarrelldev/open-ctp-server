#!/bin/sh

# This file is meant to grab the necessary dependencies for this project (including Taskfile and protoc).
# You will need bash to run this file. Linux users should have bash installed by default as a GNU system
# util, but if you are on Windows you may need to download and work with Git Bash. If this does not work
# with your system, please feel free to open a GitHub issue!

ls ./bin

RESULT=$?

echo $RESULT

if [ $RESULT -gt 0 ]; then
    echo Creating bin folder to hold the binaries...
    mkdir ./bin
fi

echo Getting the Taskfile install script...
curl -o ./bin/taskfile-install.sh https://taskfile.dev/install.sh

echo Making the Taskfile installer executable...
chmod +x ./bin/taskfile-install.sh

echo Executing the Taskfile installer...
./bin/taskfile-install.sh

echo Task is in ./bin. Either add this manually to your PATH variable or run the Task executable with "./bin/task [arguments]".

echo Finished!