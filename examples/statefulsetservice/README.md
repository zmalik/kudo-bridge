# StatefulSet Service Example

This example creates a KUDO Operator and its KUDO Bridge implementation from scratch

We will create a new operator which will create services for statefulset pods for external access.

Example:
```
apiVersion: service.statefulset.kudo.dev/v1beta1
kind: ExternalService
metadata:
  name: external-svc
  namespace: default
spec:
  statefulset:
    name: my-kafka
    namespace: default
  externalTrafficPolicy: Local
  count: 5
  ports: 9092
  type: LoadBalancer
```

This object will create 5 services of type `LoadBalancer` using the `selector` `statefulset.kubernetes.io/pod-name` for the 5 pods of the sts `my-kafka`

## Steps

### Operator

We create a simple KUDO Operator structure, read more about creating a KUDO Operator in [getting started](https://kudo.dev/docs/developing-operators/getting-started.html) docs

```
├── operator
    ├── operator.yaml
    ├── params.yaml
    └── templates
        └── service.yaml

```
In our case its a simple one task operator

```
apiVersion: kudo.dev/v1beta1
name: "external-service"
operatorVersion: "0.1.0"
kudoVersion: 0.14.0
kubernetesVersion: 1.16.0
appVersion: 1.0.0
maintainers:
  - name: Zain Malik
    email: zmalikshxhil@gmail.com
url: https://kudo.dev
tasks:
  - name: deploy
    kind: Apply
    spec:
      resources:
        - service.yaml
plans:
  deploy:
    strategy: serial
    phases:
      - name: deploy
        strategy: serial
        steps:
          - name: deploy
            tasks:
              - deploy
```
and with few parameters

```
apiVersion: kudo.dev/v1beta1
parameters:
  - name: COUNT
    description: "Services count"
    default: "1"
  - name: SERVICE_TYPE
    description: "Services type"
    default: "NodePort"
  - name: PORT
    default: "8080"
  - name: TARGET_PORT
    default: "8080"
  - name: NODE_PORT
    default: "32001"
  - name: STATEFULSET_NAME
    default: "cassandra"
  - name: TRAFFIC_POLICY
    default: "Local"
```

The main logic lies in the `templates/service.yaml` 

```
{{ range $i, $v := until (int .Params.COUNT) }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ $.Name }}-{{ $v }}
  namespace: {{ $.Namespace }}
spec:
  type: {{ $.Params.SERVICE_TYPE }}
  externalTrafficPolicy: {{ $.Params.TRAFFIC_POLICY }}
  selector:
    statefulset.kubernetes.io/pod-name: {{ $.Params.STATEFULSET_NAME }}-{{ $v }}
  ports:
    - protocol: TCP
      {{ if eq  $.Params.SERVICE_TYPE "LoadBalancer" }}
      port: {{ $.Params.PORT }}
      targetPort: {{ $.Params.TARGET_PORT }}
      {{ end }}
      {{ if eq  $.Params.SERVICE_TYPE "NodePort" }}
      port: {{ add (int $.Params.PORT) $v }}
      targetPort: {{ add (int $.Params.TARGET_PORT) $v }}
      nodePort: {{ add (int $.Params.NODE_PORT) $v }}
  {{ end }}
{{ end }}

```
### Install the KUDO Operator without Instance

```
kubect kudo install ./operator --skip-instance
```

### Bridge

The Bridge instance will use the params to map the custom values for our statefulset service

In this case its also using `inClusterOperator: true` as the operator is installed in the cluster. 

```
apiVersion: kudobridge.dev/v1alpha1
kind: BridgeInstance
metadata:
  name: ext-service-bridge
  namespace: default
  labels:
    group: "service.statefulset.kudo.dev"
    version: "v1beta1"
    kind: "ExternalService"
  finalizers:
    - finalizer.bridge.kudo.dev
spec:
  kudoOperator:
    package: external-service
    version: 0.1.0
    appVersion: 1.0.0
    inClusterOperator: true
  crdSpec:
    apiVersion: service.statefulset.kudo.dev/v1beta1
    kind: ExternalService
    metadata:
      name: external-svc
      namespace: default
    spec:
      statefulset:
        name: STATEFULSET_NAME
        namespace: default
      externalTrafficPolicy: TRAFFIC_POLICY
      count: COUNT
      port: PORT
      targetPort: TARGET_PORT
      type: SERVICE_TYPE
```

#### Initialize KUDO and KUDO Bridge 

```
kubectl kudo init --unsafe-self-signed-webhook-ca
kubectl apply -f config/deploy/deploy.yaml
```

#### Install Statefulset Service CRD and Bridge Spec
```
kubectl apply -f examples/statefulsetservice/resources/crd.yaml #install the CRD
kubectl apply -f examples/statefulsetservice/bridge.yaml #install the Bridge Spec
```

#### Install the Service
```
kubectl apply -f examples/statefulsetservice/service.yaml
```

Now we can verify if KUDO Cassandra is installed as expected:
```
$ kubectl get operator,operatorversions,instances
NAME                                 AGE
operator.kudo.dev/external-service   1m

NAME                                              AGE
operatorversion.kudo.dev/external-service-0.1.0   1m

NAME                             AGE
instance.kudo.dev/external-svc   24s
```

And the services are created correctly

```
kubectl get svc
NAME             TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
external-svc-0   LoadBalancer   10.96.145.5     <pending>     8080:30030/TCP   22s
external-svc-1   LoadBalancer   10.96.238.193   <pending>     8080:30341/TCP   22s
external-svc-2   LoadBalancer   10.96.179.74    <pending>     8080:30179/TCP   22s
external-svc-3   LoadBalancer   10.96.72.23     <pending>     8080:31685/TCP   22s
external-svc-4   LoadBalancer   10.96.43.99     <pending>     8080:32024/TCP   22s
```




**WARNING**: This example is just intended to demo KUDO Bridge
