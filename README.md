# NATS session in Go

## Setup

To have a distinct setup per partipant, it's important to call `00-setup/main.go` initially. 
This will create a .env file with `NATS_USER` and `NATS_SERVER` set. 

## Important

* The examples will work only, if the host has the correct setup running (check https://github.com/lzaugg/nats-session-setuper).

* The server is setup with token based auth, to simplify the setup process. 
    The token is the same as the `session-id` (see `setup-00`) and the user name is `gopher`.
    All participants will use the same token and there are no security measures against misuse (intentional or unintentional) of the message broker!
