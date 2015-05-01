# DOT Graph template

```
crane template -t path/to/graph.tmpl -o output/myGraph.dot
```

Generate a [DOT](http://en.wikipedia.org/wiki/DOT_(graph_description_language)) file representing the dependency graph. Bold nodes represent the containers declared in the config (as opposed to non-bold ones that are referenced in the config, but not defined). Targeted containers are highlighted with color borders. Solid edges represent links, dashed edges volumesFrom, and dotted edges net=container relations.

# Systemd template

```
crane template -t path/to/systemd.tmpl -o output/%s.service
```

This is an example template for Systemd. Edit it to adapt it to your needs. There are some assumptions to know:

* Crane needs to be installed.
* All _"data-only"_ containers must have the suffix _"-data"_ or _"-storage"_ because they are treated apart. These containers are not considered as services, so although their corresponding files will be created, you must ignore them (removing them manually).
* All the dependencies of each service are appended in `Requires` and `After` options (except the _"data-only_ containers that are removed from these options and added as `ExecStartPre` commands)

Output example for a `web` container with `mysql` and `image-storage` as dependencies (`output/mysql.service`):

```

[Unit]
Description=web container
Requires=docker.service mysql.service
After=docker.service mysql.service

[Service]
Restart=on-failure
RestartSec=10
ExecStartPre=/home/core/bin/crane run -c /home/core/crane.yml image-storage
ExecStart=/home/core/bin/crane run -c /home/core/crane.yml --recreate web
ExecStop=/home/core/bin/crane stop -c /home/core/crane.yml web

[Install]
WantedBy=multi-user.target

```

# Upstart template

```
crane template -t path/to/upstart.tmpl -o output/%s.conf
```

It's the equivalent to the Systemd's previous example

```
description "web container"
author "Me"
start on filesystem and started docker and started mysql
stop on runlevel [!2345]
respawn
script
  /home/core/bin/crane run -c /home/core/crane.yml image-storage
  /home/core/bin/crane run -c /home/core/crane.yml --recreate web
end script

```