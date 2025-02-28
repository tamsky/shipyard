# Change Log

## version v0.4.0
* Add new resource to enable copying resources

```hcl
copy "testing" {
	source = "/"
	destination = "/path"

	permissions = "0700"
}
```

* Add new resource to enable generation of certificates

```hcl
certificate_ca "testing" {
	output = "/"
}

certificate_leaf "testing" {
	ip_addresses = ["a","b"]
	dns_names = ["1","2"]

	ca_cert = "./file"
	ca_key = "./file"

	output = "/"
}
```

* Add new function to allow the creation of data directories with specific permissions

```hcl
data_with_permissions("name","0777")
```

## version v0.3.50
* Sanitize helm chart names replacing non acceptable characters

## version v0.3.43
* Update Helm library to 3.8.2

## version v0.3.42
* Add capability to install Helm charts from Helm repositories and local paths

```hcl
helm "vault" {
  depends_on = ["helm.consul"] # only install one at a time

  cluster = "k8s_cluster.k3s"

  repository {
    name = "hashicorp"
    url  = "https://helm.releases.hashicorp.com"
  }

  chart   = "hashicorp/vault" # When repository specified this is the name of the chart
  version = "v0.18.0"         # Version of the chart when repository specified

  values = "./helm/vault-values.yaml"

  health_check {
    timeout = "240s"
    pods    = ["app.kubernetes.io/name=vault"]
  }
}
```

## version v0.3.41
Shipyard now supports Podman through the Podman sock. Rootless and Root containers are both supported however, complex 
containers that require privileged access such as k8s_clusters and nomad_clusters required Root mode for podman.

## version v0.3.40
* Do not attempt to import Docker images to Nomad and Kubernetes clusters when the Name is ""

## version v0.3.37
* Adds capability for complex variable types and variable interpolation in templates. Previously,
all `vars` in `template` resources were interpreted as strings. This change now allows richer types
and where possible preserves the original type. For example, you can now define a variable with an array
of integers and then loop over this array in a template.

```javascript
variable "bool_var" {
	default = true
}

variable "num_var" {
	default = 13
}

variable "str_var" {
	default = "Abc"
}

variable "other_ports" {
	default = [2000,2001]
}

template "fetch_consul_resources" {
  source = <<EOF

bind_addr = "{{ GetPrivateInterfaces | attr \"address\" }}"
client_addr = "0.0.0.0"
data_dir = "#{{ .Vars.data_dir }}"
datacenter = "dc1"

enabled = #{{ .Vars.enabled }}
not_enabled = #{{ .Vars.not_enabled }}
bool_var = #{{ .Vars.bool_var }}
num_var = #{{ .Vars.num_var }}
string_var = "#{{ .Vars.string_var }}"

ports {
	#{{ range .Vars.ports }}
  grpc = #{{ . }}
	#{{ end }}
}

other_ports {
	#{{ range .Vars.other_ports }}
  grpc_other = #{{ . }}
	#{{ end }}
}
	EOF

	destination = "./out.txt"

  vars = {
		other_ports = var.other_ports
		bool_var = var.bool_var
		string_var = var.str_var
		num_var = var.num_var
		data_dir = "something"
		enabled = true
		not_enabled = false
		ports = [8502, 8500]
		port = 8342
		config = {
			a = 1
			names = ["foo","bar"]
		}
  }
}

* Adds new interpolation function `len`, this returns the length of an array or map variable.

* Upgrade default Kubernetes version to `v1.22.3`
* Upgrade default Nomad version to `v1.2.0`
```

## version v0.3.35
* Fix bug with Nomad clusters when using exec or java drivers
* Update Nomad base version to 1.1.6
* Fix bug with ingress selecting ports for stopped jobs

## version v0.3.34
* Fix bug where output variables were not honoring the disabled meta parameter
## version v0.3.32
* Fix bug in Nomad job where status was incorrectly reported
* Improve error messages when Nomad Job validation fails
* Add support for Docker bind propagation

## version v0.3.28
* Add bash completion for `logs` command
* Format status output to make clearer
* Add new Nomad version 1.1.3
* Return error message when resource has no name

## version v0.3.1

* Set max random ports for clusters to 64000

## version v0.3.1
### Image Caching

To save bandwidth all containers launched from Kubernetes and Nomad clusters are cached by Shipyard. Currently images
from the following registries are cached:

* k8s.gcr.io 
* gcr.io 
* asia.gcr.io 
* eu.gcr.io 
* us.gcr.io 
* quay.io
* ghcr.io"
* docker.io

