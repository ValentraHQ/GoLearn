# Kubernetes with Go
id: kubernetes
number: 18
track: devops
paths: pro
skill: kubernetes
requires: docker, concurrency
project: kubernetes-controller
summary: Talk to the Kubernetes API with client-go, then build controllers and operators.

## Kubernetes API
slug: kubernetes-api
minutes: 8
objectives: Explain that everything in Kubernetes is an API object served over REST; Build resource URLs from group, version, namespace and resource; Describe the object model (metadata, spec, status)
takeaways: Kubernetes is a REST API server (kube-apiserver) backed by etcd; Objects have apiVersion, kind, metadata, spec (desired state) and status (observed state); Paths: /api/v1/... for the core group and /apis/<group>/<version>/... for others

### Concept
`kubectl get pods` is just an HTTP GET. The API server stores **objects** and everything else (scheduler, kubelet, controllers) reads and writes them through the API.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: default
  labels: {app: web}
spec:            # desired state — you write this
  replicas: 3
status:          # observed state — controllers write this
  readyReplicas: 3
```
URL patterns:

- core group: `/api/v1/namespaces/{ns}/pods/{name}`
- named groups: `/apis/apps/v1/namespaces/{ns}/deployments`
- cluster-scoped resources (Nodes, Namespaces): no `namespaces/{ns}` segment

Verbs map to HTTP: `get`, `list`, `watch` (`?watch=true`, streaming), `create` (POST), `update` (PUT), `patch`, `delete`. Every object has a `resourceVersion` used for optimistic concurrency and for resuming watches.

### Example
```go
package main

import "fmt"

func main() {
	fmt.Println("GET /api/v1/namespaces/default/pods")
	fmt.Println("GET /apis/apps/v1/namespaces/default/deployments/web")
	fmt.Println("GET /api/v1/nodes")
}
```

### Exercise
Write `path(group, version, namespace, resource, name string) string`. An empty `group` means the core group (`/api/<version>`), otherwise `/apis/<group>/<version>`. Include `namespaces/<ns>` only when namespace is non-empty and `/<name>` only when name is non-empty.
```text expect
/api/v1/namespaces/default/pods/web-1
/apis/apps/v1/namespaces/prod/deployments
/api/v1/nodes
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func path(group, version, namespace, resource, name string) string {
	var b strings.Builder
	if group == "" {
		b.WriteString("/api/" + version)
	} else {
		b.WriteString("/apis/" + group + "/" + version)
	}
	if namespace != "" {
		b.WriteString("/namespaces/" + namespace)
	}
	b.WriteString("/" + resource)
	if name != "" {
		b.WriteString("/" + name)
	}
	return b.String()
}

// END

func main() {
	fmt.Println(path("", "v1", "default", "pods", "web-1"))
	fmt.Println(path("apps", "v1", "prod", "deployments", ""))
	fmt.Println(path("", "v1", "", "nodes", ""))
}
```

### Check
Q: Which field of a Kubernetes object holds the desired state?
T: short
A: spec
E: `spec` is what you declare; `status` is what controllers observe and report.

Q: What is the resource path for Nodes (a cluster-scoped resource)?
T: mcq
- [ ] /api/v1/namespaces/default/nodes
- [x] /api/v1/nodes
- [ ] /apis/v1/nodes
- [ ] /api/nodes/v1
E: Cluster-scoped resources have no namespace segment.

## client-go
slug: client-go
minutes: 9
objectives: Load kubeconfig or in-cluster configuration; Create a clientset and list resources; Use label selectors and handle errors
takeaways: client-go is the official Go client for the Kubernetes API; Use clientcmd for kubeconfig and rest.InClusterConfig inside a pod; List/Get/Create/Update/Delete take a context and typed options

### Concept
```go norun
import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
// inside a Pod: cfg, err := rest.InClusterConfig()
cs, err := kubernetes.NewForConfig(cfg)

pods, err := cs.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
	LabelSelector: "app=web",
})
for _, p := range pods.Items {
	fmt.Println(p.Name, p.Status.Phase)
}
```
Errors: use `apierrors.IsNotFound(err)`, `IsConflict`, `IsForbidden` (from `k8s.io/apimachinery/pkg/api/errors`). **Label selectors** filter server-side: `app=web,tier!=db`, set-based `env in (prod,staging)`. Always pass a context with a timeout, and prefer informers (later lessons) over polling.

client-go isn't available in the exercise sandbox, so this lesson's exercise models label-selector matching in plain Go.

### Example
```go
package main

