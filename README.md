# eifa-trigger-operator

A lightweight, Custom Resource-based Kubernetes Operator built using [Operator SDK](https://sdk.operatorframework.io/) and Golang.  

- 🌀 eifa-trigger-operator automatically restarts Deployments or DaemonSets when specific ConfigMaps or Secrets change — using the power of a Kubernetes Custom Resource.

---

## 🔍 What It Does

- Monitors one or more ConfigMaps or Secrets based on label selectors.
- On detected changes, triggers rollout restarts of targeted Deployments or DaemonSets.
- Tracks status and last triggered message via `.status.conditions` and `.status.lastMessage` in your `EifaTrigger` resource.
- Powered by Kubernetes-native Conditions and Operator SDK scaffolding.

---


## 🧬 Custom Resource: EifaTrigger

### API

- **Group**: `trigger.eifa.org`
- **Version**: `v1`
- **Kind**: `EifaTrigger`

## 🧬 EifaTrigger Custom Resource Definition (CRD)

| Field         | Description                                                        |
|---------------|--------------------------------------------------------------------|
| `watch`  | Defines the ConfigMap or Secret to watch using label selectors     |
| `update` | Defines the Deployment or DaemonSet to trigger on change           |
| `watchList` / `updateList` | Support multiple resources with flexible selectors     |
---

### 📨 Example Manifest

```yaml
apiVersion: trigger.eifa.org/v1
kind: EifaTrigger
metadata:
  name: example-trigger
spec:
  watchList:
    - kind: ConfigMap
      labelSelector:
        et-kind: watch
    - kind: Secret
      labelSelector:
        et-kind: watch
  updateList:
    - kind: Deployment
      labelSelector:
        et-kind: update
```
Or you can apply example with:

```bash
kubectl apply -f example/manifest.yaml
```

> ✅ Supports both single and multiple `watch`/`update` items through `watchList` and `updateList`.

---
## 📘 Status & Conditions

The operator updates status fields automatically after processing:

- `status.conditions`: Standard Kubernetes-style conditions
- `status.lastMessage`: Latest activity message for observability

---

## 📈 Status Monitoring

Each EifaTrigger resource provides a live status indicating the latest trigger activity:

```yaml
status:
  lastMessage: successfully update Deployment:et-nginx because of changes at ConfigMap:et-cm
  conditions:
    - lastTransitionTime: "2025-04-15T08:30:55Z"
      message: successfully update Deployment:et-nginx because of changes at ConfigMap:et-cm
      reason: UpdateObjectRestart
      status: "True"
      type: Success
```

This gives you visibility into what happened and when directly from the CR.

---
## 🛠️ Installation

You can install the operator in your cluster using a single command — no cloning required.

### 🔹 Quick Install (kubectl apply)

```bash
kubectl apply -f https://raw.githubusercontent.com/erfan-272758/eifa-trigger-operator/main/config/all-manifests.yaml
```

This deploys:

- The EifaTrigger CRD
- Role/RoleBinding and Service Account
- Operator deployment

---

## 🛠️ Development & Contributing

### Cloning

```bash
git clone https://github.com/erfan-272758/eifa-trigger-operator.git
cd eifa-trigger-operator
```

### Building & Generating Manifests

> Make sure you have Go, Docker, and Operator SDK installed.

#### 🔨 Build & Generate Manifests

These Makefile targets support build/deploy:

| Command              | Description                                                                 |
|----------------------|-----------------------------------------------------------------------------|
| `make deploy-file`   | Generates a ready-to-apply Kubernetes manifest at `config/all-manifests.yaml` |

#### 🐳 Example:

```bash
export IMG=your-repo/eifa-trigger-operator:latest make deploy-file
kubectl apply -f config/all-manifests.yaml
```

---
## 📂 Project Structure

- **api/v1/eifatrigger_types.go**  
  - Defines the Go structure of the `EifaTrigger` CRD, with validation, status, and CLI printing annotations.
  
- **controllers/eifatrigger_controller.go**  
  - Contains controller logic that reconciles EifaTrigger resources and watches for changes.

- **example/manifest.yaml**  
  - An example manifest demonstrating how to use the operator with a ConfigMap and Deployment.

- **config/all-manifests.yaml**  
  - Pre-built manifest that bundles operator deployment and CRDs for simple installation.

---
## ⚖️ License

Licensed under the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0).  
See [`LICENSE`](./LICENSE) for more details.

---

## 🤝 Contributing

PRs, bug reports, and ideas are welcome! To contribute:

1. ⭐ Star this repo if you find it useful
2. Fork it and create your feature branch (`git checkout -b feature/foo`)
3. Commit your changes (`git commit -am 'Add feature foo'`)
4. Push to the branch (`git push origin feature/foo`)
5. Create a new Pull Request

---

## 🙌 Acknowledgements

- ⚙️ [Operator SDK](https://sdk.operatorframework.io/)
- 🔃 [Stakater Reloader](https://github.com/stakater/Reloader) — inspiration for dynamic rollout functionality
