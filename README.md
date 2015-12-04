# commandcast
Run command on multiple hosts over SSH

### Install

```bash
$ go get github.com/jdkanani/commandcast
```

### Usage

```
$ commandcast help

NAME:
   commandcast - Run command on multiple hosts over SSH

USAGE:
   commandcast [global options] command [command options] [arguments...]

VERSION:
   1.0.0

AUTHOR(S):
   Jaynti Kanani <jdkanani@gmail.com>

COMMANDS:
   exec, e	Execute command to all hosts
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h		show help
   --version, -v	print the version
```

**Execution usage**

```
$ commandcast exec --help
NAME:
   commandcast exec - Execute command to all hosts

USAGE:
   commandcast exec [command options] [arguments...]

OPTIONS:
   --interactive, -i							Enable intereactive mode
   --hosts "localhost"							Multiple hosts (comma separated)
   --hostfile 								File containing host names
   --user, -u 								SSH auth user [$USER]
   --timeout "15"							SSH timeout (seconds)
   --keys "/Users/jdkanani/.ssh/id_dsa,/Users/jdkanani/.ssh/id_rsa"	SSH auth keys (comma separated)
```

### Examples

```bash
$ commandcast help
$ commandcast exec --help
$ commandcast exec "echo ping" --hosts host1,host2,host3
$ commandcast exec "echo $HOME" --hosts host1,host2,host3 --user root --keys /Users/jdkanani/.ssh/cluster_id_rsa
$ commandcast exec "echo $USER" --hostfile /Users/jdkanani/clusterhosts
```

**Hostfile example**

```
$ cat /Users/jdkanani/clusterhosts
node1
node2
node3
```

**Multiple commands**

```bash
$ commandcast exec -i --hosts node1,node2,node3 --user hadoop
Keys:  [/Users/jdkanani/.ssh/id_dsa /Users/jdkanani/.ssh/id_rsa]
Hosts:  [hadoop@node1 hadoop@node2 hadoop@node3]
>>> echo $HADOOP_HOME
hadoop@node1 > echo $HADOOP_HOME
/usr/local/hadoop

hadoop@node2 > echo $HADOOP_HOME
/usr/local/hadoop

hadoop@node3 > echo $HADOOP_HOME
/usr/local/hadoop

>>>
```

### License

The MIT License (MIT)