import (
	"fmt"
	"strings"
)

func parseSelector(s string) map[string]string {
	sel := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		if k, v, ok := strings.Cut(strings.TrimSpace(part), "="); ok {
			sel[k] = v
		}
	}
	return sel
}

func main() {
	fmt.Println(parseSelector("app=web, tier=frontend"))
}
```

### Exercise
Write `matches(selector string, labels map[string]string) bool` for equality selectors `k=v,k2=v2` (every pair must match; an empty selector matches everything).
```text expect
true false true
```
```go solution
package main

import (
	"fmt"
	"strings"
)

// BEGIN
func matches(selector string, labels map[string]string) bool {
	if strings.TrimSpace(selector) == "" {
		return true
	}
	for _, part := range strings.Split(selector, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || labels[k] != v {
			return false
		}
	}
	return true
}

// END

func main() {
	labels := map[string]string{"app": "web", "tier": "frontend"}
	fmt.Println(matches("app=web", labels), matches("app=web,tier=db", labels), matches("", labels))
}
```

### Check
Q: Which config loader is used when your program runs inside a Pod?
T: mcq
- [ ] clientcmd.BuildConfigFromFlags
- [x] rest.InClusterConfig
- [ ] kubernetes.NewFake
- [ ] os.Getenv("KUBECONFIG")
E: In-cluster config uses the mounted service account token and CA.

Q: Where is label-selector filtering performed when you use `ListOptions.LabelSelector`?
T: mcq
- [ ] In your program after downloading everything
- [x] On the API server
- [ ] In etcd only
- [ ] In the scheduler
E: The server filters, saving bandwidth and memory.

## Kubernetes Clients
slug: kubernetes-clients
status: planned

## Pods
slug: pods
minutes: 8
objectives: Explain what a Pod is and its lifecycle phases; Derive readiness from container statuses; Summarise pod health like kubectl does
takeaways: A Pod is the smallest deployable unit: one or more containers sharing network and storage; Phases: Pending, Running, Succeeded, Failed, Unknown; A pod is Ready only when all containers are ready — check the conditions, not just the phase

### Concept
A **Pod** runs one or more tightly coupled containers with a shared IP and volumes. Pods are **ephemeral**: they're replaced, not repaired, which is why you rarely create bare pods — a **Deployment** manages them.

`status.phase` is coarse. A pod in phase `Running` can still be crash-looping (`CrashLoopBackOff` is a container state reason) or unready. For health, read `status.containerStatuses[].ready` and the `Ready` condition.

Common failure reasons: `ImagePullBackOff`, `CrashLoopBackOff`, `OOMKilled`, `Unschedulable` (Pending, no node fits).

### Example
```go
package main

import "fmt"

type ContainerStatus struct {
	Name  string
	Ready bool
}

type Pod struct {
	Name       string
	Phase      string
	Containers []ContainerStatus
}

func ready(p Pod) bool {
	if p.Phase != "Running" || len(p.Containers) == 0 {
		return false
	}
	for _, c := range p.Containers {
		if !c.Ready {
			return false
		}
	}
	return true
}

func main() {
	p := Pod{"web-1", "Running", []ContainerStatus{{"app", true}, {"sidecar", false}}}
	fmt.Println(p.Name, ready(p))
}
```

### Exercise
Write `ready(p Pod) string` returning `n/m` (ready containers / total), like `kubectl get pods` prints in the READY column.
```text expect
1/2 2/2 0/1
```
```go solution
package main

import "fmt"

type ContainerStatus struct {
	Name  string
	Ready bool
}

type Pod struct {
	Name       string
	Containers []ContainerStatus
}

// BEGIN
func ready(p Pod) string {
	n := 0
	for _, c := range p.Containers {
		if c.Ready {
			n++
		}
	}
	return fmt.Sprintf("%d/%d", n, len(p.Containers))
}

// END

