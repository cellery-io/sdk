# Manage Cells 
Cells can be managed by,
 * [Observing cells](#observability)
 * [Scaling cells](#scaling-cells) 
 * [Updating cells](#updating-cells)
 
 ## Observability 
Cellery Observability brings along the capability to observe the Cells deployed on Cellery Mesh. It provides overview of 
the deployed cells with dependency diagrams, runtime metrics of the cells, and distributed tracing for the request that 
pass through each cells and components. By default cellery obserability component is enabled only in the 
[complete setup](installation-options.md#basic-vs-complete-installations). You can also enable the observability component 
via the [modify commands](setup/modify-setup.md). Find more details on how to observe cells [here.](cellery-observability.md)

## Scaling cells
Each components within the cells can optionally have auto scaling or zero scaling policies as explained [here](cellery-syntax.md#autoscaling). 
Therefore, a component within a cell can be scaled up to maximum replica count, and scaled down upto minimum replica count based on the metrics 
threshold defined in the policy. Auto scaling is handled by 
[kubernetes horizontal autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/), and 
zero scaling is powered by [Knative](https://knative.dev/v0.6-docs/). By default cellery installations have autoscaling, 
and zero scaling disabled in the runtime, and you can enable it by [modify commands](setup/modify-setup.md#enabledisable-autoscaling). 
Find more details on scaling up/down cells [here.](cell-scaling.md)

## Updating cells
Cells are immutable units, and hence the updates are performed on cell instances not on the component alone. A cell 
instance can be rolling updated to update the components inside the cell, and this is a viable operation when the API of the cell is unchanged. 
Further, as advanced deployment options, Cellery supports both blue/green and canary deployment options. In that case, a new cell instance with updated 
version will be spawned along with the old cell instance, and the traffic will be routed to completely or partially among them. 
Find more details on updating cells [here.](cell-update-and-adv-deployment.md)
