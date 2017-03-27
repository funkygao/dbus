/*
Package engine provides a plugin based pipeline engine cluster that decouples Input/Filter/Output plugins.

To use the cluster feature, an Input plugin should call InputRunner.DeclareResource to
show its interest in some resources, on which engine will do load balance assignment.

In this way, cluster'ed engines will have leader/standby feature.

An input plugin, after call DeclareResource will then call WaitForTicket to wait for
the assigned resources that belongs to this leader.
*/
package engine
