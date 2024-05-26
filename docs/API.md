# API ROUTES

Here are the API routes that the peer node exposes to the front end. 

## PUT /add-job

Request

{ fileHash: string, peerId: string }

Response 

{ jobID: string }

## GET /find-peer?fileHash=

Request

fileHash: string

Response 

[{ peerId: string, ip: string, region: string, price: float32 }]

## GET /job-list

Response 

[{ fileHash: string jobId: string, timeQueueued: string, status: string, accumulatedCost: string, projectedCost: string, eta: int, peerId: string }]

## GET /job-info?jobId=

Request

JobId: string

Response 

{ fileHash: string jobId: string, timeQueueued: string, status: string, accumulatedCost: string, projectedCost: string, eta: int, peerId: string }

## POST /start-jobs

Request

[{ jobId: string }]

Response 

{ status: string }

## POST /pause-jobs

Request

[{ jobId: string }]

Response 

{ status: string }

## POST /terminate-jobs

Request:

[{ jobId: string }]

Response 

{ status: string }

## GET /get-history

Response 

[{ fileHash: string jobId: string, timeQueueued: string, status: string, accumulatedCost: string, projectedCost: string, eta: int, peerId: string }]

## POST /remove-from-history
  
Request

{ jobId: string }

Response 

{ status: string }

## POST /clear-history 

Request

{ }

Response 

{ status: string }

## GET /file/:hash/info

Request

hash: string

Response 

{ name: string, size: int, numberOfPeers: int, listProducers: []string }

## POST /upload

Request

{ hash: string }

Response 

{ filePath: string, price: int64 }

## DELETE /file/:hash

Request

hash: string

Response 

{ status: string }

## GET /get-peer?peer-id=

Request

peer-id: string

Response 

{ location: string, latency: string, peerId: string, connection: string, openStream: string, flagUrl: string }

## GET /get-peers

Request

Response 

[ { location: string, latency: string, peerId: string, connection: string, openStream: string, flagUrl: string } ]

## POST /remove-peer

Request

{ peerId: string }

Response 

{ status: string }

## GET /wallet/balance

Request

{ }

Response 

{ }

## GET /wallet/revenue/daily

Request

{ }

Response 

{ }

## GET /wallet/revenue/monthly

Request

{ }

Response 

{ }

## GET /wallet/revenue/yearly

Request

{ }

Response 

{ }

## GET /wallet/transactions/complete

Request

{ }

Response 

{ }

## POST /wallet/transfer

Request

{ }

Response 

{ }

## GET /stats/network

Request

{ pub_key: string }

Response 

{ _id: string, pub_key: string, incoming_speed: string, outgoing_speed: string }

## GET /activity

Request

{ }

Response 

{ }


## The routes below are not associated with any front-end functionality. They are mainly intended for peer to peer backend communication.

## GET /requestFile/:filename

Request

fileName: string

Response

bytes, representing parts of a file

## GET /sendTransaction

## GET /writeFile

## GET /sendMoney

## GET /getLocation

## GET /getAllStored

## GET /get-file

## GET /upload-file

## GET /delete-file