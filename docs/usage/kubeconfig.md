# Handling Kubeconfigs

By default, k3d won't touch your kubeconfig without you telling it to do so.
To get a kubeconfig set up for you to connect to a k3d cluster, you can go different ways.

??? note "What is the default kubeconfig?"
    We determine the path of the used or default kubeconfig in two ways:

    1. Using the `KUBECONFIG` environment variable, if it specifies *exactly one* file
    2. Using the default path (e.g. on Linux it's `#!bash $HOME/.kube/config`)

## Getting the kubeconfig for a newly created cluster

1. Update your default kubeconfig **upon** cluster creation
    - `#!bash k3d create cluster mycluster --update-kubeconfig`
        - *Note:* this won't switch the current-context
2. Update your default kubeconfig **after** cluster creation
    - `#!bash k3d get kubeconfig mycluster`
        - *Note:* this won't switch the current-context
3. Update a different kubeconfig **after** cluster creation
    - `#!bash k3d get kubeconfig mycluster --output some/other/file.yaml`
        - *Note:* this won't switch the current-context
    - The file will be created if it doesn't exist

!!! note "Switching the current context"
    None of the above options switch the current-context.
    This is intended to be least intrusive, since the current-context has a global effect.
    You can switch the current-context directly with the `get kubeconfig` command by adding the `--switch` flag.

## Removing cluster details from the kubeconfig

`#!bash k3d delete cluster mycluster` will always remove the details for `mycluster` from the default kubeconfig.
