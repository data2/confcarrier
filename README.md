# confcarrier

[![License](http://img.shields.io/:license-apache-brightgreen.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![Build Status](https://travis-ci.org/data2/confcarrier.svg?branch=master)](https://travis-ci.org/data2/confcarrier)

GO版本分布式配置中心，支持端到端的实时通信进行配置操作，同时具备后台页面管理的功能，并能对监听指定命名空间下的配置的客户端进行消息广播通知，类似于热更新。

# architecture

![confcarrier](https://user-images.githubusercontent.com/13504729/131481175-3f4f0776-79a9-4c2c-aef7-73c533c21004.png)

# portal与confcarrier通信

+ queue message
+ cache
+ make-one-server

如果您是小型项目，可以使用make-one-server分支的代码，服务端和portal聚合为一个服务 https://github.com/data2/confcarrier/tree/make-one-big-server

