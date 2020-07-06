# Cassandra Example

This example uses the datastax CRD to create the KUDO Cassandra installation.

This is an example where the developers might need to bridge an existing CRD definition


### Steps

#### Initialize KUDO and KUDO Bridge 

```
kubectl kudo init
kubectl apply -f config/deploy/deploy.yaml
```

#### Install Cassandra CRD and Bridge Spec
```
kubectl apply -f examples/cassandra/crdSpec.yaml #install the CRD
kubectl apply -f examples/cassandra/cassandraBridge.yaml #install the Bridge Spec
```

#### Install Cassandra cluster
```
kubectl apply -f examples/cassandra/cassandra.yaml
```

Now we can verify if KUDO Cassandra is installed as expected:
```
$ kubectl get operator,operatorversions,instances
NAME                          AGE
operator.kudo.dev/cassandra   1m

NAME                                       AGE
operatorversion.kudo.dev/cassandra-1.0.0   1m

NAME                                   AGE
instance.kudo.dev/cassandra-instance   1m
```


**WARNING**: This example is just intended to demo KUDO Bridge
