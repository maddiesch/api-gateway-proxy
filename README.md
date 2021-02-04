# API Gateway HTTP Proxy

## Project Setup

### main.go

```golang
package main

import (
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/gin-gonic/gin"
  proxy "github.com/maddiesch/api-gateway-proxy"
)

var (
  app *gin.Engine
)

func init() {
  app = gin.New()
}

func main() {
  lambda.Start(proxy.Handler(app))
}

```

### template.yml

```yaml
Transform: AWS::Serverless-2016-10-31
Globals:
  Function:
    Runtime: provided.al2
    Timeout: 30
    MemorySize: 256
Resources:
  ApiResource:
    Type: AWS::Serverless::HttpApi
  AppFunctionHandler:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: bin/app_handler/
      Handler: bootstrap
      Events:
        RootRequestEvent:
          Type: HttpApi
          Properties:
            ApiId: !Ref ApiResource
            Method: any
            Path: /
        PathRequestEvent:
          Type: HttpApi
          Properties:
            ApiId: !Ref ApiResource
            Method: any
            Path: /{path+}
```
