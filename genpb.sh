#!/bin/bash

protoc -I order order/messaging/order.proto --go_out=plugins=grpc:order
protoc -I data data/messaging/data.proto --go_out=plugins=grpc:data