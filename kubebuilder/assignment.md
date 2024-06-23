## Assignment 1

Write and deploy a sample project on above cluster ([refer](https://github.com/kubernetes-sigs/kubebuilder) tool to write custom APIs + Controller based on CRD. [Example](https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata/project)) 

The goal of the assignment is to understand and learn kubebuilder and the different components like Controller, CRD, Webhook and their working/logic.


## Assignment 2:

Create a CLI tool using Golang to deploy an nginx server on the Kubernetes cluster. The CLI flags would include options for scaling an existing deployment, upgrading the version of nginx pods already deployed and authentication using a kubeconfig file. Use and explore https://github.com/kubernetes/client-go


Example command:

deploynginx --version <> --kubeconfig </path/to/file> 

This will redeploy the pods with the upgraded version.

 

deploynginx -scale 3--kubeconfig </path/to/file> 

This will scale the number of pods in the deployment to 3.