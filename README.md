# orbnet: GitHub star network

[![Build Status](https://github.com/milosgajdos/orbnet/workflows/CI/badge.svg)](https://github.com/milosgajdos/orbnet/actions?query=workflow%3ACI)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/milosgajdos/orbnet)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache--2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

`orbnet` is a collection of command line utilities that let you export your starred GitHub repositories as a network and serve it via a simple API.

There are a few command line utilties available in this repo:
* `dumper`: dumps your GitHub stars as `JSON` blobs
* `grapher`: reads the dumped `GitHub` star `JSON` blobs and builds a simple **Weighted Directed** [Graph](https://en.wikipedia.org/wiki/Graph_(discrete_mathematics))
* `apisrv`: serves the graph over JSON API

# Get started

The easiest way to get started is to build the binaries using the project `Makefile`:

```shell
make cmd
```

`cmd` target fetches all dependencies, builds all three binaries and places them into `_build` directory:

```shell
$ ls -1 ./_build
dumper
grapher
apisrv
```

## HOWTO

Once the binaries have been built you can explore the available command line options using the familiar `-help` switch. The below will demonstrate the basic usage of all project utilities.

### Prerequisites

Before you proceed furhter you must obtain a `GitHub` [API token](https://github.com/settings/tokens). Once you've got the token you can export it via an environment variable called `GITHUB_TOKEN` which is then automatically read by the project utilities.

Optionally, I'd recommend installing [GraphViz](https://graphviz.org/) toolkit that helps exploring the results visually.

### dumper: dump GitHub stars data

As described earlier, `dumper` "scrapes" `GitHub` API stars data and dumps them into `JSON` blobs. The data are dumped into standard output by default, but you can also store them in a directory of your choice by passing the path to the output directory via `-outdir` command line switch.

```shell
# dump data into standard output
./dumper -user milosgajdos
```

*NOTE:* `foo` must exist before you run the command below!
```shell
# dump data into directory foo
./dumper -user milosgajdos -paging 100 -outdir foo/
```

### grapher: build a graph of GitHub stars

`grapher` builds the graph from the dumped data. You can "feed" `grapher` either by passing the path to the directory that contains the `JSON` blobs via `-indir` command line option. Alternatively, you can also `pipe` the data to the `grapher` utility as by default it reads the data from standard input.

Building an in-memory graph is kinda fun, but there is no point of it if you can't visualise the results. `grapher` provides `-marshal` and `-format` command line switches which let you export (marshal) the graph into various data formats:
* `Graphviz` (see [here](https://graphviz.org/doc/info/lang.html))
* `SigmaJS` (see [here](http://sigmajs.org/))
* `CytoscapeJS` (see [here](https://js.cytoscape.org/))
* `networkx` (see [here](https://networkx.org/documentation/stable//reference/readwrite/json_graph.html))
* `gexf` (see [here](https://gephi.org/gexf/format/))
* `jsonapi` serializes the graph into `orbnet` API model

```shell
# load graph data from dumps directory and output it in GEXF format
./grapher -marshal -indir foo/ -format gexf > repos.gexf
```

```shell
# pipe data from dumper to grapher and dump the graph into GEXF file
./dumper -user milosgajdos | ./grapher -marshal -format gexf > repos.gexf
```

**NOTE:** if your graph is large, `sfdp` command might take a while to complete
```shell
# build the graph, dumpt it in dot format and export it to SVG
./dumper -user milosgajdos -paging 100 | ./grapher -marshal -format dot | sfdp -Tsvg > repos.svg
```

Alternatively you can run `dot` with overlay options that should build a better overlay of data:
```
./grapher -marshal -indir foo/ -format dot | sfdp -x -Goverlap=scale -Tpng > repos.png
```

**NOTE** `grapher` builds a *Weighted Directed Graph* that contains four types of nodes:
* `owner`: the repo owner
* `repo`: the name of the repo
* `topic`: the repo topic
* `lang`: the dominant programming language as returned by GitHub API

### apisrv: serve the GitHub stars graph over a JSON API

`apisrv` lets you serve the dumped graph over a JSON API. It even provides `swagger` docs on `/docs/` endpoint.
You can load the dumped graph via `-dsn _path_to_graph.json` cli switch.
