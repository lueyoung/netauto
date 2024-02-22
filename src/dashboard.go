package core

type dashboard struct{}

func (c *dashboard) sed() map[string]string {
	return make(map[string]string)
}

func (c *dashboard) path() string {
	return "/tmp/dashboard"
}

func (c *dashboard) yaml() map[string]string {
	var name string
	yaml := make(map[string]string)
	// 0 rbac
	name = "dashboard-rbac.yaml"
	yaml[name] = string(dashboardRBACYaml)
	// 1 configmap
	name = "dashboard-configmap.yaml"
	yaml[name] = string(dashboardConfigmapYaml)
	// 2 secret
	name = "dashboard-secret.yaml"
	yaml[name] = string(dashboardSecretYaml)
	// 3 service
	name = "dashboard-service.yaml"
	yaml[name] = string(dashboardServiceYaml)
	// 4 controller
	name = "dashboard-controller.yaml"
	yaml[name] = string(dashboardControllerYaml)
	return yaml
}

const dashboardRBACYaml = `kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kubernetes-dashboard
subjects:
  - kind: ServiceAccount
    name: kubernetes-dashboard
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
`

const dashboardConfigmapYaml = `apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    k8s-app: kubernetes-dashboard
    # Allows editing resource and makes sure it is created first.
    addonmanager.kubernetes.io/mode: EnsureExists
  name: kubernetes-dashboard-settings
  namespace: kube-system
`

const dashboardSecretYaml = `apiVersion: v1
kind: Secret
metadata:
  labels:
    k8s-app: kubernetes-dashboard
    # Allows editing resource and makes sure it is created first.
    addonmanager.kubernetes.io/mode: EnsureExists
  name: kubernetes-dashboard-certs
  namespace: kube-system
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  labels:
    k8s-app: kubernetes-dashboard
    # Allows editing resource and makes sure it is created first.
    addonmanager.kubernetes.io/mode: EnsureExists
  name: kubernetes-dashboard-key-holder
  namespace: kube-system
type: Opaque
`

const dashboardServiceYaml = `apiVersion: v1
kind: Service
metadata:
  name: kubernetes-dashboard
  namespace: kube-system
  labels:
    k8s-app: kubernetes-dashboard
    kubernetes.io/cluster-service: "true"
    addonmanager.kubernetes.io/mode: Reconcile
spec:
  selector:
    k8s-app: kubernetes-dashboard
  ports:
  - port: 443
    targetPort: 8443
`

const dashboardControllerYaml = `apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-app: kubernetes-dashboard
    addonmanager.kubernetes.io/mode: Reconcile
  name: kubernetes-dashboard
  namespace: kube-system
---
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: kubernetes-dashboard
  namespace: kube-system
  labels:
    k8s-app: kubernetes-dashboard
    kubernetes.io/cluster-service: "true"
    addonmanager.kubernetes.io/mode: Reconcile
spec:
  selector:
    matchLabels:
      k8s-app: kubernetes-dashboard
  template:
    metadata:
      labels:
        k8s-app: kubernetes-dashboard
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      containers:
      - name: kubernetes-dashboard
        image: lowyard/kubernetes-dashboard-amd64:v1.8.3
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 50m
            memory: 100Mi
        ports:
        - containerPort: 8443
          protocol: TCP
        args:
          - --auto-generate-certificates
        volumeMounts:
        - name: kubernetes-dashboard-certs
          mountPath: /certs
        - name: tmp-volume
          mountPath: /tmp
        livenessProbe:
          httpGet:
            scheme: HTTPS
            path: /
            port: 8443
          initialDelaySeconds: 30
          timeoutSeconds: 30
      volumes:
      - name: kubernetes-dashboard-certs
        secret:
          secretName: kubernetes-dashboard-certs
      - name: tmp-volume
        emptyDir: {}
      serviceAccountName: kubernetes-dashboard
      tolerations:
      - key: "CriticalAddonsOnly"
        operator: "Exists"
`