func main() {
	a := Pod{"a", []ContainerStatus{{"app", true}, {"side", false}}}
	b := Pod{"b", []ContainerStatus{{"app", true}, {"side", true}}}
	c := Pod{"c", []ContainerStatus{{"app", false}}}
	fmt.Println(ready(a), ready(b), ready(c))
}
```

### Check
Q: A pod's phase is `Running`. Can it still be unhealthy?
T: mcq
- [ ] No, Running means healthy
- [x] Yes — containers may be crash-looping or not ready
- [ ] Only if it is Pending
- [ ] Only for Jobs
E: Phase is coarse; readiness and container statuses tell the real story.

Q: Why do you rarely create bare Pods in production?
T: mcq
- [ ] They cannot run containers
- [x] Nothing recreates them if the node fails; a Deployment/StatefulSet manages replacement
- [ ] They're slower
- [ ] They can't be scheduled
E: Controllers give you self-healing and rolling updates.

## Deployments
slug: deployments
minutes: 8
objectives: Explain Deployments, ReplicaSets and rolling updates; Calculate maxSurge and maxUnavailable; Understand rollout status conditions
takeaways: A Deployment manages ReplicaSets, which manage Pods; Rolling updates replace pods gradually within maxSurge (extra) and maxUnavailable (missing) limits; Percentages round up for maxSurge and down for maxUnavailable

### Concept
```
Deployment ──▶ ReplicaSet (v2) ──▶ Pods
          └──▶ ReplicaSet (v1) ──▶ Pods (scaled down during rollout)
```
A change to `spec.template` creates a **new ReplicaSet**; the Deployment controller scales it up and the old one down, honouring:

- `maxSurge` — how many pods above `replicas` may exist (default 25%, **rounded up**)
- `maxUnavailable` — how many below `replicas` may be unavailable (default 25%, **rounded down**)

With 10 replicas and defaults: surge = ceil(2.5) = 3, unavailable = floor(2.5) = 2. Both can't be 0. `kubectl rollout status|undo deployment/web` inspects and reverts. Readiness probes gate progress: pods that never become Ready stall the rollout instead of taking the service down.

### Example
```go
package main

import (
	"fmt"
	"math"
)

func main() {
	replicas := 10
	surge := int(math.Ceil(float64(replicas) * 25 / 100))
	unavail := int(math.Floor(float64(replicas) * 25 / 100))
	fmt.Println(surge, unavail, "max pods:", replicas+surge, "min available:", replicas-unavail)
}
```

### Exercise
Write `limits(replicas int, surgePercent, unavailPercent int) (maxSurge, maxUnavailable int)` — surge rounds **up**, unavailable rounds **down**.
```text expect
3 2
1 0
```
```go solution
package main

import "fmt"

// BEGIN
func limits(replicas, surgePercent, unavailPercent int) (maxSurge, maxUnavailable int) {
	maxSurge = (replicas*surgePercent + 99) / 100 // ceil
	maxUnavailable = replicas * unavailPercent / 100 // floor
	return
}

// END

func main() {
	fmt.Println(limits(10, 25, 25))
	fmt.Println(limits(3, 25, 25))
}
```

### Check
Q: What happens to the Deployment's Pods when you change `spec.template`?
T: mcq
- [ ] They restart all at once
- [x] A new ReplicaSet is created and pods are rolled over gradually
- [ ] Nothing until you delete the pods
- [ ] The Deployment is recreated
E: The Deployment controller performs a rolling update between ReplicaSets.

Q: How does Kubernetes round a percentage `maxUnavailable`?
T: mcq
- [ ] Up
- [x] Down
- [ ] To nearest
- [ ] It doesn't allow percentages
E: `maxSurge` rounds up and `maxUnavailable` rounds down, favouring availability.

## Services
slug: services
status: planned

## ConfigMaps
slug: configmaps
status: planned

## Secrets
slug: secrets
status: planned

## Namespaces
slug: namespaces
status: planned

## Watches
slug: watches
minutes: 8
objectives: Explain list-then-watch; Apply ADDED/MODIFIED/DELETED events to a local cache; Recover from dropped watches with resourceVersion
takeaways: A watch is a long-lived HTTP stream of change events; The list-watch pattern lists once for a snapshot, then watches from that resourceVersion; A local store built from events lets controllers read without hammering the API server

### Concept
Polling `GET /pods` every second wastes resources. Instead:

1. **List** all objects (get a `resourceVersion`).
2. **Watch** from that version: the server streams `ADDED`, `MODIFIED`, `DELETED` (and `BOOKMARK`) events.
3. Apply events to a **local cache** keyed by `namespace/name`.
4. If the watch drops or the version is too old (`410 Gone`), **re-list** and continue.

```go norun
w, err := cs.CoreV1().Pods("default").Watch(ctx, metav1.ListOptions{ResourceVersion: rv})
for ev := range w.ResultChan() {
	switch ev.Type {
	case watch.Added, watch.Modified: /* upsert */
	case watch.Deleted:               /* remove */
	}
}
```
Informers (next lessons) implement exactly this loop, with resync and shared caches, so you rarely write it by hand.

### Example
```go
package main

