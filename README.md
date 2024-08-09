# GoShort-CLI

GoShort is a project I've put together for a software jam.

# What is GoShort-CLI

GoShort-CLI is the command line interface for interacting with a GoShort! server.  
GoShort is essentially a URL shortener for devs, we don't like web UI's :P we like sticking to our terminal.

# How to install

Run this one-liner:
```shell
some curl # windows in Powershell
some curl # linux
no mac T_T imagine using a macpoop
```

To log in, run the following  
Please note that the goshort server URL is optional, leaving it empty will use the default GoShort server.
```shell
goshort login # default server
goshort login serverURL # custom server, must start with http:// or https://
```

Shorten a URL:
```shell
goshort <url>

# example:
> goshort https://github.com/tkbstudios/GoShort-server
2024/08/09 19:37:32 Checking session token...
2024/08/09 19:37:32 Session token is valid!
2024/08/09 19:37:32 Shortened URL: http://localhost:8000/8QbXZR

```