Previously only images created from the `container` resource were cached, this change will make substantial speed and bandwidth improvements when creating Shipyard resources once the image has been cached.

Shipyard launches a pull through cache when resources are created

### Other changes
* Add max_restart_count to containers and sidecars
* Add function to retrieve the ipa
* Move terminal server to embedded in Shipyard binary
* Add capability to have a local terminal instance in Docs.

## version 0.2.10
* Add capability to run `local_exec` commands as a daemon
* Add timeout to `local_exec`
* Move `Creating TLS certificates` message to debug log

## version 0.2.0
* Add new interpolation function to get the ip address of the machine running shipyard
* Add CreateNamespace config parameter to Helm config

```javascript
// ingress exposing a local application
// would enable traffic on the K8s cluster dc1 sent to:
//      my-local-service.shipyard.svc.cluster.local:9090
// to be directed to:
//      localhost:30002
// on the shipyard host
ingress "k8s-to-local" {
  source {
    driver = "k8s"
    
    config {
      cluster = "k8s_cluster.dc1"
      port = 9091
    }
  }
  
  destination {
    driver = "local"
    
    config { 
      address = "localhost"
      port = 30002
    }
  
  }
}

// ingress exposing an application on one K8s cluster to the shipyard host
// would enable traffic on the shipyard host sent to:
//      localhost:9090
// to be directed to:
//      dc1-service.mynamespace.svc.cluster.local:30002
// on the dc1 cluster
ingress "local-to-k8s" {
  source {
    driver = "local"
    
    config {
      port = 9092
    }
  }
  
  destination {
    driver = "k8s"
    
    config {
      cluster = "k8s_cluster.dc1"
      address = "k8s-to-local.shipyard.svc"
      port = 9091
    }
  }
}
```

* Add ability to define defaults for variables

Variables can also be defined in modules, variables specified as 
flags, files, or environment override defaults.

```
variable "mod_network" {
  default = "modulenetwork"
}

* New Interpolation functions file_path and file_dir

```

### Ingress
Major refactor of ingress, implemented K8s and Local

## version 0.1.18
* Abiility to run with external Docker engine. Shipyard now respects the 
environment variable `DOCKER_HOST` and can be run on external docker engines.

## version 0.1.17
* Compress Shipyard binary to reduce distribution size
* Add ability to add variables overides with test command

## version 0.1.15
Add updates for Nomad ingress

## version 0.1.14
Various bug fixes

* Fix Nomad health checks
* Move version manager package

## version 0.1.8

### Fix local exec capability 
Fixes bugs and extends the local exec capability for running commands local as part of a run

```javascript
exec_remote "exec_standalone" {
  cmd = "consul"
  args = [
    "services",
    "register",
    "./config/redis.hcl"
  ]

  env {
    key = "CONSUL_HTTP_ADDR"
    value = "http://consul.container.shipyard.run:8500"
  }
}
```

## version 0.0.32

### Remote Exec
* Ensure remote exec containers use a unique ID

## version 0.0.31

### Containers

Add ability to configure network aliases for containers

```
  network { 
    name = "network.onprem"
    ip_address = "10.5.0.200"
    // Add network aliases for the container
    aliases = ["web.ingress.container.shipyard.run", "api.ingress.container.shipyard.run"]
  }
```

## version 0.0.30

### Bugfixes Exec
* Fix log output when executing commands in containers

## version 0.0.28

## Remote Exec
* Add `working_directory` parameter to allow setting of the execution folder when running commands

## version 0.0.26

### Push Command
* Changes to caching volumes inadvertently changed the behaviour of the `push` command. When an image was already cached
  in the cluster image volume `push` would not overwrite this image. If the user specifies the `--force-update` command
  then the pushed image would be overwritten regardles of local cache; however, force-update would also attempt to pull
  the image from a remote registry. Push was designed to push a local image to a cluster as part of the development
  flow, pulling remote images is not desireable behaviour.
* This change ensures that the `push` command will always push an image regardless of local cache.

## version 0.0.25

### Bugfixes
* Environment variables were not being correctly passed to remote container executions when attaching to an exising container


## version 0.0.24

## Env Command
Added a new command to simplify the export of environment variables. Environment variables can be defined in a blueprint
readme, after the blueprint has completed the variables will be presented to the user.

