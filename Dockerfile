FROM golang:1.17.2 as deps

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /src

# Copy and download dependencies
COPY go.mod .
COPY go.sum .
# does not add new requirements or update existing requirements.
RUN go mod download 

# Copy all code into container
COPY . .

# using deps now
FROM deps as build

# build all binaries
RUN go build -o cats ./...

# set a dist directory
WORKDIR /dist

# copy generated binaries individually to dist
RUN cp /src/cats .

# run the binary on 8080
FROM scratch
WORKDIR /root/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /dist/cats ./

EXPOSE 8080
CMD ["./cats"]