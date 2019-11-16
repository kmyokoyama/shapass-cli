# shapass-cli

CLI version of Shapass: The password manager that does not store your passwords.

[Official website](https://shapass.com/)

## Installation

```
$ go get -d github.com/kmyokoyama/shapass-cli
$ cd $GOPATH/src/github.com/kmyokoyama/shapass-cli
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

## Usage

The last argument must always be the target service (e.g., `facebook`, `twitter` etc).
```
$ shapass -help

  -clip
        Should copy output password to system clipboard? (default true)
  -len int
        Length of the password (default 10)
  -show
        Should show output password? (default false) (default true)
  -suffix string
        Suffix to generate the output password (default "")
```



## Examples

```
$ shapass facebook
> Enter master password: [mysecret]
<No output showed. Output password sent to clipboard>
```

```
$ shapass -show facebook
> Enter master password: [mysecret]
7quoXGYb8b
```

```
$ shapass -suffix=xpto -show facebook
> Enter master password: [mysecret]
7quoXGYb8bxpto
```

```
$ shapass -len=30 -suffix=xpto -show facebook
> Enter master password: [mysecret]
7quoXGYb8bMwZKCdIV9s8vSVT68rSWxpto
```

```
$ shapass -len=30 -clip=false -show facebook
> Enter master password: [mysecret]
7quoXGYb8bMwZKCdIV9s8vSVT68rSWxpto <Does not copy to clipboard>
```

## Contributing

Contributions are welcome.

Get the required dependencies with:

```
$ go get -d github.com/atotto/clipboard
```

And start hacking!

## License

MIT License