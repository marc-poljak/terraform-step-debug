Infrastructure Details - ${environment}
=================================

Application Name: ${app_name}
Storage Bucket: ${bucket_name}
Service Port: ${port}
Environment: ${environment}

Resource Tags:
%{ for key, value in tags ~}
  - ${key}: ${value}
%{ endfor ~}

This is a simulated infrastructure report that would typically contain 
information about your deployed resources. In a real environment, this
might include IP addresses, DNS names, cluster details, etc.