**Blueprint**
```
---
title: Single Container Example
author: Nic Jackson
slug: container
browser_windows: http://consul-http.ingress.shipyard.run:8500
env:
  - KUBECONFIG=$HOME/.shipyard/config/k3s/kubeconfig.yaml
  - VAULT_ADDR=http://localhost:18200
  - VAULT_TOKEN=root
---

# Single Container

This blueprint shows how you can create a single container with Shipyard
```

**Output**
```
######################################################

Environment Variables

######################################################

This blueprint exports the following environment variables:

KUBECONFIG=$HOME/.shipyard/config/k3s/kubeconfig.yaml=
VAULT_ADDR=http://localhost:18200=
VAULT_TOKEN=root=

You can set exported environment variables for your current terminal session using the following command:

eval $(shipyard env)
```

The `env` command will write the defined environment variables to the command line...

```
➜ shipyard env
export KUBECONFIG=$HOME/.shipyard/config/k3s/kubeconfig.yaml
export VAULT_ADDR=http://localhost:18200
export VAULT_TOKEN=root
```

To set the variable to the command line the user can either copy and paste the output or evaulate it.

```
eval $(shipyard env)
```

## Bugfixes
* Update terminal server in Docs, there was a bug with the latest Chrome where the terminal was not immediately
  displaying the terminal.

## version 0.0.20

## Clusters
* New feature to persist and share cache for importing images for clusters. Previously each cluster has its own cache
  which was created every time the cluster was started. Images were copied from the local machine to a Docker volume
  before being imported to the clsuter. If multiple clusters were defined in a config this process would run for each
  cluster unecessarilly slowed the startup time. Now images are cached in a persistant volume
  `images.volume.shipyard.run` improving startup times. When a cluster starts it first checks to see if an image exits,
  if it does not it will import the image from the local Docker instance.
* Added a new feature to the `purge` command which removes the persistent image cache.

## Modules
* New resource type `modules` which can import other Shipyard configuration blueprints from a file or from GitHub.
* Modules do not respect dependencies, this feature is a TODO item.

```
module "k8s" {
	source = "github.com/shipyard-run/shipyard/examples//single_k3s_cluster"
}

module "consul" {
	source = "../container"
}
```


## version 0.0.19

### Sidecar resource
Added a new resource `sidecar`. A `sidecar` is a special container which shares the network and ip address of the target container.

```ruby
container "consul" {
  image   {
    name = "consul:1.6.1"
  }

  command = ["consul", "agent", "-config-file=/config/consul.hcl"]
}

sidecar "consul" {
  target = "container.consul"

  image   {
    name = "consul:1.6.1"
  }

  command = ["consul", "connect", "envoy", "-sidecar-for", "myservice"]
}
```

### Helm charts
* Added ability to set the namespace

### Purge Command
* Added `shipyard purge` command to remove all cached images, blueprints, and helm charts

### Bugfixes
* Remote exec now correctly pulls a container

## version 0.0.18

### Helm charts
* Added capability to reference a github repo as well as local charts
* Added the ability to specify values as a map addition to a local file

```
helm "vault" {
  cluster = "k8s_cluster.k3s"
  chart = "github.com/hashicorp/vault-helm"

  values_string = {
    "server.dataStorage.size" = "128Mb",
    "server.dev.enabled" = "true",
    "server.standalone.enabled" = "true",
    "server.authDelegator.enabled" = "true"
  }

  health_check {
    timeout = "120s"
    pods = ["app.kubernetes.io/name=vault"]
  } 
}
```

### General
* Added `--force-update` flag to `run` and `get`. The user can specify that any local cache for images and helm charts is ignored at runtime.


## version 0.0.16

### Bugfixes
* By default when looking up the id of a Docker container Docker does a greedy match, this caused issues where we would
  grab an incorrect id on destroy. Changed to use a regex.

## version 0.0.16

### Bugfixes
* Fixes a bug where local folders were being created for types other than `bind`
* Fix a bug where Volumes for Nomad Clusters were not changed to absolute paths

## version 0.0.15

### General
* When defining a volume mount for a container, if the source does not exist, creat it rather than erroring


## version 0.0.14

### Bugfixes
* Fix bug where Documentation resource was failing to create on OSX. Docker could not read temporary file. 
  Moved temp files to .shipyard folder to avoid conflict.


## version 0.0.11

### Browser windows
Change `OpenInBrowser` to a string so that the user can specify the path

### Version check
Check the latest verison on startup