import "fmt"

type Event struct {
	Type string
	Key  string
	Val  string
}

func apply(store map[string]string, ev Event) {
	switch ev.Type {
	case "ADDED", "MODIFIED":
		store[ev.Key] = ev.Val
	case "DELETED":
		delete(store, ev.Key)
	}
}

func main() {
	store := map[string]string{}
	for _, ev := range []Event{{"ADDED", "default/a", "v1"}, {"MODIFIED", "default/a", "v2"}, {"ADDED", "default/b", "v1"}, {"DELETED", "default/b", ""}} {
		apply(store, ev)
	}
	fmt.Println(store)
}
```

### Exercise
Write `replay(events []Event) map[string]string` that applies the events to a fresh store (ADDED/MODIFIED upsert, DELETED removes) and returns it.
```text expect
map[default/a:v2 default/c:v1]
```
```go solution
package main

import "fmt"

type Event struct {
	Type string
	Key  string
	Val  string
}

// BEGIN
func replay(events []Event) map[string]string {
	store := map[string]string{}
	for _, ev := range events {
		switch ev.Type {
		case "ADDED", "MODIFIED":
			store[ev.Key] = ev.Val
		case "DELETED":
			delete(store, ev.Key)
		}
	}
	return store
}

// END

func main() {
	fmt.Println(replay([]Event{
		{"ADDED", "default/a", "v1"},
		{"ADDED", "default/b", "v1"},
		{"MODIFIED", "default/a", "v2"},
		{"DELETED", "default/b", ""},
		{"ADDED", "default/c", "v1"},
	}))
}
```

### Check
Q: What should a controller do when its watch fails with `410 Gone` (resource version too old)?
T: mcq
- [ ] Exit
- [x] Re-list to get a fresh snapshot, then watch again
- [ ] Ignore it
- [ ] Retry the same watch forever
E: The server no longer has the history; a new list-watch cycle recovers.

Q: Why keep a local cache built from watch events?
T: mcq
- [ ] To avoid writing controllers
- [x] Reads become local and cheap, reducing load on the API server
- [ ] Watches only work with caches
- [ ] It replaces etcd
E: Controllers read from the cache and only write to the API.

## Informers
slug: informers
status: planned

## Controllers
slug: controllers
minutes: 9
challenges: workqueue-dedupe
objectives: Describe the controller pattern (observe → diff → act); Explain work queues, deduplication and retries; Recognise level-triggered versus edge-triggered logic
takeaways: A controller watches objects and continuously drives actual state toward desired state; Events enqueue object keys; a work queue deduplicates them and workers reconcile one key at a time; Failed reconciles are re-queued with rate-limited backoff

### Concept
```
watch events ─▶ event handlers ─▶ workqueue (keys, deduped)
                                       │
                                  worker goroutines
                                       │
                           Reconcile(key): read cache, compare, act via API
```
Handlers don't do work; they enqueue the **key** (`namespace/name`). The queue **deduplicates**: if a key is added ten times while waiting, it is processed once — the reconcile reads the *current* state anyway. If a key is re-added while being processed, it is queued again after the worker calls `Done`, so no update is lost. On error the worker calls `AddRateLimited(key)` for exponential backoff.

This is **level-triggered** design: react to *what the state is now*, not to *what event just happened*, so missed or duplicated events don't matter.

### Example
```go
package main

import "fmt"

type Queue struct {
	order   []string
	pending map[string]bool
}

func NewQueue() *Queue { return &Queue{pending: map[string]bool{}} }

func (q *Queue) Add(key string) {
	if q.pending[key] {
		return // deduplicate
	}
	q.pending[key] = true
	q.order = append(q.order, key)
}

