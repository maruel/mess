# Third party code

I'll want to replace these later. Leveraging to get a MVP sooner.

## swarming_bot

- URL: https://chromium.googlesource.com/infra/luci/luci-py
- Commit: 9dc0a10f2d7d03e666052838354ad2ec35be9443.

```
rsync -Lav <ROOT>/luci-py/appengine/swarming/swarming_bot .
```

## ui2

- URL: https://chromium.googlesource.com/infra/luci/luci-py
- Commit: 9dc0a10f2d7d03e666052838354ad2ec35be9443.

```
rsync -Lav <ROOT>/luci-py/appengine/swarming/ui2 .
```

## remote

- URL: https://github.com/bazelbuild/remote-apis
- Commit: 04784f4a830cc0df1f419a492cde9fc323f728db

```
rsync -Lav <ROOT>/remote-apis/build .
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/googleapis/googleapis@latest
# Patch all the go_options lines in the proto files to be relative to this
# directory.
find build -name "*.pb.go" -delete
protoc --proto_path=. --go_out=. \
  -I=$GOPATH/pkg/mod/github.com/googleapis/googleapis@v0.0.0-20220329183022-151e02bdb281 \
  --go_opt=paths=source_relative \
  build/bazel/remote/asset/v1/remote_asset.proto \
  build/bazel/remote/execution/v2/remote_execution.proto \
  build/bazel/remote/logstream/v1/remote_logstream.proto \
  build/bazel/semver/semver.proto
```
