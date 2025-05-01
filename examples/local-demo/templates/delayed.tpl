DELAYED REPORT - ${environment}
===========================

This file was created after a ${delay}-second delay to simulate 
a longer-running resource. The delay is configurable through 
the variables:
  - enable_delay: ${delay > 0 ? "true" : "false"}
  - delay_seconds: ${delay}

When using different var-files, this resource might be skipped
entirely if enable_delay is set to false.

Timestamp: ${timestamp}

Note: In a real environment, longer running resources might
include database instances, clusters, or large storage systems.