```
########################################################
                   SHIPYARD UPDATE
########################################################

The current version of shipyard is "0.0.10", you have "841f3b0445cc01f44ea2728dd5113f6b0f611e1f".

To upgrade Shipyard please use your package manager or, 
see the documentation at:
https://shipyard.run/docs/install for other options.
```


## version 0.0.9

### Browser windows
Ensure that browser windows are openened correctly on Windows platforms.

### Preflight
Check that Docker is installed and running


## version 0.0.8

### Bugfixes

* Fix problem with home folder resolution on Windows
* Add ipv6 DNS entries for resources

## version 0.0.7

### Ingress
Add three new ingress types `nomad_ingress`, `k8s_ingress`, `container_ingress`, these new type have cluster specific options
rather than trying to shoehorn all into `ingress`. For now `ingress` exists and is backward compatible.

```
k8s_ingress "consul-http" {
  cluster = "k8s_cluster.k3s"
  service  = "consul-consul-server"

  network {
    name = "network.cloud"
  }

  port {
    local  = 8500
    remote = 8500
    host   = 18500
  }
}

container_ingress "consul-http" {
  target  = "container.consul"

  port {
    local  = 8500
    remote = 8500
    host   = 18500
  }

  network  {
    name = "network.cloud"
  }
}

nomad_ingress "nomad-http" {
  cluster  = "nomad_cluster.dev"
  job = ""
  group = ""
  task = ""

  port {
    local  = 4646
    remote = 4646
    host   = 14646
    open_in_browser = true
  }

  network  {
    name = "network.cloud"
  }
}
```

## Browser Windows
* Added ability to open browser windows to ingress, containers, docs
* Only open browser windows if they have not been opened

## General
* Can now do `shipyard run` or `shipyard run .` and this equals `shipyard run ./`
* Sidebar configuration is now optional in docs
* Changed names of Docker resources and FQDN for resources to `[name].[type].shipyard.run`
* Resources are now resolvable via DNS
* Added interpolation helper `${home()}` to resolve the home folder from config
* Added interpolation helper `${shipyard()}` to resolve the shipyard folder from config

### Bug fixes
* Fix serialisation of blueprint in state
* Fix bug with rolled back resources being in an inconsistent state
* Fix bug with state not serialzing correctly


## version 0.0.6

### Bug fixes
* Fix bug where heirachy in state was lost on incremental applies
* Minor UX tweak for run command, use current directory as default

## version 0.0.5

### Container
HTTP health checks can now be defined for containers. At present HTTP health checks are executed on the client and 
require forwarded ports or an ingress to be configured. Future updates will move this functionality to the network
removing the requirement for external access.

```javascript
container "vault" {
  image {
    name = "hashicorp/vault-enterprise:1.4.0-rc1_ent"
  }

  command = [
    "vault",
    "server",
    "-dev",
    "-dev-root-token-id=root",
    "-dev-listen-address=0.0.0.0:8200",
  ]

  port {
    local = 8200
    remote = 8200
    host = 8200
  }

  health_check {
    timeout = "30s"
    http = "http://localhost:8200/v1/sys/health"
  }
}
```

### Nomad Job
Nomad job is a new resource type which allows the running of Nomad Jobs on a cluster. Health checks can be applied to jobs
a health check will only be marked as passed when all the tasks in side the job are reported as "Running".

Nomad job is a dependent resource and will not apply until the cluster is up and healthy.

```javascript
nomad_cluster "dev" {
  version = "v0.10.2"

  nodes = 1 // default

  network {
    name = "network.cloud"
  }

  image {
    name = "consul:1.7.1"
  }
}

nomad_job "redis" {
  cluster = "nomad_cluster.dev"

  paths = ["./app_config/example2.nomad"]
  
  health_check {
    timeout = "60s"
    nomad_jobs = ["example_2"]
  }
}
```

## version 0.0.4

### Container
Allow HTTP health checks to be added to containers

## version 0.0.3
This version was skipped due to issues getting Chocolately distributions setup

## version 0.0.2

## Docs
Improve UX with documentation, Shipyard now autogenerates the JSON files required for Docusarus, the user
only needs to author the markdown

## Nomad
Implmeneted ability to push a local container to the Nomad cluster
Allow mounting of custom volumes for Nomad clusters

## Build process
Added Chocolatey and Brew, Deb and Rpm instalation sources

## Yard files
Yard files are to be depricated in favor of Markdown files for blueprints.
The information which was previously added to the .yard file can now be added as frontdown in a `README.md` file which resides in the root of your blueprint.

