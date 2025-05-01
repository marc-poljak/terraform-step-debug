{
  "application": {
    "name": "${app_name}",
    "port": ${port},
    "priority": ${priority},
    "environment": "${environment}"
  },
  "storage": {
    "bucket": "${bucket_name}"
  },
  "security": {
    "database_password": "${db_password}",
    "api_key": "${api_key}",
    "api_token": "${api_token}"
  },
  "tags": {
    %{ for key, value in tags ~}
    "${key}": "${value}"%{ if key != keys(tags)[length(keys(tags)) - 1] },%{ endif }
    %{ endfor ~}
  }
}