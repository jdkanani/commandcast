# commandcast
Run command on multiple hosts over SSH

# Install

```bash
    $ go get github.com/jdkanani/commandcast
```

# Usage

```bash
    $ commandcast help
    $ commandcast exec "echo ping" --hosts host1,host2,host3
    $ commandcast exec "echo $HOME" --hosts host1,host2,host3 --user root --keys /home/jdkanani/.ssh/cluster_id_rsa
    $ commandcast exec "echo $USER" --hostfile /home/jdkanani/clusterhosts
```

# License

The MIT License (MIT)