````
---
title: Single Container Example
author: Nic Jackson
slug: container
browser_windows: http://localhost:8080
---

# Single Container

This blueprint shows how you can create a single container with Shipyard

```shell
curl localhost:8080
```
````

When the user runs `shipyard run`, this renders to the terminal as:

```
########################################################

Title Single Container Example
Author Nic Jackson

########################################################


1 Single Container
────────────────────────────────────────────────────────────────────────────────

This blueprint shows how you can create a single container with Shipyard

┃ curl localhost:8080
```

### Bugfixes
* Move create shipyard home directory to run or get, this was generating with invalid permissions when using the quick install

## version 0.0.0-beta.12

### Bugfixes
* Ensure downloaded blueprints are stored with their full path
* Increase test coverage

## version 0.0.0-beta.11

### Bugfixes
* Correct log line in Kubernetes controller

## version 0.0.0-beta.11

### Bugfixes
* Fix bug where Kubernetes Config was not returning an error when applying bad config
* Add log output for Helm charts and Kubernetes config

## version 0.0.0-beta.10

### Added no-browser to `run` command
The `run` command now has a `no-browser` flag which supresses any browser windows from opening if they are defined in the stack


## version 0.0.0-beta.9

### Updated exec command
Exec command now uses lighter ingress container instead of tools

### Container resource
Add "entrypoint" configuration to set the containers entrypoint

### Docs
New documentation contianer which proxies terminal websockets using Envoy. This was an issue when the docs site was 
running behind a proxy server such as Instruqt.


## version 0.0.0-beta.8

### Updated status command
The status command now pretty prints the resources

```shell
➜ shipyard status

 [ CREATED ] docs.docs
 [ CREATED ] container.tools
 [ CREATED ] helm.consul
 [ CREATED ] ingress.consul-http
 [ CREATED ] k8s_cluster.k3s
 [ CREATED ] network.cloud
 [ CREATED ] container.vscode

Pending: 0 Created: 7 Failed: 0
```

To view status in json format use the `--json` flag

### Bug fixes
* alpine/linux was pulled every time when importing images regardless of local cache
* fix `push` to use new configuration

## version 0.0.0-beta.7

### Helm provider
* Helm provider now uninstalls the chart when deleting a resource, previousuly it was assumed that a chart and cluster would be deleted together
* Added `exec` command to allow the creation of a shell or execution of a command in a container or pod
```
➜ yard-dev exec k8s_cluster.k3s consul-consul-227vz               
parameters: []string{"k8s_cluster.k3s", "consul-consul-227vz"} - command: []string{}
2020-02-19T11:45:28.523Z [DEBUG] Image exists in local cache: image=shipyardrun/tools:latest
2020-02-19T11:45:28.524Z [INFO]  Creating Container: ref=exec-524329800
2020-02-19T11:45:28.641Z [DEBUG] Attaching container to network: ref=exec-524329800 network=network.cloud
/ # ls -las
total 68
     4 drwxr-xr-x    1 root     root          4096 Feb 19 11:38 .
     4 drwxr-xr-x    1 root     root          4096 Feb 19 11:38 ..
     4 drwxr-xr-x    1 root     root          4096 Sep 13 06:21 bin
```
* Added `version` command to return the current application verion
* When restarting from pause, health check all containers, helm charts and k8s config
* Update status command to pretty print status and add `--json` flag for detail

### Bug fixes
* Improve test quality


## version 0.0.0-beta.6

### Bug fixes
* Alpine container not pulled when copying images to cluster
* Health check for pod was only looking at status not ready checks
* Check Network exists before removing
* Upgrade Helm dependency

## Version 0.0.0-beta.5

### Introduce taint command and the ability to re-create resources.

Resources can now be tainted using the command `shipyard taint [type] [name]`

When a resource is marked as tained the next run of `shipyard run` will destroy the resource and re-create it.
This feature is especailly useful when building blueprints, often you require a change to a particular container you run `shipyard destroy`
to destroy the stack and then `shipyard run` to re-create. Now it is possible to destroy only the affected resource with `shipyard taint`.

### Change behaviour when processing folders

Previously `shipyard run` would recurse into folders, this behaviour causes problems when the sub-folders contain `*.hcl` files which are not
Shipyard resources. `shipyard run` now only process the top level folder. Sub folder support will be added when we add the `module` feature.

### Improve handling for failed resources

Resources which fail to create can now be retired by re-running `shipyard run`, any depended resources which were not created due to the failure
will also be created when the command is run.
