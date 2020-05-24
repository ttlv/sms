# 1.What is it?

This is a sms service which is a micro service,it can be used in your service for decoupling your service,Golang is the main language for this project.Http and RPC are supported,So you can choose any method you like.

# 2.How to run it

## 1. First Step Initialize Your Env 🎁

You can find the dev_env file in this project,let's have a look.

### 1. SMS_ADDRESS

This is a address for your server,port is very important,don't forget🙂

### 2. SMS_DB

This is the database address for sms service,I suggest use mysql.🚀

#### 3. SMS_AMQPDIAL

This is the address for rabbitMQ,We use RabbitMQ as message middleware.😈

### 4.  SMS_APITOKEN & SMS_APIKEY

Both of them are for my Monitor Server,Because I use PushOver to notice me that if the sms service is healthy.😃

### 5.  SMS_Port

It's the port for Webhook service,and what's Webhook?it's just used to handle the api server's callback.🥰

```bash
source dev_env
## path: github.com/ttlv/sms
```

## 2. Use Docker-Compose To Run Some Dependent Services 🐋

````bash
docker-compose -f docker-compose.yml up -d
````

### Notice:

we use mysql and rabbitmq as other dependent services and you can see them in the dcoker-compose file.

Use docker ps -a command to make sure dependent services are available.

## 3. How To Run All Of The Servers

### 1. Http Server

```bash
cd $GOPATH/github.com/ttlv/sms/service/producer_http/main
go run main.go
```

### 2. GRPC Server

```
cd $GOPATH/github.com/ttlv/sms/service/producer_grpc/main
go run main.go
```

### 3. Run Consumer Server

```
cd $GOPATH/github.com/ttlv/sms/service/consumer
go run main.go
```

### 3. Run Webhook Server

````
cd $GOPATH/github.com/ttlv/sms/webhook/main
go run main.go
````

### 5. Run Monitor Server

````
cd $GOPATH/github.com/ttlv/sms/monitor
go run main.go
````

### 6. Run Admin Server

please jump to https://github.com/ttlv/sms_admin to have a look.

# 4. API Doc

## 1. Requset Method

### Method: POST

## 2. Request Params

### 1.  HTTP Headers

#### Authorization:	XXXXXXXXXXXXXXXXX**(required)**

### 2. HTTP BODY

#### 1. brand: XXXX**(required)**

#### 2. phone: XXXX**(required)**

#### 3. content: XXXX(required)




