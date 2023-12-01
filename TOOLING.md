# Tooling

There are many tools which are used within this project. This document seeks to de-mystify many of the tools that are used (or which will be used) during this process.

## gRPC

### What is gRPC?

From their website (https://grpc.io/): 
```
gRPC is a modern open source high performance Remote Procedure Call (RPC) framework that can run in any environment.
```

gRPC can best be thought of as an alternative to HTTPS which uses Protobuffers for communication.

## Okay, what are Protobuffers?

From the protobuffers website (https://protobuf.dev/): 
```
Protocol Buffers are language-neutral, platform-neutral extensible mechanisms for serializing structured data.
```

In essence, you can use protobuffers to set up services within gRPC as well as pass types between languages. For examples of the protobuffer output in different languages, see the site above.

## Why are you using gRPC and Protobuffers? Why not HTTPS?

One of the main goals of this project is to be usable for many different languages (to allow for more clients to possibly be crafted). This would be difficult on the client-side developer, as the types returned by the server would not be defined in their language of choice. This is where protobuffers excel - they compile to many different languages, but roughly the same types. So, if I wanted to map a Hoplite, I may have something similar to this protobuffer: 
```protobuffer
message Hoplite {
  int32 health = 1;
  int32 attack = 2;
  string custom_name = 3;
  // more fields below
}
```

From this one protobuffer, you can derive a C++ "Hoplite" type: 
```C++
Hoplite hoplite;
fstream input(argv[1], ios::in | ios::binary);
hoplite.ParseFromIstream(&input);
health = hoplite.health();
name = hoplite.attack();
custom_name = hoplite.custom_name();
// more fields below
```

... or a Java builder: 
```Java
Hoplite hoplite = Hoplite.newBuilder()
    .setHealth(100)
    .setAttack(10)
    .setCustomName("Hoplite McHopliteFace")
    // more fields here
    .build();
output = new FileOutputStream(args[0]);
hoplite.writeTo(output);
```

... or any other language which Protobuffers support. 

This will allow Call to Power clients to be built with whichever language is desired by the authors.

## But what about HTTPS? Wasn't there a main goal to be cURL-able?

Yes! To do that, we will use our next technology...

## gRPC-Gateway

gRPC-Gateway allows you to serve both gRPC and RESTful requests at the same time. You can read more here: https://github.com/grpc-ecosystem/grpc-gateway

## What about the server architecture of this project? How will this be deployed?

There is information about the deployment process and server architecture within `ARCHITECTURE.md`.