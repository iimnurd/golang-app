############################
# STEP 1 build executable binary
# pull official base image
############################
FROM golang:1.15-alpine3.12 as builder

# Install dependencies
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
# tzdata is for timezone info
RUN apk update && \
    apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

# Fetch dependencies.
# Using go mod with go > 1.11
# will also be cached if we won't change mod/sum
WORKDIR /app
ENV GO111MODULE=on
COPY go.mod go.sum ./

RUN go mod download 
RUN go mod verify

# Copy the source code
# Build the binary
COPY . ./
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-w -s" \
    -o /app/main 

############################
# STEP 2 build a small image
############################

# Dynatrace can't do deep scanning in alpine
FROM debian:buster-20200908-slim

# Create appuser
WORKDIR /app
ENV USER=appuser
ENV UID=10001
# See https://stackoverflow.com/a/55757473/12429735RUN 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

# Copy our static executable
COPY --from=builder --chown=appuser:appuser /app/main ./main



# Use an unprivileged user.
USER appuser:appuser

# Tell docker how the process PID 1 handle gracefully shutdown
# Signal Interupt for gracefully shutdown echo server
STOPSIGNAL SIGINT

# Tell if our container will open this port
# Set app to use this port too
EXPOSE 9000

# There are 2 different method how docker run our program
# Shell form: `CMD command param1 param2` equivalent with 
#             `/bin/sh -c command param1 param2`
# Exec form: `CMD ["command", "param1", "param2"]`
#
# Since shell form running our program inside sh, so the PID 1
# is sh, not our program. Hence, our program never receive stop signal
# So we will use exec form.

# Entrypoint is how docker will execute our program by default
# CMD is the default param and can be replaced by docker params, even though
# we can use it to execute our program.

ENTRYPOINT ["./main"]