func (q *Queue) Get() (string, bool) {
	if len(q.order) == 0 {
		return "", false
	}
	k := q.order[0]
	q.order = q.order[1:]
	delete(q.pending, k)
	return k, true
}

func main() {
	q := NewQueue()
	for _, k := range []string{"default/a", "default/b", "default/a", "default/a"} {
		q.Add(k)
	}
	for k, ok := q.Get(); ok; k, ok = q.Get() {
		fmt.Println("reconcile", k)
	}
}
```

### Exercise
Write `Len()` on `Queue` returning the number of distinct keys waiting. After adding `a, b, a, c, b` print the length and then drain the queue printing each key on one line, space-separated.
```text expect
3
a b c
```
```go solution
package main

import (
	"fmt"
	"strings"
)

type Queue struct {
	order   []string
	pending map[string]bool
}

func NewQueue() *Queue { return &Queue{pending: map[string]bool{}} }

func (q *Queue) Add(key string) {
	if q.pending[key] {
		return
	}
	q.pending[key] = true
	q.order = append(q.order, key)
}

// BEGIN
func (q *Queue) Len() int { return len(q.order) }

// END

func (q *Queue) Get() (string, bool) {
	if len(q.order) == 0 {
		return "", false
	}
	k := q.order[0]
	q.order = q.order[1:]
	delete(q.pending, k)
	return k, true
}

func main() {
	q := NewQueue()
	for _, k := range []string{"a", "b", "a", "c", "b"} {
		q.Add(k)
	}
	fmt.Println(q.Len())
	var keys []string
	for k, ok := q.Get(); ok; k, ok = q.Get() {
		keys = append(keys, k)
	}
	fmt.Println(strings.Join(keys, " "))
}
```

### Check
Q: What do informer event handlers typically put on the work queue?
T: mcq
- [ ] The whole object
- [x] The object's key (namespace/name)
- [ ] The event type only
- [ ] A copy of the diff
E: Reconcile reads the latest state from the cache, so only the key is needed.

Q: What does "level-triggered" mean for a controller?
T: mcq
- [ ] It reacts only to the last event
- [x] It compares current desired and actual state each time, so missed or duplicate events don't matter
- [ ] It polls every second
- [ ] It runs at cluster level only
E: Reconciliation is based on state, not on the history of events.

## Custom Resources
slug: custom-resources
status: planned

## CRDs
slug: crds
minutes: 9
objectives: Explain what a CustomResourceDefinition adds; Describe an OpenAPI schema for spec validation; Design a spec and status for your own resource
takeaways: A CRD registers a new resource type (kind) with the API server; The schema validates spec and prunes unknown fields; Custom resources plus a controller form an operator

### Concept
```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: webapps.example.com
spec:
  group: example.com
  scope: Namespaced
  names: {plural: webapps, singular: webapp, kind: WebApp}
  versions:
  - name: v1alpha1
    served: true
    storage: true
    subresources: {status: {}}
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            required: [image]
            properties:
              image:    {type: string}
              replicas: {type: integer, minimum: 1, maximum: 20, default: 2}
```
After applying it, `kubectl get webapps` works and `kubectl apply -f my-webapp.yaml` stores your objects in etcd — validated by the schema before any controller sees them. The `status` subresource lets controllers update status separately from spec.

Tools like **kubebuilder** and **controller-gen** generate CRDs from Go struct markers (`// +kubebuilder:validation:Minimum=1`).

### Example
```go
package main

import "fmt"

type WebAppSpec struct {
	Image    string
	Replicas int
}

func main() {
	spec := WebAppSpec{Image: "nginx:1.27", Replicas: 3}
	fmt.Printf("%+v\n", spec)
}
```

### Exercise
Write `validate(spec map[string]any) []string` mirroring the schema: `image` is required and a string (`image is required`), `replicas` (optional, default ok) must be an integer between 1 and 20 (`replicas must be between 1 and 20`). Return messages in that order; empty when valid. JSON numbers arrive as `float64`.
```text expect
[image is required replicas must be between 1 and 20]
[]
```
```go solution
package main

import "fmt"

// BEGIN
func validate(spec map[string]any) []string {
	var errs []string
	if img, ok := spec["image"].(string); !ok || img == "" {
		errs = append(errs, "image is required")
	}
	if v, present := spec["replicas"]; present {
		n, ok := v.(float64)
		if !ok || n != float64(int(n)) || n < 1 || n > 20 {
			errs = append(errs, "replicas must be between 1 and 20")
		}
	}
	return errs
}

// END

func main() {
	fmt.Println(validate(map[string]any{"replicas": 0.0}))
	fmt.Println(validate(map[string]any{"image": "nginx", "replicas": 3.0}))
}
```

