# backend server

Provide feeds update , article extract and search functions.

# Directory structure
```
server
|-- api                  # api service
|   |-- entry
|   |-- feed
|   |-- search
|-- cli                   
|-- common               # public module
|-- crawler              # entry crawler service
|-- database                
|-- http                 # 
|-- knowledge            # knowledge api
|-- model                
|-- reader               # parse entry data from different feed types 
|-- scheduler            # select schedule feed to update
|-- service              # process feed and entry data
|-- storage              # save data to db 
|-- worker               # work pool 

```

# Run
```
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_NAME="document"
export MONGODB_FEED_COLL="feeds"
export MONGODB_ENTRY_COLL="entries"
export LISTEN_ADDR="127.0.0.1:8080"
export RSS_HUB_URL="http://127.0.0.1:3010/rss"
export WE_CHAT_REFRESH_FEED_URL="http://127.0.0.1:8080/api/wechat/entries"
go run main.go
```

    
