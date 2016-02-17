# Using crane for github-like docker development

> **Note** The version of `crane` described below is a modified one. Find it on [my github](http://www.github.com/bivas/crane) fork.

If you're custom to working with [github](http://www.github.com) that you'll recognize the following flow:

1. You fork a repo (e.g. `user/foo`)
2. You commit and push code to your fork (e.g. `bivas/foo`)
3. You create a pull request (PR) to merge your changes with original repo.

This approach is usually referred to as [`upstream`](http://stackoverflow.com/questions/9257533/what-is-the-difference-between-origin-and-upstream-in-github) remote and branching model is taken from [git flow](http://nvie.com/posts/a-successful-git-branching-model/).

## The `docker` flow problem

When working with [`docker`](http://www.docker.io), you pull and push images to a registry. Usually, when using your own registry, your push, pull, tag, build cycle might get messy.

1. You pull `company/foo` from `registry.company.co`
2. The pulled imaged is now locally tagged as `registry.company.co/company/foo`
3. You write a `Dockerfile` to build a dependent image:
	a. Start with `FROM company/foo` ?
	b. Start with `FROM registry.company.co/company/foo`?
4. So you decide to `tag` the pulled image as `company/foo` **Remember**: You can't use [`ARG`](https://docs.docker.com/engine/reference/builder/#arg) in `FROM` directive, so your registry prefix is there for good (or until refactored) 
5. Completed your local tests and you wish to `push`. Again, your `tag` your image and `push` it as `registry.company.co/company/foo`

Why can't we simply do - pull, build, and push? Let someone else deal with the registry and tagging.

## The `crane` flow
So we have our registry at `registry.company.co`, our images are prefixed as `company` and all our `Dockerfile`s begin with something like `company/foo`. We extended [`crane`](https://github.com/michaelsauter/crane) to support our flow of `pull` (from official registry), `build` (locally on developer workspace, and `push` (to a user registry for backup). The entire flow keeps the original image name as `company/foo` but knows when to change tags to support other flows.
Here's an example `crane.yaml` file:
```
containers:
  os:
    image: company/os
    build:
      context: os/
    pull:
      registry: ${DOCKER_REGISTRY}
 etcd:
     image: coreos/etcd:2
     push:
       skip: true
 service:
     image: company/service
     build:
       context: service/
     pull:
       registry: registry.company.co
     push:
      registry: registry.${USER}.company.co
      override_user: ${USER}
```

The added directives:

- `pull.registry` = which registry to `pull`from
	- `crane` will `pull` from `registry.company.co/company/foo` and locally tag it as `company/foo`
- `pull.override_user` =  which user to `pull` as
    - `crane` will `pull` the specified image as `${USER}/foo` and locally tag it as `company/foo` (useful when testing against your version of the container stack. 
- `push.registry` = which registry to `push` to
 - `crane` will `push` to  `registry.company.co/company/foo` the image locally tagged as `company/foo`
- `push.override_user`= which user to `push` as
 - `crane` will `push` the specified image as `${USER}/foo` the image locally tagged as `company/foo` (useful to backup your images in registry)
- `push.skip` = skip the image when attempting `push` (useful when dealing with community images)

Now, a developer can easily work with `FROM company/foo` directive without dealing with re-tag. She can also modify `company/foo` and keep a backups without the need to re-tag the entire images stack. 
We have a github-like development cycle of "forking" (`pull`) the official registry, `push`ing images to your private registry for development and backup (`push`). Once the developer is satisfied with the modifications - she creates a _normal_ pull request to the modified `Dockerfile`.
