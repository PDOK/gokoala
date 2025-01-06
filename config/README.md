# Config

This config package is used to validate and unmarshal a GoKoala YAML config file to Go structs.

In addition, this package is imported as a _library_ in the PDOK [OGCAPI operator](https://github.com/PDOK/ogcapi-operator)
to validate the `OGCAPI` Custom Resource (CR) in order to orchestrate GoKoala in Kubernetes.
For this reason the structs in the package are annotated with [Kubebuilder markers](https://book.kubebuilder.io/reference/markers).