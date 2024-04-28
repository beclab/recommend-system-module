# system workflow

Provider sync and crawler task.

# Directory structure
```
system workflow
|-- api                  # knowledge api     
|-- common               # public module
|-- crawler              # crawler module
|-- model                
|-- protobuf_entity      # protobuf data format       # 
|-- storge               # data access module 
|-- sync                 # sync module

```

## sync
### environment
```
export TERMINUS_RECOMMEND_MONGODB_URI="mongodb://localhost:27017"
export TERMINUS_RECOMMEND_MONGODB_NAME="document"
export TERMINUS_RECOMMEND_REDIS_ADDR="feeds"
export TERMINUS_RECOMMEND_REDIS_PASSOWRD="entries"
export NFS_ROOT_DIRECTORY="127.0.0.1:8080"
export KNOWLEDGE_BASE_API_URL="http://127.0.0.1:3010/rss"
export KNOWLEDGE_BASE_API_URL="http://localhost:3010/knowledge/feed/algorithm/"
export TERMIUS_USER_NAME="user1"
```

### local run
```
docker run  --name  sync  -v /tmp/data/nfs:/nfs -v /tmp/data/juicefs:/juicefs -e TERMINUS_RECOMMEND_REDIS_ADDR=$TERMINUS_RECOMMEND_REDIS_ADDR  -e TERMINUS_RECOMMEND_REDIS_PASSOWRD=$TERMINUS_RECOMMEND_REDIS_PASSOWRD -e NFS_ROOT_DIRECTORY=$NFS_ROOT_DIRECTORY -e JUICEFS_ROOT_DIRECTORY=$JUICEFS_ROOT_DIRECTORY -e FEED_MONGO_API_URL=$FEED_MONGO_API_URL  -e ALGORITHM_FILE_CONFIG_PATH=$ALGORITHM_FILE_CONFIG_PATH  --net=host -d  beclab/recommend-sync
```
## crawler
### environment
```
export KNOWLEDGE_BASE_API_URL="http://localhost:3010/knowledge/feed/algorithm/"
export TERMIUS_USER_NAME="user1"
```

### local run
```
docker run  --name  crawler  -e TERMINUS_RECOMMEND_MONGODB_URI=$TERMINUS_RECOMMEND_MONGODB_URI -e TERMIUS_USER_NAME=$TERMIUS_USER_NAME --net=host -d  beclab/recommend-crawler
```

    