### Check
Q: What does a CRD do?
T: mcq
- [ ] Creates a new cluster
- [x] Registers a new resource type with the API server
- [ ] Runs a controller
- [ ] Schedules pods
E: A CRD teaches the API server about your kind; a controller gives it behaviour.

Q: Which combination is commonly called an operator?
T: mcq
- [ ] A Deployment and a Service
- [x] A custom resource plus a controller that reconciles it
- [ ] A ConfigMap and a Secret
- [ ] Two clusters
E: Operators encode operational knowledge in custom controllers.

## Operators
slug: operators
status: planned

## Reconciliation
slug: reconciliation
minutes: 9
challenges: reconcile-plan
objectives: Write an idempotent Reconcile function; Compute the actions needed to converge actual state to desired state; Reason about what happens when reconcile runs twice
takeaways: Reconcile(key) reads desired and actual state and performs the minimum actions to converge them; It must be idempotent: running it again with nothing changed does nothing; Report progress in status and never assume a previous run's result

### Concept
```go norun
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var app v1alpha1.WebApp
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // deleted: nothing to do
	}
	// 1. observe actual state (the child Deployment)
	// 2. compare with app.Spec
	// 3. create / update / delete children to match
	// 4. update app.Status
	return ctrl.Result{}, nil
}
```
Design rules:

- **Idempotent** — create only what's missing, update only what differs.
- **Stateless** — derive everything from the objects, not from memory of past events.
- **Owner references** on children so deleting the parent garbage-collects them.
- Return an error (→ retry with backoff) for transient failures; `RequeueAfter` to check again later.

The exercise models the core diff: given desired and actual replica counts, decide which pods to create or delete.

### Example
```go
package main

import "fmt"

func plan(desired, actual int) (create, remove int) {
	switch {
	case actual < desired:
		return desired - actual, 0
	case actual > desired:
		return 0, actual - desired
	}
	return 0, 0
}

func main() {
	fmt.Println(plan(3, 1))
	fmt.Println(plan(3, 3))
	fmt.Println(plan(1, 4))
}
```

### Exercise
Write `reconcile(desired int, pods []string) (actions []string)` returning `create <name>` actions for missing pods named `pod-0 … pod-(desired-1)` and `delete <name>` for surplus pods (any existing pod not in that set), creates first, then deletes in slice order. Running it on an already correct set returns no actions.
```text expect
[create pod-1 create pod-2 delete old-x]
[]
```
```go solution
package main

import "fmt"

// BEGIN
func reconcile(desired int, pods []string) (actions []string) {
	want := map[string]bool{}
	for i := 0; i < desired; i++ {
		want[fmt.Sprintf("pod-%d", i)] = true
	}
	have := map[string]bool{}
	for _, p := range pods {
		have[p] = true
	}
	for i := 0; i < desired; i++ {
		name := fmt.Sprintf("pod-%d", i)
		if !have[name] {
			actions = append(actions, "create "+name)
		}
	}
	for _, p := range pods {
		if !want[p] {
			actions = append(actions, "delete "+p)
		}
	}
	return actions
}

// END

func main() {
	fmt.Println(reconcile(3, []string{"pod-0", "old-x"}))
	fmt.Println(reconcile(2, []string{"pod-0", "pod-1"}))
}
```

### Check
Q: What does it mean for Reconcile to be idempotent?
T: mcq
- [ ] It runs only once
- [x] Running it repeatedly with unchanged inputs makes no further changes
- [ ] It never fails
- [ ] It ignores errors
E: Because reconciles are retried and duplicated, they must be safe to repeat.

Q: What is the purpose of owner references on child objects?
T: mcq
- [ ] Access control
- [x] Kubernetes garbage-collects the children when the owner is deleted
- [ ] Faster scheduling
- [ ] Labelling
E: Deleting the custom resource automatically removes the Deployment it created.
