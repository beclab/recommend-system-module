# Recommend System Module

Provide recommend system services   

## Table of Contents
- [system-workflow](#system-workflow)
- [backend-server](#backend-server)
- [argo-task](#argo-task)

## system-workflow
This part of the code is about the system workflow of the recommend,including data synchronization(sync) and data crawler.
- According to the data source configured in the algorithms,data synchronization synchronizes data from the cloud regularly.
- According to the needs of the algorithms,data crawler regularly crawls the raw content from the Internet.

more detail system-workflow/README.md


## backend-server
This part of the code is about feed update and article extractor of the library.

more detail backend-server/README.md


## argo-task
This part of the code is to submit sync task and crawler task to the argo workflows.

more detail argo-task/README.md
