sequenceDiagram
autonumber
participant App as Application
participant Client as Client
participant Req as Request
participant Core as execFunc
participant FH as fasthttp.Client (shared)
participant Temp as fasthttp.Client (temp)
participant Resp as Response

App->>Client: Build Request
App->>Req: SetStreamResponseBody? / other config
App->>Core: Execute(req)

Core->>Client: read client.streamResponseBody
Core->>Req: req.StreamResponseBody()
alt request override differs
Core->>Core: clone shared fasthttp.Client -> Temp\nset Temp.StreamResponseBody = request value
Core->>Temp: Do / DoRedirects(req)
Temp-->>Core: RawResponse
else matches
Core->>FH: Use shared fasthttp.Client.Do / DoRedirects(req)
FH-->>Core: RawResponse
end

Core-->>Resp: wrap RawResponse
App->>Resp: BodyStream() or Body()
alt BodyStream available
Resp-->>App: io.Reader (stream)
else fallback
Resp-->>App: bytes.Reader over Body()
end
