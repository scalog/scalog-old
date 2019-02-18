#!/bin/bash

protoc -I proto proto/client.proto --go_out=plugins=grpc:messages
protoc -I proto proto/order.proto --go_out=plugins=grpc:messages
protoc -I proto proto/data.proto --go_out=plugins=grpc:messages
