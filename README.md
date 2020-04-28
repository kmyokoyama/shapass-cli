# shapass-cli

CLI version of Shapass: The password manager that does not store your passwords.

Current version: `v1.1.0`

[Official website](https://shapass.com/)

## Installation

```
$ go get -d github.com/kmyokoyama/shapass-cli
$ cd $GOPATH/src/github.com/kmyokoyama/shapass-cli
$ go get -d ./... # Get all dependencies.
$ go install
```

You can later create an alias:

```
$ echo "alias shapass=$GOPATH/bin/shapass-cli" >> ~/.bashrc
```

Or copy the binary to somewhere in `$PATH`, for instance:

```
$ cp $GOPATH/bin/shapass-cli /usr/local/bin/shapass
```

## Integration with shapass.com API

Since `v1.1.0` shapass-cli is able to fetch data from [shapass.com](https://shapass.com/) API
regarding your service parameters so you don't have to repeat yourself when using the
site and shapass-cli. All your service parameters are available for use in shapass-cli
as soon as you create and save them on [shapass.com](https://shapass.com/).

All you have to do is to pass the `-api` argument to shapass-cli
and provide your master password and email to login to the API. **Your master password
is not persisted in anyway! It is only kept in memory while shapass-cli is running.**

Your email is persisted in `${HOME}/.shapass` so you don't have to type it every time.

## Usage

The last argument must always be the target service (e.g., `facebook`, `twitter` etc).
```
$ shapass -help

  -api
        Should fetch configurations from shapass.com? (default false)
  -copy
        Should copy output password to system clipboard? (default true)
  -display
        Should show output password? (default false)
  -length int
        Length of the password (default 32)
  -prefix string
        Prefix to generate the output password (default "")
  -suffix string
        Suffix to generate the output password (default "")
```

## Examples

```
$ shapass facebook
> Enter master password: [mysecret]
Password copied to clipboard successfully!
<No output displayed. Output password sent to clipboard>
```

```
$ shapass -display facebook
> Enter master password: [mysecret]
7quoXGYb8b
Password copied to clipboard successfully!
```

```
$ shapass -suffix=xpto -display facebook
> Enter master password: [mysecret]
7quoXGYb8bxpto
Password copied to clipboard successfully!
```

```
$ shapass -length=30 -suffix=xpto -display facebook
> Enter master password: [mysecret]
7quoXGYb8bMwZKCdIV9s8vSVT68rSWxpto
Password copied to clipboard successfully!
```

```
$ shapass -length=30 -copy=false -display facebook
> Enter master password: [mysecret]
7quoXGYb8bMwZKCdIV9s8vSVT68rSW
<Does not copy to clipboard>
```

```
$ shapass -api
> Enter master password: [mysecret]
> Email: [my@email.com]
> Choose a service [1-3]:
[1] twitter
[2] facebook
[3] linkedin
> 1
Password copied to clipboard successfully!
```

```
$ shapass -api
> Enter master password: [mysecret]
> Should use email 'my@email.com'? [Y/n] 
> Choose a service [1-3]:
[1] facebook
[2] linkedin
[3] twitter
> 2
Password copied to clipboard successfully!
```

```
$ shapass -api facebook
> Enter master password: [mysecret]
> Should use email 'my@email.com'? [Y/n] 
Password copied to clipboard successfully!
```

Note: the final length of the password may exceed 44 characters. In fact, it will have
`lenFlag + len(suffix)` characters.

## Contributing

Contributions are welcome.

Get the required dependencies with:

```
$ go get -d ./... # Get all dependencies.
```

And start hacking!

## License

MIT License
