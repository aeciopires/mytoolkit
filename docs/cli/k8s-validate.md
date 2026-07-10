<!-- TOC -->

- [Kubernetes YAML Validator — CLI](#kubernetes-yaml-validator--cli)
  - [Examples](#examples)

<!-- TOC -->

# Kubernetes YAML Validator — CLI

```
mytoolkit k8s-validate --in <file|-> [--out <file|->]
```

Exit code `0` if every document is valid, `1` otherwise (including on a YAML syntax error) — usable directly in CI.

## Examples

```
$ printf 'apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm1\ndata:\n  a: b\n' | mytoolkit k8s-validate
Document 1: VALID (apiVersion=v1, kind=ConfigMap, name=cm1)

1/1 document(s) valid.
```

Multi-document streams (`---`-separated, as accepted by `kubectl apply -f`) are fully checked, one document at a time — a problem in one document doesn't stop the others from being reported:

```
$ printf 'apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm1\n---\napiVersion: apps/v1\nmetadata:\n  name: dep1\n' | mytoolkit k8s-validate
Document 1: VALID (apiVersion=v1, kind=ConfigMap, name=cm1)
Document 2: INVALID: missing required field "kind" (apiVersion=apps/v1)

1/2 document(s) valid.
Error: 1 of 2 document(s) failed Kubernetes validation
```

A hard YAML syntax error aborts the whole check (no partial report):

```
$ printf 'apiVersion: [1, 2\n' | mytoolkit k8s-validate
Error: yaml: line 1: did not find expected ',' or ']'
```

This does **not** validate against a specific resource's full schema (Deployment, Service, a CRD, ...) — only the universal `apiVersion`/`kind`/`metadata` shape every Kubernetes API object needs. For full schema validation, use `kubectl apply --dry-run=server` against a real cluster, or a dedicated tool like `kubeconform`.
