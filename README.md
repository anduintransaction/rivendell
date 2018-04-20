# Kubernetes resources management tools

# Table of Contents

 - [Installation](#installation)
 - [How to use](#how-to-use)
 - [Configuration](#configuration)
   - [Config file format](#config-file-format)
   - [Environment variables substitution](#environment-variables-substitution)
   - [Command line variables and variable files](#command-line-variables-and-variable-files)
   - [Default variables](#default-variables)
   - [Sample configuration file](#sample-configuration-file)
   - [Configuration keys](#configuration-keys)
 - [Resource groups](#resource-groups)
   - [Resources dependency](#resources-dependency)
   - [Resource group configuration](#resource-group-configuration)
   - [Resource files glob](#resource-files-glob)
   - [Waiting for pods or jobs](#waiting-for-pods-or-jobs) 

## Installation

Check https://github.com/anduintransaction/rivendell/releases

## How to use

 - `rivendell generate project.yml` to generate a project file. See `Configuration` section for more detail
 about how to configure a project.
 - Add your own kubernetes configuration files.
 - Run `rivendell up project.yml` to create all resources.
 - Run `rivendell down project.yml`to destroy all resources.
 - Run `rivendell update project.yml` to update all resources other than `pod` or `job`.
 - Run `rivendell upgrade project.yml` to upgrade all resources, including `pod` and `job`. The `pods` and `jobs` must be stopped before upgrading
 
## Configuration

### Config file format

The configuration format is `YAML`

### Environment variables substitution

Environment variables referenced with the $(...) notation within the configuration file will be replaced with the value of the environment variable, for instance:

```YAML
namespace: $(KUBERNETES_NAMESPACE)
```

### Command line variables and variable files

The configuration file also supports `go-template`:

```YAML
namespace: {{.namespace}}
```

Variables can be passed to the configuration file by command line flags:

`rivendell up project.yml --variable namespace=my-namespace`

Or using variable file:

`rivendell up project.yml --variableFile varFile`

Sample `varFile` content:

```
namespace=my-namespace
```

If a variable key is specified more than once, the order of important is `--variable` > `variable file` > `configuration file` > `default value`

### Default variables

- `rivendellVarNamespace`: Current kubernetes namespace
- `rivendellVarContext`: Current kubernetes context
- `rivendellVarKubeConfig`: Current kubernetes config file
- `rivendellVarRootDir`: Root directory of project

### Sample configuration file

```YAML
root_dir: .
namespace: coruscant
variables:
  postgresTag: {{.postgresImageTag}}
  redisTag: 4-alpine
  postgresSidecarImage: postgres-sidecar:{{.appTag}}
  redisSidecarImage: redis-sidecar:{{.appTag}}
resource_groups:
  - name: configs
    resources:
      - ./configs/*.yml
    excludes:
      - ./configs/*ignore*
  - name: secrets
    resources:
      - ./secrets/*.yml
  - name: databases
    resources:
      - ./databases/*.yml
    depend:
      - configs
      - secrets
  - name: init-jobs
    resources:
      - ./jobs/*.yml
    depend:
      - databases
  - name: services
    resources:
      - ./services/*.yml
    depend:
      - init-jobs
    wait:
      - name: init-postgres
        kind: job
      - name: init-redis
        kind: job
delete_namespace: true
```

### Configuration keys

| Key | Type | Description |
|-----|------|-------------|
| root\_dir | string | Root dir, relative to the configuration file. All kubernetes configuration files will be relative to this directory |
| namespace | string | Kubernetes namespace, value from command line flag will override this value |
| variables | map | Variables map, value from command line flags will override these values |
| resource\_groups | array | See [Resource groups](#resource-groups) |
| delete\_namespace | string | Delete the namespace in `down` command or not |

## Resource groups

### Resources dependency

Rivendell manages a graph of *resource groups*, with each group contains multiple resource files, and can depend
on each others. If group A depends on group B, rivendell wait for all resources in group B to become ready before 
creating resources in group A. A resource is defined as ready when its kubernetes life cycle status is *Active*.

A resource group can also be configured to wait for some jobs or pods to complete.

### Resource group configuration

| Key | Type | Description |
|-----|------|-------------|
| name | string | Name of the group |
| resources | string array | List of resource files. See [Resource files glob](#resource-files-glob) |
| excludes | string array | List of excluded resources file |
| depend | string array | List of groups this group depends on |
| wait | array | See [Waiting for pods or jobs](#waiting-for-pods-or-jobs) |


### Resource files glob

A list of glob-based files can be add to resource group. Supported glob patterns are:

 - `path/to/*.yml` matches all `yml` files under `path/to`
 - `path/*/*.yml` matches all `yml` files under subfolder of `path`
 - `path/**/*.yml` matches all `yml` files under any level of subfolder of `path`
 
### Waiting for pods or jobs

A resource can be configured to wait for some pods or jobs to complete with:

```YAML
wait:
  - name: job1
    kind: job
    timeout: 300
  - name: pod1
    kind: pod
```
