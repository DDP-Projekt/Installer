# Ein Installierer für Die Deutsche Programmiersprache

## create-ddp-release

This programm creates the compressed Release folder of
DDP (DDP-\<version-info>) from all locally built components.

It consumes a config.json (create-ddp-release/config.json) file which should look like this:

	{
		"Kompilierer": "<Directory to the Kompilierer repo>",
		"vscode-ddp": "<Directory to the vscode-ddp repo>",
		"DDPLS": "<Directory to the DDPLS repo>"
		"mingw": "<Directory to the mingw64 installation that should be shiped>"
	}

The "mingw" value only needs to be present on windows.
All the git-repos should be up-to-date.

To create the Release simply run create-ddp-release:

```
cd create-ddp-release
go run .
```

## ddp-setup

This program installs DDP.
It is present in the release folder which is shipped to the user.
It should be executed from said folder:

```
cd DDP-<version-info>
./ddp-setup
```

## ddp-rm

This program is a very simple implementation of the unix `rm` command.
It is shipped on windows for the installer to call `make clean` if the
runtime and stdlib had to be rebuilt
