# GO version distributed configuration center confcarrier

[![License](http://img.shields.io/:license-apache-brightgreen.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![Build Status](https://travis-ci.com/data2/confcarrier.svg?branch=master)](https://travis-ci.com/data2/confcarrier)

GO version distributed configuration center confcarrier
+ Support end-to-end real-time communication, reduce resource consumption with long connections
+ Can meet the rich configuration operations of the business
+ Publish and subscribe mode - the server listens to changes in portal configuration and broadcasts message notifications to the client
+ Equipped with backend page management and web-based operation configuration

# architecture

![image](架构图.png)

# Communication between portal and confcarrier

+ queue message
+ cache
+ make-one-server

If you are a small project, you can use the code of the make one server branch to aggregate the server and portal into one service https://github.com/data2/confcarrier/tree/make-one-big-server

# use
### start serve
```
go run carrier.go  util.go db.go queue.go tcpPort  mysqlUrl redisUrl
```
### start distributed client 
```
go run client.go  port  
```
###  start portal
 ```
 go run portal.go  util.go db.go queue.go port  mysqlUrl redisUrl
 ```
